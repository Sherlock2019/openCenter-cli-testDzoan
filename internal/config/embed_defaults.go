package config

import _ "embed"

// Default structured IaC values (locals as main, and modules) parsed from Terraform
// and expressed as YAML. Used to seed iac.main and iac.modules during cluster init.
//
//go:embed defaults/openstack.yaml
var defaultIACYAML string
