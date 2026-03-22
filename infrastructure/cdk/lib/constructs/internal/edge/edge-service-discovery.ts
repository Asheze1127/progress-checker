import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import * as route53 from "aws-cdk-lib/aws-route53";
import * as route53Targets from "aws-cdk-lib/aws-route53-targets";
import { Construct } from "constructs";

export type EdgeServiceDiscoveryProps = {
  vpc: ec2.IVpc;
  internalAlb: elbv2.IApplicationLoadBalancer;
  privateHostedZoneName: string;
  issueApiRecordName: string;
};

export type EdgeServiceDiscovery = {
  issueApiFqdn: string;
};

export function createEdgeServiceDiscovery(scope: Construct, props: EdgeServiceDiscoveryProps): EdgeServiceDiscovery {
  const privateHostedZone = new route53.PrivateHostedZone(scope, "InternalHostedZone", {
    vpc: props.vpc,
    zoneName: props.privateHostedZoneName,
  });

  const issueApiFqdn = `${props.issueApiRecordName}.${props.privateHostedZoneName}`;

  new route53.ARecord(scope, "IssueApiAliasRecord", {
    recordName: props.issueApiRecordName,
    target: route53.RecordTarget.fromAlias(new route53Targets.LoadBalancerTarget(props.internalAlb)),
    zone: privateHostedZone,
  });

  return {
    issueApiFqdn,
  };
}
