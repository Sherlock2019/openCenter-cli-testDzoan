package services

import (
	"github.com/opencenter-cloud/opencenter-cli/internal/config/registry"
)

// CiliumConfig extends BaseConfig with Cilium-specific configuration
type CiliumConfig struct {
	BaseConfig `yaml:",inline"`

	OperatorEnabled      bool   `yaml:"operator_enabled" json:"operator_enabled,omitempty" jsonschema:"description=Enable Cilium operator for advanced features"`
	KubeProxyReplacement bool   `yaml:"kubeProxyReplacement" json:"kubeProxyReplacement,omitempty" jsonschema:"description=Replace kube-proxy with Cilium's eBPF implementation"`
	ModuleSource         string `yaml:"module_source" json:"module_source,omitempty" jsonschema:"description=Cilium module source location"`
}

func init() {
	registry.RegisterServiceConfig("cilium", CiliumConfig{})
}
