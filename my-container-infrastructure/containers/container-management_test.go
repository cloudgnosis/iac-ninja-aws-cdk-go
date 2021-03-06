package containers

import (
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	assertions "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/jsii-runtime-go"
	"strconv"
	"testing"
)

func TestEcsClusterDefinedExistingVpc(t *testing.T) {
	stack := cdk.NewStack(nil, nil, nil)
	vpc := ec2.NewVpc(stack, jsii.String("vpc"), nil)

	cluster := NewCluster(stack, jsii.String("test-cluster"), vpc)

	template := assertions.Template_FromStack(stack)
	template.ResourceCountIs(jsii.String("AWS::ECS::Cluster"), jsii.Number(1))
	if cluster.Vpc() != vpc {
		t.Errorf("Expected cluster VPC to be same as existing VPC")
	}
}

func TestEcsFargateTaskDefinitionDefined(t *testing.T) {
	stack := cdk.NewStack(nil, nil, nil)
	cpuval, memval, familyval := jsii.Number(512), jsii.Number(1024), jsii.String("test")
	taskCfg := TaskConfig{Cpu: cpuval, MemoryLimitMB: memval, Family: familyval}
	imageName := jsii.String("httpd")
	containerCfg := ContainerConfig{DockerhubImage: imageName, TcpPorts: []*float64{jsii.Number(80)}}

	taskdef := NewTaskDefinitionWithContainer(stack, jsii.String("test-taskdef"), taskCfg, containerCfg)

	if !*taskdef.IsFargateCompatible() {
		t.Errorf("Expected task definition to be Fargate compatible, but it isn't")
	}
	gotItem := false
	for _, item := range *stack.Node().Children() {
		if item == taskdef {
			gotItem = true
			break
		}
	}
	if !gotItem {
		t.Errorf("Excpected Task Defintion to be in stack, but it isn't")
	}

	template := assertions.Template_FromStack(stack)
	template.ResourceCountIs(jsii.String("AWS::ECS::TaskDefinition"), jsii.Number(1))
	template.HasResourceProperties(jsii.String("AWS::ECS::TaskDefinition"), &map[string]interface{}{
		"Cpu":    jsii.String(strconv.FormatInt(int64(*cpuval), 10)), // Convert to string instead of float64, due to buggy CDK
		"Memory": jsii.String(strconv.FormatInt(int64(*memval), 10)), // Convert to string instead of float64, due to buggy CDK
		"Family": familyval,
	})
}

func TestContainerDefinitionAddedToTaskDefinition(t *testing.T) {
	stack := cdk.NewStack(nil, nil, nil)
	cpuval, memval, familyval := jsii.Number(512), jsii.Number(1024), jsii.String("test")
	taskCfg := TaskConfig{Cpu: cpuval, MemoryLimitMB: memval, Family: familyval}
	imageName := jsii.String("httpd")
	containerCfg := ContainerConfig{DockerhubImage: imageName, TcpPorts: []*float64{jsii.Number(80)}}

	taskdef := NewTaskDefinitionWithContainer(stack, jsii.String("test-taskdef"), taskCfg, containerCfg)

	containerDef := taskdef.DefaultContainer()
	if containerDef == nil {
		t.Errorf("Expected task definition default container to be defined")
	}
	if containerDef.ImageName() != nil && *containerDef.ImageName() != *imageName {
		t.Errorf("Expected task definition default container to have image name %s, but got %s",
			*imageName, *containerDef.ImageName())
	}

	// Not needed with the test above, included for historical purposes
	template := assertions.Template_FromStack(stack)
	template.HasResourceProperties(jsii.String("AWS::ECS::TaskDefinition"), &map[string]interface{}{
		"ContainerDefinitions": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Image": imageName,
			}),
		}),
	})
}

func setupServiceTest() (cdk.Stack, ecs.Cluster, ecs.TaskDefinition) {
	stack := cdk.NewStack(nil, nil, nil)
	vpc := ec2.NewVpc(stack, jsii.String("vpc"), nil)
	cluster := NewCluster(stack, jsii.String("test-cluster"), vpc)
	cpuval, memval, familyval := jsii.Number(512), jsii.Number(1024), jsii.String("test")
	taskCfg := TaskConfig{Cpu: cpuval, MemoryLimitMB: memval, Family: familyval}
	imageName := jsii.String("httpd")
	containerCfg := ContainerConfig{DockerhubImage: imageName, TcpPorts: []*float64{jsii.Number(80)}}

	taskdef := NewTaskDefinitionWithContainer(stack, jsii.String("test-taskdef"), taskCfg, containerCfg)
	return stack, cluster, taskdef
}

