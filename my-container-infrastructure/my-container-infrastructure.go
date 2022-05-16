package main

import (
	"fmt"
	"os"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	//"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"my-container-infrastructure/containers"
)

func main() {
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, jsii.String("my-container-infrastructure"), &cdk.StackProps {
		Env: env(),
	})

	vpc := ec2.Vpc_FromLookup(stack, jsii.String("vpc"), &ec2.VpcLookupOptions {
		IsDefault: jsii.Bool(true),
	})

	var id = "my-test-cluster"
	cluster := containers.NewCluster(stack, jsii.String(id), vpc)

	taskConfig := containers.TaskConfig {
		Cpu: jsii.Number(512),
		MemoryLimitMB: jsii.Number(1024),
		Family: jsii.String("webserver"),
	}
	containerConfig := containers.ContainerConfig {
		DockerhubImage: jsii.String("httpd"),
	}
	taskDefId := fmt.Sprintf("taskdef-%s", *taskConfig.Family)
	taskdef := containers.NewTaskDefinitionWithContainer(stack, &taskDefId, taskConfig, containerConfig)
	serviceId := fmt.Sprintf("service-%s", *taskConfig.Family)
	containers.NewService(
		stack,
		&serviceId,
		cluster,
		taskdef, 
		jsii.Number(80),
		jsii.Number(0),
		jsii.Bool(true),
		nil)

	app.Synth(nil)
}

func env() *cdk.Environment {
	return &cdk.Environment{
	  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}