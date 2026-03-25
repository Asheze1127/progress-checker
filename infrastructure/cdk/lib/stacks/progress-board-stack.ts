import * as cdk from "aws-cdk-lib";
import type { Construct } from "constructs";
import { ApiService } from "../constructs/api-service";
import { Database } from "../constructs/database";
import { Edge } from "../constructs/edge";
import { LambdaWithSqs } from "../constructs/lambda-with-sqs";
import { Network } from "../constructs/network";
import { QueueWithDlq } from "../constructs/queue-with-dlq";
import type { StageStackConfiguration } from "./stage-config";

const APP_SUBNET_NAME = "app";
const DATABASE_SUBNET_NAME = "database";
const API_CONTAINER_NAME = "ApiContainer";
const API_CONTAINER_PORT = 8080;
const DATABASE_PORT = 5432;
const API_HEALTH_CHECK_PATH = "/healthz";
const HEALTH_CHECK_SUCCESS_CODES = "200-399";

export interface ProgressBoardStackProps extends cdk.StackProps {
  configuration: StageStackConfiguration;
}

export class ProgressBoardStack extends cdk.Stack {
  public constructor(scope: Construct, id: string, props: ProgressBoardStackProps) {
    const { configuration, ...stackProps } = props;

    super(scope, id, stackProps);

    cdk.Tags.of(this).add("Environment", configuration.stageName);
    cdk.Tags.of(this).add("Service", "kcl-progress-board");

    const apiContainerImageUri = new cdk.CfnParameter(this, "ApiContainerImageUri", {
      description: "Container image URI for the ECS Go API service.",
      type: "String",
    });

    const databaseName = new cdk.CfnParameter(this, "DatabaseName", {
      default: configuration.databaseName,
      description: "PostgreSQL database name for the application.",
      type: "String",
    });

    const network = new Network(this, "Network", {
      apiContainerPort: API_CONTAINER_PORT,
      appSubnetName: APP_SUBNET_NAME,
      availabilityZones: configuration.availabilityZones,
      databasePort: DATABASE_PORT,
      databaseSubnetName: DATABASE_SUBNET_NAME,
      natGatewayCount: configuration.natGatewayCount,
    });

    const questionQueue = new QueueWithDlq(this, "QuestionQueue", {
      deadLetterQueueName: configuration.queueNames.question.deadLetterQueueName,
      queueName: configuration.queueNames.question.queueName,
    });

    const issueQueue = new QueueWithDlq(this, "IssueQueue", {
      deadLetterQueueName: configuration.queueNames.issue.deadLetterQueueName,
      queueName: configuration.queueNames.issue.queueName,
    });

    const issueApiHostname = `${configuration.issueApiRecordName}.${configuration.privateHostedZoneName}`;

    const apiService = new ApiService(this, "ApiService", {
      apiContainerImageUri: apiContainerImageUri.valueAsString,
      apiContainerPort: API_CONTAINER_PORT,
      apiServiceSecurityGroup: network.apiServiceSecurityGroup,
      appEnvironment: configuration.stageName,
      appSubnetName: APP_SUBNET_NAME,
      containerName: API_CONTAINER_NAME,
      databaseName: databaseName.valueAsString,
      issueApiHostname,
      vpc: network.vpc,
    });

    const database = new Database(this, "Database", {
      allocatedStorage: configuration.databaseAllocatedStorageGib,
      appEnvironment: configuration.stageName,
      databaseName: databaseName.valueAsString,
      deletionProtection: configuration.databaseDeletionProtection,
      maxAllocatedStorage: configuration.databaseMaxAllocatedStorageGib,
      removalPolicy: configuration.databaseRemovalPolicy,
      securityGroups: [network.databaseSecurityGroup],
      vpc: network.vpc,
      vpcSubnets: {
        subnetGroupName: DATABASE_SUBNET_NAME,
      },
    });

    const edge = new Edge(this, "Edge", {
      apiContainerName: API_CONTAINER_NAME,
      apiContainerPort: API_CONTAINER_PORT,
      apiHealthCheckPath: API_HEALTH_CHECK_PATH,
      apiService: apiService.service,
      appSubnetName: APP_SUBNET_NAME,
      healthCheckSuccessCodes: HEALTH_CHECK_SUCCESS_CODES,
      internalAlbSecurityGroup: network.internalAlbSecurityGroup,
      issueApiRecordName: configuration.issueApiRecordName,
      privateHostedZoneName: configuration.privateHostedZoneName,
      publicAlbSecurityGroup: network.publicAlbSecurityGroup,
      publicWebAclName: configuration.publicWebAclName,
      vpc: network.vpc,
    });

    new LambdaWithSqs(this, "QuestionLambda", {
      appEnvironment: configuration.stageName,
      appSubnetName: APP_SUBNET_NAME,
      databaseHost: database.instance.instanceEndpoint.hostname,
      databaseName: databaseName.valueAsString,
      databaseSecret: database.instance.secret,
      lambdaName: "question",
      lambdaSecurityGroup: network.lambdaSecurityGroup,
      queue: questionQueue.queue,
      vpc: network.vpc,
    });

    new LambdaWithSqs(this, "IssueLambda", {
      appEnvironment: configuration.stageName,
      appSubnetName: APP_SUBNET_NAME,
      databaseHost: database.instance.instanceEndpoint.hostname,
      databaseName: databaseName.valueAsString,
      databaseSecret: database.instance.secret,
      extraEnvironment: {
        ISSUE_API_HOSTNAME: edge.issueApiFqdn,
      },
      lambdaName: "issue",
      lambdaSecurityGroup: network.lambdaSecurityGroup,
      queue: issueQueue.queue,
      vpc: network.vpc,
    });

    new cdk.CfnOutput(this, "PublicAlbDnsName", {
      value: edge.publicAlb.loadBalancerDnsName,
    });

    new cdk.CfnOutput(this, "IssueApiFqdn", {
      value: edge.issueApiFqdn,
    });

    new cdk.CfnOutput(this, "QuestionQueueUrl", {
      value: questionQueue.queue.queueUrl,
    });

    new cdk.CfnOutput(this, "IssueQueueUrl", {
      value: issueQueue.queue.queueUrl,
    });

    if (database.instance.secret) {
      new cdk.CfnOutput(this, "DatabaseSecretArn", {
        value: database.instance.secret.secretArn,
      });
    }
  }
}
