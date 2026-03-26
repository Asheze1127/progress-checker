"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
exports.buildResponseBlocks = buildResponseBlocks;
exports.processNewQuestion = processNewQuestion;
const amazon_bedrock_1 = require("@ai-sdk/amazon-bedrock");
const web_api_1 = require("@slack/web-api");
const ai_1 = require("ai");
// --- Constants ---
const AI_DISCLAIMER = "※ この回答はAIによる自動生成です。正確でない場合があります。";
const BEDROCK_MODEL_ID = "anthropic.claude-3-5-sonnet-20241022-v2:0";
const SYSTEM_PROMPT = `You are a helpful technical mentor assisting hackathon participants.
Provide clear, actionable answers. When the question is ambiguous, ask a focused clarifying question.
Always respond in the same language as the question.`;
// --- Slack helpers ---
function createSlackClient() {
    const token = process.env.SLACK_BOT_TOKEN;
    if (!token) {
        throw new Error("SLACK_BOT_TOKEN environment variable is not set");
    }
    return new web_api_1.WebClient(token);
}
function buildResponseBlocks(responseText, questionId) {
    const actionButtons = [
        {
            type: "button",
            text: { type: "plain_text", text: "解決済み" },
            action_id: "question_resolve",
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
async function processNewQuestion(message, deps) {
    const logger = {
        questionId: message.question_id,
        participantId: message.participant_id,
        type: message.type,
    };
    console.log(JSON.stringify({ event: "processing_start", ...logger }));
    const responseText = await deps.generateAiResponse(message.text);
    console.log(JSON.stringify({
        event: "ai_response_generated",
        responseLength: responseText.length,
        ...logger,
    }));
    const blocks = buildResponseBlocks(responseText, message.question_id);
    await deps.postToSlack(message.slack_channel_id, message.slack_thread_ts, responseText, blocks);
    console.log(JSON.stringify({ event: "slack_response_posted", ...logger }));
}
// --- Production dependencies ---
function createProductionDependencies() {
    return {
        generateAiResponse: async (questionText) => {
            const { text } = await (0, ai_1.generateText)({
                model: (0, amazon_bedrock_1.bedrock)(BEDROCK_MODEL_ID),
                system: SYSTEM_PROMPT,
                prompt: questionText,
            });
            return text;
        },
        postToSlack: async (channelId, threadTs, text, blocks) => {
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
const handler = async (event) => {
    const deps = createProductionDependencies();
    for (const record of event.Records) {
        try {
            const message = JSON.parse(record.body);
            await processNewQuestion(message, deps);
        }
        catch (error) {
            const errorMessage = error instanceof Error ? error.message : String(error);
            console.error(JSON.stringify({
                event: "record_processing_error",
                messageId: record.messageId,
                error: errorMessage,
            }));
            // Re-throw so SQS can retry this specific record
            throw error;
        }
    }
};
exports.handler = handler;
