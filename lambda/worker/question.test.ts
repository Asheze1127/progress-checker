import { describe, it, beforeEach, expect, vi } from "vitest";
import type { KnownBlock } from "@slack/web-api";
import {
  processNewQuestion,
  buildResponseBlocks,
  type QuestionMessage,
  type Dependencies,
} from "./question";

// --- Test helpers ---

function createValidMessage(): QuestionMessage {
  return {
    type: "question_new",
    question_id: "q-001",
    participant_id: "p-001",
    title: "How to use React hooks?",
    text: "I am trying to use useEffect but it keeps running in an infinite loop. How do I fix this?",
    slack_channel_id: "C1234567890",
    slack_thread_ts: "1234567890.123456",
  };
}

interface SlackPostCall {
  channelId: string;
  threadTs: string;
  text: string;
  blocks: KnownBlock[];
}

function createMockDependencies() {
  const aiCalls: string[] = [];
  const slackCalls: SlackPostCall[] = [];

  const deps: Dependencies = {
    generateAiResponse: async (questionText: string) => {
      aiCalls.push(questionText);
      return "Here is the AI response to your question.";
    },
    postToSlack: async (
      channelId: string,
      threadTs: string,
      text: string,
      blocks: KnownBlock[],
    ) => {
      slackCalls.push({ channelId, threadTs, text, blocks });
    },
  };

  return { deps, aiCalls, slackCalls };
}

// --- Tests ---

describe("processNewQuestion", () => {
  it("should call AI with the question text and post response to Slack", async () => {
    const message = createValidMessage();
    const { deps, aiCalls, slackCalls } = createMockDependencies();

    await processNewQuestion(message, deps);

    expect(aiCalls).toHaveLength(1);
    expect(aiCalls[0]).toBe(message.text);

    expect(slackCalls).toHaveLength(1);
    expect(slackCalls[0].channelId).toBe(message.slack_channel_id);
    expect(slackCalls[0].threadTs).toBe(message.slack_thread_ts);
    expect(slackCalls[0].text).toBe(
      "Here is the AI response to your question.",
    );
  });

  it("should include Block Kit blocks in Slack response", async () => {
    const message = createValidMessage();
    const { deps, slackCalls } = createMockDependencies();

    await processNewQuestion(message, deps);

    const blocks = slackCalls[0].blocks;
    expect(blocks).toHaveLength(3);
    expect(blocks[0].type).toBe("section");
    expect(blocks[1].type).toBe("context");
    expect(blocks[2].type).toBe("actions");
  });

  it("should propagate AI generation errors", async () => {
    const message = createValidMessage();
    const { deps } = createMockDependencies();
    deps.generateAiResponse = async () => {
      throw new Error("Bedrock service unavailable");
    };

    await expect(processNewQuestion(message, deps)).rejects.toThrow(
      "Bedrock service unavailable",
    );
  });

  it("should propagate Slack posting errors", async () => {
    const message = createValidMessage();
    const { deps } = createMockDependencies();
    deps.postToSlack = async () => {
      throw new Error("Slack API error: channel_not_found");
    };

    await expect(processNewQuestion(message, deps)).rejects.toThrow(
      "channel_not_found",
    );
  });
});

describe("buildResponseBlocks", () => {
  it("should return section, context, and actions blocks", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    expect(blocks).toHaveLength(3);
    expect(blocks[0].type).toBe("section");
    expect(blocks[1].type).toBe("context");
    expect(blocks[2].type).toBe("actions");
  });

  it("should include AI disclaimer in context block", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    const contextBlock = blocks[1] as { elements: Array<{ text: string }> };
    expect(contextBlock.elements[0].text).toContain("AIによる自動生成");
  });

  it("should include resolve, continue, and escalate action buttons", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    const actionsBlock = blocks[2] as {
      elements: Array<{ action_id: string; value: string }>;
    };
    expect(actionsBlock.elements).toHaveLength(3);

    const actionIds = actionsBlock.elements.map((e) => e.action_id);
    expect(actionIds).toContain("question_resolve");
    expect(actionIds).toContain("question_continue");
    expect(actionIds).toContain("question_escalate");

    for (const element of actionsBlock.elements) {
      expect(element.value).toBe("q-001");
    }
  });

  it("should include the response text in the section block", () => {
    const blocks = buildResponseBlocks("My detailed answer", "q-002");

    const sectionBlock = blocks[0] as { text: { text: string } };
    expect(sectionBlock.text.text).toBe("My detailed answer");
  });
});
