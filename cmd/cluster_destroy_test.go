package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
)

func TestClusterDestroyKindProvider(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	stateDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("FAKE_KIND_STATE_DIR", stateDir)
	installFakeKindBinary(t, binDir)
	prependTestPath(t, binDir)

	cfg, clusterPaths := saveKindConfigForCommandTest(t, dir, "destroy-kind", "opencenter")
	cfg.OpenCenter.Meta.Stage = config.StageBootstrap
	cfg.OpenCenter.Meta.Status = config.StatusSuccess
	if err := saveConfig(context.Background(), cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(stateDir, "clusters"), []byte("destroy-kind\n"), 0o644); err != nil {
		t.Fatalf("write cluster state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clusterPaths.GitOpsDir, "README.md"), []byte("gitops"), 0o644); err != nil {
		t.Fatalf("write gitops marker: %v", err)
	}

	cmd := newClusterDestroyCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"destroy-kind", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster destroy failed: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(clusterPaths.ConfigPath); !os.IsNotExist(err) {
		t.Fatalf("expected config file to be removed")
	}
	if _, err := os.Stat(clusterPaths.GitOpsDir); !os.IsNotExist(err) {
		t.Fatalf("expected gitops directory to be removed")
	}

	kindLog, err := os.ReadFile(filepath.Join(stateDir, "kind.log"))
	if err != nil {
		t.Fatalf("read fake kind log: %v", err)
	}
	if !strings.Contains(string(kindLog), "kind delete cluster --name destroy-kind") {
		t.Fatalf("expected delete cluster invocation\nlog:\n%s", string(kindLog))
	}
}

func TestClusterDestroyKindProviderAlreadyMissing(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	stateDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("FAKE_KIND_STATE_DIR", stateDir)
	installFakeKindBinary(t, binDir)
	prependTestPath(t, binDir)

	_, clusterPaths := saveKindConfigForCommandTest(t, dir, "destroy-missing-kind", "opencenter")
	if err := os.WriteFile(filepath.Join(clusterPaths.GitOpsDir, "README.md"), []byte("gitops"), 0o644); err != nil {
		t.Fatalf("write gitops marker: %v", err)
	}

	cmd := newClusterDestroyCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"destroy-missing-kind", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster destroy failed: %v", err)
	}

	kindLog, err := os.ReadFile(filepath.Join(stateDir, "kind.log"))
	if err != nil {
		t.Fatalf("read fake kind log: %v", err)
	}
	logText := string(kindLog)
	if !strings.Contains(logText, "kind get clusters") {
		t.Fatalf("expected destroy to check existing clusters\nlog:\n%s", logText)
	}
	if strings.Contains(logText, "kind delete cluster") {
		t.Fatalf("expected destroy to skip delete when cluster is absent\nlog:\n%s", logText)
	}
}

func TestClusterDestroyKindProviderDeleteFailure(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	stateDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("FAKE_KIND_STATE_DIR", stateDir)
	installFakeKindBinary(t, binDir)
	prependTestPath(t, binDir)

	_, clusterPaths := saveKindConfigForCommandTest(t, dir, "destroy-kind-fail", "opencenter")
	if err := os.WriteFile(filepath.Join(stateDir, "clusters"), []byte("destroy-kind-fail\n"), 0o644); err != nil {
		t.Fatalf("write cluster state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "delete_fail"), []byte("1\n"), 0o644); err != nil {
		t.Fatalf("write delete fail flag: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clusterPaths.GitOpsDir, "README.md"), []byte("gitops"), 0o644); err != nil {
		t.Fatalf("write gitops marker: %v", err)
	}

	cmd := newClusterDestroyCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"destroy-kind-fail", "--force"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected cluster destroy to fail when kind delete fails")
	}
	if !strings.Contains(err.Error(), "failed to destroy kind cluster") {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(clusterPaths.ConfigPath); statErr != nil {
		t.Fatalf("expected config file to remain after delete failure: %v", statErr)
	}
	if _, statErr := os.Stat(clusterPaths.GitOpsDir); statErr != nil {
		t.Fatalf("expected gitops directory to remain after delete failure: %v", statErr)
	}
}
