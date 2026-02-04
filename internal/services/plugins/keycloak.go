package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
	svc "github.com/rackerlabs/opencenter-cli/internal/services"
)

// KeycloakPlugin implements the ServicePlugin interface for Keycloak
// using composition with BaseServicePlugin
type KeycloakPlugin struct {
	*svc.BaseServicePlugin
}

// NewKeycloakPlugin creates a new KeycloakPlugin
func NewKeycloakPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "keycloak",
		Version:     "1.0.0",
		Description: "Identity and access management service",
		Type:        svc.ServiceTypeSecurity,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &KeycloakPlugin{
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

// validate implements keycloak specific validation
func (p *KeycloakPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.KeycloakConfig)
	if !ok {
		return fmt.Errorf("invalid config type for keycloak: expected *KeycloakConfig")
	}

	// Basic validation
	if cfg.IsEnabled() {
		// Validate frontend URL
		if cfg.FrontendURL != "" && !strings.HasPrefix(cfg.FrontendURL, "https://") && !strings.HasPrefix(cfg.FrontendURL, "http://") {
			return fmt.Errorf("keycloak_frontend_url must be a valid HTTP(S) URL")
		}

		// Validate production mode requires HA
		if cfg.StartOptimized && cfg.Instances < 2 {
			return fmt.Errorf("production mode (start_optimized=true) requires at least 2 instances for high availability")
		}

		// Validate autoscaling configuration
		if cfg.MinReplicas > 0 && cfg.MaxReplicas > 0 && cfg.MinReplicas > cfg.MaxReplicas {
			return fmt.Errorf("min_replicas (%d) cannot exceed max_replicas (%d)", cfg.MinReplicas, cfg.MaxReplicas)
		}

		// Validate DB pool configuration
		if cfg.DBPoolMaxSize > 0 && cfg.DBPoolMinSize > cfg.DBPoolMaxSize {
			return fmt.Errorf("db_pool_min_size (%d) cannot exceed db_pool_max_size (%d)", cfg.DBPoolMinSize, cfg.DBPoolMaxSize)
		}
	}

	return nil
}

// render implements keycloak specific rendering
func (p *KeycloakPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	return nil
}

// status implements keycloak specific status logic
func (p *KeycloakPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.KeycloakConfig)
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
		Message: "Keycloak identity service",
		Details: map[string]interface{}{
			"realm":        cfg.Realm,
			"frontend_url": cfg.FrontendURL,
			"client_id":    cfg.ClientID,
		},
	}
}
