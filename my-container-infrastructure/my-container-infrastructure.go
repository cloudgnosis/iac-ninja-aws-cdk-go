package main

import (
	"os"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	//"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, jsii.String("my-container-infrastructure"), &cdk.StackProps {
		Env: env(),
	})

	ec2.Vpc_FromLookup(stack, jsii.String("vpc"), &ec2.VpcLookupOptions {
		IsDefault: jsii.Bool(true),
	})

	app.Synth(nil)
}

func env() *cdk.Environment {
	return &cdk.Environment{
	  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}