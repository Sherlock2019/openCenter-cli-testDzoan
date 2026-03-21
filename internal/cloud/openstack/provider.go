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

package openstack

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	floatingips "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	secgroups "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	secrules "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"

	"github.com/opencenter-cloud/opencenter-cli/internal/cloud"
	"github.com/opencenter-cloud/opencenter-cli/internal/config"
)

type serverLookup struct {
	idToName      map[string]string
	addressToName map[string]string
}

// Provider implements the CloudProvider interface for OpenStack.
type Provider struct {
	authOpts gophercloud.AuthOptions
	region   string
}

// NewProvider creates a new OpenStack cloud provider.
func NewProvider(authOpts gophercloud.AuthOptions, region string) *Provider {
	return &Provider{
		authOpts: authOpts,
		region:   region,
	}
}

func (p *Provider) getProviderClient() (*gophercloud.ProviderClient, error) {
	client, err := openstack.AuthenticatedClient(p.authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with OpenStack: %w", err)
	}
	return client, nil
}

func (p *Provider) getComputeClient() (*gophercloud.ServiceClient, error) {
	provider, err := p.getProviderClient()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: p.region})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}
	return client, nil
}

func (p *Provider) getNetworkClient() (*gophercloud.ServiceClient, error) {
	provider, err := p.getProviderClient()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{Region: p.region})
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %w", err)
	}
	return client, nil
}

func (p *Provider) getLoadBalancerClient() (*gophercloud.ServiceClient, error) {
	provider, err := p.getProviderClient()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewLoadBalancerV2(provider, gophercloud.EndpointOpts{Region: p.region})
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer client: %w", err)
	}
	return client, nil
}

func (p *Provider) getBlockStorageClient() (*gophercloud.ServiceClient, error) {
	provider, err := p.getProviderClient()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{Region: p.region})
	if err != nil {
		return nil, fmt.Errorf("failed to create block storage client: %w", err)
	}
	return client, nil
}

