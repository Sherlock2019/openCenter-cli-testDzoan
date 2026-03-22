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

package vmware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/opencenter-cloud/opencenter-cli/internal/cloud"
	"github.com/opencenter-cloud/opencenter-cli/internal/config"
)

type clientFactory func(context.Context, *url.URL, bool) (*govmomi.Client, error)

// Provider implements the CloudProvider interface for VMware vSphere.
type Provider struct {
	newClient    clientFactory
	logoutClient bool
}

// NewProvider creates a new VMware cloud provider.
func NewProvider() *Provider {
	return &Provider{
		newClient:    govmomi.NewClient,
		logoutClient: true,
	}
}

// GetCurrentState retrieves the current infrastructure state from VMware vSphere.
func (p *Provider) GetCurrentState(ctx context.Context, cfg config.Config) (*cloud.InfrastructureState, error) {
	sdkURL, insecure, err := buildSDKURL(cfg)
	if err != nil {
		return nil, err
	}

	client, err := p.newClient(ctx, sdkURL, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with VMware: %w", err)
	}
	if p.logoutClient {
		defer func() {
			_ = client.Logout(ctx)
		}()
	}

	finder := find.NewFinder(client.Client, false)
	dc, err := resolveDatacenter(ctx, finder, cfg)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(dc)

	virtualMachines, err := listVirtualMachines(ctx, client, dc)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMware virtual machines: %w", err)
	}

	byName := make(map[string]mo.VirtualMachine, len(virtualMachines))
	byUUID := make(map[string]mo.VirtualMachine, len(virtualMachines))
	for _, vm := range virtualMachines {
		name := virtualMachineName(vm)
		if name != "" {
			byName[name] = vm
		}
		if uuid := strings.ToLower(strings.TrimSpace(vm.Summary.Config.Uuid)); uuid != "" {
			byUUID[uuid] = vm
		}
	}

	state := &cloud.InfrastructureState{
		Servers:        []cloud.Server{},
		Networks:       []cloud.Network{},
		SecurityGroups: []cloud.SecurityGroup{},
		LoadBalancers:  []cloud.LoadBalancer{},
		Volumes:        []cloud.Volume{},
		FloatingIPs:    []cloud.FloatingIP{},
	}

	networkName := strings.TrimSpace(cfg.OpenCenter.Infrastructure.Cloud.VMware.Network)
	if networkName != "" {
		network, err := finder.Network(ctx, networkName)
		if err == nil {
			state.Networks = append(state.Networks, cloud.Network{
				ID:   network.Reference().Value,
				Name: networkName,
			})
		} else if !isFindNotFound(err) {
			return nil, fmt.Errorf("failed to lookup VMware network %q: %w", networkName, err)
		}
	}

	clusterName := cfg.OpenCenter.Cluster.ClusterName
	for _, node := range cfg.OpenCenter.Infrastructure.Cloud.VMware.Nodes {
		vm, ok := matchVirtualMachine(node, byUUID, byName)
		if !ok {
			continue
		}

		serverNetworks, err := resolveReferenceNames(ctx, client, vm.Network)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve networks for VM %s: %w", virtualMachineName(vm), err)
		}
		volumes, err := resolveDatastoreVolumes(ctx, client, vm)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve datastores for VM %s: %w", virtualMachineName(vm), err)
		}

		state.Servers = append(state.Servers, cloud.Server{
			ID:       vm.Reference().Value,
			Name:     virtualMachineName(vm),
			Status:   normalizePowerState(vm.Summary.Runtime.PowerState),
			Networks: serverNetworks,
			Tags: map[string]string{
				"cluster": clusterName,
				"role":    nodeRole(node.Role),
			},
		})
		state.Volumes = append(state.Volumes, volumes...)
	}

	sort.Slice(state.Servers, func(i, j int) bool { return state.Servers[i].Name < state.Servers[j].Name })
	sort.Slice(state.Networks, func(i, j int) bool { return state.Networks[i].Name < state.Networks[j].Name })
	sort.Slice(state.Volumes, func(i, j int) bool { return state.Volumes[i].Name < state.Volumes[j].Name })

	return state, nil
}

// DetectDrift compares desired state with actual state and returns a drift report.
func (p *Provider) DetectDrift(ctx context.Context, desired, actual *cloud.InfrastructureState) (*cloud.DriftReport, error) {
	return cloud.CompareInfrastructureState(desired, actual), nil
}

// ReconcileDrift reports that VMware reconciliation requires manual intervention.
func (p *Provider) ReconcileDrift(ctx context.Context, drift *cloud.DriftReport) error {
	if drift == nil || len(drift.Drifts) == 0 {
		return nil
	}
	return fmt.Errorf("vmware drift reconciliation is not supported; resolve the reported drift manually")
}

