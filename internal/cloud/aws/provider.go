// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/opencenter-cloud/opencenter-cli/internal/cloud"
	"github.com/opencenter-cloud/opencenter-cli/internal/config"
)

type instanceLookup struct {
	idToName         map[string]string
	addressToName    map[string]string
	securityGroupIDs map[string]struct{}
}

// Provider implements the CloudProvider interface for AWS.
type Provider struct {
	region  string
	profile string
}

// NewProvider creates a new AWS cloud provider.
func NewProvider(region, profile string) *Provider {
	return &Provider{
		region:  region,
		profile: profile,
	}
}

func (p *Provider) loadAWSConfig(ctx context.Context) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if strings.TrimSpace(p.region) != "" {
		opts = append(opts, awsconfig.WithRegion(p.region))
	}
	if strings.TrimSpace(p.profile) != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(p.profile))
	}
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

// GetCurrentState retrieves the current infrastructure state from AWS.
func (p *Provider) GetCurrentState(ctx context.Context, cfg config.Config) (*cloud.InfrastructureState, error) {
	if cfg.OpenCenter.Infrastructure.Cloud.AWS.Region != "" {
		p.region = cfg.OpenCenter.Infrastructure.Cloud.AWS.Region
	}
	if cfg.OpenCenter.Infrastructure.Cloud.AWS.Profile != "" {
		p.profile = cfg.OpenCenter.Infrastructure.Cloud.AWS.Profile
	}

	awsCfg, err := p.loadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	ec2Client := ec2.NewFromConfig(awsCfg)
	elbClient := elbv2.NewFromConfig(awsCfg)

	clusterName := cfg.OpenCenter.Cluster.ClusterName
	vpcID := cfg.OpenCenter.Infrastructure.Cloud.AWS.VPCID
	state := &cloud.InfrastructureState{
		Servers:        []cloud.Server{},
		Networks:       []cloud.Network{},
		SecurityGroups: []cloud.SecurityGroup{},
		LoadBalancers:  []cloud.LoadBalancer{},
		Volumes:        []cloud.Volume{},
		FloatingIPs:    []cloud.FloatingIP{},
	}

	serverList, lookup, err := p.listInstances(ctx, ec2Client, clusterName, vpcID)
	if err != nil {
		return nil, fmt.Errorf("failed to list EC2 instances: %w", err)
	}
	state.Servers = serverList

	networkList, err := p.listNetworks(ctx, ec2Client, vpcID, cfg.OpenCenter.Infrastructure.Cloud.AWS.PrivateSubnets, cfg.OpenCenter.Infrastructure.Cloud.AWS.PublicSubnets)
	if err != nil {
		return nil, fmt.Errorf("failed to list VPC networks: %w", err)
	}
	state.Networks = networkList

	securityGroupList, err := p.listSecurityGroups(ctx, ec2Client, clusterName, vpcID, lookup.securityGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list security groups: %w", err)
	}
	state.SecurityGroups = securityGroupList

	loadBalancerList, err := p.listLoadBalancers(ctx, elbClient, clusterName, vpcID, lookup.idToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %w", err)
	}
	state.LoadBalancers = loadBalancerList

	volumeList, err := p.listVolumes(ctx, ec2Client, clusterName, lookup.idToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list EBS volumes: %w", err)
	}
	state.Volumes = volumeList

	floatingIPList, err := p.listElasticIPs(ctx, ec2Client, clusterName, lookup.idToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list elastic IPs: %w", err)
	}
	state.FloatingIPs = floatingIPList

	return state, nil
}

