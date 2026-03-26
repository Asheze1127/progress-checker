import { bedrock } from "@ai-sdk/amazon-bedrock";
import type { KnownBlock } from "@slack/web-api";
import { WebClient } from "@slack/web-api";
import { generateText } from "ai";
import type { SQSEvent, SQSHandler } from "aws-lambda";

// --- Types ---

export interface QuestionMessage {
  type: "question_new";
  question_id: string;
  participant_id: string;
  title: string;
  text: string;
  slack_channel_id: string;
  slack_thread_ts: string;
}

export interface Dependencies {
  generateAiResponse: (questionText: string) => Promise<string>;
  postToSlack: (
    channelId: string,
    threadTs: string,
    text: string,
    blocks: KnownBlock[],
  ) => Promise<void>;
}

// --- Constants ---

const AI_DISCLAIMER =
  "※ この回答はAIによる自動生成です。正確でない場合があります。";

const BEDROCK_MODEL_ID = "anthropic.claude-3-5-sonnet-20241022-v2:0";

const SYSTEM_PROMPT = `You are a helpful technical mentor assisting hackathon participants.
Provide clear, actionable answers. When the question is ambiguous, ask a focused clarifying question.
Always respond in the same language as the question.`;

// --- Slack helpers ---

function createSlackClient(): WebClient {
  const token = process.env.SLACK_BOT_TOKEN;
  if (!token) {
    throw new Error("SLACK_BOT_TOKEN environment variable is not set");
  }
  return new WebClient(token);
}

interface ActionButton {
  type: "button";
  text: { type: "plain_text"; text: string };
  action_id: string;
  style?: "primary" | "danger";
  value: string;
}

export function buildResponseBlocks(
  responseText: string,
  questionId: string,
): KnownBlock[] {
  const actionButtons: ActionButton[] = [
    {
      type: "button",
      text: { type: "plain_text", text: "解決済み" },
      action_id: "question_resolved",
      style: "primary",
      value: questionId,
    },
    {
      type: "button",
      text: { type: "plain_text", text: "追加質問する" },
      action_id: "question_continue",
      value: questionId,
    },
    {
      type: "button",
      text: { type: "plain_text", text: "メンターに聞く" },
      action_id: "question_escalate",
      style: "danger",
      value: questionId,
    },
  ];

  return [
    {
      type: "section",
      text: {
        type: "mrkdwn",
        text: responseText,
      },
    },
    {
      type: "context",
      elements: [
        {
          type: "mrkdwn",
          text: AI_DISCLAIMER,
        },
      ],
    },
    {
      type: "actions",
      elements: actionButtons,
    },
  ];
}

// --- Core processing (testable with injected dependencies) ---

export async function processNewQuestion(
  message: QuestionMessage,
  deps: Dependencies,
): Promise<void> {
  const logger = {
    questionId: message.question_id,
    participantId: message.participant_id,
    type: message.type,
  };

  console.log(JSON.stringify({ event: "processing_start", ...logger }));

  const responseText = await deps.generateAiResponse(message.text);

  console.log(
    JSON.stringify({
      event: "ai_response_generated",
      responseLength: responseText.length,
      ...logger,
    }),
  );

  const blocks = buildResponseBlocks(responseText, message.question_id);

  await deps.postToSlack(
    message.slack_channel_id,
    message.slack_thread_ts,
    responseText,
    blocks,
  );

  console.log(JSON.stringify({ event: "slack_response_posted", ...logger }));
}

// --- Production dependencies ---

function createProductionDependencies(): Dependencies {
  return {
    generateAiResponse: async (questionText: string): Promise<string> => {
      const { text } = await generateText({
        model: bedrock(BEDROCK_MODEL_ID),
        system: SYSTEM_PROMPT,
        prompt: questionText,
      });
      return text;
    },
    postToSlack: async (
      channelId: string,
      threadTs: string,
      text: string,
      blocks: KnownBlock[],
    ): Promise<void> => {
      const slackClient = createSlackClient();
      await slackClient.chat.postMessage({
        channel: channelId,
        thread_ts: threadTs,
        text,
        blocks,
      });
    },
  };
}

// --- SQS Handler ---

export const handler: SQSHandler = async (event: SQSEvent) => {
  const deps = createProductionDependencies();

  for (const record of event.Records) {
    try {
      const message: QuestionMessage = JSON.parse(record.body);
      await processNewQuestion(message, deps);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      console.error(
        JSON.stringify({
          event: "record_processing_error",
          messageId: record.messageId,
          error: errorMessage,
        }),
      );
      // Re-throw so SQS can retry this specific record
      throw error;
    }
  }
};
