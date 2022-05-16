/*
Package containers implements simple management of cloud infrastructure for containers
*/
package containers

import (
	"fmt"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
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
	Cpu *float64
	MemoryLimitMB *float64
	Family *string
}

// Configuration for container image
type ContainerConfig struct {
	DockerhubImage *string
}

// Add an ECS task definition
func NewTaskDefinitionWithContainer(
	scope constructs.Construct,
	id *string,
	taskConfig TaskConfig,
	containerConfig ContainerConfig) ecs.FargateTaskDefinition {
		taskdef := ecs.NewFargateTaskDefinition(scope, id, &ecs.FargateTaskDefinitionProps{
			Cpu: taskConfig.Cpu,
			MemoryLimitMiB: taskConfig.MemoryLimitMB,
			Family: taskConfig.Family,
		});

		image := ecs.ContainerImage_FromRegistry(containerConfig.DockerhubImage, nil)
		containerId := jsii.String(fmt.Sprintf("container-%s", *containerConfig.DockerhubImage))
		taskdef.AddContainer(containerId, &ecs.ContainerDefinitionOptions{ Image: image })

	return taskdef

}