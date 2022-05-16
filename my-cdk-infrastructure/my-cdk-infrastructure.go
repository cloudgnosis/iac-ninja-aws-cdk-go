package main

import (
	"os"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	iam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	//"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, jsii.String("my-stack"), &cdk.StackProps {
		Env: env(),
	})

	var policies = []iam.IManagedPolicy {
		iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
	}

	role := iam.NewRole(stack, jsii.String("ec2-role"), &iam.RoleProps{
		AssumedBy: iam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &policies,
	})

	vpc := ec2.Vpc_FromLookup(stack, jsii.String("my-vpc"), &ec2.VpcLookupOptions {
		IsDefault: jsii.Bool(true),
	})

	ec2.NewInstance(stack, jsii.String("my-ec2"), &ec2.InstanceProps {
		InstanceType: ec2.InstanceType_Of(ec2.InstanceClass_BURSTABLE2, ec2.InstanceSize_MICRO),
		MachineImage: ec2.MachineImage_LatestAmazonLinux(nil),
		Role: role,
		Vpc: vpc,
	})

	app.Synth(nil)
}

func env() *cdk.Environment {
	return &cdk.Environment{
	  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
