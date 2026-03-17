package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	svc "github.com/opencenter-cloud/opencenter-cli/internal/services"
)

// KubeOVNPlugin implements the ServicePlugin interface for Kube-OVN
// using composition with BaseServicePlugin
type KubeOVNPlugin struct {
	*svc.BaseServicePlugin
}

// NewKubeOVNPlugin creates a new KubeOVNPlugin
func NewKubeOVNPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "kube-ovn",
		Version:     "1.0.0",
		Description: "Kube-OVN overlay networking for Kubernetes",
		Type:        svc.ServiceTypeNetworking,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &KubeOVNPlugin{
		BaseServicePlugin: base,
	}

	// Inject service-specific validation logic
	base.SetValidator(plugin.validate)

	// Inject service-specific rendering logic
	base.SetRenderer(plugin.render)

	// Inject service-specific status logic
	base.SetStatusFunc(plugin.status)

	return plugin
}

// validate implements kube-ovn specific validation
func (p *KubeOVNPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.KubeOVNConfig)
	if !ok {
		return fmt.Errorf("invalid config type for kube-ovn: expected *KubeOVNConfig")
	}

	// Basic validation
	if cfg.IsEnabled() {
		// Validate version format if specified
		if cfg.Version != "" && !strings.Contains(cfg.Version, ".") {
			return fmt.Errorf("invalid kube-ovn version format: %s, expected semantic version", cfg.Version)
		}

		// Validate default subnet format if specified
		if cfg.DefaultSubnet != "" {
			// Basic CIDR validation - should contain a slash
			if !strings.Contains(cfg.DefaultSubnet, "/") {
				return fmt.Errorf("invalid default_subnet format: %s, expected CIDR notation (e.g., 10.16.0.0/16)", cfg.DefaultSubnet)
			}
		}
	}

	return nil
}

// render implements kube-ovn specific rendering
func (p *KubeOVNPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	// This is a placeholder for future implementation
	return nil
}

// status implements kube-ovn specific status logic
func (p *KubeOVNPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.KubeOVNConfig)
	if !ok {
		return svc.ServiceStatus{
			State:   "failed",
			Message: "Invalid configuration type",
		}
	}

	if !cfg.IsEnabled() {
		return svc.ServiceStatus{
			State:   "disabled",
			Message: "Service is disabled",
		}
	}

	// Get status from config, default to "pending" if not set
	state := cfg.GetStatus()
	if state == "" {
		state = "pending"
	}

	return svc.ServiceStatus{
		State:   state,
		Message: "Kube-OVN networking service",
		Details: map[string]interface{}{
			"cilium_integration": cfg.CiliumIntegration,
			"default_subnet":     cfg.DefaultSubnet,
			"version":            cfg.Version,
			"enable_lb":          cfg.EnableLB,
		},
	}
}
