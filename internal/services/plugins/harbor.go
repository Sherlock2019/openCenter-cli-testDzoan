package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
	svc "github.com/rackerlabs/opencenter-cli/internal/services"
)

// HarborPlugin implements the ServicePlugin interface for Harbor
// using composition with BaseServicePlugin
type HarborPlugin struct {
	*svc.BaseServicePlugin
}

// NewHarborPlugin creates a new HarborPlugin
func NewHarborPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "harbor",
		Version:     "1.0.0",
		Description: "Container registry and artifact management",
		Type:        svc.ServiceTypeStorage,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &HarborPlugin{
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

// validate implements harbor specific validation
func (p *HarborPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.HarborConfig)
	if !ok {
		return fmt.Errorf("invalid config type for harbor: expected *HarborConfig")
	}

	// Basic validation
	if cfg.IsEnabled() {
		// Validate external URL format
		if cfg.ExternalURL != "" && !strings.HasPrefix(cfg.ExternalURL, "https://") && !strings.HasPrefix(cfg.ExternalURL, "http://") {
			return fmt.Errorf("harbor external_url must be a valid HTTP(S) URL")
		}

		// Validate S3 storage configuration
		if cfg.StorageType == "s3" {
			if cfg.S3Bucket == "" {
				return fmt.Errorf("s3_bucket is required when storage_type is s3")
			}
			if cfg.S3Region == "" {
				return fmt.Errorf("s3_region is required when storage_type is s3")
			}
		}

		// Validate external database configuration
		if cfg.DatabaseType == "external" {
			if cfg.DatabaseHost == "" {
				return fmt.Errorf("database_host is required when database_type is external")
			}
			if cfg.DatabasePort <= 0 {
				return fmt.Errorf("database_port must be positive when database_type is external")
			}
			if cfg.DatabaseName == "" {
				return fmt.Errorf("database_name is required when database_type is external")
			}
			if cfg.DatabaseUser == "" {
				return fmt.Errorf("database_user is required when database_type is external")
			}
		}

		// Validate registry volume size
		if cfg.RegistryVolumeSize < 0 {
			return fmt.Errorf("registry_volume_size must be non-negative")
		}
	}

	return nil
}

// render implements harbor specific rendering
func (p *HarborPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	return nil
}

// status implements harbor specific status logic
func (p *HarborPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.HarborConfig)
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
		Message: "Harbor container registry",
		Details: map[string]interface{}{
			"external_url":   cfg.ExternalURL,
			"storage_type":   cfg.StorageType,
			"database_type":  cfg.DatabaseType,
			"volume_size_gb": cfg.RegistryVolumeSize,
		},
	}
}
