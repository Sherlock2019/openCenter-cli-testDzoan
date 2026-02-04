package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
	svc "github.com/rackerlabs/opencenter-cli/internal/services"
)

// CertManagerPlugin implements the ServicePlugin interface for cert-manager
// using composition with BaseServicePlugin
type CertManagerPlugin struct {
	*svc.BaseServicePlugin
}

// NewCertManagerPlugin creates a new CertManagerPlugin
func NewCertManagerPlugin() svc.ServicePlugin {
	// Create base plugin with metadata
	base := svc.NewBasePlugin(svc.PluginMetadata{
		Name:        "cert-manager",
		Version:     "1.0.0",
		Description: "Certificate management for Kubernetes",
		Type:        svc.ServiceTypeSecurity,
		Author:      "opencenter",
		License:     "Apache-2.0",
	})

	plugin := &CertManagerPlugin{
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

// validate implements cert-manager specific validation
func (p *CertManagerPlugin) validate(config interface{}) error {
	cfg, ok := config.(*services.CertManagerConfig)
	if !ok {
		return fmt.Errorf("invalid config type for cert-manager: expected *CertManagerConfig")
	}

	// Basic validation
	if cfg.IsEnabled() {
		if cfg.LetsEncryptServer != "" && !strings.HasPrefix(cfg.LetsEncryptServer, "https://") {
			return fmt.Errorf("letsencrypt_server must be an HTTPS URL")
		}
		if cfg.Email != "" && !strings.Contains(cfg.Email, "@") {
			return fmt.Errorf("email must be a valid email address")
		}
	}

	return nil
}

// render implements cert-manager specific rendering
func (p *CertManagerPlugin) render(ctx context.Context, config interface{}, workspace interface{}) error {
	// Template rendering will be handled by the template system
	// This is a placeholder for future implementation
	return nil
}

// status implements cert-manager specific status logic
func (p *CertManagerPlugin) status(config interface{}) svc.ServiceStatus {
	cfg, ok := config.(*services.CertManagerConfig)
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
		Message: "Cert-manager service",
		Details: map[string]interface{}{
			"letsencrypt_server": cfg.LetsEncryptServer,
			"email":              cfg.Email,
		},
	}
}
