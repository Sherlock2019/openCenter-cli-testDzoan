package config

// Security represents security-related configuration
type Security struct {
	CACertificates        string   `yaml:"ca_certificates" json:"ca_certificates"`
	K8sHardening          bool     `yaml:"k8s_hardening" json:"k8s_hardening"`
	OSHardening           bool     `yaml:"os_hardening" json:"os_hardening"`
	PodSecurityExemptions []string `yaml:"pod_security_exemptions" json:"pod_security_exemptions"`
}

// ClusterSecurityConfig represents cluster-level security configuration
type ClusterSecurityConfig struct {
	CACertificates string `yaml:"ca_certificates" json:"ca_certificates"`
	OSHardening    bool   `yaml:"os_hardening" json:"os_hardening"`
}

// KubernetesSecurityConfig represents Kubernetes-level security configuration
type KubernetesSecurityConfig struct {
	K8sHardening          bool     `yaml:"k8s_hardening" json:"k8s_hardening"`
	PodSecurityExemptions []string `yaml:"pod_security_exemptions" json:"pod_security_exemptions"`
	PodSecurityStandards  string   `yaml:"pod_security_standards,omitempty" json:"pod_security_standards,omitempty" jsonschema:"description=Pod security standards enforcement level,enum=privileged,enum=baseline,enum=restricted"`
	AuditLogging          bool     `yaml:"audit_logging" json:"audit_logging" jsonschema:"description=Enable Kubernetes API audit logging"`
	EncryptionAtRest      bool     `yaml:"encryption_at_rest" json:"encryption_at_rest" jsonschema:"description=Enable encryption of secrets at rest"`
}
