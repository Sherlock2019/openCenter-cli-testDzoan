package gitops

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

func TestRenderInfrastructureClusterAtomicTalosOpenStack(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewWorkspaceManager(tempDir)
	ctx := context.Background()

	cfg := newDefault("talos-render")
	cfg.OpenCenter.Cluster.ClusterName = "talos-render"
	cfg.OpenCenter.Infrastructure.Provider = "openstack"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL = "https://auth.example.local/v3/"
	v2.ApplyTalosDeploymentDefaults(&cfg)

	workspace, err := manager.CreateWorkspace(ctx, cfg)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	defer manager.CleanupWorkspace(ctx, workspace)

	if err := RenderInfrastructureClusterAtomic(cfg, workspace); err != nil {
		t.Fatalf("RenderInfrastructureClusterAtomic() error = %v", err)
	}

	clusterRoot := filepath.Join("infrastructure", "clusters", cfg.ClusterName())
	mainTF := readWorkspaceFile(t, workspace, filepath.Join(clusterRoot, "main.tf"))
	for _, want := range []string{
		`module "openstack-nova"`,
		`resource "local_file" "talos_inventory"`,
		`filename = "${path.module}/talos/inventory.yaml"`,
		"talos_api_port: 50000",
		"control_plane:",
		"workers:",
	} {
		if !strings.Contains(mainTF, want) {
			t.Fatalf("Talos main.tf missing %q\n%s", want, mainTF)
		}
	}
	for _, forbidden := range []string{
		`module "kubespray-cluster"`,
		"kubespray_version",
		"inventory/inventory.yaml",
		"run_kubespray",
	} {
		if strings.Contains(mainTF, forbidden) {
			t.Fatalf("Talos main.tf should not contain %q\n%s", forbidden, mainTF)
		}
	}

	patches := []string{
		"disable-cni.yaml",
		"disable-kubeproxy.yaml",
		"disable-node-cidr-allocator.yaml",
		"ntp.yaml.tmpl",
		"network-subnets.yaml.tmpl",
	}
	for _, patch := range patches {
		path := filepath.Join(clusterRoot, "talos", "patches", patch)
		if !workspace.Exists(path) {
			t.Fatalf("expected Talos patch artifact %s", path)
		}
	}
}

func readWorkspaceFile(t *testing.T, workspace *GitOpsWorkspace, rel string) string {
	t.Helper()

	data, err := os.ReadFile(workspace.GetPath(rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}
