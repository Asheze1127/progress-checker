import * as cdk from "aws-cdk-lib";
import type { QueueNames } from "../constructs/queue-with-dlq";

export type StageName = "stg" | "prod";

export interface StageStackConfiguration {
  stageName: StageName;
  stackId: string;
  description: string;
  region: string;
  availabilityZones: string[];
  natGatewayCount: number;
  databaseName: string;
  databaseAllocatedStorageGib: number;
  databaseMaxAllocatedStorageGib: number;
  databaseDeletionProtection: boolean;
  databaseRemovalPolicy: cdk.RemovalPolicy;
  issueApiRecordName: string;
  privateHostedZoneName: string;
  publicWebAclName: string;
  internalApiTokenSecretName?: string;
  queueNames: {
    question: QueueNames;
    issue: QueueNames;
  };
}

const AVAILABILITY_ZONES = ["ap-northeast-1a", "ap-northeast-1c"];
const NAT_GATEWAY_COUNT = 2;

export const STAGE_NAMES: StageName[] = ["stg", "prod"];

export const STAGE_CONFIGS: Record<StageName, StageStackConfiguration> = {
  stg: {
    availabilityZones: AVAILABILITY_ZONES,
    databaseAllocatedStorageGib: 50,
    databaseDeletionProtection: false,
    databaseMaxAllocatedStorageGib: 200,
    databaseName: "progress_checker_stg",
    databaseRemovalPolicy: cdk.RemovalPolicy.DESTROY,
    description: "Staging infrastructure for the KCL Progress Board.",
    issueApiRecordName: "issue-api",
    natGatewayCount: NAT_GATEWAY_COUNT,
    privateHostedZoneName: "stg.internal.example.com",
    publicWebAclName: "kcl-progress-board-stg-public-alb-web-acl",
    internalApiTokenSecretName: "kcl-progress-board/stg/internal-api-token",
    region: "ap-northeast-1",
    stackId: "KclProgressBoardStgStack",
    queueNames: {
      issue: createStageQueueNames("issue", "stg"),
      question: createStageQueueNames("question", "stg"),
    },
    stageName: "stg",
  },
  prod: {
    availabilityZones: AVAILABILITY_ZONES,
    databaseAllocatedStorageGib: 100,
    databaseDeletionProtection: true,
    databaseMaxAllocatedStorageGib: 500,
    databaseName: "progress_checker",
    databaseRemovalPolicy: cdk.RemovalPolicy.RETAIN,
    description: "Production infrastructure for the KCL Progress Board.",
    issueApiRecordName: "issue-api",
    natGatewayCount: NAT_GATEWAY_COUNT,
    privateHostedZoneName: "internal.example.com",
    publicWebAclName: "kcl-progress-board-prod-public-alb-web-acl",
    internalApiTokenSecretName: "kcl-progress-board/prod/internal-api-token",
    region: "ap-northeast-1",
    stackId: "KclProgressBoardProdStack",
    queueNames: {
      issue: createStageQueueNames("issue", "prod"),
      question: createStageQueueNames("question", "prod"),
    },
    stageName: "prod",
  },
};

function createStageQueueNames(queueName: string, stageName: StageName): QueueNames {
  return {
    deadLetterQueueName: `${queueName}-${stageName}-dlq`,
    queueName: `${queueName}-${stageName}`,
  };
}
