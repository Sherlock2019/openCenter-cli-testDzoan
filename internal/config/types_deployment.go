package config

// Deployment represents deployment behavior configuration
type Deployment struct {
	AutoDeploy bool `yaml:"auto_deploy" json:"auto_deploy"`
}
