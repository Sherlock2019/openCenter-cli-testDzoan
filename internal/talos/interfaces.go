package talos

import (
	"context"
	"time"

	"github.com/rackerlabs/opencenter-cli/internal/config"
)

// Validator performs pre-flight validation checks on OpenStack environments.
type Validator interface {
	// ValidateEnvironment checks all OpenStack prerequisites
	ValidateEnvironment(ctx context.Context, config *config.Config) (*ValidationReport, error)

	// ValidateKeystone checks Keystone availability and MFA
	ValidateKeystone(ctx context.Context) error

	// ValidateBarbican tests secret creation/retrieval
	ValidateBarbican(ctx context.Context) error

	// ValidateOctavia checks load balancer service
	ValidateOctavia(ctx context.Context) error

	// ValidateQuotas verifies tenant resource quotas
	ValidateQuotas(ctx context.Context, required ResourceRequirements) error

	// ValidateGlance checks image signature verification
	ValidateGlance(ctx context.Context) error
}

// Generator creates declarative artifacts for cluster deployment.
type Generator interface {
	// GenerateClusterConfig creates complete cluster configuration
	GenerateClusterConfig(ctx context.Context, config *config.Config) (*ClusterArtifacts, error)

	// GenerateTalosMachineConfig creates Talos machine configurations
	GenerateTalosMachineConfig(ctx context.Context, nodeType NodeType) ([]byte, error)

	// GeneratePulumiStack creates Pulumi stack configuration
	GeneratePulumiStack(ctx context.Context, config *config.Config) ([]byte, error)

	// GenerateWireGuardConfig creates VPN configuration
	GenerateWireGuardConfig(ctx context.Context) (*WireGuardConfig, error)

	// GenerateNetworkTopology creates network definitions
	GenerateNetworkTopology(ctx context.Context, config *config.Config) (*NetworkTopology, error)

	// GenerateSecurityGroups creates security group rules
	GenerateSecurityGroups(ctx context.Context, config *config.Config) ([]SecurityGroup, error)

	// GenerateGitOpsStructure creates directory layout
	GenerateGitOpsStructure(ctx context.Context, basePath string) error
}

// PulumiManager handles Pulumi operations for infrastructure lifecycle management.
type PulumiManager interface {
	// Initialize sets up Pulumi stack and backend
	Initialize(ctx context.Context, config *TalosPulumiConfig) error

	// Preview shows planned infrastructure changes
	Preview(ctx context.Context) (*PulumiPreview, error)

	// Apply provisions or updates infrastructure
	Apply(ctx context.Context) (*PulumiResult, error)

	// Refresh detects configuration drift
	Refresh(ctx context.Context) (*DriftReport, error)

	// Destroy tears down all resources
	Destroy(ctx context.Context) error

	// GetOutputs retrieves stack outputs
	GetOutputs(ctx context.Context) (map[string]interface{}, error)
}

// ValidationReport contains validation results.
type ValidationReport struct {
	Passed       bool                `json:"passed"`
	Checks       []ValidationCheck   `json:"checks"`
	Remediations []RemediationAction `json:"remediations"`
	Timestamp    time.Time           `json:"timestamp"`
}

// ValidationCheck represents a single validation check.
type ValidationCheck struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// RemediationAction provides guidance for failed checks.
type RemediationAction struct {
	Check       string   `json:"check"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

// ResourceRequirements defines required OpenStack quotas.
type ResourceRequirements struct {
	Instances      int
	Cores          int
	RAM            int // in MB
	Networks       int
	Routers        int
	SecurityGroups int
	Volumes        int
	VolumeStorage  int // in GB
	Snapshots      int
	LoadBalancers  int
}

// ClusterArtifacts contains all generated artifacts.
type ClusterArtifacts struct {
	TalosMachineConfigs map[NodeType][]byte
	PulumiStack         []byte
	WireGuardConfig     *WireGuardConfig
	NetworkTopology     *NetworkTopology
	SecurityGroups      []SecurityGroup
	SOPSConfig          []byte
	GitOpsStructure     map[string][]byte
}

// NodeType represents different node roles.
type NodeType string

const (
	NodeTypeControlPlane NodeType = "control-plane"
	NodeTypeWorker       NodeType = "worker"
	NodeTypeBastion      NodeType = "bastion"
)

// PulumiPreview contains planned changes.
type PulumiPreview struct {
	Creates  []ResourceChange `json:"creates"`
	Updates  []ResourceChange `json:"updates"`
	Deletes  []ResourceChange `json:"deletes"`
	Replaces []ResourceChange `json:"replaces"`
}

// ResourceChange describes a planned infrastructure change.
type ResourceChange struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
	Reason     string                 `json:"reason"`
}

// PulumiResult contains the result of a Pulumi operation.
type PulumiResult struct {
	Success bool                   `json:"success"`
	Outputs map[string]interface{} `json:"outputs"`
	Summary string                 `json:"summary"`
}

// DriftReport contains drift detection results.
type DriftReport struct {
	HasDrift     bool                `json:"has_drift"`
	Drifted      []DriftedResource   `json:"drifted"`
	Remediations []RemediationAction `json:"remediations"`
}

// DriftedResource represents a resource with configuration drift.
type DriftedResource struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Expected map[string]interface{} `json:"expected"`
	Actual   map[string]interface{} `json:"actual"`
}

// TalosPulumiConfig holds Pulumi-specific settings.
type TalosPulumiConfig struct {
	StackName         string `yaml:"stack_name" json:"stack_name"`
	SwiftContainer    string `yaml:"swift_container" json:"swift_container"`
	SwiftPrefix       string `yaml:"swift_prefix" json:"swift_prefix"`
	SecretsPassphrase string `yaml:"secrets_passphrase" json:"secrets_passphrase"`
}