func buildSDKURL(cfg config.Config) (*url.URL, bool, error) {
	host := strings.TrimSpace(cfg.Secrets.VSphereCsi.VCenterHost)
	if host == "" {
		host = strings.TrimSpace(cfg.OpenCenter.Infrastructure.Cloud.VMware.VCenterServer)
	}
	if host == "" {
		return nil, false, fmt.Errorf("vmware drift detection requires secrets.vsphere_csi.vcenter_host or infrastructure.cloud.vmware.vcenter_server")
	}

	username := strings.TrimSpace(cfg.Secrets.VSphereCsi.Username)
	password := cfg.Secrets.VSphereCsi.Password
	if username == "" || password == "" {
		return nil, false, fmt.Errorf("vmware drift detection requires secrets.vsphere_csi.username and secrets.vsphere_csi.password")
	}

	port := strings.TrimSpace(cfg.Secrets.VSphereCsi.Port)
	if port == "" {
		port = "443"
	}

	insecure := false
	rawInsecure := strings.TrimSpace(cfg.Secrets.VSphereCsi.InsecureFlag)
	if rawInsecure != "" {
		value, err := strconv.ParseBool(rawInsecure)
		if err != nil {
			return nil, false, fmt.Errorf("invalid secrets.vsphere_csi.insecure_flag %q: %w", rawInsecure, err)
		}
		insecure = value
	}

	target := host
	if !strings.Contains(target, "://") {
		target = "https://" + target
	}

	sdkURL, err := url.Parse(target)
	if err != nil {
		return nil, false, fmt.Errorf("invalid VMware endpoint %q: %w", host, err)
	}

	if sdkURL.Scheme == "" {
		sdkURL.Scheme = "https"
	}

	hostname := sdkURL.Hostname()
	if hostname == "" {
		hostname = sdkURL.Host
	}
	sdkURL.Host = net.JoinHostPort(hostname, port)
	if sdkURL.Path == "" || sdkURL.Path == "/" {
		sdkURL.Path = "/sdk"
	}
	sdkURL.User = url.UserPassword(username, password)

	return sdkURL, insecure, nil
}

func resolveDatacenter(ctx context.Context, finder *find.Finder, cfg config.Config) (*object.Datacenter, error) {
	datacenter := strings.TrimSpace(cfg.OpenCenter.Infrastructure.Cloud.VMware.Datacenter)
	if datacenter == "" {
		for _, candidate := range strings.Split(cfg.Secrets.VSphereCsi.Datacenters, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate != "" {
				datacenter = candidate
				break
			}
		}
	}
	if datacenter == "" {
		return nil, fmt.Errorf("vmware drift detection requires infrastructure.cloud.vmware.datacenter or secrets.vsphere_csi.datacenters")
	}

	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to find VMware datacenter %q: %w", datacenter, err)
	}
	return dc, nil
}

func listVirtualMachines(ctx context.Context, client *govmomi.Client, dc *object.Datacenter) ([]mo.VirtualMachine, error) {
	manager := view.NewManager(client.Client)
	containerView, err := manager.CreateContainerView(ctx, dc.Reference(), []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	defer containerView.Destroy(ctx)

	var virtualMachines []mo.VirtualMachine
	err = containerView.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "network", "datastore"}, &virtualMachines)
	if err != nil {
		return nil, err
	}
	return virtualMachines, nil
}

func matchVirtualMachine(node config.VMNode, byUUID map[string]mo.VirtualMachine, byName map[string]mo.VirtualMachine) (mo.VirtualMachine, bool) {
	if node.UUID != "" {
		vm, ok := byUUID[strings.ToLower(strings.TrimSpace(node.UUID))]
		if ok {
			return vm, true
		}
	}

	vm, ok := byName[node.Name]
	return vm, ok
}

func resolveReferenceNames(ctx context.Context, client *govmomi.Client, refs []types.ManagedObjectReference) ([]string, error) {
	seen := make(map[string]struct{}, len(refs))
	names := make([]string, 0, len(refs))

	for _, ref := range refs {
		key := ref.Type + ":" + ref.Value
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		name, err := resolveReferenceName(ctx, client, ref)
		if err != nil {
			return nil, err
		}
		if name != "" {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

func resolveReferenceName(ctx context.Context, client *govmomi.Client, ref types.ManagedObjectReference) (string, error) {
	nameable, ok := object.NewReference(client.Client, ref).(interface {
		ObjectName(context.Context) (string, error)
	})
	if !ok {
		return "", nil
	}

	name, err := nameable.ObjectName(ctx)
	if err != nil {
		return "", err
	}
	return name, nil
}

func resolveDatastoreVolumes(ctx context.Context, client *govmomi.Client, vm mo.VirtualMachine) ([]cloud.Volume, error) {
	datastoreNames, err := resolveReferenceNames(ctx, client, vm.Datastore)
	if err != nil {
		return nil, err
	}

	volumes := make([]cloud.Volume, 0, len(datastoreNames))
	vmName := virtualMachineName(vm)
	for _, datastoreName := range datastoreNames {
		volumes = append(volumes, cloud.Volume{
			Name:   fmt.Sprintf("%s@%s", vmName, datastoreName),
			Status: "in-use",
		})
	}

	return volumes, nil
}

func virtualMachineName(vm mo.VirtualMachine) string {
	if name := strings.TrimSpace(vm.Summary.Config.Name); name != "" {
		return name
	}
	return strings.TrimSpace(vm.Name)
}

func normalizePowerState(state types.VirtualMachinePowerState) string {
	switch state {
	case types.VirtualMachinePowerStatePoweredOn:
		return "ACTIVE"
	case types.VirtualMachinePowerStatePoweredOff:
		return "POWERED_OFF"
	case types.VirtualMachinePowerStateSuspended:
		return "SUSPENDED"
	default:
		value := strings.TrimSpace(string(state))
		if value == "" {
			return ""
		}
		return strings.ToUpper(value)
	}
}

func nodeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "master", "control-plane":
		return "control-plane"
	default:
		return "worker"
	}
}

func isFindNotFound(err error) bool {
	var notFound *find.NotFoundError
	return errors.As(err, &notFound)
}