func (p *Provider) listInstances(ctx context.Context, client *ec2.Client, clusterName, vpcID string) ([]cloud.Server, instanceLookup, error) {
	instances, err := p.describeInstances(ctx, client, []ec2types.Filter{
		{Name: aws.String("tag:cluster"), Values: []string{clusterName}},
	})
	if err != nil {
		return nil, instanceLookup{}, err
	}

	if len(instances) == 0 && vpcID != "" {
		instances, err = p.describeInstances(ctx, client, []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		})
		if err != nil {
			return nil, instanceLookup{}, err
		}
	}

	result := make([]cloud.Server, 0, len(instances))
	lookup := instanceLookup{
		idToName:         make(map[string]string),
		addressToName:    make(map[string]string),
		securityGroupIDs: make(map[string]struct{}),
	}

	for _, instance := range instances {
		if !matchesClusterInstance(instance, clusterName) {
			continue
		}

		name := instanceName(instance)
		instanceID := aws.ToString(instance.InstanceId)
		lookup.idToName[instanceID] = name

		for _, address := range instanceAddresses(instance) {
			lookup.addressToName[address] = name
		}
		for _, group := range instance.SecurityGroups {
			if group.GroupId != nil {
				lookup.securityGroupIDs[*group.GroupId] = struct{}{}
			}
		}

		networks := []string{}
		if instance.SubnetId != nil && *instance.SubnetId != "" {
			networks = append(networks, *instance.SubnetId)
		}

		result = append(result, cloud.Server{
			ID:       instanceID,
			Name:     name,
			Flavor:   string(instance.InstanceType),
			Image:    aws.ToString(instance.ImageId),
			Status:   normalizeInstanceState(instance.State),
			Networks: networks,
			Tags:     tagsToMap(instance.Tags),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, lookup, nil
}

func (p *Provider) describeInstances(ctx context.Context, client *ec2.Client, filters []ec2types.Filter) ([]ec2types.Instance, error) {
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{Filters: filters})
	result := make([]ec2types.Instance, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		for _, reservation := range page.Reservations {
			result = append(result, reservation.Instances...)
		}
	}
	return result, nil
}

func (p *Provider) listNetworks(ctx context.Context, client *ec2.Client, vpcID string, privateSubnets, publicSubnets []string) ([]cloud.Network, error) {
	subnetIDs := append([]string(nil), privateSubnets...)
	subnetIDs = append(subnetIDs, publicSubnets...)

	subnetInput := &ec2.DescribeSubnetsInput{}
	if len(subnetIDs) > 0 {
		subnetInput.SubnetIds = subnetIDs
	} else if vpcID != "" {
		subnetInput.Filters = []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		}
	}

	subnetOutput, err := client.DescribeSubnets(ctx, subnetInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	if vpcID == "" && len(subnetOutput.Subnets) > 0 {
		vpcID = aws.ToString(subnetOutput.Subnets[0].VpcId)
	}
	if vpcID == "" {
		return []cloud.Network{}, nil
	}

	vpcOutput, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{VpcIds: []string{vpcID}})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPC %s: %w", vpcID, err)
	}

	cidr := ""
	if len(vpcOutput.Vpcs) > 0 {
		cidr = aws.ToString(vpcOutput.Vpcs[0].CidrBlock)
	}

	subnets := make([]cloud.Subnet, 0, len(subnetOutput.Subnets))
	for _, subnet := range subnetOutput.Subnets {
		if aws.ToString(subnet.VpcId) != vpcID {
			continue
		}
		subnets = append(subnets, cloud.Subnet{
			ID:   aws.ToString(subnet.SubnetId),
			Name: aws.ToString(subnet.SubnetId),
			CIDR: aws.ToString(subnet.CidrBlock),
		})
	}

	sort.Slice(subnets, func(i, j int) bool { return subnets[i].Name < subnets[j].Name })
	return []cloud.Network{{
		ID:      vpcID,
		Name:    vpcID,
		CIDR:    cidr,
		Subnets: subnets,
	}}, nil
}

