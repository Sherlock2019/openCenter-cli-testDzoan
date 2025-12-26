package services

import (
	"github.com/rackerlabs/openCenter-cli/internal/config/registry"
)

// WeaveGitOpsConfig extends BaseConfig with Weave GitOps configuration
type WeaveGitOpsConfig struct {
	BaseConfig `yaml:",inline"`
}

func init() {
	registry.RegisterServiceConfig("weave-gitops", WeaveGitOpsConfig{})
}
