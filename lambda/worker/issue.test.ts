import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import type { SQSEvent, SQSRecord, Context } from "aws-lambda";

vi.mock("ai", () => ({
  generateText: vi.fn(),
}));

vi.mock("@ai-sdk/amazon-bedrock", () => ({
  bedrock: vi.fn(() => "mocked-model"),
}));

import { handler, formatThreadMessages, parseIssueSummary, parseIssueMessage } from "./issue";
import { generateText } from "ai";

const MOCK_ISSUE_URL = "https://github.com/org/repo/issues/42";

const MOCK_LLM_RESPONSE = [
  "# Title",
  "Fix login timeout issue",
  "",
  "# Description",
  "Users are experiencing login timeouts when the server is under heavy load.",
  "",
  "# Action Items",
  "- Investigate server timeout configuration",
  "- Add retry logic to the login flow",
].join("\n");

function createSqsRecord(body: Record<string, unknown>): SQSRecord {
  return {
    messageId: "test-message-id",
    receiptHandle: "test-receipt-handle",
    body: JSON.stringify(body),
    attributes: {
      ApproximateReceiveCount: "1",
      SentTimestamp: "1234567890",
      SenderId: "test-sender",
      ApproximateFirstReceiveTimestamp: "1234567890",
    },
    messageAttributes: {},
    md5OfBody: "test-md5",
    eventSource: "aws:sqs",
    eventSourceARN: "arn:aws:sqs:us-east-1:123456789012:issue-queue",
    awsRegion: "us-east-1",
  };
}

function createValidIssueMessage(): Record<string, unknown> {
  return {
    type: "issue",
    channel_id: "C123456",
    thread_ts: "1234567890.123456",
    trigger_user_id: "U123456",
    trigger_type: "message_action",
    thread_messages: [
      { user: "U111", text: "Login is broken", ts: "1234567890.000001" },
      { user: "U222", text: "I see the same issue, timeout after 30s", ts: "1234567890.000002" },
      { user: "U111", text: "Let's create a ticket for this", ts: "1234567890.000003" },
    ],
  };
}

describe("formatThreadMessages", () => {
  it("formats thread messages into readable text", () => {
    const messages = [
      { user: "U111", text: "Hello", ts: "1234567890.000001" },
      { user: "U222", text: "World", ts: "1234567890.000002" },
    ];

    const result = formatThreadMessages(messages);

    expect(result).toBe("[U111] (1234567890.000001): Hello\n[U222] (1234567890.000002): World");
  });

  it("handles a single message", () => {
    const messages = [{ user: "U111", text: "Solo message", ts: "1234567890.000001" }];

    const result = formatThreadMessages(messages);

    expect(result).toBe("[U111] (1234567890.000001): Solo message");
  });
});

describe("parseIssueSummary", () => {
  it("parses a well-formatted LLM response into title and body", () => {
    const result = parseIssueSummary(MOCK_LLM_RESPONSE);

    expect(result.title).toBe("Fix login timeout issue");
    expect(result.body).toContain("## Description");
    expect(result.body).toContain("Users are experiencing login timeouts");
    expect(result.body).toContain("## Action Items");
    expect(result.body).toContain("Investigate server timeout configuration");
  });

  it("returns 'Untitled Issue' when title section is missing", () => {
    const summary = "Some text without proper formatting";
    const result = parseIssueSummary(summary);

    expect(result.title).toBe("Untitled Issue");
  });

  it("uses full summary as body when sections are missing", () => {
    const summary = "Just a plain text summary with no sections";
    const result = parseIssueSummary(summary);

    expect(result.body).toBe(summary);
  });

  it("handles summary with only title and description", () => {
    const summary = "# Title\nMy Title\n\n# Description\nSome description\n\n# Action Items\n";
    const result = parseIssueSummary(summary);

    expect(result.title).toBe("My Title");
    expect(result.body).toContain("## Description");
    expect(result.body).toContain("Some description");
  });
});

