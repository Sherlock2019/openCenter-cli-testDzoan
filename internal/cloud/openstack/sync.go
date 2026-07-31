package openstack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud"
	gopenstack "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumetypes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/ec2credentials"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	networkexternal "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	clusteropenstacksync "github.com/opencenter-cloud/opencenter-cli/internal/cluster/sync/openstack"
	"gopkg.in/yaml.v3"
)

// DefaultCloudsYAMLPath returns the standard OpenStack client configuration
// path, respecting the conventional override used by openstacksdk and OSC.
func DefaultCloudsYAMLPath() string {
	if path := strings.TrimSpace(os.Getenv("OS_CLIENT_CONFIG_FILE")); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "clouds.yaml"
	}
	return filepath.Join(home, ".config", "openstack", "clouds.yaml")
}

type cloudsFile struct {
	Clouds map[string]cloudProfile `yaml:"clouds"`
}

type cloudProfile struct {
	Auth       cloudAuth `yaml:"auth"`
	RegionName string    `yaml:"region_name"`
}

type cloudAuth struct {
	AuthURL                     string `yaml:"auth_url"`
	Username                    string `yaml:"username"`
	UserID                      string `yaml:"user_id"`
	Password                    string `yaml:"password"`
	ProjectID                   string `yaml:"project_id"`
	ProjectName                 string `yaml:"project_name"`
	ProjectDomainName           string `yaml:"project_domain_name"`
	UserDomainName              string `yaml:"user_domain_name"`
	DomainName                  string `yaml:"domain_name"`
	ApplicationCredentialID     string `yaml:"application_credential_id"`
	ApplicationCredentialName   string `yaml:"application_credential_name"`
	ApplicationCredentialSecret string `yaml:"application_credential_secret"`
}

type clusterSyncClient struct{ profile cloudProfile }

// NewClusterSyncDependencies loads a clouds.yaml profile and adapts Gophercloud
// to the portable cluster reconciliation service.
func NewClusterSyncDependencies(cloudsPath, cloudName string) (clusteropenstacksync.Dependencies, error) {
	data, err := os.ReadFile(cloudsPath)
	if err != nil {
		return clusteropenstacksync.Dependencies{}, fmt.Errorf("read clouds.yaml: %w", err)
	}
	var clouds cloudsFile
	if err := yaml.Unmarshal(data, &clouds); err != nil {
		return clusteropenstacksync.Dependencies{}, fmt.Errorf("parse clouds.yaml: %w", err)
	}
	profile, ok := clouds.Clouds[cloudName]
	if !ok {
		return clusteropenstacksync.Dependencies{}, fmt.Errorf("cloud profile %q not found in %s", cloudName, cloudsPath)
	}
	if strings.TrimSpace(profile.Auth.AuthURL) == "" {
		return clusteropenstacksync.Dependencies{}, fmt.Errorf("cloud profile %q has no auth.auth_url", cloudName)
	}
	client := &clusterSyncClient{profile: profile}
	return clusteropenstacksync.Dependencies{
		DiscoverCore:         client.discoverCore,
		EnsureContainer:      client.ensureContainer,
		CreateEC2Credentials: client.createEC2Credentials,
		CreateAppCredential:  client.createAppCredential,
	}, nil
}

func (c *clusterSyncClient) provider() (*gophercloud.ProviderClient, error) {
	auth := c.profile.Auth
	opts := gophercloud.AuthOptions{
		IdentityEndpoint:            strings.TrimSpace(auth.AuthURL),
		Username:                    strings.TrimSpace(auth.Username),
		UserID:                      strings.TrimSpace(auth.UserID),
		Password:                    auth.Password,
		DomainName:                  firstNonEmpty(strings.TrimSpace(auth.UserDomainName), strings.TrimSpace(auth.DomainName)),
		TenantID:                    strings.TrimSpace(auth.ProjectID),
		TenantName:                  strings.TrimSpace(auth.ProjectName),
		ApplicationCredentialID:     strings.TrimSpace(auth.ApplicationCredentialID),
		ApplicationCredentialName:   strings.TrimSpace(auth.ApplicationCredentialName),
		ApplicationCredentialSecret: auth.ApplicationCredentialSecret,
		AllowReauth:                 true,
	}
	if opts.TenantID == "" && opts.TenantName != "" && strings.TrimSpace(auth.ProjectDomainName) != "" {
		opts.Scope = &gophercloud.AuthScope{ProjectName: opts.TenantName, DomainName: strings.TrimSpace(auth.ProjectDomainName)}
	}
	provider, err := gopenstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("authenticate OpenStack client: %w", err)
	}
	return provider, nil
}

