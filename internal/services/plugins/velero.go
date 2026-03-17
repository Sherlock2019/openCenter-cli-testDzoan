package plugins

import (
	"context"
	"fmt"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	svc "github.com/opencenter-cloud/opencenter-cli/internal/services"
)

// VeleroPlugin implements the ServicePlugin interface for Velero
// using composition with BaseServicePlugin
type VeleroPlugin struct {
	*svc.BaseServicePlugin
}

// NewVeleroPlugin creates a new VeleroPlugin
func NewVeleroPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "velero",
		Version:     "1.0.0",
		Description: "Backup and disaster recovery for Kubernetes",
		Type:        svc.ServiceTypeStorage,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &VeleroPlugin{
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

// validate implements velero specific validation
func (p *VeleroPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.VeleroConfig)
	if !ok {
		return fmt.Errorf("invalid config type for velero: expected *VeleroConfig")
	}

	// Basic validation
	if cfg.IsEnabled() && cfg.BackupBucket == "" {
		return fmt.Errorf("velero_backup_bucket is required when velero is enabled")
	}

	return nil
}

// render implements velero specific rendering
func (p *VeleroPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	return nil
}

// status implements velero specific status logic
func (p *VeleroPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.VeleroConfig)
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
		Message: "Velero backup service",
		Details: map[string]interface{}{
			"backup_bucket": cfg.BackupBucket,
			"region":        cfg.Region,
		},
	}
}
