package aws

import (
	"context"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opencenter-cloud/opencenter-cli/internal/cloud"
)

func TestProvider_DetectDrift_SecurityGroupRuleMismatch(t *testing.T) {
	provider := NewProvider("us-east-1", "default")

	desired := &cloud.InfrastructureState{
		SecurityGroups: []cloud.SecurityGroup{
			{
				ID:   "sg-1",
				Name: "prod-cluster-control-plane-sg",
				Rules: []cloud.SecurityRule{
					{
						ID:        "rule-expected",
						Direction: "ingress",
						Protocol:  "tcp",
						PortRange: "6443",
						RemoteIP:  "10.0.0.0/8",
					},
				},
			},
		},
	}

	actual := &cloud.InfrastructureState{
		SecurityGroups: []cloud.SecurityGroup{
			{
				ID:   "sg-1",
				Name: "prod-cluster-control-plane-sg",
				Rules: []cloud.SecurityRule{
					{
						ID:        "rule-actual",
						Direction: "ingress",
						Protocol:  "tcp",
						PortRange: "6443",
						RemoteIP:  "0.0.0.0/0",
					},
				},
			},
		},
	}

	report, err := provider.DetectDrift(context.Background(), desired, actual)
	require.NoError(t, err)
	require.Len(t, report.Drifts, 1)

	drift := report.Drifts[0]
	assert.Equal(t, "security_group", drift.ResourceType)
	assert.Equal(t, "rules", drift.Field)
	assert.Equal(t, cloud.SeverityCritical, drift.Severity)
	assert.True(t, drift.Reconcilable)
}

func TestConvertAWSRules(t *testing.T) {
	group := ec2types.SecurityGroup{
		GroupId:   stringPtr("sg-1"),
		GroupName: stringPtr("prod-cluster-control-plane-sg"),
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: stringPtr("tcp"),
				FromPort:   int32Ptr(6443),
				ToPort:     int32Ptr(6443),
				IpRanges: []ec2types.IpRange{
					{CidrIp: stringPtr("10.0.0.0/8")},
				},
			},
		},
		IpPermissionsEgress: []ec2types.IpPermission{
			{
				IpProtocol: stringPtr("-1"),
				IpRanges: []ec2types.IpRange{
					{CidrIp: stringPtr("0.0.0.0/0")},
				},
			},
		},
	}

	rules := convertAWSRules(group)
	require.Len(t, rules, 2)

	byDirection := map[string]cloud.SecurityRule{}
	for _, rule := range rules {
		byDirection[rule.Direction] = rule
	}

	assert.Equal(t, "tcp", byDirection["ingress"].Protocol)
	assert.Equal(t, "6443", byDirection["ingress"].PortRange)
	assert.Equal(t, "10.0.0.0/8", byDirection["ingress"].RemoteIP)
	assert.Equal(t, "-1", byDirection["egress"].Protocol)
	assert.Equal(t, "0.0.0.0/0", byDirection["egress"].RemoteIP)
}

func TestNormalizeInstanceState(t *testing.T) {
	assert.Equal(t, "ACTIVE", normalizeInstanceState(&ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}))
	assert.Equal(t, "BUILD", normalizeInstanceState(&ec2types.InstanceState{Name: ec2types.InstanceStateNamePending}))
	assert.Equal(t, "STOPPED", normalizeInstanceState(&ec2types.InstanceState{Name: ec2types.InstanceStateNameStopped}))
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
