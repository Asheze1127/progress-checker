import * as cdk from "aws-cdk-lib";
import type * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";
import type { StageName } from "../stacks/stage-config";

export type ApiServiceProps = {
  vpc: ec2.IVpc;
  apiServiceSecurityGroup: ec2.ISecurityGroup;
  appEnvironment: StageName;
  appSubnetName: string;
  containerName: string;
  apiContainerPort: number;
  apiContainerImageUri: string;
  databaseName: string;
  issueApiHostname: string;
};

export class ApiService extends Construct {
  public readonly cluster: ecs.Cluster;
  public readonly taskDefinition: ecs.FargateTaskDefinition;
  public readonly service: ecs.FargateService;

  public constructor(scope: Construct, id: string, props: ApiServiceProps) {
    super(scope, id);

    this.cluster = new ecs.Cluster(this, "Cluster", {
      vpc: props.vpc,
    });

    const apiLogGroup = new logs.LogGroup(this, "LogGroup", {
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      retention: logs.RetentionDays.ONE_MONTH,
    });

    this.taskDefinition = new ecs.FargateTaskDefinition(this, "TaskDefinition", {
      cpu: props.appEnvironment === "prod" ? 1024 : 512,
      memoryLimitMiB: props.appEnvironment === "prod" ? 2048 : 1024,
    });

    this.taskDefinition.addContainer(props.containerName, {
      environment: {
        APP_ENV: props.appEnvironment,
        DATABASE_NAME: props.databaseName,
        ISSUE_API_HOSTNAME: props.issueApiHostname,
        PORT: props.apiContainerPort.toString(),
      },
      image: ecs.ContainerImage.fromRegistry(props.apiContainerImageUri),
      logging: ecs.LogDrivers.awsLogs({
        logGroup: apiLogGroup,
        streamPrefix: "api",
      }),
      portMappings: [
        {
          containerPort: props.apiContainerPort,
        },
      ],
    });

    this.service = new ecs.FargateService(this, "Service", {
      assignPublicIp: false,
      cluster: this.cluster,
      desiredCount: props.appEnvironment === "prod" ? 4 : 2,
      healthCheckGracePeriod: cdk.Duration.seconds(60),
      maxHealthyPercent: 200,
      minHealthyPercent: 100,
      securityGroups: [props.apiServiceSecurityGroup],
      taskDefinition: this.taskDefinition,
      vpcSubnets: {
        subnetGroupName: props.appSubnetName,
      },
    });
  }
}
