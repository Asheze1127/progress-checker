import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as rds from "aws-cdk-lib/aws-rds";
import { Construct } from "constructs";
import type { StageName } from "../stacks/stage-config";

export type DatabaseProps = {
  allocatedStorage: number;
  appEnvironment: StageName;
  databaseName: string;
  deletionProtection: boolean;
  maxAllocatedStorage: number;
  removalPolicy: cdk.RemovalPolicy;
  securityGroups: ec2.ISecurityGroup[];
  vpc: ec2.IVpc;
  vpcSubnets: ec2.SubnetSelection;
};

export class Database extends Construct {
  public readonly instance: rds.DatabaseInstance;

  public constructor(scope: Construct, id: string, props: DatabaseProps) {
    super(scope, id);

    this.instance = new rds.DatabaseInstance(this, "PostgresDatabase", {
      allocatedStorage: props.allocatedStorage,
      backupRetention: props.appEnvironment === "prod" ? cdk.Duration.days(7) : cdk.Duration.days(1),
      credentials: rds.Credentials.fromGeneratedSecret("postgres"),
      databaseName: props.databaseName,
      deletionProtection: props.deletionProtection,
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_16_4,
      }),
      instanceType: ec2.InstanceType.of(
        ec2.InstanceClass.T4G,
        props.appEnvironment === "prod" ? ec2.InstanceSize.SMALL : ec2.InstanceSize.MICRO,
      ),
      maxAllocatedStorage: props.maxAllocatedStorage,
      multiAz: props.appEnvironment === "prod",
      publiclyAccessible: false,
      removalPolicy: props.removalPolicy,
      securityGroups: props.securityGroups,
      storageEncrypted: true,
      storageType: rds.StorageType.GP3,
      vpc: props.vpc,
      vpcSubnets: props.vpcSubnets,
    });
  }
}
