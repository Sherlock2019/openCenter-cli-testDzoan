package config

import (
	"os"
	"path/filepath"
	"testing"
)

func createSecureConfigTestCluster(t testing.TB, baseDir, organization, clusterName string) string {
	t.Helper()
	blueprintsDir := filepath.Join(baseDir, "blueprints", organization, clusterName)
	if err := os.MkdirAll(blueprintsDir, 0o755); err != nil {
		t.Fatalf("create blueprints dir: %v", err)
	}
	return filepath.Join(blueprintsDir, clusterName+"-config.yaml")
}
