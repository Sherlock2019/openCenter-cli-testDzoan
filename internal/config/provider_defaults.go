package config

import (
	"fmt"
	"strings"

	registrydefaults "github.com/opencenter-cloud/opencenter-cli/internal/config/defaults"
)

const defaultOpenStackRegion = "dfw3"

// NewProviderDefault returns a v2 configuration initialized with provider-aware defaults.
// It is the shared builder for commands that need a default cluster configuration as a
// starting point, such as `cluster init` and `cluster template`.
func NewProviderDefault(name, provider string) (Config, error) {
	cfg := defaultConfig(name)

	selectedProvider := strings.TrimSpace(provider)
	if selectedProvider == "" {
		selectedProvider = cfg.OpenCenter.Infrastructure.Provider
	}

	cfg.OpenCenter.Infrastructure.Provider = selectedProvider
	if err := ApplyProviderDefaults(&cfg, selectedProvider); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyOpenStackDefaults(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	cfg.OpenCenter.Infrastructure.Provider = "openstack"

	region := strings.TrimSpace(cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region)
	if region == "" {
		region = strings.TrimSpace(cfg.OpenCenter.Meta.Region)
	}
	if region == "" {
		region = defaultOpenStackRegion
	}

	cfg.OpenCenter.Meta.Region = region
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region = region

	hydrator := registrydefaults.NewHydrator(registrydefaults.GetGlobalRegistry())
	if err := hydrator.Hydrate(cfg, "openstack", strings.ToLower(region)); err != nil {
		return fmt.Errorf("apply openstack defaults for region %q: %w", region, err)
	}

	cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Cinder.Enabled = true
	cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Vsphere.Enabled = false

	return nil
}