func (c *clusterSyncClient) discoverCore(ctx context.Context, _ string) (clusteropenstacksync.CoreDiscovery, error) {
	_ = ctx
	provider, err := c.provider()
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	endpoint := gophercloud.EndpointOpts{Region: strings.TrimSpace(c.profile.RegionName)}
	compute, err := gopenstack.NewComputeV2(provider, endpoint)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, fmt.Errorf("create compute client: %w", err)
	}
	network, err := gopenstack.NewNetworkV2(provider, endpoint)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, fmt.Errorf("create network client: %w", err)
	}
	image, err := gopenstack.NewImageServiceV2(provider, endpoint)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, fmt.Errorf("create image client: %w", err)
	}
	object, err := gopenstack.NewObjectStorageV1(provider, endpoint)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, fmt.Errorf("create object storage client: %w", err)
	}

	allImages, err := listSyncImages(image)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	allFlavors, err := listSyncFlavors(compute)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	zones, err := listSyncZones(compute)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	internalNetworks, externalNetworkID, err := listSyncNetworks(network)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	internalSubnets, err := listSyncSubnets(network, internalNetworks)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}
	volumeTypes, err := listSyncVolumeTypes(provider, endpoint)
	if err != nil {
		return clusteropenstacksync.CoreDiscovery{}, err
	}

	linux, windows := chooseSyncImages(allImages)
	auth := c.profile.Auth
	return clusteropenstacksync.CoreDiscovery{
		AuthURL: strings.TrimSpace(auth.AuthURL), Region: strings.TrimSpace(c.profile.RegionName),
		ProjectID: strings.TrimSpace(auth.ProjectID), ProjectName: strings.TrimSpace(auth.ProjectName),
		ApplicationCredentialID: strings.TrimSpace(auth.ApplicationCredentialID), ApplicationCredentialSecret: auth.ApplicationCredentialSecret,
		UserDomainName:    firstNonEmpty(strings.TrimSpace(auth.UserDomainName), strings.TrimSpace(auth.DomainName)),
		ProjectDomainName: firstNonEmpty(strings.TrimSpace(auth.ProjectDomainName), strings.TrimSpace(auth.DomainName)),
		ImageID:           linux, ImageIDWindows: windows, ExternalNetworkID: externalNetworkID, FloatingIPPool: externalNetworkID,
		AvailabilityZones: zones, InternalNetworks: internalNetworks, InternalSubnets: internalSubnets,
		SwiftEndpoint: object.Endpoint, Flavors: allFlavors, VolumeTypes: volumeTypes,
	}, nil
}

func (c *clusterSyncClient) ensureContainer(ctx context.Context, req clusteropenstacksync.ContainerRequest) error {
	_ = ctx
	provider, err := c.provider()
	if err != nil {
		return err
	}
	client, err := gopenstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{Region: req.Region})
	if err != nil {
		return fmt.Errorf("create object storage client: %w", err)
	}
	_, err = containers.Create(client, req.Name, containers.CreateOpts{VersionsEnabled: req.EnableVersioning}).Extract()
	return err
}

