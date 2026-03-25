import { describe, it, beforeEach } from "node:test";
import assert from "node:assert/strict";
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

    // Verify AI was called with the question text
    assert.equal(aiCalls.length, 1);
    assert.equal(aiCalls[0], message.text);

    // Verify Slack was called with correct channel and thread
    assert.equal(slackCalls.length, 1);
    assert.equal(slackCalls[0].channelId, message.slack_channel_id);
    assert.equal(slackCalls[0].threadTs, message.slack_thread_ts);
    assert.equal(
      slackCalls[0].text,
      "Here is the AI response to your question.",
    );
  });

  it("should include Block Kit blocks in Slack response", async () => {
    const message = createValidMessage();
    const { deps, slackCalls } = createMockDependencies();

    await processNewQuestion(message, deps);

    const blocks = slackCalls[0].blocks;
    assert.equal(blocks.length, 3);
    assert.equal(blocks[0].type, "section");
    assert.equal(blocks[1].type, "context");
    assert.equal(blocks[2].type, "actions");
  });

  it("should propagate AI generation errors", async () => {
    const message = createValidMessage();
    const { deps } = createMockDependencies();
    deps.generateAiResponse = async () => {
      throw new Error("Bedrock service unavailable");
    };

    await assert.rejects(
      () => processNewQuestion(message, deps),
      (error: unknown) => {
        assert.ok(error instanceof Error);
        assert.ok(error.message.includes("Bedrock service unavailable"));
        return true;
      },
    );
  });

  it("should propagate Slack posting errors", async () => {
    const message = createValidMessage();
    const { deps } = createMockDependencies();
    deps.postToSlack = async () => {
      throw new Error("Slack API error: channel_not_found");
    };

    await assert.rejects(
      () => processNewQuestion(message, deps),
      (error: unknown) => {
        assert.ok(error instanceof Error);
        assert.ok(error.message.includes("channel_not_found"));
        return true;
      },
    );
  });
});

describe("buildResponseBlocks", () => {
  it("should return section, context, and actions blocks", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    assert.equal(blocks.length, 3);
    assert.equal(blocks[0].type, "section");
    assert.equal(blocks[1].type, "context");
    assert.equal(blocks[2].type, "actions");
  });

  it("should include AI disclaimer in context block", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    const contextBlock = blocks[1] as { elements: Array<{ text: string }> };
    assert.ok(
      contextBlock.elements[0].text.includes("AIによる自動生成"),
      "Disclaimer should mention AI-generated content",
    );
  });

  it("should include resolve, continue, and escalate action buttons", () => {
    const blocks = buildResponseBlocks("Test response", "q-001");

    const actionsBlock = blocks[2] as {
      elements: Array<{ action_id: string; value: string }>;
    };
    assert.equal(actionsBlock.elements.length, 3);

    const actionIds = actionsBlock.elements.map((e) => e.action_id);
    assert.ok(actionIds.includes("question_resolve"));
    assert.ok(actionIds.includes("question_continue"));
    assert.ok(actionIds.includes("question_escalate"));

    // All buttons should carry the question ID
    for (const element of actionsBlock.elements) {
      assert.equal(element.value, "q-001");
    }
  });

  it("should include the response text in the section block", () => {
    const blocks = buildResponseBlocks("My detailed answer", "q-002");

    const sectionBlock = blocks[0] as { text: { text: string } };
    assert.equal(sectionBlock.text.text, "My detailed answer");
  });
});
