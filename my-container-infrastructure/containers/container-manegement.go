/*
Package containers implements simple management of cloud infrastructure for containers
*/
package containers

import (
	"fmt"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
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
		logdriver := ecs.LogDriver_AwsLogs(&ecs.AwsLogDriverProps{
			StreamPrefix: taskConfig.Family,
			LogRetention: logs.RetentionDays_ONE_DAY,
		})
		containerId := jsii.String(fmt.Sprintf("container-%s", *containerConfig.DockerhubImage))
		taskdef.AddContainer(containerId, &ecs.ContainerDefinitionOptions{
			Image: image,
			Logging: logdriver,
		 })

	return taskdef
}

func NewService(
	scope constructs.Construct,
	id *string,
	cluster ecs.Cluster,
	taskDef ecs.FargateTaskDefinition,
	port *float64,
	desiredCount *float64,
	assignPublicIp *bool,
	serviceName *string) ecs.FargateService {
		sgid := fmt.Sprintf("%s-security-group", *id)
		sgdesc := "Security group for service "
		if serviceName != nil {
			sgdesc += *serviceName
		}
		sg := ec2.NewSecurityGroup(scope, jsii.String(sgid), &ec2.SecurityGroupProps{
			Description: jsii.String(sgdesc),
			Vpc: cluster.Vpc(),
		})
		sg.AddIngressRule(ec2.Peer_AnyIpv4(), ec2.Port_Tcp(port), nil, nil)

		service := ecs.NewFargateService(scope, id, &ecs.FargateServiceProps{
			Cluster: cluster,
			TaskDefinition: taskDef,
			DesiredCount: desiredCount,
			ServiceName: serviceName,
			SecurityGroups: &[]ec2.ISecurityGroup{ sg },
			CircuitBreaker: &ecs.DeploymentCircuitBreaker{
				Rollback: jsii.Bool(true),
			},
			AssignPublicIp: assignPublicIp,
		})
		return service
	}