func (c *clusterSyncClient) createEC2Credentials(ctx context.Context, req clusteropenstacksync.EC2CredentialRequest) (clusteropenstacksync.EC2Credentials, error) {
	_ = ctx
	if strings.TrimSpace(c.profile.Auth.UserID) == "" {
		return clusteropenstacksync.EC2Credentials{}, fmt.Errorf("cloud profile must specify auth.user_id to create EC2 credentials")
	}
	if strings.TrimSpace(req.ProjectID) == "" {
		return clusteropenstacksync.EC2Credentials{}, fmt.Errorf("cloud profile must specify auth.project_id to create EC2 credentials")
	}
	provider, err := c.provider()
	if err != nil {
		return clusteropenstacksync.EC2Credentials{}, err
	}
	identity, err := gopenstack.NewIdentityV3(provider, gophercloud.EndpointOpts{Region: c.profile.RegionName})
	if err != nil {
		return clusteropenstacksync.EC2Credentials{}, fmt.Errorf("create identity client: %w", err)
	}
	credential, err := ec2credentials.Create(identity, c.profile.Auth.UserID, ec2credentials.CreateOpts{TenantID: req.ProjectID}).Extract()
	if err != nil {
		return clusteropenstacksync.EC2Credentials{}, err
	}
	return clusteropenstacksync.EC2Credentials{AccessKeyID: credential.Access, SecretAccessKey: credential.Secret}, nil
}

func (c *clusterSyncClient) createAppCredential(ctx context.Context, req clusteropenstacksync.AppCredentialRequest) (clusteropenstacksync.AppCredential, error) {
	_ = ctx
	if strings.TrimSpace(c.profile.Auth.UserID) == "" {
		return clusteropenstacksync.AppCredential{}, fmt.Errorf("cloud profile must specify auth.user_id to create application credentials")
	}
	provider, err := c.provider()
	if err != nil {
		return clusteropenstacksync.AppCredential{}, err
	}
	identity, err := gopenstack.NewIdentityV3(provider, gophercloud.EndpointOpts{Region: c.profile.RegionName})
	if err != nil {
		return clusteropenstacksync.AppCredential{}, fmt.Errorf("create identity client: %w", err)
	}
	rules := make([]applicationcredentials.AccessRule, 0, len(req.AccessRules))
	for _, rule := range req.AccessRules {
		rules = append(rules, applicationcredentials.AccessRule{Service: rule.Service, Method: rule.Method, Path: rule.Path})
	}
	credential, err := applicationcredentials.Create(identity, c.profile.Auth.UserID, applicationcredentials.CreateOpts{Name: req.Name, Description: req.Description, AccessRules: rules}).Extract()
	if err != nil {
		return clusteropenstacksync.AppCredential{}, err
	}
	return clusteropenstacksync.AppCredential{ID: credential.ID, Secret: credential.Secret}, nil
}

func listSyncImages(client *gophercloud.ServiceClient) ([]images.Image, error) {
	pages, err := images.List(client, images.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("list OpenStack images: %w", err)
	}
	items, err := images.ExtractImages(pages)
	if err != nil {
		return nil, fmt.Errorf("extract OpenStack images: %w", err)
	}
	active := make([]images.Image, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(string(item.Status), "active") {
			active = append(active, item)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		if active[i].Name == active[j].Name {
			return active[i].ID < active[j].ID
		}
		return active[i].Name < active[j].Name
	})
	return active, nil
}

func chooseSyncImages(items []images.Image) (linux, windows string) {
	for _, item := range items {
		name := strings.ToLower(item.Name)
		if windows == "" && strings.Contains(name, "windows") {
			windows = item.ID
			continue
		}
		if linux == "" && !strings.Contains(name, "windows") {
			linux = item.ID
		}
	}
	return linux, windows
}

func listSyncFlavors(client *gophercloud.ServiceClient) ([]clusteropenstacksync.Flavor, error) {
	pages, err := flavors.ListDetail(client, nil).AllPages()
	if err != nil {
		return nil, fmt.Errorf("list OpenStack flavors: %w", err)
	}
	items, err := flavors.ExtractFlavors(pages)
	if err != nil {
		return nil, fmt.Errorf("extract OpenStack flavors: %w", err)
	}
	out := make([]clusteropenstacksync.Flavor, 0, len(items))
	for _, item := range items {
		out = append(out, clusteropenstacksync.Flavor{ID: item.ID, Name: item.Name, VCPUs: item.VCPUs, RAM: item.RAM, Disk: item.Disk})
	}
	return out, nil
}

