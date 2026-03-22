import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import { Construct } from "constructs";
import { createEdgeLoadBalancers } from "./internal/edge/edge-load-balancers";
import { createEdgeServiceDiscovery } from "./internal/edge/edge-service-discovery";
import { attachEdgeWaf } from "./internal/edge/edge-waf";

export interface EdgeProps {
  vpc: ec2.IVpc;
  publicAlbSecurityGroup: ec2.ISecurityGroup;
  internalAlbSecurityGroup: ec2.ISecurityGroup;
  appSubnetName: string;
  apiService: ecs.FargateService;
  apiContainerName: string;
  apiContainerPort: number;
  apiHealthCheckPath: string;
  healthCheckSuccessCodes: string;
  privateHostedZoneName: string;
  issueApiRecordName: string;
  publicWebAclName: string;
}

export class Edge extends Construct {
  public readonly publicAlb: elbv2.ApplicationLoadBalancer;
  public readonly internalAlb: elbv2.ApplicationLoadBalancer;
  public readonly issueApiFqdn: string;

  public constructor(scope: Construct, id: string, props: EdgeProps) {
    super(scope, id);

    const loadBalancers = createEdgeLoadBalancers(this, {
      apiContainerName: props.apiContainerName,
      apiContainerPort: props.apiContainerPort,
      apiHealthCheckPath: props.apiHealthCheckPath,
      apiService: props.apiService,
      appSubnetName: props.appSubnetName,
      healthCheckSuccessCodes: props.healthCheckSuccessCodes,
      internalAlbSecurityGroup: props.internalAlbSecurityGroup,
      publicAlbSecurityGroup: props.publicAlbSecurityGroup,
      vpc: props.vpc,
    });

    this.publicAlb = loadBalancers.publicAlb;
    this.internalAlb = loadBalancers.internalAlb;

    const serviceDiscovery = createEdgeServiceDiscovery(this, {
      internalAlb: this.internalAlb,
      issueApiRecordName: props.issueApiRecordName,
      privateHostedZoneName: props.privateHostedZoneName,
      vpc: props.vpc,
    });

    this.issueApiFqdn = serviceDiscovery.issueApiFqdn;

    attachEdgeWaf(this, {
      publicAlb: this.publicAlb,
      publicWebAclName: props.publicWebAclName,
    });
  }
}
