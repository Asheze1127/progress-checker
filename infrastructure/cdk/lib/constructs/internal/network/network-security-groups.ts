import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";

export type NetworkSecurityGroupsProps = {
  vpc: ec2.Vpc;
  apiContainerPort: number;
  databasePort: number;
};

export type NetworkSecurityGroups = {
  publicAlbSecurityGroup: ec2.SecurityGroup;
  internalAlbSecurityGroup: ec2.SecurityGroup;
  apiServiceSecurityGroup: ec2.SecurityGroup;
  lambdaSecurityGroup: ec2.SecurityGroup;
  databaseSecurityGroup: ec2.SecurityGroup;
  endpointSecurityGroup: ec2.SecurityGroup;
};

export function createNetworkSecurityGroups(
  scope: Construct,
  props: NetworkSecurityGroupsProps,
): NetworkSecurityGroups {
  const publicAlbSecurityGroup = new ec2.SecurityGroup(scope, "PublicAlbSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for the public application load balancer.",
    vpc: props.vpc,
  });
  publicAlbSecurityGroup.addIngressRule(
    ec2.Peer.anyIpv4(),
    ec2.Port.tcp(80),
    "Allow inbound HTTP traffic from the internet.",
  );

  const internalAlbSecurityGroup = new ec2.SecurityGroup(scope, "InternalAlbSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for the internal application load balancer.",
    vpc: props.vpc,
  });
  internalAlbSecurityGroup.addIngressRule(
    ec2.Peer.ipv4(props.vpc.vpcCidrBlock),
    ec2.Port.tcp(80),
    "Allow internal HTTP traffic from the VPC.",
  );

  const apiServiceSecurityGroup = new ec2.SecurityGroup(scope, "ApiServiceSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for the ECS API service.",
    vpc: props.vpc,
  });
  apiServiceSecurityGroup.addIngressRule(
    publicAlbSecurityGroup,
    ec2.Port.tcp(props.apiContainerPort),
    "Allow the public ALB to reach the API container.",
  );
  apiServiceSecurityGroup.addIngressRule(
    internalAlbSecurityGroup,
    ec2.Port.tcp(props.apiContainerPort),
    "Allow the internal ALB to reach the API container.",
  );

  const lambdaSecurityGroup = new ec2.SecurityGroup(scope, "LambdaSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for the Lambda functions.",
    vpc: props.vpc,
  });

  const databaseSecurityGroup = new ec2.SecurityGroup(scope, "DatabaseSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for the PostgreSQL instance.",
    vpc: props.vpc,
  });
  databaseSecurityGroup.addIngressRule(
    apiServiceSecurityGroup,
    ec2.Port.tcp(props.databasePort),
    "Allow the ECS API service to connect to PostgreSQL.",
  );
  databaseSecurityGroup.addIngressRule(
    lambdaSecurityGroup,
    ec2.Port.tcp(props.databasePort),
    "Allow the Lambda functions to connect to PostgreSQL.",
  );

  const endpointSecurityGroup = new ec2.SecurityGroup(scope, "EndpointSecurityGroup", {
    allowAllOutbound: true,
    description: "Security group for interface VPC endpoints.",
    vpc: props.vpc,
  });
  endpointSecurityGroup.addIngressRule(
    apiServiceSecurityGroup,
    ec2.Port.tcp(443),
    "Allow the ECS API service to reach interface VPC endpoints.",
  );
  endpointSecurityGroup.addIngressRule(
    lambdaSecurityGroup,
    ec2.Port.tcp(443),
    "Allow the Lambda functions to reach interface VPC endpoints.",
  );

  return {
    publicAlbSecurityGroup,
    internalAlbSecurityGroup,
    apiServiceSecurityGroup,
    lambdaSecurityGroup,
    databaseSecurityGroup,
    endpointSecurityGroup,
  };
}