func TestFargateLoadbalancedServiceWithMandatoryProperties(t *testing.T) {
	stack, cluster, taskdef := setupServiceTest()
	port, desiredCount := jsii.Number(80), jsii.Number(1)

	service := NewLoadBalancedService(stack, jsii.String("test-service"), cluster, taskdef, port, desiredCount, nil, nil)

	if service.Cluster() != cluster {
		t.Errorf("Service cluster is not the same as the created cluster")
	}
	if service.TaskDefinition() != taskdef {
		t.Errorf("Service task definition is not the same as the created task definition")
	}

	sgCapture := assertions.NewCapture(nil)
	template := assertions.Template_FromStack(stack)

	template.ResourceCountIs(jsii.String("AWS::ECS::Service"), jsii.Number(1))
	template.HasResourceProperties(jsii.String("AWS::ECS::Service"), &map[string]interface{}{
		"DesiredCount": desiredCount,
		"LaunchType":   jsii.String("FARGATE"),
		"NetworkConfiguration": assertions.Match_ObjectLike(&map[string]interface{}{
			"AwsvpcConfiguration": assertions.Match_ObjectLike(&map[string]interface{}{
				"AssignPublicIp": jsii.String("DISABLED"),
				"SecurityGroups": assertions.Match_ArrayWith(&[]interface{}{sgCapture}),
			}),
		}),
	})

	template.ResourceCountIs(jsii.String("AWS::ElasticLoadBalancingV2::LoadBalancer"), jsii.Number(1))
	template.HasResourceProperties(jsii.String("AWS::ElasticLoadBalancingV2::LoadBalancer"), &map[string]interface{}{
		"Type":   jsii.String("application"),
		"Scheme": jsii.String("internet-facing"),
	})

	template.HasResourceProperties(jsii.String("AWS::EC2::SecurityGroup"), &map[string]interface{}{
		"SecurityGroupIngress": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"CidrIp":     jsii.String("0.0.0.0/0"),
				"FromPort":   port,
				"IpProtocol": jsii.String("tcp"),
			}),
		}),
	})
}

func TestFargateLoadbalancedServiceWithoutPublicAccess(t *testing.T) {
	stack, cluster, taskdef := setupServiceTest()
	port, desiredCount := jsii.Number(80), jsii.Number(1)

	NewLoadBalancedService(stack, jsii.String("test-service"), cluster, taskdef, port, desiredCount, jsii.Bool(false), nil)

	template := assertions.Template_FromStack(stack)

	template.ResourceCountIs(jsii.String("AWS::ElasticLoadBalancingV2::LoadBalancer"), jsii.Number(1))
	template.HasResourceProperties(jsii.String("AWS::ElasticLoadBalancingV2::LoadBalancer"), &map[string]interface{}{
		"Type":   jsii.String("application"),
		"Scheme": jsii.String("internal"),
	})
}

func TestScalingSettingsForLoadBalancedService(t *testing.T) {
	stack, cluster, taskdef := setupServiceTest()
	port, desiredCount := jsii.Number(80), jsii.Number(2)

	service := NewLoadBalancedService(stack, jsii.String("test-service"), cluster, taskdef, port, desiredCount, jsii.Bool(false), nil)

	config := ServiceScalingConfig{ 
		MinCount: jsii.Number(1),
		MaxCount: jsii.Number(5),
		ScaleCpuTarget: &ScalingThreshold{ Percent: jsii.Number(50)},
		ScaleMemoryTarget: &ScalingThreshold{ Percent: jsii.Number(50)},
	}

	SetServiceScaling(service.Service(), &config)

	scaleResource := assertions.NewCapture(nil)
	template := assertions.Template_FromStack(stack)
	template.ResourceCountIs(jsii.String("AWS::ApplicationAutoScaling::ScalableTarget"), jsii.Number(1))
	template.HasResourceProperties(jsii.String("AWS::ApplicationAutoScaling::ScalableTarget"), &map[string]interface{} {
		"MaxCapacity": config.MaxCount,
		"MinCapacity": config.MinCount,
		"ResourceId": scaleResource,
		"ScalableDimension": jsii.String("ecs:service:DesiredCount"),
		"ServiceNamespace": jsii.String("ecs"),
	})

	template.ResourceCountIs(jsii.String("AWS::ApplicationAutoScaling::ScalingPolicy"), jsii.Number(2))
	template.HasResourceProperties(jsii.String("AWS::ApplicationAutoScaling::ScalingPolicy"), &map[string]interface{} {
		"PolicyType": jsii.String("TargetTrackingScaling"),
		"TargetTrackingScalingPolicyConfiguration": assertions.Match_ObjectLike(&map[string]interface{} {
			"PredefinedMetricSpecification": assertions.Match_ObjectEquals(&map[string]interface{} {
				"PredefinedMetricType": jsii.String("ECSServiceAverageCPUUtilization"),
			}),
			"TargetValue": config.ScaleCpuTarget.Percent,
		}),
	})
	template.HasResourceProperties(jsii.String("AWS::ApplicationAutoScaling::ScalingPolicy"), &map[string]interface{} {
		"PolicyType": jsii.String("TargetTrackingScaling"),
		"TargetTrackingScalingPolicyConfiguration": assertions.Match_ObjectLike(&map[string]interface{} {
			"PredefinedMetricSpecification": assertions.Match_ObjectEquals(&map[string]interface{} {
				"PredefinedMetricType": jsii.String("ECSServiceAverageMemoryUtilization"),
			}),
			"TargetValue": config.ScaleMemoryTarget.Percent,
		}),
	})

}