func listSyncZones(client *gophercloud.ServiceClient) ([]string, error) {
	pages, err := availabilityzones.List(client).AllPages()
	if err != nil {
		return nil, fmt.Errorf("list OpenStack availability zones: %w", err)
	}
	items, err := availabilityzones.ExtractAvailabilityZones(pages)
	if err != nil {
		return nil, fmt.Errorf("extract OpenStack availability zones: %w", err)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.ZoneState.Available {
			out = append(out, item.ZoneName)
		}
	}
	sort.Strings(out)
	return out, nil
}

func listSyncNetworks(client *gophercloud.ServiceClient) ([]clusteropenstacksync.Network, string, error) {
	pages, err := networks.List(client, networks.ListOpts{}).AllPages()
	if err != nil {
		return nil, "", fmt.Errorf("list OpenStack networks: %w", err)
	}
	items, err := networks.ExtractNetworks(pages)
	if err != nil {
		return nil, "", fmt.Errorf("extract OpenStack networks: %w", err)
	}
	external := true
	externalPages, err := networks.List(client, networkexternal.ListOptsExt{ListOptsBuilder: networks.ListOpts{}, External: &external}).AllPages()
	if err != nil {
		return nil, "", fmt.Errorf("list OpenStack external networks: %w", err)
	}
	externalItems, err := networks.ExtractNetworks(externalPages)
	if err != nil {
		return nil, "", fmt.Errorf("extract OpenStack external networks: %w", err)
	}
	externalIDs := map[string]bool{}
	for _, item := range externalItems {
		externalIDs[item.ID] = true
	}
	sort.Slice(externalItems, func(i, j int) bool {
		if externalItems[i].Name == externalItems[j].Name {
			return externalItems[i].ID < externalItems[j].ID
		}
		return externalItems[i].Name < externalItems[j].Name
	})
	internal := make([]clusteropenstacksync.Network, 0, len(items))
	for _, item := range items {
		if !externalIDs[item.ID] {
			internal = append(internal, clusteropenstacksync.Network{ID: item.ID, Name: item.Name})
		}
	}
	sort.Slice(internal, func(i, j int) bool {
		if internal[i].Name == internal[j].Name {
			return internal[i].ID < internal[j].ID
		}
		return internal[i].Name < internal[j].Name
	})
	if len(externalItems) == 0 {
		return internal, "", nil
	}
	return internal, externalItems[0].ID, nil
}

func listSyncSubnets(client *gophercloud.ServiceClient, internal []clusteropenstacksync.Network) ([]clusteropenstacksync.Subnet, error) {
	pages, err := subnets.List(client, subnets.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("list OpenStack subnets: %w", err)
	}
	items, err := subnets.ExtractSubnets(pages)
	if err != nil {
		return nil, fmt.Errorf("extract OpenStack subnets: %w", err)
	}
	internalIDs := map[string]bool{}
	for _, network := range internal {
		internalIDs[network.ID] = true
	}
	out := make([]clusteropenstacksync.Subnet, 0, len(items))
	for _, item := range items {
		if internalIDs[item.NetworkID] {
			out = append(out, clusteropenstacksync.Subnet{ID: item.ID, Name: item.Name, NetworkID: item.NetworkID})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].ID < out[j].ID
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func listSyncVolumeTypes(provider *gophercloud.ProviderClient, endpoint gophercloud.EndpointOpts) ([]clusteropenstacksync.VolumeType, error) {
	client, err := gopenstack.NewBlockStorageV3(provider, endpoint)
	if err != nil {
		return nil, fmt.Errorf("create block storage client: %w", err)
	}
	pages, err := volumetypes.List(client, nil).AllPages()
	if err != nil {
		return nil, fmt.Errorf("list OpenStack volume types: %w", err)
	}
	items, err := volumetypes.ExtractVolumeTypes(pages)
	if err != nil {
		return nil, fmt.Errorf("extract OpenStack volume types: %w", err)
	}
	out := make([]clusteropenstacksync.VolumeType, 0, len(items))
	for _, item := range items {
		out = append(out, clusteropenstacksync.VolumeType{ID: item.ID, Name: item.Name, IsPublic: item.IsPublic})
	}
	return out, nil
}
