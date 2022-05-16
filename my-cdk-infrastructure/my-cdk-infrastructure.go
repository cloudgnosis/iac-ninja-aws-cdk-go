package main

import (
	"os"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	//"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)

	stack := awscdk.NewStack(app, jsii.String("my-stack"), &awscdk.StackProps {
		Env: env(),
	})

	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("my-vpc"), &awsec2.VpcLookupOptions {
		IsDefault: jsii.Bool(true),
	})

	awsec2.NewInstance(stack, jsii.String("my-ec2"), &awsec2.InstanceProps {
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE2, awsec2.InstanceSize_MICRO),
		MachineImage: awsec2.MachineImage_LatestAmazonLinux(nil),
		Vpc: vpc,
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
	  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
