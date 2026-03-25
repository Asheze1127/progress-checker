import * as cdk from "aws-cdk-lib";
import type * as ec2 from "aws-cdk-lib/aws-ec2";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as eventsources from "aws-cdk-lib/aws-lambda-event-sources";
import type * as secretsmanager from "aws-cdk-lib/aws-secretsmanager";
import type * as sqs from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import * as path from "node:path";
import type { StageName } from "../stacks/stage-config";

const LAMBDA_CODE_PATH = path.resolve(__dirname, "../../../../../lambda/worker");

export type LambdaName = "question" | "issue";

export type LambdaWithSqsProps = {
  appEnvironment: StageName;
  vpc: ec2.IVpc;
  appSubnetName: string;
  lambdaSecurityGroup: ec2.ISecurityGroup;
  queue: sqs.IQueue;
  databaseName: string;
  databaseHost: string;
  databaseSecret?: secretsmanager.ISecret;
  lambdaName: LambdaName;
  extraEnvironment?: Record<string, string>;
};

export class LambdaWithSqs extends Construct {
  public readonly lambdaFunction: lambda.Function;

  public constructor(scope: Construct, id: string, props: LambdaWithSqsProps) {
    super(scope, id);

    this.lambdaFunction = new lambda.Function(this, "Function", {
      code: lambda.Code.fromAsset(LAMBDA_CODE_PATH),
      environment: {
        APP_ENV: props.appEnvironment,
        DATABASE_HOST: props.databaseHost,
        DATABASE_NAME: props.databaseName,
        DATABASE_SECRET_ARN: props.databaseSecret?.secretArn ?? "",
        LAMBDA_NAME: props.lambdaName,
        ...props.extraEnvironment,
      },
      handler: `${props.lambdaName}.handler`,
      memorySize: props.appEnvironment === "prod" ? 1024 : 512,
      runtime: lambda.Runtime.NODEJS_20_X,
      securityGroups: [props.lambdaSecurityGroup],
      timeout: cdk.Duration.seconds(60 * 5),
      vpc: props.vpc,
      vpcSubnets: {
        subnetGroupName: props.appSubnetName,
      },
    });

    this.lambdaFunction.addEventSource(
      new eventsources.SqsEventSource(props.queue, {
        batchSize: 1,
      }),
    );
    props.queue.grantConsumeMessages(this.lambdaFunction);

    if (props.databaseSecret) {
      props.databaseSecret.grantRead(this.lambdaFunction);
    }
  }
}
