package config

import overlaycfg "github.com/opencenter-cloud/opencenter-cli/internal/config/overlay"

// GitOpsConfig holds configuration related to GitOps scaffolding and repositories.
type GitOpsConfig struct {
	GitDir    string     `yaml:"git_dir" json:"git_dir" validate:"required"`
	GitURL    string     `yaml:"git_url" json:"git_url" validate:"omitempty,url"`
	GitSSHKey string     `yaml:"git_ssh_key,omitempty" json:"git_ssh_key,omitempty"`
	GitSSHPub string     `yaml:"git_ssh_pub,omitempty" json:"git_ssh_pub,omitempty"`
	GitBranch string     `yaml:"git_branch,omitempty" json:"git_branch,omitempty"`
	Release   string     `yaml:"release,omitempty" json:"release,omitempty"`
	Branch    string     `yaml:"branch,omitempty" json:"branch,omitempty"`
	Uri       string     `yaml:"uri,omitempty" json:"uri,omitempty" validate:"omitempty,url"`
	Flux      GitOpsFlux `yaml:"flux,omitempty" json:"flux,omitempty"`

	// Token-based authentication for HTTPS Git access (e.g. Gitea, GitHub, GitLab).
	// When git_token is set, SSH key fields are ignored for bootstrap operations.
	GitToken         string `yaml:"git_token,omitempty" json:"git_token,omitempty" jsonschema:"description=Path to file containing Git access token for HTTPS auth"`
	GitTokenProvider string `yaml:"git_token_provider,omitempty" json:"git_token_provider,omitempty" jsonschema:"description=Token provider type (gitea, github, gitlab)"`

	// New fields for GitOps base repository configuration
	GitOpsBaseRepo    string                 `yaml:"gitops_base_repo,omitempty" json:"gitops_base_repo,omitempty" jsonschema:"description=URL of the GitOps base repository" validate:"required,url"`
	GitOpsBaseRelease string                 `yaml:"gitops_base_release,omitempty" json:"gitops_base_release,omitempty" jsonschema:"description=Release tag of the GitOps base repository" validate:"omitempty"`
	GitOpsBranch      string                 `yaml:"gitops_branch,omitempty" json:"gitops_branch,omitempty" jsonschema:"description=Branch of the GitOps base repository,default=main" validate:"required"`
	SecretName        string                 `yaml:"secret_name,omitempty" json:"secret_name,omitempty" jsonschema:"description=Name of the GitOps secret for repository access,default=opencenter-base"`
	OverlayUnits      overlaycfg.UnitsConfig `yaml:"overlay_units,omitempty" json:"overlay_units,omitempty"`
}

// GitOpsFlux holds optional FluxCD settings for reconciliation behavior.
type GitOpsFlux struct {
	Interval string `yaml:"interval" json:"interval"`
	Prune    bool   `yaml:"prune" json:"prune"`
}
