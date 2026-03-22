import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import { Construct } from "constructs";

export type EdgeLoadBalancersProps = {
  vpc: ec2.IVpc;
  publicAlbSecurityGroup: ec2.ISecurityGroup;
  internalAlbSecurityGroup: ec2.ISecurityGroup;
  appSubnetName: string;
  apiService: ecs.FargateService;
  apiContainerName: string;
  apiContainerPort: number;
  apiHealthCheckPath: string;
  healthCheckSuccessCodes: string;
};

export type EdgeLoadBalancers = {
  publicAlb: elbv2.ApplicationLoadBalancer;
  internalAlb: elbv2.ApplicationLoadBalancer;
};

export function createEdgeLoadBalancers(scope: Construct, props: EdgeLoadBalancersProps): EdgeLoadBalancers {
  const publicAlb = new elbv2.ApplicationLoadBalancer(scope, "PublicAlb", {
    internetFacing: true,
    securityGroup: props.publicAlbSecurityGroup,
    vpc: props.vpc,
    vpcSubnets: {
      subnetType: ec2.SubnetType.PUBLIC,
    },
  });

  publicAlb
    .addListener("PublicHttpListener", {
      open: false,
      port: 80,
    })
    .addTargets("PublicApiTargets", {
      healthCheck: {
        healthyHttpCodes: props.healthCheckSuccessCodes,
        path: props.apiHealthCheckPath,
        timeout: cdk.Duration.seconds(5),
      },
      port: props.apiContainerPort,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targets: [
        props.apiService.loadBalancerTarget({
          containerName: props.apiContainerName,
          containerPort: props.apiContainerPort,
        }),
      ],
    });

  const internalAlb = new elbv2.ApplicationLoadBalancer(scope, "InternalAlb", {
    internetFacing: false,
    securityGroup: props.internalAlbSecurityGroup,
    vpc: props.vpc,
    vpcSubnets: {
      subnetGroupName: props.appSubnetName,
    },
  });

  internalAlb
    .addListener("InternalHttpListener", {
      open: false,
      port: 80,
    })
    .addTargets("InternalApiTargets", {
      healthCheck: {
        healthyHttpCodes: props.healthCheckSuccessCodes,
        path: props.apiHealthCheckPath,
        timeout: cdk.Duration.seconds(5),
      },
      port: props.apiContainerPort,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targets: [
        props.apiService.loadBalancerTarget({
          containerName: props.apiContainerName,
          containerPort: props.apiContainerPort,
        }),
      ],
    });

  return {
    publicAlb,
    internalAlb,
  };
}