func (p *Provider) listSecurityGroups(ctx context.Context, client *ec2.Client, clusterName, vpcID string, instanceSecurityGroups map[string]struct{}) ([]cloud.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{}
	if vpcID != "" {
		input.Filters = []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		}
	}

	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, input)
	result := make([]cloud.SecurityGroup, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe security groups: %w", err)
		}
		for _, group := range page.SecurityGroups {
			groupID := aws.ToString(group.GroupId)
			_, attachedToCluster := instanceSecurityGroups[groupID]
			if !attachedToCluster && !matchesClusterEC2Tags(group.Tags, clusterName) && !strings.Contains(strings.ToLower(aws.ToString(group.GroupName)), strings.ToLower(clusterName)) {
				continue
			}

			result = append(result, cloud.SecurityGroup{
				ID:    groupID,
				Name:  aws.ToString(group.GroupName),
				Rules: convertAWSRules(group),
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listLoadBalancers(ctx context.Context, client *elbv2.Client, clusterName, vpcID string, instanceIDToName map[string]string) ([]cloud.LoadBalancer, error) {
	paginator := elbv2.NewDescribeLoadBalancersPaginator(client, &elbv2.DescribeLoadBalancersInput{})
	result := make([]cloud.LoadBalancer, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe load balancers: %w", err)
		}

		for _, lb := range page.LoadBalancers {
			name := aws.ToString(lb.LoadBalancerName)
			if !strings.Contains(strings.ToLower(name), strings.ToLower(clusterName)) && aws.ToString(lb.VpcId) != vpcID {
				continue
			}

			listenersOutput, err := client.DescribeListeners(ctx, &elbv2.DescribeListenersInput{
				LoadBalancerArn: lb.LoadBalancerArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe listeners for %s: %w", name, err)
			}

			targetGroupsOutput, err := client.DescribeTargetGroups(ctx, &elbv2.DescribeTargetGroupsInput{
				LoadBalancerArn: lb.LoadBalancerArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe target groups for %s: %w", name, err)
			}

			members := make([]string, 0)
			for _, targetGroup := range targetGroupsOutput.TargetGroups {
				healthOutput, err := client.DescribeTargetHealth(ctx, &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: targetGroup.TargetGroupArn,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe target health for %s: %w", name, err)
				}
				for _, target := range healthOutput.TargetHealthDescriptions {
					targetID := aws.ToString(target.Target.Id)
					if memberName := instanceIDToName[targetID]; memberName != "" {
						members = append(members, memberName)
						continue
					}
					if targetID != "" {
						members = append(members, targetID)
					}
				}
			}

			protocol := ""
			port := 0
			if len(listenersOutput.Listeners) > 0 {
				protocol = string(listenersOutput.Listeners[0].Protocol)
				port = int(aws.ToInt32(listenersOutput.Listeners[0].Port))
			}

			sort.Strings(members)
			canonicalName := name
			if strings.Contains(strings.ToLower(name), strings.ToLower(clusterName)) {
				canonicalName = fmt.Sprintf("%s-api", clusterName)
			}

			result = append(result, cloud.LoadBalancer{
				ID:       aws.ToString(lb.LoadBalancerArn),
				Name:     canonicalName,
				VIP:      aws.ToString(lb.DNSName),
				Members:  dedupeStrings(members),
				Protocol: protocol,
				Port:     port,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listVolumes(ctx context.Context, client *ec2.Client, clusterName string, instanceIDToName map[string]string) ([]cloud.Volume, error) {
	paginator := ec2.NewDescribeVolumesPaginator(client, &ec2.DescribeVolumesInput{})
	result := make([]cloud.Volume, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe volumes: %w", err)
		}

		for _, volume := range page.Volumes {
			attachedTo := ""
			if len(volume.Attachments) > 0 {
				attachedTo = instanceIDToName[aws.ToString(volume.Attachments[0].InstanceId)]
				if attachedTo == "" {
					attachedTo = aws.ToString(volume.Attachments[0].InstanceId)
				}
			}

			if !matchesClusterEC2Tags(volume.Tags, clusterName) && !strings.Contains(strings.ToLower(attachedTo), strings.ToLower(clusterName)) {
				continue
			}

			result = append(result, cloud.Volume{
				ID:         aws.ToString(volume.VolumeId),
				Name:       aws.ToString(tagValue(volume.Tags, "Name")),
				Size:       int(aws.ToInt32(volume.Size)),
				Status:     string(volume.State),
				AttachedTo: attachedTo,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return cloudVolumeKey(result[i]) < cloudVolumeKey(result[j]) })
	return result, nil
}

func (p *Provider) listElasticIPs(ctx context.Context, client *ec2.Client, clusterName string, instanceIDToName map[string]string) ([]cloud.FloatingIP, error) {
	output, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe elastic IPs: %w", err)
	}

	result := make([]cloud.FloatingIP, 0, len(output.Addresses))
	for _, address := range output.Addresses {
		attachedTo := instanceIDToName[aws.ToString(address.InstanceId)]
		if !matchesClusterEC2Tags(address.Tags, clusterName) && !strings.Contains(strings.ToLower(attachedTo), strings.ToLower(clusterName)) {
			continue
		}

		status := "DOWN"
		if aws.ToString(address.AssociationId) != "" || aws.ToString(address.InstanceId) != "" {
			status = "ACTIVE"
		}

		result = append(result, cloud.FloatingIP{
			ID:         aws.ToString(address.AllocationId),
			Address:    aws.ToString(address.PublicIp),
			Status:     status,
			AttachedTo: attachedTo,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result, nil
}

// DetectDrift compares desired state with actual state for AWS resources.
func (p *Provider) DetectDrift(ctx context.Context, desired, actual *cloud.InfrastructureState) (*cloud.DriftReport, error) {
	return cloud.CompareInfrastructureState(desired, actual), nil
}

// ReconcileDrift applies safe mutable drift changes in AWS.
func (p *Provider) ReconcileDrift(ctx context.Context, drift *cloud.DriftReport) error {
	if !drift.Reconcilable {
		return fmt.Errorf("drift report contains non-reconcilable items")
	}

	awsCfg, err := p.loadAWSConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	ec2Client := ec2.NewFromConfig(awsCfg)

	for _, item := range drift.Drifts {
		if !item.Reconcilable {
			continue
		}

		switch item.ResourceType {
		case "server":
			if err := p.reconcileServerDrift(ctx, ec2Client, item); err != nil {
				return fmt.Errorf("failed to reconcile server drift: %w", err)
			}
		case "security_group":
			if err := p.reconcileSecurityGroupDrift(ctx, ec2Client, item); err != nil {
				return fmt.Errorf("failed to reconcile security group drift: %w", err)
			}
		default:
			return fmt.Errorf("unsupported resource type for reconciliation: %s", item.ResourceType)
		}
	}

	return nil
}

func (p *Provider) reconcileServerDrift(ctx context.Context, client *ec2.Client, item cloud.DriftItem) error {
	if !strings.HasPrefix(item.Field, "tags.") {
		return fmt.Errorf("unsupported server field for reconciliation: %s", item.Field)
	}

	tagKey := strings.TrimPrefix(item.Field, "tags.")
	_, err := client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{item.ResourceID},
		Tags: []ec2types.Tag{{
			Key:   aws.String(tagKey),
			Value: aws.String(fmt.Sprint(item.Expected)),
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to update EC2 instance tag: %w", err)
	}

	return nil
}

func (p *Provider) reconcileSecurityGroupDrift(ctx context.Context, client *ec2.Client, item cloud.DriftItem) error {
	if item.Field != "rules" {
		return fmt.Errorf("unsupported security group field for reconciliation: %s", item.Field)
	}

	expectedRules, ok := item.Expected.([]cloud.SecurityRule)
	if !ok {
		return fmt.Errorf("unexpected expected rules type %T", item.Expected)
	}
	actualRules, ok := item.Actual.([]cloud.SecurityRule)
	if !ok {
		return fmt.Errorf("unexpected actual rules type %T", item.Actual)
	}

	expectedSet := make(map[string]cloud.SecurityRule, len(expectedRules))
	for _, rule := range expectedRules {
		expectedSet[normalizeRule(rule)] = rule
	}

	actualSet := make(map[string]cloud.SecurityRule, len(actualRules))
	for _, rule := range actualRules {
		actualSet[normalizeRule(rule)] = rule
	}

	for key, actualRule := range actualSet {
		if _, exists := expectedSet[key]; exists || !isHighRiskRule(actualRule) {
			continue
		}

		permission, direction, err := toEC2Permission(actualRule)
		if err != nil {
			return err
		}

		if direction == "egress" {
			if _, err := client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
				GroupId:       aws.String(item.ResourceID),
				IpPermissions: []ec2types.IpPermission{permission},
			}); err != nil {
				return fmt.Errorf("failed to revoke security group egress rule: %w", err)
			}
			continue
		}

		if _, err := client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(item.ResourceID),
			IpPermissions: []ec2types.IpPermission{permission},
		}); err != nil {
			return fmt.Errorf("failed to revoke security group ingress rule: %w", err)
		}
	}

	for key, expectedRule := range expectedSet {
		if _, exists := actualSet[key]; exists {
			continue
		}

		permission, direction, err := toEC2Permission(expectedRule)
		if err != nil {
			return err
		}

		if direction == "egress" {
			if _, err := client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId:       aws.String(item.ResourceID),
				IpPermissions: []ec2types.IpPermission{permission},
			}); err != nil {
				return fmt.Errorf("failed to authorize security group egress rule: %w", err)
			}
			continue
		}

		if _, err := client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(item.ResourceID),
			IpPermissions: []ec2types.IpPermission{permission},
		}); err != nil {
			return fmt.Errorf("failed to authorize security group ingress rule: %w", err)
		}
	}

	return nil
}

func convertAWSRules(group ec2types.SecurityGroup) []cloud.SecurityRule {
	result := make([]cloud.SecurityRule, 0)
	result = append(result, convertAWSPermissions("ingress", group.IpPermissions)...)
	result = append(result, convertAWSPermissions("egress", group.IpPermissionsEgress)...)
	sort.Slice(result, func(i, j int) bool { return normalizeRule(result[i]) < normalizeRule(result[j]) })
	return result
}

func convertAWSPermissions(direction string, permissions []ec2types.IpPermission) []cloud.SecurityRule {
	result := make([]cloud.SecurityRule, 0)
	for _, permission := range permissions {
		protocol := aws.ToString(permission.IpProtocol)
		portRange := permissionPortRange(permission)

		if len(permission.IpRanges) == 0 && len(permission.Ipv6Ranges) == 0 {
			result = append(result, cloud.SecurityRule{
				Direction: direction,
				Protocol:  protocol,
				PortRange: portRange,
			})
		}

		for _, ipRange := range permission.IpRanges {
			result = append(result, cloud.SecurityRule{
				Direction:   direction,
				Protocol:    protocol,
				PortRange:   portRange,
				RemoteIP:    aws.ToString(ipRange.CidrIp),
				Description: aws.ToString(ipRange.Description),
			})
		}
		for _, ipRange := range permission.Ipv6Ranges {
			result = append(result, cloud.SecurityRule{
				Direction:   direction,
				Protocol:    protocol,
				PortRange:   portRange,
				RemoteIP:    aws.ToString(ipRange.CidrIpv6),
				Description: aws.ToString(ipRange.Description),
			})
		}
	}
	return result
}

func permissionPortRange(permission ec2types.IpPermission) string {
	fromPort := permission.FromPort
	toPort := permission.ToPort
	if fromPort == nil || toPort == nil {
		return ""
	}
	if *fromPort == *toPort {
		return strconv.FormatInt(int64(*fromPort), 10)
	}
	return fmt.Sprintf("%d-%d", *fromPort, *toPort)
}

func tagsToMap(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		if tag.Key == nil {
			continue
		}
		result[*tag.Key] = aws.ToString(tag.Value)
	}
	return result
}

func tagValue(tags []ec2types.Tag, key string) *string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return tag.Value
		}
	}
	return nil
}

func matchesClusterInstance(instance ec2types.Instance, clusterName string) bool {
	if matchesClusterEC2Tags(instance.Tags, clusterName) {
		return true
	}
	return strings.Contains(strings.ToLower(instanceName(instance)), strings.ToLower(clusterName))
}

func matchesClusterEC2Tags(tags []ec2types.Tag, clusterName string) bool {
	lowerCluster := strings.ToLower(clusterName)
	for _, tag := range tags {
		key := strings.ToLower(aws.ToString(tag.Key))
		value := strings.ToLower(aws.ToString(tag.Value))
		if value == lowerCluster {
			return true
		}
		if key == "cluster" && value == lowerCluster {
			return true
		}
		if key == "name" && strings.Contains(value, lowerCluster) {
			return true
		}
	}
	return false
}

func instanceName(instance ec2types.Instance) string {
	if name := aws.ToString(tagValue(instance.Tags, "Name")); name != "" {
		return name
	}
	return aws.ToString(instance.InstanceId)
}

func instanceAddresses(instance ec2types.Instance) []string {
	addresses := make([]string, 0)
	if instance.PrivateIpAddress != nil {
		addresses = append(addresses, *instance.PrivateIpAddress)
	}
	if instance.PublicIpAddress != nil {
		addresses = append(addresses, *instance.PublicIpAddress)
	}
	for _, iface := range instance.NetworkInterfaces {
		if iface.PrivateIpAddress != nil {
			addresses = append(addresses, *iface.PrivateIpAddress)
		}
		for _, addr := range iface.PrivateIpAddresses {
			if addr.PrivateIpAddress != nil {
				addresses = append(addresses, *addr.PrivateIpAddress)
			}
		}
		if iface.Association != nil && iface.Association.PublicIp != nil {
			addresses = append(addresses, *iface.Association.PublicIp)
		}
	}

	sort.Strings(addresses)
	return dedupeStrings(addresses)
}

func normalizeInstanceState(state *ec2types.InstanceState) string {
	if state == nil {
		return ""
	}
	switch state.Name {
	case ec2types.InstanceStateNameRunning:
		return "ACTIVE"
	case ec2types.InstanceStateNamePending:
		return "BUILD"
	default:
		return strings.ToUpper(string(state.Name))
	}
}

func cloudVolumeKey(volume cloud.Volume) string {
	if volume.AttachedTo != "" {
		return volume.AttachedTo
	}
	if volume.Name != "" {
		return volume.Name
	}
	return volume.ID
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}

	result := make([]string, 0, len(values))
	last := ""
	for _, value := range values {
		if value == last {
			continue
		}
		result = append(result, value)
		last = value
	}
	return result
}

