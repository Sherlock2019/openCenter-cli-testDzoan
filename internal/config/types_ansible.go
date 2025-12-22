package config

// Ansible holds Ansible-specific settings.
type Ansible struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	Path      string   `yaml:"path" json:"path"`
	Inventory string   `yaml:"inventory,omitempty" json:"inventory,omitempty"`
	Playbooks []string `yaml:"playbooks,omitempty" json:"playbooks,omitempty"`
}
