import * as cdk from "aws-cdk-lib";
import * as sqs from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";

export type QueueNames = {
  queueName: string;
  deadLetterQueueName: string;
};

export class QueueWithDlq extends Construct {
  public readonly queue: sqs.Queue;
  public readonly deadLetterQueue: sqs.Queue;

  public constructor(scope: Construct, id: string, props: QueueNames) {
    super(scope, id);

    const maxReceiveCount = 3;
    const retentionPeriod = cdk.Duration.days(14);
    // Visibility timeout should be greater than the processing time of the message in the lambda multiplied by 6 to avoid multiple lambdas processing the same message simultaneously.
    const visibilityTimeout = cdk.Duration.seconds(60 * 5 * 6);

    this.deadLetterQueue = new sqs.Queue(this, "DeadLetterQueue", {
      queueName: props.deadLetterQueueName,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod,
    });

    this.queue = new sqs.Queue(this, "Queue", {
      queueName: props.queueName,
      deadLetterQueue: {
        queue: this.deadLetterQueue,
        maxReceiveCount,
      },
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod,
      visibilityTimeout,
    });
  }
}
