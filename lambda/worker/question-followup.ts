import { generateText } from "ai";
import { bedrock } from "@ai-sdk/amazon-bedrock";
import type { SQSEvent, SQSHandler, SQSRecord } from "aws-lambda";
import { postSlackMessage } from "./shared/slack";

// Bedrock model identifier for Claude 3.5 Sonnet v2
const BEDROCK_MODEL_ID = "anthropic.claude-3-5-sonnet-20241022-v2:0";

// Maximum number of thread messages to include in context
const MAX_THREAD_MESSAGES = 50;

// Status values for question sessions
const SESSION_STATUS_AWAITING_USER = "awaiting_user";
const SESSION_STATUS_RESOLVED = "resolved";

// Keywords that indicate a resolution signal from AI
const RESOLUTION_INDICATORS = [
  "解決",
  "問題が解消",
  "うまくいくはず",
  "試してみてください",
  "これで動作するはず",
  "修正できるはず",
];

// Keywords that indicate AI is asking for more information
const FOLLOWUP_INDICATORS = [
  "教えていただけますか",
  "確認させてください",
  "共有していただけますか",
  "以下の情報",
  "詳しく教えて",
  "もう少し情報",
  "エラーログ",
  "環境を教えて",
];

interface ThreadMessage {
  user: string;
  text: string;
  ts: string;
}

export interface FollowupMessage {
  type: "question_followup";
  question_id: string;
  participant_id: string;
  text: string;
  slack_channel_id: string;
  slack_thread_ts: string;
  thread_messages: ThreadMessage[];
}

// AI-generated content disclaimer appended to every response
const AI_DISCLAIMER =
  "\n\n---\n_※ この回答はAIによる自動生成です。正確でない場合があります。_";

const SYSTEM_PROMPT = `You are a helpful technical mentor for hackathon participants.
Review the conversation history and the user's latest message.

Your priorities:
1. If critical information is missing (error logs, environment details, steps already tried, expected vs actual behavior), ask specific clarifying questions before providing a solution.
2. If you have enough context, provide a clear, actionable solution.
3. Keep responses concise and focused.
4. When asking clarifying questions, explain WHY each piece of information is needed.

Respond in Japanese.`;

/**
 * Build the message array for the AI model from thread history and the latest user message.
 */
export function buildMessages(
  threadMessages: ThreadMessage[],
  latestText: string,
): Array<{ role: "user" | "assistant"; content: string }> {
  const messages: Array<{ role: "user" | "assistant"; content: string }> = [];

  // Include thread messages as context, limited to avoid token overflow
  const limitedMessages = threadMessages.slice(-MAX_THREAD_MESSAGES);

  for (const msg of limitedMessages) {
    // Bot messages are treated as assistant; everything else as user
    const role = msg.user === "bot" ? "assistant" : "user";
    messages.push({ role, content: msg.text });
  }

  // Add the latest followup message from the user
  messages.push({ role: "user", content: latestText });

  return messages;
}

/**
 * Determine whether the AI response indicates it needs more info or considers the issue resolved.
 */
export function detectResponseIntent(responseText: string): "followup_needed" | "resolved" {
  const hasFollowupIndicator = FOLLOWUP_INDICATORS.some((indicator) => responseText.includes(indicator));
  const hasResolutionIndicator = RESOLUTION_INDICATORS.some((indicator) => responseText.includes(indicator));

  // If the response asks for more information, treat as followup needed
  if (hasFollowupIndicator && !hasResolutionIndicator) {
    return "followup_needed";
  }

  return "resolved";
}

/**
 * Update the question session status in the session store.
 */
async function updateQuestionSessionStatus(questionId: string, status: string): Promise<void> {
  // TODO: gh-66 Implement actual database update via internal API or direct DB access
  console.log(JSON.stringify({ action: "update_session_status", questionId, status }));
}

/**
 * Process a single followup question message from SQS.
 */
export async function processFollowupQuestion(message: FollowupMessage): Promise<void> {
  const { question_id, text, slack_channel_id, slack_thread_ts, thread_messages } = message;

  console.log(
    JSON.stringify({
      action: "process_followup",
      questionId: question_id,
      threadMessageCount: thread_messages.length,
    }),
  );

  const messages = buildMessages(thread_messages, text);

  const { text: aiResponse } = await generateText({
    model: bedrock(BEDROCK_MODEL_ID),
    system: SYSTEM_PROMPT,
    messages,
  });

  const responseWithDisclaimer = aiResponse + AI_DISCLAIMER;

  await postSlackMessage({ channelId: slack_channel_id, threadTs: slack_thread_ts, text: responseWithDisclaimer });

  const intent = detectResponseIntent(aiResponse);

  if (intent === "followup_needed") {
    await updateQuestionSessionStatus(question_id, SESSION_STATUS_AWAITING_USER);
    console.log(JSON.stringify({ action: "session_awaiting_user", questionId: question_id }));
  } else {
    await updateQuestionSessionStatus(question_id, SESSION_STATUS_RESOLVED);
    console.log(JSON.stringify({ action: "session_resolved", questionId: question_id }));
  }
}

/**
 * Parse and validate the SQS record body as a FollowupMessage.
 */
export function parseFollowupMessage(body: string): FollowupMessage {
  const parsed: unknown = JSON.parse(body);

  if (typeof parsed !== "object" || parsed === null) {
    throw new Error("Invalid message: not an object");
  }

  const msg = parsed as Record<string, unknown>;

  if (msg.type !== "question_followup") {
    throw new Error(`Unexpected message type: ${String(msg.type)}`);
  }

  if (typeof msg.question_id !== "string" || msg.question_id.length === 0) {
    throw new Error("Missing or invalid question_id");
  }

  if (typeof msg.participant_id !== "string") {
    throw new Error("Missing or invalid participant_id");
  }

  if (typeof msg.text !== "string" || msg.text.length === 0) {
    throw new Error("Missing or invalid text");
  }

  if (typeof msg.slack_channel_id !== "string") {
    throw new Error("Missing or invalid slack_channel_id");
  }

  if (typeof msg.slack_thread_ts !== "string") {
    throw new Error("Missing or invalid slack_thread_ts");
  }

  if (!Array.isArray(msg.thread_messages)) {
    throw new Error("Missing or invalid thread_messages");
  }

  return parsed as FollowupMessage;
}

/**
 * Process a single SQS record, logging errors per-record without failing the batch.
 */
async function processRecord(record: SQSRecord): Promise<void> {
  const message = parseFollowupMessage(record.body);
  await processFollowupQuestion(message);
}

/**
 * SQS Lambda handler for question:followup messages.
 */
export const handler: SQSHandler = async (event: SQSEvent): Promise<void> => {
  console.log(
    JSON.stringify({
      appEnvironment: process.env.APP_ENV,
      databaseHost: process.env.DATABASE_HOST,
      lambdaName: process.env.LAMBDA_NAME,
      recordCount: event.Records?.length ?? 0,
    }),
  );

  const errors: Array<{ messageId: string; error: unknown }> = [];

  for (const record of event.Records) {
    try {
      await processRecord(record);
    } catch (error) {
      console.error(
        JSON.stringify({
          action: "record_processing_error",
          messageId: record.messageId,
          error: error instanceof Error ? error.message : String(error),
        }),
      );
      errors.push({ messageId: record.messageId, error });
    }
  }

  if (errors.length > 0) {
    throw new Error(`Failed to process ${errors.length} record(s)`);
  }
};
