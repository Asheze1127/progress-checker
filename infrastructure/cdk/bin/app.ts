#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import { ProgressBoardStack } from "../lib/stacks/progress-board-stack";
import { STAGE_CONFIGS, STAGE_NAMES, type StageName } from "../lib/stacks/stage-config";

const app = new cdk.App();

const requestedStage = app.node.tryGetContext("stage");

if (typeof requestedStage !== "string" || !Object.hasOwn(STAGE_CONFIGS, requestedStage)) {
  throw new Error(`Specify the deployment stage with "-c stage=<${STAGE_NAMES.join("|")}>".`);
}

const stageName = requestedStage as StageName;
const stageConfiguration = STAGE_CONFIGS[stageName];

new ProgressBoardStack(app, stageConfiguration.stackId, {
  configuration: stageConfiguration,
  description: stageConfiguration.description,
  env: {
    region: stageConfiguration.region,
  },
});
