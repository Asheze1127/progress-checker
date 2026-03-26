/**
 * Shared Slack messaging module for Lambda workers.
 *
 * Uses raw `fetch` to avoid pulling in the @slack/web-api SDK at runtime.
 */

export interface SlackPostResult {
  ok: boolean;
  error?: string;
}

export interface SlackMessageOptions {
  channelId: string;
  threadTs: string;
  text: string;
  blocks?: unknown[];
}

/**
 * Post a message to a Slack channel/thread using the chat.postMessage API.
 *
 * Validates `SLACK_BOT_TOKEN` from environment, sends the request via fetch,
 * and checks both HTTP status and the Slack API `ok` field.
 */
export async function postSlackMessage(options: SlackMessageOptions): Promise<SlackPostResult> {
  const slackToken = process.env.SLACK_BOT_TOKEN;
  if (!slackToken) {
    throw new Error("SLACK_BOT_TOKEN environment variable is not set");
  }

  const payload: Record<string, unknown> = {
    channel: options.channelId,
    thread_ts: options.threadTs,
    text: options.text,
  };

  if (options.blocks !== undefined) {
    payload.blocks = options.blocks;
  }

  const response = await fetch("https://slack.com/api/chat.postMessage", {
    method: "POST",
    headers: {
      "Content-Type": "application/json; charset=utf-8",
      Authorization: `Bearer ${slackToken}`,
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    throw new Error(`Slack API returned HTTP ${String(response.status)}`);
  }

  const result = (await response.json()) as SlackPostResult;
  if (!result.ok) {
    throw new Error(`Slack API error: ${result.error ?? "unknown"}`);
  }

  return result;
}