// GetCurrentState retrieves the current infrastructure state from OpenStack.
func (p *Provider) GetCurrentState(ctx context.Context, cfg config.Config) (*cloud.InfrastructureState, error) {
	clusterName := cfg.OpenCenter.Cluster.ClusterName
	state := &cloud.InfrastructureState{
		Servers:        []cloud.Server{},
		Networks:       []cloud.Network{},
		SecurityGroups: []cloud.SecurityGroup{},
		LoadBalancers:  []cloud.LoadBalancer{},
		Volumes:        []cloud.Volume{},
		FloatingIPs:    []cloud.FloatingIP{},
	}

	serverList, lookup, err := p.listServers(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	state.Servers = serverList

	networkID := cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Networking.NetworkID
	networkList, err := p.listNetworks(ctx, clusterName, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}
	state.Networks = networkList

	securityGroupList, err := p.listSecurityGroups(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list security groups: %w", err)
	}
	state.SecurityGroups = securityGroupList

	loadBalancerList, err := p.listLoadBalancers(ctx, clusterName, lookup.addressToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %w", err)
	}
	state.LoadBalancers = loadBalancerList

	volumeList, err := p.listVolumes(ctx, clusterName, lookup.idToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	state.Volumes = volumeList

	floatingIPList, err := p.listFloatingIPs(ctx, clusterName, lookup.addressToName)
	if err != nil {
		return nil, fmt.Errorf("failed to list floating IPs: %w", err)
	}
	state.FloatingIPs = floatingIPList

	return state, nil
}

func (p *Provider) listServers(ctx context.Context, clusterName string) ([]cloud.Server, serverLookup, error) {
	client, err := p.getComputeClient()
	if err != nil {
		return nil, serverLookup{}, err
	}

	allPages, err := servers.List(client, servers.ListOpts{}).AllPages()
	if err != nil {
		return nil, serverLookup{}, fmt.Errorf("failed to list servers: %w", err)
	}

	serverList, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, serverLookup{}, fmt.Errorf("failed to extract servers: %w", err)
	}

	result := make([]cloud.Server, 0, len(serverList))
	lookup := serverLookup{
		idToName:      make(map[string]string),
		addressToName: make(map[string]string),
	}

	for _, server := range serverList {
		if !matchesClusterServer(server, clusterName) {
			continue
		}

		networks := extractNetworkNames(server.Addresses)
		addresses := extractServerAddresses(server)
		for _, address := range addresses {
			lookup.addressToName[address] = server.Name
		}
		lookup.idToName[server.ID] = server.Name

		result = append(result, cloud.Server{
			ID:       server.ID,
			Name:     server.Name,
			Flavor:   mapValue(server.Flavor, "id"),
			Image:    mapValue(server.Image, "id"),
			Status:   server.Status,
			Networks: networks,
			Tags:     copyStringMap(server.Metadata),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, lookup, nil
}

func (p *Provider) listNetworks(ctx context.Context, clusterName, networkID string) ([]cloud.Network, error) {
	client, err := p.getNetworkClient()
	if err != nil {
		return nil, err
	}

	networksToConvert := make([]networks.Network, 0)
	if networkID != "" {
		network, err := networks.Get(client, networkID).Extract()
		if err != nil {
			return nil, fmt.Errorf("failed to get network %s: %w", networkID, err)
		}
		networksToConvert = append(networksToConvert, *network)
	} else {
		allPages, err := networks.List(client, networks.ListOpts{}).AllPages()
		if err != nil {
			return nil, fmt.Errorf("failed to list networks: %w", err)
		}

		allNetworks, err := networks.ExtractNetworks(allPages)
		if err != nil {
			return nil, fmt.Errorf("failed to extract networks: %w", err)
		}

		for _, network := range allNetworks {
			if matchesClusterResource(clusterName, network.Tags, network.Name, network.Description) {
				networksToConvert = append(networksToConvert, network)
			}
		}
	}

	result := make([]cloud.Network, 0, len(networksToConvert))
	for _, network := range networksToConvert {
		subnetList, err := p.listSubnets(ctx, network.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list subnets for network %s: %w", network.ID, err)
		}

		networkCIDR := ""
		if len(subnetList) == 1 {
			networkCIDR = subnetList[0].CIDR
		}

		networkName := network.Name
		if networkID != "" {
			networkName = network.ID
		}

		result = append(result, cloud.Network{
			ID:      network.ID,
			Name:    networkName,
			CIDR:    networkCIDR,
			Subnets: subnetList,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listSubnets(ctx context.Context, networkID string) ([]cloud.Subnet, error) {
	client, err := p.getNetworkClient()
	if err != nil {
		return nil, err
	}

	allPages, err := subnets.List(client, subnets.ListOpts{NetworkID: networkID}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %w", err)
	}

	subnetList, err := subnets.ExtractSubnets(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract subnets: %w", err)
	}

	result := make([]cloud.Subnet, 0, len(subnetList))
	for _, subnet := range subnetList {
		result = append(result, cloud.Subnet{
			ID:   subnet.ID,
			Name: subnet.Name,
			CIDR: subnet.CIDR,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listSecurityGroups(ctx context.Context, clusterName string) ([]cloud.SecurityGroup, error) {
	client, err := p.getNetworkClient()
	if err != nil {
		return nil, err
	}

	allPages, err := secgroups.List(client, secgroups.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list security groups: %w", err)
	}

	groupList, err := secgroups.ExtractGroups(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract security groups: %w", err)
	}

	result := make([]cloud.SecurityGroup, 0, len(groupList))
	for _, group := range groupList {
		if !matchesClusterResource(clusterName, group.Tags, group.Name, group.Description) {
			continue
		}

		result = append(result, cloud.SecurityGroup{
			ID:    group.ID,
			Name:  group.Name,
			Rules: convertSecurityRules(group.Rules),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listLoadBalancers(ctx context.Context, clusterName string, addressToName map[string]string) ([]cloud.LoadBalancer, error) {
	client, err := p.getLoadBalancerClient()
	if err != nil {
		return nil, err
	}

	allPages, err := loadbalancers.List(client, loadbalancers.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %w", err)
	}

	loadBalancerList, err := loadbalancers.ExtractLoadBalancers(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract load balancers: %w", err)
	}

	result := make([]cloud.LoadBalancer, 0, len(loadBalancerList))
	for _, lb := range loadBalancerList {
		if !matchesClusterResource(clusterName, lb.Tags, lb.Name, lb.Description) {
			continue
		}

		members, err := p.listPoolMembers(ctx, client, lb.ID, addressToName)
		if err != nil {
			return nil, fmt.Errorf("failed to list pool members for load balancer %s: %w", lb.ID, err)
		}

		protocol := ""
		port := 0
		if len(lb.Listeners) > 0 {
			protocol = lb.Listeners[0].Protocol
			port = lb.Listeners[0].ProtocolPort
		}

		canonicalName := lb.Name
		if strings.Contains(strings.ToLower(lb.Name), strings.ToLower(clusterName)) {
			canonicalName = fmt.Sprintf("%s-api", clusterName)
		}

		result = append(result, cloud.LoadBalancer{
			ID:       lb.ID,
			Name:     canonicalName,
			VIP:      lb.VipAddress,
			Members:  members,
			Protocol: protocol,
			Port:     port,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (p *Provider) listPoolMembers(ctx context.Context, client *gophercloud.ServiceClient, loadBalancerID string, addressToName map[string]string) ([]string, error) {
	allPages, err := pools.List(client, pools.ListOpts{LoadbalancerID: loadBalancerID}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	poolList, err := pools.ExtractPools(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract pools: %w", err)
	}

	members := make([]string, 0)
	for _, pool := range poolList {
		for _, member := range pool.Members {
			if member.Address == "" {
				continue
			}
			if name := addressToName[member.Address]; name != "" {
				members = append(members, name)
				continue
			}
			members = append(members, member.Address)
		}
	}

	sort.Strings(members)
	return dedupeStrings(members), nil
}

func (p *Provider) listVolumes(ctx context.Context, clusterName string, serverIDToName map[string]string) ([]cloud.Volume, error) {
	client, err := p.getBlockStorageClient()
	if err != nil {
		return nil, err
	}

	allPages, err := volumes.List(client, volumes.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	volumeList, err := volumes.ExtractVolumes(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract volumes: %w", err)
	}

	result := make([]cloud.Volume, 0, len(volumeList))
	for _, volume := range volumeList {
		if !matchesClusterVolume(clusterName, volume) {
			continue
		}

		attachedTo := ""
		if len(volume.Attachments) > 0 {
			attachedTo = serverIDToName[volume.Attachments[0].ServerID]
			if attachedTo == "" {
				attachedTo = volume.Attachments[0].ServerID
			}
		}

		result = append(result, cloud.Volume{
			ID:         volume.ID,
			Name:       volume.Name,
			Size:       volume.Size,
			Status:     volume.Status,
			AttachedTo: attachedTo,
		})
	}

	sort.Slice(result, func(i, j int) bool { return cloudVolumeKey(result[i]) < cloudVolumeKey(result[j]) })
	return result, nil
}

func (p *Provider) listFloatingIPs(ctx context.Context, clusterName string, addressToName map[string]string) ([]cloud.FloatingIP, error) {
	client, err := p.getNetworkClient()
	if err != nil {
		return nil, err
	}

	allPages, err := floatingips.List(client, floatingips.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list floating IPs: %w", err)
	}

	floatingIPList, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract floating IPs: %w", err)
	}

	result := make([]cloud.FloatingIP, 0, len(floatingIPList))
	for _, fip := range floatingIPList {
		attachedTo := addressToName[fip.FixedIP]
		if !matchesClusterFloatingIP(clusterName, fip, attachedTo) {
			continue
		}

		result = append(result, cloud.FloatingIP{
			ID:         fip.ID,
			Address:    fip.FloatingIP,
			Status:     fip.Status,
			AttachedTo: attachedTo,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result, nil
}

// DetectDrift compares desired state with actual state and returns a drift report.
func (p *Provider) DetectDrift(ctx context.Context, desired, actual *cloud.InfrastructureState) (*cloud.DriftReport, error) {
	return cloud.CompareInfrastructureState(desired, actual), nil
}

// ReconcileDrift applies safe mutable changes only.
func (p *Provider) ReconcileDrift(ctx context.Context, drift *cloud.DriftReport) error {
	if !drift.Reconcilable {
		return fmt.Errorf("drift report contains non-reconcilable items")
	}

	for _, item := range drift.Drifts {
		if !item.Reconcilable {
			continue
		}

		if err := p.reconcileDriftItem(ctx, item); err != nil {
			return fmt.Errorf("failed to reconcile %s %s: %w", item.ResourceType, item.ResourceName, err)
		}
	}

	return nil
}

func (p *Provider) reconcileDriftItem(ctx context.Context, item cloud.DriftItem) error {
	switch item.ResourceType {
	case "server":
		return p.reconcileServerDrift(ctx, item)
	case "security_group":
		return p.reconcileSecurityGroupDrift(ctx, item)
	default:
		return fmt.Errorf("unsupported resource type for reconciliation: %s", item.ResourceType)
	}
}

func (p *Provider) reconcileServerDrift(ctx context.Context, item cloud.DriftItem) error {
	if !strings.HasPrefix(item.Field, "tags.") {
		return fmt.Errorf("unsupported server field for reconciliation: %s", item.Field)
	}

	client, err := p.getComputeClient()
	if err != nil {
		return err
	}

	tagKey := strings.TrimPrefix(item.Field, "tags.")
	updateOpts := servers.MetadataOpts{
		tagKey: fmt.Sprint(item.Expected),
	}

	_, err = servers.UpdateMetadata(client, item.ResourceID, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("failed to update server metadata: %w", err)
	}

	return nil
}

func (p *Provider) reconcileSecurityGroupDrift(ctx context.Context, item cloud.DriftItem) error {
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

	client, err := p.getNetworkClient()
	if err != nil {
		return err
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
		if _, exists := expectedSet[key]; exists {
			continue
		}
		if !isReconcilableOpenStackRule(actualRule) || actualRule.ID == "" {
			continue
		}
		if err := secrules.Delete(client, actualRule.ID).ExtractErr(); err != nil {
			return fmt.Errorf("failed to delete security group rule %s: %w", actualRule.ID, err)
		}
	}

	for key, expectedRule := range expectedSet {
		if _, exists := actualSet[key]; exists {
			continue
		}

		createOpts, err := toSecurityRuleCreateOpts(item.ResourceID, expectedRule)
		if err != nil {
			return err
		}
		if _, err := secrules.Create(client, createOpts).Extract(); err != nil {
			return fmt.Errorf("failed to create security group rule for %s: %w", item.ResourceName, err)
		}
	}

	return nil
}

func matchesClusterServer(server servers.Server, clusterName string) bool {
	if server.Metadata["cluster"] == clusterName {
		return true
	}
	return strings.Contains(strings.ToLower(server.Name), strings.ToLower(clusterName))
}

func matchesClusterResource(clusterName string, tags []string, name, description string) bool {
	lowerCluster := strings.ToLower(clusterName)
	if strings.Contains(strings.ToLower(name), lowerCluster) || strings.Contains(strings.ToLower(description), lowerCluster) {
		return true
	}

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)
		if lowerTag == lowerCluster || lowerTag == "cluster:"+lowerCluster || lowerTag == "cluster="+lowerCluster {
			return true
		}
	}

	return false
}

func matchesClusterVolume(clusterName string, volume volumes.Volume) bool {
	if volume.Metadata["cluster"] == clusterName {
		return true
	}
	if strings.Contains(strings.ToLower(volume.Name), strings.ToLower(clusterName)) {
		return true
	}
	return false
}

func matchesClusterFloatingIP(clusterName string, fip floatingips.FloatingIP, attachedTo string) bool {
	if matchesClusterResource(clusterName, fip.Tags, fip.Description, fip.Description) {
		return true
	}
	if attachedTo != "" && strings.Contains(strings.ToLower(attachedTo), strings.ToLower(clusterName)) {
		return true
	}
	return false
}

func convertSecurityRules(rules []secrules.SecGroupRule) []cloud.SecurityRule {
	result := make([]cloud.SecurityRule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, cloud.SecurityRule{
			ID:          rule.ID,
			Direction:   rule.Direction,
			Protocol:    rule.Protocol,
			PortRange:   portRange(rule.PortRangeMin, rule.PortRangeMax),
			RemoteIP:    rule.RemoteIPPrefix,
			Description: rule.Description,
		})
	}
	sort.Slice(result, func(i, j int) bool { return normalizeRule(result[i]) < normalizeRule(result[j]) })
	return result
}

func extractNetworkNames(addresses map[string]interface{}) []string {
	result := make([]string, 0, len(addresses))
	for name := range addresses {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func extractServerAddresses(server servers.Server) []string {
	addresses := make([]string, 0)
	if server.AccessIPv4 != "" {
		addresses = append(addresses, server.AccessIPv4)
	}
	if server.AccessIPv6 != "" {
		addresses = append(addresses, server.AccessIPv6)
	}

	for _, value := range server.Addresses {
		entries, ok := value.([]interface{})
		if !ok {
			continue
		}
		for _, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			address, _ := entryMap["addr"].(string)
			if address != "" {
				addresses = append(addresses, address)
			}
		}
	}

	sort.Strings(addresses)
	return dedupeStrings(addresses)
}

func copyStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}

	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func mapValue(values map[string]interface{}, key string) string {
	value, ok := values[key]
	if !ok {
		return ""
	}
	stringValue, _ := value.(string)
	return stringValue
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

func portRange(minPort, maxPort int) string {
	if minPort == 0 && maxPort == 0 {
		return ""
	}
	if minPort == maxPort {
		return strconv.Itoa(minPort)
	}
	return fmt.Sprintf("%d-%d", minPort, maxPort)
}

func normalizeRule(rule cloud.SecurityRule) string {
	return strings.ToLower(strings.Join([]string{
		rule.Direction,
		rule.Protocol,
		rule.PortRange,
		rule.RemoteIP,
	}, "|"))
}

func isReconcilableOpenStackRule(rule cloud.SecurityRule) bool {
	if !strings.EqualFold(rule.Direction, "ingress") {
		return false
	}

	remote := strings.TrimSpace(rule.RemoteIP)
	return remote == "0.0.0.0/0" || remote == "::/0"
}

func toSecurityRuleCreateOpts(groupID string, rule cloud.SecurityRule) (secrules.CreateOpts, error) {
	minPort, maxPort, err := parsePortRange(rule.PortRange)
	if err != nil {
		return secrules.CreateOpts{}, fmt.Errorf("failed to parse port range %q: %w", rule.PortRange, err)
	}

	etherType := secrules.RuleEtherType("IPv4")
	if strings.Contains(rule.RemoteIP, ":") {
		etherType = secrules.RuleEtherType("IPv6")
	}

	return secrules.CreateOpts{
		Direction:      secrules.RuleDirection(strings.ToLower(rule.Direction)),
		Description:    rule.Description,
		EtherType:      etherType,
		SecGroupID:     groupID,
		PortRangeMin:   minPort,
		PortRangeMax:   maxPort,
		Protocol:       secrules.RuleProtocol(strings.ToLower(rule.Protocol)),
		RemoteIPPrefix: rule.RemoteIP,
	}, nil
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
