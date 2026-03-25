import * as ec2 from "aws-cdk-lib/aws-ec2";
import type { Construct } from "constructs";

export type NetworkVpcProps = {
  availabilityZones: string[];
  natGatewayCount: number;
  appSubnetName: string;
  databaseSubnetName: string;
};

export function createNetworkVpc(scope: Construct, props: NetworkVpcProps): ec2.Vpc {
  return new ec2.Vpc(scope, "Vpc", {
    availabilityZones: props.availabilityZones,
    natGateways: props.natGatewayCount,
    subnetConfiguration: [
      {
        cidrMask: 24,
        name: "public",
        subnetType: ec2.SubnetType.PUBLIC,
      },
      {
        cidrMask: 24,
        name: props.appSubnetName,
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      {
        cidrMask: 24,
        name: props.databaseSubnetName,
        subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
      },
    ],
  });
}
