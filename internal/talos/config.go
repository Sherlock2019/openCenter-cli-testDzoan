package talos

// TalosConfig represents Talos-specific configuration.
// This extends the existing config.Config structure with Talos provider settings.
type TalosConfig struct {
	Enabled        bool                `yaml:"enabled" json:"enabled"`
	Version        string              `yaml:"version" json:"version"`
	ImageURL       string              `yaml:"image_url" json:"image_url"`
	ImageSignature string              `yaml:"image_signature" json:"image_signature"`
	MachineConfig  TalosMachineConfig  `yaml:"machine_config" json:"machine_config"`
	NetworkConfig  TalosNetworkConfig  `yaml:"network_config" json:"network_config"`
	SecurityConfig TalosSecurityConfig `yaml:"security_config" json:"security_config"`
	PulumiConfig   TalosPulumiConfig   `yaml:"pulumi_config" json:"pulumi_config"`
}

// TalosMachineConfig holds Talos machine configuration settings.
type TalosMachineConfig struct {
	AppArmorEnabled  bool     `yaml:"apparmor_enabled" json:"apparmor_enabled"`
	SeccompEnabled   bool     `yaml:"seccomp_enabled" json:"seccomp_enabled"`
	DiskEncryption   bool     `yaml:"disk_encryption" json:"disk_encryption"`
	KubePrismEnabled bool     `yaml:"kubeprism_enabled" json:"kubeprism_enabled"`
	SystemExtensions []string `yaml:"system_extensions" json:"system_extensions"`
	LogDestination   string   `yaml:"log_destination" json:"log_destination"`
}

// TalosNetworkConfig holds network topology settings.
type TalosNetworkConfig struct {
	ManagementSubnet string   `yaml:"management_subnet" json:"management_subnet"`
	ControlSubnet    string   `yaml:"control_subnet" json:"control_subnet"`
	DataSubnet       string   `yaml:"data_subnet" json:"data_subnet"`
	WireGuardPort    int      `yaml:"wireguard_port" json:"wireguard_port"`
	TalosAPIPort     int      `yaml:"talos_api_port" json:"talos_api_port"`
	AllowedCIDRs     []string `yaml:"allowed_cidrs" json:"allowed_cidrs"`
}

// TalosSecurityConfig holds security-related settings.
type TalosSecurityConfig struct {
	VTPMEnabled       bool   `yaml:"vtpm_enabled" json:"vtpm_enabled"`
	BarbicanKeyID     string `yaml:"barbican_key_id" json:"barbican_key_id"`
	ImageVerification bool   `yaml:"image_verification" json:"image_verification"`
	MFARequired       bool   `yaml:"mfa_required" json:"mfa_required"`
	AuditLogEnabled   bool   `yaml:"audit_log_enabled" json:"audit_log_enabled"`
}

// DefaultTalosConfig returns a TalosConfig with secure defaults.
func DefaultTalosConfig() *TalosConfig {
	return &TalosConfig{
		Enabled:        false,
		Version:        "v1.7.0",
		ImageURL:       "",
		ImageSignature: "",
		MachineConfig: TalosMachineConfig{
			AppArmorEnabled:  true,
			SeccompEnabled:   true,
			DiskEncryption:   true,
			KubePrismEnabled: true,
			SystemExtensions: []string{},
			LogDestination:   "",
		},
		NetworkConfig: TalosNetworkConfig{
			ManagementSubnet: "10.0.1.0/24",
			ControlSubnet:    "10.0.2.0/24",
			DataSubnet:       "10.0.3.0/24",
			WireGuardPort:    51820,
			TalosAPIPort:     50000,
			AllowedCIDRs:     []string{},
		},
		SecurityConfig: TalosSecurityConfig{
			VTPMEnabled:       true,
			BarbicanKeyID:     "",
			ImageVerification: true,
			MFARequired:       true,
			AuditLogEnabled:   true,
		},
		PulumiConfig: TalosPulumiConfig{
			StackName:         "",
			SwiftContainer:    "",
			SwiftPrefix:       "",
			SecretsPassphrase: "",
		},
	}
}
