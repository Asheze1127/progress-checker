/**
 * Local SQS worker for development.
 *
 * Polls ElasticMQ (SQS-compatible) queues and invokes Lambda handlers directly,
 * replicating the SQS → Lambda integration locally.
 *
 * Usage:
 *   SQS_ENDPOINT=http://localhost:9324 tsx worker/local-worker.ts
 */

import {
  SQSClient,
  ReceiveMessageCommand,
  DeleteMessageCommand,
  GetQueueUrlCommand,
} from "@aws-sdk/client-sqs";
import type { Context, SQSEvent, SQSHandler, SQSRecord } from "aws-lambda";

import { handler as questionHandler } from "./question";
import { handler as issueHandler } from "./issue";
import { handler as followupHandler } from "./question-followup";

// --- Configuration ---

const SQS_ENDPOINT = process.env.SQS_ENDPOINT ?? "http://localhost:9324";
const AWS_REGION = process.env.AWS_REGION ?? "ap-northeast-1";
const POLL_WAIT_SECONDS = 20;
const MAX_MESSAGES = 1;

interface QueueWorkerConfig {
  queueName: string;
  handler: SQSHandler;
}

const QUEUE_CONFIGS: QueueWorkerConfig[] = [
  { queueName: "question-new", handler: questionHandler },
  { queueName: "issue", handler: issueHandler },
  { queueName: "question-followup", handler: followupHandler },
];

// --- SQS Client ---

const sqsClient = new SQSClient({
  region: AWS_REGION,
  endpoint: SQS_ENDPOINT,
  credentials: {
    accessKeyId: "local",
    secretAccessKey: "local",
  },
});

// --- Helpers ---

async function resolveQueueUrl(queueName: string): Promise<string> {
  const result = await sqsClient.send(
    new GetQueueUrlCommand({ QueueName: queueName }),
  );
  if (!result.QueueUrl) {
    throw new Error(`Could not resolve URL for queue: ${queueName}`);
  }
  return result.QueueUrl;
}

function buildSqsRecord(
  messageId: string,
  receiptHandle: string,
  body: string,
  queueArn: string,
): SQSRecord {
  return {
    messageId,
    receiptHandle,
    body,
    attributes: {
      ApproximateReceiveCount: "1",
      SentTimestamp: String(Date.now()),
      SenderId: "local",
      ApproximateFirstReceiveTimestamp: String(Date.now()),
    },
    messageAttributes: {},
    md5OfBody: "",
    eventSource: "aws:sqs",
    eventSourceARN: queueArn,
    awsRegion: AWS_REGION,
  };
}

// --- Poll loop for one queue ---

async function pollQueue(config: QueueWorkerConfig): Promise<void> {
  const queueUrl = await resolveQueueUrl(config.queueName);
  const queueArn = `arn:aws:sqs:${AWS_REGION}:000000000000:${config.queueName}`;

  console.log(
    `[local-worker] Polling queue: ${config.queueName} (${queueUrl})`,
  );

  while (true) {
    try {
      const response = await sqsClient.send(
        new ReceiveMessageCommand({
          QueueUrl: queueUrl,
          MaxNumberOfMessages: MAX_MESSAGES,
          WaitTimeSeconds: POLL_WAIT_SECONDS,
        }),
      );

      const messages = response.Messages ?? [];
      if (messages.length === 0) continue;

      for (const message of messages) {
        if (!message.MessageId || !message.ReceiptHandle || !message.Body) {
          console.warn(
            `[local-worker] Skipping malformed message on ${config.queueName}:`,
            JSON.stringify({ messageId: message.MessageId, hasBody: !!message.Body }),
          );
          continue;
        }

        const record = buildSqsRecord(
          message.MessageId,
          message.ReceiptHandle,
          message.Body,
          queueArn,
        );

        const event: SQSEvent = { Records: [record] };

        try {
          console.log(
            `[local-worker] Processing message ${message.MessageId} from ${config.queueName}`,
          );
          const stubContext: Context = {
            callbackWaitsForEmptyEventLoop: true,
            functionName: `local-${config.queueName}`,
            functionVersion: "$LATEST",
            invokedFunctionArn: `arn:aws:lambda:${AWS_REGION}:000000000000:function:local-${config.queueName}`,
            memoryLimitInMB: "512",
            awsRequestId: crypto.randomUUID(),
            logGroupName: `/aws/lambda/local-${config.queueName}`,
            logStreamName: "local",
            getRemainingTimeInMillis: () => 300000,
            done: () => {},
            fail: () => {},
            succeed: () => {},
          };
          await config.handler(event, stubContext, () => {});

          await sqsClient.send(
            new DeleteMessageCommand({
              QueueUrl: queueUrl,
              ReceiptHandle: message.ReceiptHandle,
            }),
          );
          console.log(
            `[local-worker] Successfully processed and deleted ${message.MessageId}`,
          );
        } catch (error) {
          console.error(
            `[local-worker] Handler error for ${message.MessageId} on ${config.queueName}:`,
            error instanceof Error ? error.message : String(error),
          );
          // Back off before re-polling to avoid tight loop on poison messages
          await new Promise((resolve) => setTimeout(resolve, 5000));
        }
      }
    } catch (error) {
      console.error(
        `[local-worker] Poll error on ${config.queueName}:`,
        error instanceof Error ? error.message : String(error),
      );
      await new Promise((resolve) => setTimeout(resolve, 5000));
    }
  }
}

// --- Main ---

async function main(): Promise<void> {
  console.log("[local-worker] Starting local SQS worker");
  console.log(`[local-worker] SQS endpoint: ${SQS_ENDPOINT}`);
  console.log(
    `[local-worker] Queues: ${QUEUE_CONFIGS.map((q) => q.queueName).join(", ")}`,
  );

  await Promise.all(QUEUE_CONFIGS.map((config) => pollQueue(config)));
}

main().catch((error) => {
  console.error("[local-worker] Fatal error:", error);
  process.exit(1);
});
