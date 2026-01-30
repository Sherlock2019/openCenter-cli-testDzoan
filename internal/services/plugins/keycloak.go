package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
	svc "github.com/rackerlabs/opencenter-cli/internal/services"
)

// KeycloakPlugin implements the ServicePlugin interface for Keycloak
type KeycloakPlugin struct{}

// NewKeycloakPlugin creates a new KeycloakPlugin
func NewKeycloakPlugin() svc.ServicePlugin {
	return &KeycloakPlugin{}
}

// Name returns the service name
func (p *KeycloakPlugin) Name() string {
	return "keycloak"
}

// Type returns the service type
func (p *KeycloakPlugin) Type() svc.ServiceType {
	return svc.ServiceTypeSecurity
}

// Validate validates the service configuration
func (p *KeycloakPlugin) Validate(config interface{}) error {
	cfg, ok := config.(*services.KeycloakConfig)
	if !ok {
		return fmt.Errorf("invalid config type for keycloak: expected *KeycloakConfig")
	}

	if !cfg.IsEnabled() {
		return nil
	}

	// Validate production mode settings
	if cfg.StartOptimized && cfg.Instances < 2 {
		return fmt.Errorf("production mode (start_optimized=true) requires at least 2 instances for high availability")
	}

	// Validate instance count
	if cfg.Instances < 1 {
		return fmt.Errorf("instances must be at least 1")
	}

	// Validate autoscaling configuration
	if cfg.MinReplicas > 0 && cfg.MaxReplicas > 0 {
		if cfg.MinReplicas > cfg.MaxReplicas {
			return fmt.Errorf("min_replicas (%d) cannot exceed max_replicas (%d)", cfg.MinReplicas, cfg.MaxReplicas)
		}
		if cfg.MinReplicas < 2 {
			return fmt.Errorf("min_replicas should be at least 2 for high availability")
		}
	}

	// Validate DB pool configuration
	if cfg.DBPoolMaxSize > 0 && cfg.DBPoolMinSize > cfg.DBPoolMaxSize {
		return fmt.Errorf("db_pool_min_size (%d) cannot exceed db_pool_max_size (%d)", cfg.DBPoolMinSize, cfg.DBPoolMaxSize)
	}

	// Validate frontend URL - require HTTPS in production
	if cfg.FrontendURL != "" {
		if !strings.HasPrefix(cfg.FrontendURL, "https://") && !strings.HasPrefix(cfg.FrontendURL, "http://") {
			return fmt.Errorf("keycloak_frontend_url must be a valid HTTP(S) URL")
		}
		if cfg.StartOptimized && !strings.HasPrefix(cfg.FrontendURL, "https://") {
			return fmt.Errorf("keycloak_frontend_url must use HTTPS when start_optimized is enabled")
		}
	}

	// Validate log level
	if cfg.LogLevel != "" {
		validLevels := map[string]bool{"INFO": true, "DEBUG": true, "WARN": true, "ERROR": true, "TRACE": true}
		if !validLevels[strings.ToUpper(cfg.LogLevel)] {
			return fmt.Errorf("invalid log_level: %s (must be INFO, DEBUG, WARN, ERROR, or TRACE)", cfg.LogLevel)
		}
	}

	// Validate log format
	if cfg.LogFormat != "" {
		validFormats := map[string]bool{"default": true, "json": true}
		if !validFormats[cfg.LogFormat] {
			return fmt.Errorf("invalid log_format: %s (must be 'default' or 'json')", cfg.LogFormat)
		}
	}

	// Validate cache stack
	if cfg.CacheStack != "" {
		validStacks := map[string]bool{"kubernetes": true, "ispn": true}
		if !validStacks[cfg.CacheStack] {
			return fmt.Errorf("invalid cache_stack: %s (must be 'kubernetes' or 'ispn')", cfg.CacheStack)
		}
	}

	return nil
}

// Render renders the service templates to the workspace
func (p *KeycloakPlugin) Render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	return nil
}

// Status returns the current status of the service
func (p *KeycloakPlugin) Status(config interface{}) svc.ServiceStatus {
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
