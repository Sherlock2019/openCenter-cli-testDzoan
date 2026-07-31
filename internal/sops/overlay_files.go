package sops

import v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"

// overlayFilesToEncrypt returns the ordered overlay files that should be encrypted.
func overlayFilesToEncrypt(cfg *v2.Config) []string {
	files := []string{
		"flux-system/gotk-sync.yaml",
		"managed-services/sources/base-repo.yaml",
	}

	switch cfg.OpenCenter.Infrastructure.Provider {
	case "openstack":
		files = append(files, "secrets/openstack-credentials.yaml")
	case "vsphere":
		files = append(files,
			"secrets/vsphere-credentials.yaml",
			"customer-managed/services/cloud-provider-vsphere/secret.yaml",
		)
	}

	return files
}
