import * as ec2 from "aws-cdk-lib/aws-ec2";

export type NetworkVpcEndpointsProps = {
  vpc: ec2.IVpc;
  appSubnetName: string;
  endpointSecurityGroup: ec2.ISecurityGroup;
};

export function createNetworkVpcEndpoints(props: NetworkVpcEndpointsProps): void {
  props.vpc.addInterfaceEndpoint("SqsVpcEndpoint", {
    securityGroups: [props.endpointSecurityGroup],
    service: ec2.InterfaceVpcEndpointAwsService.SQS,
    subnets: {
      subnetGroupName: props.appSubnetName,
    },
  });

  props.vpc.addInterfaceEndpoint("SecretsManagerVpcEndpoint", {
    securityGroups: [props.endpointSecurityGroup],
    service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER,
    subnets: {
      subnetGroupName: props.appSubnetName,
    },
  });

  props.vpc.addInterfaceEndpoint("BedrockRuntimeVpcEndpoint", {
    securityGroups: [props.endpointSecurityGroup],
    service: ec2.InterfaceVpcEndpointAwsService.BEDROCK_RUNTIME,
    subnets: {
      subnetGroupName: props.appSubnetName,
    },
  });
}
