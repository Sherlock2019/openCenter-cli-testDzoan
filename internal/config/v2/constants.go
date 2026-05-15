package v2

// Cluster lifecycle stages.
const (
	StageInit      = "init"
	StagePreflight = "preflight"
	StageSetup     = "setup"
	StageBootstrap = "bootstrap"
	StageValidate  = "validate"
	StageDestroy   = "destroy"
	StageRender    = "render"
	StagePlan      = "plan"
	StageApply     = "apply"
)

// Cluster lifecycle statuses.
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// Validation modes.
const (
	ValidationModeOffline = "offline"
	ValidationModeOnline  = "online"
)

// GitOps authentication methods.
const (
	GitopsAuthMethodSSH   = "ssh"
	GitopsAuthMethodToken = "token"
)

// DefaultSSHAuthorizedKeyPlaceholder is the placeholder SSH key used in templates.
const DefaultSSHAuthorizedKeyPlaceholder = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExamplePublicKeyDataHere user@example.com"
