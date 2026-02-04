package plugins

import (
	"context"
	"fmt"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
	svc "github.com/rackerlabs/opencenter-cli/internal/services"
)

// LokiPlugin implements the ServicePlugin interface for Loki
// using composition with BaseServicePlugin
type LokiPlugin struct {
	*svc.BaseServicePlugin
}

// NewLokiPlugin creates a new LokiPlugin
func NewLokiPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "loki",
		Version:     "1.0.0",
		Description: "Loki logging and log aggregation service",
		Type:        svc.ServiceTypeLogging,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &LokiPlugin{
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

// validate implements loki specific validation
func (p *LokiPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.LokiConfig)
	if !ok {
		return fmt.Errorf("invalid config type for loki: expected *LokiConfig")
	}

	// Basic validation
	if cfg.IsEnabled() && cfg.StorageType != "" {
		if cfg.StorageType != "s3" && cfg.StorageType != "swift" {
			return fmt.Errorf("storage_type must be 's3' or 'swift', got '%s'", cfg.StorageType)
		}

		// Swift-specific validation
		if cfg.StorageType == "swift" && cfg.SwiftAuthURL == "" {
			return fmt.Errorf("swift_auth_url is required when storage_type is 'swift'")
		}

		// S3-specific validation
		if cfg.StorageType == "s3" && cfg.S3Endpoint == "" {
			return fmt.Errorf("s3_endpoint is required when storage_type is 's3'")
		}
	}

	return nil
}

// render implements loki specific rendering
func (p *LokiPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	// This is a placeholder for future implementation
	return nil
}

// status implements loki specific status logic
func (p *LokiPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.LokiConfig)
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
		Message: "Loki logging service",
		Details: map[string]interface{}{
			"storage_type": cfg.StorageType,
			"bucket_name":  cfg.BucketName,
			"volume_size":  cfg.VolumeSize,
		},
	}
}
