package services

import (
	"github.com/rackerlabs/openCenter-cli/internal/config/registry"
)

// VSphereCSIConfig extends BaseConfig with vSphere CSI configuration
type VSphereCSIConfig struct {
	BaseConfig `yaml:",inline"`
}

func init() {
	registry.RegisterServiceConfig("vsphere-csi", VSphereCSIConfig{})
}
