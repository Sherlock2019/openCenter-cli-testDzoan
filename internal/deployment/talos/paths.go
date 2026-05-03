package talos

import (
	"path/filepath"

	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
)

type ArtifactPaths struct {
	TalosDir           string
	PatchesDir         string
	InventoryPath      string
	SecretsDir         string
	MachineSecretsPath string
	TalosConfigPath    string
	KubeconfigPath     string
	SOPSConfigPath     string
}

func ResolveArtifactPaths(clusterPaths *paths.ClusterPaths, clusterName string) ArtifactPaths {
	if clusterPaths == nil {
		return ArtifactPaths{}
	}
	talosDir := filepath.Join(clusterPaths.ClusterDir, "talos")
	secretsDir := filepath.Join(clusterPaths.SecretsDir, "talos", clusterName)
	return ArtifactPaths{
		TalosDir:           talosDir,
		PatchesDir:         filepath.Join(talosDir, "patches"),
		InventoryPath:      filepath.Join(talosDir, "inventory.yaml"),
		SecretsDir:         secretsDir,
		MachineSecretsPath: filepath.Join(secretsDir, "machine-secrets.yaml"),
		TalosConfigPath:    filepath.Join(secretsDir, "talosconfig.yaml"),
		KubeconfigPath:     clusterPaths.KubeconfigPath,
		SOPSConfigPath:     clusterPaths.SOPSConfigPath,
	}
}
