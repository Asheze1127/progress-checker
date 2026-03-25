import type * as ec2 from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";
import { createNetworkSecurityGroups } from "./internal/network/network-security-groups";
import { createNetworkVpcEndpoints } from "./internal/network/network-vpc-endpoints";
import { createNetworkVpc } from "./internal/network/network-vpc";

export type NetworkProps = {
  availabilityZones: string[];
  natGatewayCount: number;
  appSubnetName: string;
  databaseSubnetName: string;
  apiContainerPort: number;
  databasePort: number;
};

export class Network extends Construct {
  public readonly vpc: ec2.Vpc;
  public readonly publicAlbSecurityGroup: ec2.SecurityGroup;
  public readonly internalAlbSecurityGroup: ec2.SecurityGroup;
  public readonly apiServiceSecurityGroup: ec2.SecurityGroup;
  public readonly lambdaSecurityGroup: ec2.SecurityGroup;
  public readonly databaseSecurityGroup: ec2.SecurityGroup;

  public constructor(scope: Construct, id: string, props: NetworkProps) {
    super(scope, id);

    this.vpc = createNetworkVpc(this, {
      appSubnetName: props.appSubnetName,
      availabilityZones: props.availabilityZones,
      databaseSubnetName: props.databaseSubnetName,
      natGatewayCount: props.natGatewayCount,
    });

    const securityGroups = createNetworkSecurityGroups(this, {
      apiContainerPort: props.apiContainerPort,
      databasePort: props.databasePort,
      vpc: this.vpc,
    });

    this.publicAlbSecurityGroup = securityGroups.publicAlbSecurityGroup;
    this.internalAlbSecurityGroup = securityGroups.internalAlbSecurityGroup;
    this.apiServiceSecurityGroup = securityGroups.apiServiceSecurityGroup;
    this.lambdaSecurityGroup = securityGroups.lambdaSecurityGroup;
    this.databaseSecurityGroup = securityGroups.databaseSecurityGroup;

    createNetworkVpcEndpoints({
      appSubnetName: props.appSubnetName,
      endpointSecurityGroup: securityGroups.endpointSecurityGroup,
      vpc: this.vpc,
    });
  }
}
