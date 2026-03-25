import test from "node:test";
import assert from "node:assert/strict";
import * as cdk from "aws-cdk-lib";
import { Match, Template } from "aws-cdk-lib/assertions";
import { ProgressBoardStack } from "../lib/stacks/progress-board-stack";
import { STAGE_CONFIGS } from "../lib/stacks/stage-config";

test("the staging stack keeps stage-specific names and tags", () => {
  const app = new cdk.App();
  const stack = new ProgressBoardStack(app, "TestStgStack", {
    configuration: STAGE_CONFIGS.stg,
  });
  const template = Template.fromStack(stack);

  template.hasParameter("DatabaseName", {
    Default: "progress_checker_stg",
  });

  template.hasResourceProperties("AWS::Route53::HostedZone", {
    Name: "stg.internal.example.com.",
  });

  template.hasResourceProperties("AWS::RDS::DBInstance", {
    BackupRetentionPeriod: 1,
    DBName: {
      Ref: "DatabaseName",
    },
    DBInstanceClass: "db.t4g.micro",
    Engine: "postgres",
    MultiAZ: false,
    PubliclyAccessible: false,
    StorageEncrypted: true,
  });

  template.hasResourceProperties("AWS::WAFv2::WebACL", {
    Scope: "REGIONAL",
  });

  template.hasResourceProperties("AWS::EC2::SecurityGroup", {
    GroupDescription: "Security group for the Lambda functions.",
  });

  template.resourceCountIs("AWS::Lambda::Function", 2);
  template.resourceCountIs("AWS::Lambda::EventSourceMapping", 2);

  template.hasResourceProperties("AWS::Lambda::Function", {
    Handler: "issue.handler",
    MemorySize: 512,
    Runtime: "nodejs20.x",
    Timeout: 300,
    Environment: {
      Variables: Match.objectLike({
        APP_ENV: "stg",
        ISSUE_API_HOSTNAME: "issue-api.stg.internal.example.com",
        LAMBDA_NAME: "issue",
      }),
    },
  });

  template.hasResourceProperties("AWS::SQS::Queue", {
    QueueName: "issue-stg",
  });

  assert.equal(
    Object.values(template.findResources("AWS::Lambda::Function")).some((resource) => {
      const lambdaResource = resource as { Properties?: { Handler?: string } };
      return lambdaResource.Properties?.Handler === "progress.handler";
    }),
    false,
  );

  template.hasResourceProperties("AWS::ECS::TaskDefinition", {
    Cpu: "512",
    ContainerDefinitions: Match.arrayWith([
      Match.objectLike({
        Environment: Match.arrayWith([
          {
            Name: "APP_ENV",
            Value: "stg",
          },
        ]),
      }),
    ]),
    Memory: "1024",
  });

  template.hasResourceProperties("AWS::EC2::VPC", {
    Tags: Match.arrayWith([
      {
        Key: "Environment",
        Value: "stg",
      },
    ]),
  });

  assert.ok(template.findResources("AWS::ECS::Service"));
});

test("the production stack uses production sizing defaults", () => {
  const app = new cdk.App();
  const stack = new ProgressBoardStack(app, "TestProdStack", {
    configuration: STAGE_CONFIGS.prod,
  });
  const template = Template.fromStack(stack);

  template.hasParameter("DatabaseName", {
    Default: "progress_checker",
  });

  template.hasResourceProperties("AWS::RDS::DBInstance", {
    AllocatedStorage: "100",
    BackupRetentionPeriod: 7,
    DBInstanceClass: "db.t4g.small",
    DeletionProtection: true,
    MultiAZ: true,
  });

  template.hasResourceProperties("AWS::ECS::Service", {
    DesiredCount: 4,
  });

  template.hasResourceProperties("AWS::ECS::TaskDefinition", {
    Cpu: "1024",
    Memory: "2048",
  });

  template.hasResourceProperties("AWS::SQS::Queue", {
    QueueName: "question-prod",
  });

  template.hasResourceProperties("AWS::Lambda::Function", {
    Handler: "issue.handler",
    MemorySize: 1024,
  });

  assert.equal(
    Object.values(template.findResources("AWS::SQS::Queue")).some((resource) => {
      const queueResource = resource as { Properties?: { QueueName?: string } };
      return queueResource.Properties?.QueueName === "progress-prod";
    }),
    false,
  );
});
