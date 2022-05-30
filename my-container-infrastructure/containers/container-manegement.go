/*
Package containers implements simple management of cloud infrastructure for containers
*/
package containers

import (
	"fmt"
	appscale "github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	ecspatterns "github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	logs "github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Add an ECS cluster to an existing construct, for a specified VPC
func NewCluster(scope constructs.Construct, id *string, vpc ec2.IVpc) ecs.Cluster {
	return ecs.NewCluster(scope, id, &ecs.ClusterProps{
		Vpc: vpc,
	})
}

// Configuration fo task definition
type TaskConfig struct {
	Cpu           *float64
	MemoryLimitMB *float64
	Family        *string
}

// Configuration for container image
type ContainerConfig struct {
	DockerhubImage *string
	TcpPorts       []*float64
}

// Add an ECS task definition
func NewTaskDefinitionWithContainer(
	scope constructs.Construct,
	id *string,
	taskConfig TaskConfig,
	containerConfig ContainerConfig) ecs.FargateTaskDefinition {
	taskdef := ecs.NewFargateTaskDefinition(scope, id, &ecs.FargateTaskDefinitionProps{
		Cpu:            taskConfig.Cpu,
		MemoryLimitMiB: taskConfig.MemoryLimitMB,
		Family:         taskConfig.Family,
	})

	image := ecs.ContainerImage_FromRegistry(containerConfig.DockerhubImage, nil)
	logdriver := ecs.LogDriver_AwsLogs(&ecs.AwsLogDriverProps{
		StreamPrefix: taskConfig.Family,
		LogRetention: logs.RetentionDays_ONE_DAY,
	})
	containerId := jsii.String(fmt.Sprintf("container-%s", *containerConfig.DockerhubImage))
	containerdef := taskdef.AddContainer(containerId, &ecs.ContainerDefinitionOptions{
		Image:   image,
		Logging: logdriver,
	})

	for _, port := range containerConfig.TcpPorts {
		containerdef.AddPortMappings(&ecs.PortMapping{ContainerPort: port, Protocol: ecs.Protocol_TCP})
	}

	return taskdef
}

func NewLoadBalancedService(
	scope constructs.Construct,
	id *string,
	cluster ecs.Cluster,
	taskDef ecs.FargateTaskDefinition,
	port *float64,
	desiredCount *float64,
	publicEndpoint *bool,
	serviceName *string) ecspatterns.ApplicationLoadBalancedFargateService {
	service := ecspatterns.NewApplicationLoadBalancedFargateService(scope, id, &ecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: taskDef,
		DesiredCount:   desiredCount,
		ServiceName:    serviceName,
		CircuitBreaker: &ecs.DeploymentCircuitBreaker{
			Rollback: jsii.Bool(true),
		},
		PublicLoadBalancer: publicEndpoint,
		ListenerPort:       port,
	})
	return service
}

type ScalingThreshold struct {
	Percent *float64
}

type ServiceScalingConfig struct {
	MinCount *float64
	MaxCount *float64
	ScaleCpuTarget *ScalingThreshold
	ScaleMemoryTarget *ScalingThreshold
}

func SetServiceScaling(service ecs.FargateService, config *ServiceScalingConfig) {
	scaling := service.AutoScaleTaskCount( &appscale.EnableScalingProps{
		MinCapacity: config.MinCount,
		MaxCapacity: config.MaxCount,
	})

	scaling.ScaleOnCpuUtilization(jsii.String("CpuScaling"), &ecs.CpuUtilizationScalingProps{
		TargetUtilizationPercent: config.ScaleCpuTarget.Percent,
	})
	scaling.ScaleOnMemoryUtilization(jsii.String("MemoryScaling"), &ecs.MemoryUtilizationScalingProps{
		TargetUtilizationPercent: config.ScaleMemoryTarget.Percent,
	})
}
