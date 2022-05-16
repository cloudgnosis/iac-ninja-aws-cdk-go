/*
Package containers implements simple management of cloud infrastructure for containers
*/
package containers

import (
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/constructs-go/constructs/v10"
)

// Add an ECS cluster to an existing construct, for a specified VPC
func NewCluster(scope constructs.Construct, id *string, vpc ec2.IVpc) ecs.Cluster {
	return ecs.NewCluster(scope, id, &ecs.ClusterProps{
		Vpc: vpc,
	})
}