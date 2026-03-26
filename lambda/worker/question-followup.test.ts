import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import {
  buildMessages,
  detectResponseIntent,
  parseFollowupMessage,
  processFollowupQuestion,
  handler,
  type FollowupMessage,
} from "./question-followup";
import type { SQSEvent, Context, Callback } from "aws-lambda";

// Mock external dependencies
vi.mock("ai", () => ({
  generateText: vi.fn(),
}));

vi.mock("@ai-sdk/amazon-bedrock", () => ({
  bedrock: vi.fn(() => "mocked-model"),
}));

// Mock global fetch for Slack API calls
const mockFetch = vi.fn();
global.fetch = mockFetch;

import { generateText } from "ai";

const mockGenerateText = generateText as Mock;

function createValidMessage(overrides: Partial<FollowupMessage> = {}): FollowupMessage {
  return {
    type: "question_followup",
    question_id: "q-123",
    participant_id: "u-456",
    text: "Here are my error logs",
    slack_channel_id: "C12345",
    slack_thread_ts: "1234567890.123456",
    thread_messages: [
      { user: "u-456", text: "I have a problem with my code", ts: "1234567890.000001" },
      { user: "bot", text: "Could you share the error logs?", ts: "1234567890.000002" },
    ],
    ...overrides,
  };
}

function createSqsEvent(messages: FollowupMessage[]): SQSEvent {
  return {
    Records: messages.map((msg, index) => ({
      messageId: `msg-${index}`,
      receiptHandle: `handle-${index}`,
      body: JSON.stringify(msg),
      attributes: {
        ApproximateReceiveCount: "1",
        SentTimestamp: "1234567890",
        SenderId: "sender",
        ApproximateFirstReceiveTimestamp: "1234567890",
      },
      messageAttributes: {},
      md5OfBody: "abc123",
      eventSource: "aws:sqs",
      eventSourceARN: "arn:aws:sqs:us-east-1:123456789:question-queue",
      awsRegion: "us-east-1",
    })),
  };
}

describe("buildMessages", () => {
  it("should convert thread messages and latest text into AI message format", () => {
    const threadMessages = [
      { user: "u-1", text: "My app crashes", ts: "1" },
      { user: "bot", text: "Can you share the error?", ts: "2" },
    ];

    const result = buildMessages(threadMessages, "Here is the error log");

    expect(result).toEqual([
      { role: "user", content: "My app crashes" },
      { role: "assistant", content: "Can you share the error?" },
      { role: "user", content: "Here is the error log" },
    ]);
  });

  it("should handle empty thread messages", () => {
    const result = buildMessages([], "First message");

    expect(result).toEqual([{ role: "user", content: "First message" }]);
  });

  it("should limit thread messages to prevent token overflow", () => {
    const threadMessages = Array.from({ length: 100 }, (_, i) => ({
      user: `u-${i}`,
      text: `Message ${i}`,
      ts: `${i}`,
    }));

    const result = buildMessages(threadMessages, "Latest");

    // 50 thread messages + 1 latest = 51 total
    expect(result).toHaveLength(51);
    // Should include the last 50 thread messages, not the first 50
    expect(result[0].content).toBe("Message 50");
  });

  it("should treat all non-bot users as user role", () => {
    const threadMessages = [
      { user: "u-admin", text: "Admin message", ts: "1" },
      { user: "u-mentor", text: "Mentor message", ts: "2" },
    ];

    const result = buildMessages(threadMessages, "Latest");

    expect(result[0].role).toBe("user");
    expect(result[1].role).toBe("user");
  });
});

describe("detectResponseIntent", () => {
  it("should detect followup needed when AI asks for information", () => {
    const response = "エラーログを共有していただけますか？どのような環境で動かしていますか？";
    expect(detectResponseIntent(response)).toBe("followup_needed");
  });

  it("should detect resolved when AI provides a solution", () => {
    const response = "このコードを修正すれば問題が解消されます。変数名を修正してください。";
    expect(detectResponseIntent(response)).toBe("resolved");
  });

  it("should default to resolved when no indicators match", () => {
    const response = "こちらのコードを参考にしてください。";
    expect(detectResponseIntent(response)).toBe("resolved");
  });

  it("should prefer resolved when both indicators are present", () => {
    const response = "エラーログを確認させてください。これで解決するはずです。";
    expect(detectResponseIntent(response)).toBe("resolved");
  });

  it("should detect followup for missing environment details", () => {
    const response = "もう少し情報が必要です。開発環境を教えてください。";
    expect(detectResponseIntent(response)).toBe("followup_needed");
  });
});