describe("parseIssueMessage", () => {
  it("parses a valid issue message", () => {
    const validMessage = createValidIssueMessage();
    const result = parseIssueMessage(JSON.stringify(validMessage));

    expect(result.type).toBe("issue");
    expect(result.channel_id).toBe("C123456");
    expect(result.thread_ts).toBe("1234567890.123456");
    expect(result.thread_messages).toHaveLength(3);
  });

  it("throws on invalid JSON", () => {
    expect(() => parseIssueMessage("not-json")).toThrow();
  });

  it("throws on non-object body", () => {
    expect(() => parseIssueMessage('"just a string"')).toThrow("expected an object");
  });

  it("throws on wrong message type", () => {
    const message = { ...createValidIssueMessage(), type: "question" };
    expect(() => parseIssueMessage(JSON.stringify(message))).toThrow("Unexpected message type");
  });

  it("throws on missing channel_id", () => {
    const message = { ...createValidIssueMessage(), channel_id: "" };
    expect(() => parseIssueMessage(JSON.stringify(message))).toThrow("Missing or invalid channel_id");
  });

  it("throws on missing thread_ts", () => {
    const message = { ...createValidIssueMessage(), thread_ts: "" };
    expect(() => parseIssueMessage(JSON.stringify(message))).toThrow("Missing or invalid thread_ts");
  });

  it("throws on empty thread_messages", () => {
    const message = { ...createValidIssueMessage(), thread_messages: [] };
    expect(() => parseIssueMessage(JSON.stringify(message))).toThrow("Missing or empty thread_messages");
  });
});

describe("handler", () => {
  let mockFetch: Mock;

  beforeEach(() => {
    vi.clearAllMocks();

    (generateText as Mock).mockResolvedValue({ text: MOCK_LLM_RESPONSE });

    mockFetch = vi.fn();
    global.fetch = mockFetch;

    // Mock internal API response (create issue)
    mockFetch.mockImplementation((url: string) => {
      if (typeof url === "string" && url.includes("/internal/issues")) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ issue_url: MOCK_ISSUE_URL }),
        });
      }
      // Mock Slack API response (post message)
      if (typeof url === "string" && url.includes("slack.com")) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ ok: true }),
        });
      }
      return Promise.resolve({ ok: false, status: 404, text: () => Promise.resolve("Not found") });
    });
  });

  it("processes an SQS event and creates a GitHub issue", async () => {
    const event: SQSEvent = {
      Records: [createSqsRecord(createValidIssueMessage())],
    };

    await handler(event, {} as Context, () => {});

    expect(generateText).toHaveBeenCalledOnce();
    expect(mockFetch).toHaveBeenCalledTimes(2);

    const issueApiCall = mockFetch.mock.calls.find((call: unknown[]) =>
      String(call[0]).includes("/internal/issues"),
    );
    expect(issueApiCall).toBeDefined();

    const issueApiBody = JSON.parse((issueApiCall![1] as RequestInit).body as string);
    expect(issueApiBody.channel_id).toBe("C123456");
    expect(issueApiBody.title).toBe("Fix login timeout issue");
    expect(issueApiBody.body).toContain("## Description");

    const slackCall = mockFetch.mock.calls.find((call: unknown[]) => String(call[0]).includes("slack.com"));
    expect(slackCall).toBeDefined();

    const slackBody = JSON.parse((slackCall![1] as RequestInit).body as string);
    expect(slackBody.channel).toBe("C123456");
    expect(slackBody.thread_ts).toBe("1234567890.123456");
    expect(slackBody.text).toContain(MOCK_ISSUE_URL);
  });

  it("throws when internal API returns an error", async () => {
    mockFetch.mockImplementation((url: string) => {
      if (typeof url === "string" && url.includes("/internal/issues")) {
        return Promise.resolve({
          ok: false,
          status: 500,
          text: () => Promise.resolve("Internal Server Error"),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
    });

    const event: SQSEvent = {
      Records: [createSqsRecord(createValidIssueMessage())],
    };

    await expect(handler(event, {} as Context, () => {})).rejects.toThrow("Internal API returned 500");
  });

  it("throws when Slack API returns an error", async () => {
    mockFetch.mockImplementation((url: string) => {
      if (typeof url === "string" && url.includes("/internal/issues")) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ issue_url: MOCK_ISSUE_URL }),
        });
      }
      if (typeof url === "string" && url.includes("slack.com")) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ ok: false, error: "channel_not_found" }),
        });
      }
      return Promise.resolve({ ok: false, status: 404, text: () => Promise.resolve("Not found") });
    });

    const event: SQSEvent = {
      Records: [createSqsRecord(createValidIssueMessage())],
    };

    await expect(handler(event, {} as Context, () => {})).rejects.toThrow("Slack API error: channel_not_found");
  });

  it("processes multiple records sequentially", async () => {
    const event: SQSEvent = {
      Records: [createSqsRecord(createValidIssueMessage()), createSqsRecord(createValidIssueMessage())],
    };

    await handler(event, {} as Context, () => {});

    expect(generateText).toHaveBeenCalledTimes(2);
    expect(mockFetch).toHaveBeenCalledTimes(4);
  });
});
