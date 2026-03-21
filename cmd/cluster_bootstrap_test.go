package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, dir, name, provider, gitDir string) {
	t.Helper()
	if gitDir != "" {
		if err := os.MkdirAll(gitDir, 0o755); err != nil {
			t.Fatalf("mkdir git dir: %v", err)
		}
	}
	orgDir := filepath.Join(dir, "clusters", "opencenter")
	clusterDir := filepath.Join(orgDir, "infrastructure", "clusters", name)
	if err := os.MkdirAll(clusterDir, 0o755); err != nil {
		t.Fatalf("create cluster directory: %v", err)
	}

	content := fmt.Sprintf(`schema_version: "2.0"
opencenter:
  meta:
    organization: opencenter
    name: %s
    stage: init
    status: success
  infrastructure:
    provider: %s
%s  cluster:
    cluster_name: %s
  gitops:
    git_dir: %q
`, name, provider, kindConfigBlock(provider), name, gitDir)

	path := filepath.Join(orgDir, "."+name+"-config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func kindConfigBlock(provider string) string {
	if provider != "kind" {
		return ""
	}

	return `    kind:
      cluster_name: demo
      kubernetes_version: "1.30.4"
      control_plane_count: 1
      worker_count: 2
      api_server_address: "127.0.0.1"
      api_server_port: 6443
      pod_subnet: "10.244.0.0/16"
      service_subnet: "10.96.0.0/16"
      kubeconfig_path_policy: "cluster-owned"
`
}

func TestClusterBootstrapDryRunMake(t *testing.T) {
	t.Setenv("CONTAINER_RUNTIME", "")

	cfgDir := t.TempDir()
	gitDir := filepath.Join(t.TempDir(), "repo")
	clusterDir := filepath.Join(gitDir, "infrastructure", "clusters", "demo")
	if err := os.MkdirAll(clusterDir, 0o755); err != nil {
		t.Fatalf("mkdir cluster dir: %v", err)
	}

	writeTestConfig(t, cfgDir, "demo", "openstack", gitDir)
	prepareCommandTestEnv(t, cfgDir)

	cmd := newClusterBootstrapCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"demo", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bootstrap command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Bootstrap complete in") {
		t.Fatalf("expected completion message in output, got: %s", output)
	}
}

func TestClusterBootstrapDryRunKind(t *testing.T) {
	cfgDir := t.TempDir()
	gitDir := filepath.Join(t.TempDir(), "repo")
	writeTestConfig(t, cfgDir, "demo", "kind", gitDir)
	prepareCommandTestEnv(t, cfgDir)
	t.Setenv("CONTAINER_RUNTIME", "docker")

	cmd := newClusterBootstrapCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"demo", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bootstrap command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Bootstrap complete in") {
		t.Fatalf("expected completion message in output, got: %s", output)
	}
}
