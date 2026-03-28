import type * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import * as wafv2 from "aws-cdk-lib/aws-wafv2";
import type { Construct } from "constructs";

export type EdgeWafProps = {
  publicAlb: elbv2.IApplicationLoadBalancer;
  publicWebAclName: string;
};

export function attachEdgeWaf(scope: Construct, props: EdgeWafProps): void {
  const publicWebAcl = new wafv2.CfnWebACL(scope, "PublicAlbWebAcl", {
    defaultAction: {
      allow: {},
    },
    description: "Managed web ACL for the public ALB.",
    name: props.publicWebAclName,
    rules: [
      {
        name: "AWSManagedRulesCommonRuleSet",
        overrideAction: {
          none: {},
        },
        priority: 1,
        statement: {
          managedRuleGroupStatement: {
            name: "AWSManagedRulesCommonRuleSet",
            vendorName: "AWS",
          },
        },
        visibilityConfig: {
          cloudWatchMetricsEnabled: true,
          metricName: "awsManagedRulesCommonRuleSet",
          sampledRequestsEnabled: true,
        },
      },
      {
        name: "AWSManagedRulesKnownBadInputsRuleSet",
        priority: 2,
        overrideAction: { none: {} },
        statement: {
          managedRuleGroupStatement: {
            vendorName: "AWS",
            name: "AWSManagedRulesKnownBadInputsRuleSet",
          },
        },
        visibilityConfig: {
          cloudWatchMetricsEnabled: true,
          metricName: `${props.publicWebAclName}-known-bad-inputs`,
          sampledRequestsEnabled: true,
        },
      },
      {
        name: "AWSManagedRulesSQLiRuleSet",
        priority: 3,
        overrideAction: { none: {} },
        statement: {
          managedRuleGroupStatement: {
            vendorName: "AWS",
            name: "AWSManagedRulesSQLiRuleSet",
          },
        },
        visibilityConfig: {
          cloudWatchMetricsEnabled: true,
          metricName: `${props.publicWebAclName}-sqli`,
          sampledRequestsEnabled: true,
        },
      },
      {
        name: "AWSManagedRulesAmazonIpReputationList",
        priority: 4,
        overrideAction: { none: {} },
        statement: {
          managedRuleGroupStatement: {
            vendorName: "AWS",
            name: "AWSManagedRulesAmazonIpReputationList",
          },
        },
        visibilityConfig: {
          cloudWatchMetricsEnabled: true,
          metricName: `${props.publicWebAclName}-ip-reputation`,
          sampledRequestsEnabled: true,
        },
      },
      {
        name: "RateLimitRule",
        priority: 5,
        action: { block: {} },
        statement: {
          rateBasedStatement: {
            limit: 2000,
            aggregateKeyType: "IP",
          },
        },
        visibilityConfig: {
          cloudWatchMetricsEnabled: true,
          metricName: `${props.publicWebAclName}-rate-limit`,
          sampledRequestsEnabled: true,
        },
      },
    ],
    scope: "REGIONAL",
    visibilityConfig: {
      cloudWatchMetricsEnabled: true,
      metricName: "publicAlbWebAcl",
      sampledRequestsEnabled: true,
    },
  });

  new wafv2.CfnWebACLAssociation(scope, "PublicAlbWebAclAssociation", {
    resourceArn: props.publicAlb.loadBalancerArn,
    webAclArn: publicWebAcl.attrArn,
  });
}
