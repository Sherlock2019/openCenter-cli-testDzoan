package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	"gopkg.in/yaml.v3"
)

func TestClusterInitKindDefaults(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	cmd := newClusterInitCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"kind-cluster", "--type", "kind", "--no-keygen"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster init failed: %v", err)
	}

	configPath := filepath.Join(dir, "clusters", "opencenter", ".kind-cluster-config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}

	if cfg.OpenCenter.Infrastructure.Provider != "kind" {
		t.Fatalf("expected provider kind, got %s", cfg.OpenCenter.Infrastructure.Provider)
	}
	if cfg.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected opencenter.infrastructure.kind to be populated")
	}
	if cfg.OpenTofu.Enabled {
		t.Fatal("expected opentofu to be disabled for kind")
	}
	if cfg.OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL != "" {
		t.Fatalf("expected openstack auth_url to be cleared for kind, got %q", cfg.OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL)
	}
	if cfg.OpenCenter.Meta.Stage != config.StageInit || cfg.OpenCenter.Meta.Status != config.StatusSuccess {
		t.Fatalf("unexpected lifecycle state: %s/%s", cfg.OpenCenter.Meta.Stage, cfg.OpenCenter.Meta.Status)
	}
}

func TestClusterTemplateKindDefaults(t *testing.T) {
	cmd := newClusterTemplateCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--provider", "kind"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster template failed: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(stdout.Bytes(), &cfg); err != nil {
		t.Fatalf("unmarshal template: %v", err)
	}

	if cfg.OpenCenter.Infrastructure.Provider != "kind" {
		t.Fatalf("expected provider kind, got %s", cfg.OpenCenter.Infrastructure.Provider)
	}
	if cfg.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected opencenter.infrastructure.kind in template output")
	}
	if cfg.OpenTofu.Enabled {
		t.Fatal("expected opentofu to be disabled in kind template output")
	}
}

func TestClusterStatusUsesExplicitClusterName(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	writeTestConfig(t, dir, "active-cluster", "openstack", filepath.Join(dir, "clusters", "opencenter"))
	writeTestConfig(t, dir, "requested-cluster", "openstack", filepath.Join(dir, "clusters", "opencenter"))
	t.Setenv("OPENCENTER_CLUSTER", "active-cluster")

	cmd := newClusterStatusCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"requested-cluster", "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster status failed: %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "requested-cluster" {
		t.Fatalf("expected requested-cluster, got %q", got)
	}
}

func TestClusterStatusShowsKindStatusDetails(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	stateDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("FAKE_KIND_STATE_DIR", stateDir)
	installFakeKindBinary(t, binDir)
	installFakeKubectlBinary(t, binDir)
	prependTestPath(t, binDir)

	cfg, clusterPaths := saveKindConfigForCommandTest(t, dir, "status-kind-cluster", "opencenter")
	cfg.OpenCenter.Meta.Stage = config.StageBootstrap
	cfg.OpenCenter.Meta.Status = config.StatusSuccess
	if err := saveConfig(context.Background(), cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(clusterPaths.ClusterDir, "kind-config.yaml"), []byte("kind: Cluster\n"), 0o644); err != nil {
		t.Fatalf("write kind-config.yaml: %v", err)
	}
	if err := os.WriteFile(clusterPaths.KubeconfigPath, []byte("apiVersion: v1\n"), 0o644); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "clusters"), []byte("status-kind-cluster\n"), 0o644); err != nil {
		t.Fatalf("write cluster state: %v", err)
	}

	cmd := newClusterStatusCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status-kind-cluster"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster status failed: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Kind Status:",
		"GitOps Setup:      ✓ Ready",
		"kind-config.yaml:  ✓ Present",
		"Kubeconfig:        ✓ Present",
		"Cluster Exists:    ✓ Present",
		"API Ready:         ✓ Ready",
		"API Endpoint:      https://127.0.0.1:6443",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q\noutput:\n%s", expected, output)
		}
	}
}