func normalizeRule(rule cloud.SecurityRule) string {
	return strings.ToLower(strings.Join([]string{
		rule.Direction,
		rule.Protocol,
		rule.PortRange,
		rule.RemoteIP,
	}, "|"))
}

func isHighRiskRule(rule cloud.SecurityRule) bool {
	if !strings.EqualFold(rule.Direction, "ingress") {
		return false
	}
	remote := strings.TrimSpace(rule.RemoteIP)
	if remote != "0.0.0.0/0" && remote != "::/0" {
		return false
	}

	portRange := strings.TrimSpace(rule.PortRange)
	if portRange == "" {
		return true
	}
	if !strings.Contains(portRange, "-") {
		port, err := strconv.Atoi(portRange)
		return err == nil && port == 6443
	}

	parts := strings.SplitN(portRange, "-", 2)
	if len(parts) != 2 {
		return false
	}
	minPort, errMin := strconv.Atoi(parts[0])
	maxPort, errMax := strconv.Atoi(parts[1])
	return errMin == nil && errMax == nil && minPort <= 6443 && 6443 <= maxPort
}

func toEC2Permission(rule cloud.SecurityRule) (ec2types.IpPermission, string, error) {
	fromPort, toPort, err := parsePortRange(rule.PortRange)
	if err != nil {
		return ec2types.IpPermission{}, "", fmt.Errorf("failed to parse port range %q: %w", rule.PortRange, err)
	}

	permission := ec2types.IpPermission{
		IpProtocol: aws.String(rule.Protocol),
	}
	if rule.PortRange != "" {
		permission.FromPort = aws.Int32(int32(fromPort))
		permission.ToPort = aws.Int32(int32(toPort))
	}

	if strings.Contains(rule.RemoteIP, ":") {
		permission.Ipv6Ranges = []ec2types.Ipv6Range{{
			CidrIpv6:    aws.String(rule.RemoteIP),
			Description: aws.String(rule.Description),
		}}
	} else if rule.RemoteIP != "" {
		permission.IpRanges = []ec2types.IpRange{{
			CidrIp:      aws.String(rule.RemoteIP),
			Description: aws.String(rule.Description),
		}}
	}

	return permission, strings.ToLower(rule.Direction), nil
}

func parsePortRange(value string) (int, int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, nil
	}
	if !strings.Contains(value, "-") {
		port, err := strconv.Atoi(value)
		if err != nil {
			return 0, 0, err
		}
		return port, port, nil
	}

	parts := strings.SplitN(value, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid port range")
	}
	minPort, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	maxPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return minPort, maxPort, nil
}
