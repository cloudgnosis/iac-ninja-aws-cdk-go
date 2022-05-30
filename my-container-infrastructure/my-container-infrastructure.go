package main

import (
	"fmt"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"os"
	//"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"my-container-infrastructure/containers"
)

func main() {
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, jsii.String("my-container-infrastructure"), &cdk.StackProps{
		Env: env(),
	})

	var vpc ec2.IVpc
	maybeVpcName := app.Node().TryGetContext(jsii.String("vpcname"))
	if maybeVpcName == nil {
		vpc = ec2.NewVpc(stack, jsii.String("vpc"), &ec2.VpcProps{
			VpcName:     jsii.String("my-vpc"),
			NatGateways: jsii.Number(1),
			MaxAzs:      jsii.Number(2),
		})
	} else {
		vpc = ec2.Vpc_FromLookup(stack, jsii.String("vpc"), &ec2.VpcLookupOptions{
			VpcName: jsii.String(maybeVpcName.(string)),
		})
	}

	var id = "my-test-cluster"
	cluster := containers.NewCluster(stack, jsii.String(id), vpc)

	taskConfig := containers.TaskConfig{
		Cpu:           jsii.Number(512),
		MemoryLimitMB: jsii.Number(1024),
		Family:        jsii.String("webserver"),
	}
	containerConfig := containers.ContainerConfig{
		DockerhubImage: jsii.String("httpd"),
		TcpPorts:       []*float64{jsii.Number(80)},
	}
	taskDefId := fmt.Sprintf("taskdef-%s", *taskConfig.Family)
	taskdef := containers.NewTaskDefinitionWithContainer(stack, &taskDefId, taskConfig, containerConfig)
	serviceId := fmt.Sprintf("service-%s", *taskConfig.Family)
	service := containers.NewLoadBalancedService(
		stack,
		&serviceId,
		cluster,
		taskdef,
		jsii.Number(80),
		jsii.Number(2),
		jsii.Bool(true),
		nil)

	containers.SetServiceScaling(service.Service(), &containers.ServiceScalingConfig{
		MinCount: jsii.Number(1),
		MaxCount: jsii.Number(4),
		ScaleCpuTarget: &containers.ScalingThreshold{
			Percent: jsii.Number(50),
		},
		ScaleMemoryTarget: &containers.ScalingThreshold{
			Percent: jsii.Number(70),
		},
	})

	app.Synth(nil)
}

func env() *cdk.Environment {
	return &cdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