describe("parseFollowupMessage", () => {
  it("should parse a valid message", () => {
    const message = createValidMessage();
    const result = parseFollowupMessage(JSON.stringify(message));

    expect(result).toEqual(message);
  });

  it("should reject non-object bodies", () => {
    expect(() => parseFollowupMessage('"just a string"')).toThrow("not an object");
  });

  it("should reject wrong message type", () => {
    const message = { ...createValidMessage(), type: "question_new" };
    expect(() => parseFollowupMessage(JSON.stringify(message))).toThrow("Unexpected message type");
  });

  it("should reject missing question_id", () => {
    const message = { ...createValidMessage(), question_id: "" };
    expect(() => parseFollowupMessage(JSON.stringify(message))).toThrow("question_id");
  });

  it("should reject missing text", () => {
    const message = { ...createValidMessage(), text: "" };
    expect(() => parseFollowupMessage(JSON.stringify(message))).toThrow("text");
  });

  it("should reject missing slack_channel_id", () => {
    const { slack_channel_id, ...rest } = createValidMessage();
    expect(() => parseFollowupMessage(JSON.stringify(rest))).toThrow("slack_channel_id");
  });

  it("should reject non-array thread_messages", () => {
    const message = { ...createValidMessage(), thread_messages: "not-an-array" };
    expect(() => parseFollowupMessage(JSON.stringify(message))).toThrow("thread_messages");
  });

  it("should reject invalid JSON", () => {
    expect(() => parseFollowupMessage("not-json")).toThrow();
  });
});

describe("processFollowupQuestion", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    process.env.SLACK_BOT_TOKEN = "xoxb-test-token";

    mockFetch.mockResolvedValue({
      json: () => Promise.resolve({ ok: true }),
    });
  });

  it("should generate AI response and post to Slack", async () => {
    mockGenerateText.mockResolvedValue({
      text: "この問題を解決するには、以下の手順を試してみてください。",
    });

    const message = createValidMessage();
    await processFollowupQuestion(message);

    expect(mockGenerateText).toHaveBeenCalledOnce();
    expect(mockGenerateText).toHaveBeenCalledWith(
      expect.objectContaining({
        system: expect.stringContaining("technical mentor"),
        messages: expect.arrayContaining([expect.objectContaining({ role: "user" })]),
      }),
    );

    expect(mockFetch).toHaveBeenCalledWith(
      "https://slack.com/api/chat.postMessage",
      expect.objectContaining({
        method: "POST",
        body: expect.stringContaining(message.slack_channel_id),
      }),
    );
  });

  it("should append AI disclaimer to the response", async () => {
    mockGenerateText.mockResolvedValue({ text: "AI response text" });

    const message = createValidMessage();
    await processFollowupQuestion(message);

    const fetchBody = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(fetchBody.text).toContain("AI response text");
    expect(fetchBody.text).toContain("generated by AI");
  });

  it("should throw when SLACK_BOT_TOKEN is missing", async () => {
    delete process.env.SLACK_BOT_TOKEN;
    mockGenerateText.mockResolvedValue({ text: "response" });

    const message = createValidMessage();
    await expect(processFollowupQuestion(message)).rejects.toThrow("SLACK_BOT_TOKEN");
  });

  it("should throw when Slack API returns an error", async () => {
    mockGenerateText.mockResolvedValue({ text: "response" });
    mockFetch.mockResolvedValue({
      json: () => Promise.resolve({ ok: false, error: "channel_not_found" }),
    });

    const message = createValidMessage();
    await expect(processFollowupQuestion(message)).rejects.toThrow("channel_not_found");
  });
});

describe("handler", () => {
  const mockContext = {} as Context;
  const mockCallback = (() => {}) as Callback;

  beforeEach(() => {
    vi.clearAllMocks();
    process.env.SLACK_BOT_TOKEN = "xoxb-test-token";
    process.env.APP_ENV = "test";
    process.env.DATABASE_HOST = "localhost";
    process.env.LAMBDA_NAME = "question";

    mockGenerateText.mockResolvedValue({
      text: "解決策を提案します。試してみてください。",
    });

    mockFetch.mockResolvedValue({
      json: () => Promise.resolve({ ok: true }),
    });
  });

  it("should process all records in the event", async () => {
    const event = createSqsEvent([createValidMessage(), createValidMessage({ question_id: "q-789" })]);

    await handler(event, mockContext, mockCallback);

    expect(mockGenerateText).toHaveBeenCalledTimes(2);
    expect(mockFetch).toHaveBeenCalledTimes(2);
  });

  it("should throw on record processing failure", async () => {
    mockGenerateText.mockRejectedValue(new Error("Model unavailable"));

    const event = createSqsEvent([createValidMessage()]);

    await expect(handler(event, mockContext, mockCallback)).rejects.toThrow("Model unavailable");
  });

  it("should handle events with no records", async () => {
    const event: SQSEvent = { Records: [] };

    await handler(event, mockContext, mockCallback);

    expect(mockGenerateText).not.toHaveBeenCalled();
  });
});
