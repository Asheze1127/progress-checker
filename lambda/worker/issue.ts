import { generateText } from "ai";
import { bedrock } from "@ai-sdk/amazon-bedrock";
import type { SQSEvent, SQSHandler } from "aws-lambda";
import { postSlackMessage } from "./shared/slack";

interface ThreadMessage {
  user: string;
  text: string;
  ts: string;
}

interface IssueMessage {
  type: "issue";
  channel_id: string;
  thread_ts: string;
  trigger_user_id: string;
  trigger_type: string;
  thread_messages: ThreadMessage[];
}

interface ParsedIssueSummary {
  title: string;
  body: string;
}

interface IssueApiResponse {
  issue_url: string;
}

const ISSUE_API_HOSTNAME = process.env.ISSUE_API_HOSTNAME ?? "";
const INTERNAL_API_TOKEN = process.env.INTERNAL_API_TOKEN ?? "";
const LLM_MODEL_ID = "anthropic.claude-3-5-sonnet-20241022-v2:0";

const SYSTEM_PROMPT = [
  "Summarize the following Slack discussion into a GitHub Issue.",
  "Include: Title, Description, and Action Items.",
  "Write in Japanese.",
  "",
  "Format your response exactly as follows:",
  "# Title",
  "<issue title here>",
  "",
  "# Description",
  "<issue description here>",
  "",
  "# Action Items",
  "<action items here>",
].join("\n");

function formatThreadMessages(messages: ThreadMessage[]): string {
  return messages.map((message) => `[${message.user}] (${message.ts}): ${message.text}`).join("\n");
}

function parseIssueSummary(summary: string): ParsedIssueSummary {
  const titleMatch = summary.match(/^#\s*Title\s*\n(.+)/m);
  const title = titleMatch?.[1]?.trim() ?? "Untitled Issue";

  const descriptionMatch = summary.match(/^#\s*Description\s*\n([\s\S]*?)(?=^#\s*Action Items)/m);
  const actionItemsMatch = summary.match(/^#\s*Action Items\s*\n([\s\S]*?)$/m);

  const description = descriptionMatch?.[1]?.trim() ?? "";
  const actionItems = actionItemsMatch?.[1]?.trim() ?? "";

  const bodyParts: string[] = [];
  if (description) {
    bodyParts.push("## Description\n\n" + description);
  }
  if (actionItems) {
    bodyParts.push("## Action Items\n\n" + actionItems);
  }

  const body = bodyParts.length > 0 ? bodyParts.join("\n\n") : summary;

  return { title, body };
}

function parseIssueMessage(rawBody: string): IssueMessage {
  const parsed: unknown = JSON.parse(rawBody);
  if (typeof parsed !== "object" || parsed === null) {
    throw new Error("Invalid SQS message body: expected an object");
  }

  const message = parsed as Record<string, unknown>;

  if (message.type !== "issue") {
    throw new Error(`Unexpected message type: ${String(message.type)}`);
  }
  if (typeof message.channel_id !== "string" || message.channel_id === "") {
    throw new Error("Missing or invalid channel_id");
  }
  if (typeof message.thread_ts !== "string" || message.thread_ts === "") {
    throw new Error("Missing or invalid thread_ts");
  }
  if (!Array.isArray(message.thread_messages) || message.thread_messages.length === 0) {
    throw new Error("Missing or empty thread_messages");
  }

  return parsed as IssueMessage;
}

async function summarizeThread(threadMessages: ThreadMessage[]): Promise<string> {
  const formattedMessages = formatThreadMessages(threadMessages);

  const { text } = await generateText({
    model: bedrock(LLM_MODEL_ID),
    system: SYSTEM_PROMPT,
    prompt: formattedMessages,
  });

  return text;
}

async function createGitHubIssue(channelId: string, title: string, body: string): Promise<IssueApiResponse> {
  const url = `https://${ISSUE_API_HOSTNAME}/internal/issues`;

  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Token": INTERNAL_API_TOKEN,
    },
    body: JSON.stringify({
      channel_id: channelId,
      title,
      body,
    }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Internal API returned ${String(response.status)}: ${errorText}`);
  }

  const result: unknown = await response.json();
  return result as IssueApiResponse;
}

async function processIssueCreation(message: IssueMessage): Promise<void> {
  console.log(
    JSON.stringify({
      event: "issue_creation_started",
      channelId: message.channel_id,
      threadTs: message.thread_ts,
      triggerUserId: message.trigger_user_id,
      triggerType: message.trigger_type,
      messageCount: message.thread_messages.length,
    }),
  );

  const summary = await summarizeThread(message.thread_messages);
  const { title, body } = parseIssueSummary(summary);

  console.log(JSON.stringify({ event: "summary_generated", title }));

  const { issue_url: issueUrl } = await createGitHubIssue(message.channel_id, title, body);

  console.log(JSON.stringify({ event: "issue_created", issueUrl }));

  await postSlackMessage({ channelId: message.channel_id, threadTs: message.thread_ts, text: `GitHub Issue created: ${issueUrl}` });

  console.log(JSON.stringify({ event: "issue_creation_completed", issueUrl }));
}

export const handler: SQSHandler = async (event: SQSEvent): Promise<void> => {
  console.log(
    JSON.stringify({
      appEnvironment: process.env.APP_ENV,
      lambdaName: process.env.LAMBDA_NAME,
      recordCount: event.Records?.length ?? 0,
    }),
  );

  const errors: Array<{ messageId: string; error: unknown }> = [];

  for (const record of event.Records) {
    try {
      const message = parseIssueMessage(record.body);
      await processIssueCreation(message);
    } catch (error) {
      console.error(
        JSON.stringify({
          event: "issue_creation_failed",
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

export { formatThreadMessages, parseIssueSummary, parseIssueMessage, processIssueCreation };
