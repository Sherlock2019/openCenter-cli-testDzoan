package talos

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
)

type fakeClient struct {
	appliedConfigs map[string][]byte
	bootstrapped   []string
	healthNodes    [][]string
}

func (f *fakeClient) ApplyMachineConfig(ctx context.Context, node string, config []byte) error {
	if f.appliedConfigs == nil {
		f.appliedConfigs = map[string][]byte{}
	}
	f.appliedConfigs[node] = append([]byte(nil), config...)
	return nil
}

func (f *fakeClient) Bootstrap(ctx context.Context, node string) error {
	f.bootstrapped = append(f.bootstrapped, node)
	return nil
}

func (f *fakeClient) Kubeconfig(ctx context.Context, node string) ([]byte, error) {
	return []byte("apiVersion: v1\nclusters: []\n"), nil
}

func (f *fakeClient) Health(ctx context.Context, nodes []string) error {
	f.healthNodes = append(f.healthNodes, append([]string(nil), nodes...))
	return nil
}

func TestRuntimeDeployFlowDoesNotPersistMachineConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	clusterPaths := &paths.ClusterPaths{
		ClusterDir:      filepath.Join(tmpDir, "infrastructure", "clusters", "demo"),
		SecretsDir:      filepath.Join(tmpDir, "secrets"),
		KubeconfigPath:  filepath.Join(tmpDir, "infrastructure", "clusters", "demo", "kubeconfig.yaml"),
		SOPSConfigPath:  filepath.Join(tmpDir, ".sops.yaml"),
		OrganizationDir: tmpDir,
	}
	if err := os.MkdirAll(filepath.Join(clusterPaths.ClusterDir, "talos"), 0o755); err != nil {
		t.Fatalf("mkdir talos dir: %v", err)
	}
	writeRuntimeInventory(t, filepath.Join(clusterPaths.ClusterDir, "talos", "inventory.yaml"))

	cfgPtr, err := v2.NewV2Default("demo", "openstack")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}
	cfg := *cfgPtr
	v2.ApplyTalosDeploymentDefaults(&cfg)
	cfg.Deployment.Talos.Patches.Static = nil
	cfg.Deployment.Talos.Endpoint = "https://10.2.128.5:443"

	fake := &fakeClient{}
	var factoryEndpoints []string
	runtime, err := NewRuntime(&cfg, clusterPaths, WithClientFactory(func(_ context.Context, _ *clientconfig.Config, endpoints []string) (Client, error) {
		factoryEndpoints = append([]string(nil), endpoints...)
		return fake, nil
	}))
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}

	ctx := context.Background()
	for name, run := range map[string]func(context.Context) error{
		"read-inventory":         runtime.ReadInventory,
		"generate-secrets":       runtime.GenerateSecrets,
		"apply-machine-configs":  runtime.ApplyMachineConfigs,
		"bootstrap-controlplane": runtime.BootstrapControlPlane,
		"export-talosconfig":     runtime.ExportTalosConfig,
		"export-kubeconfig":      runtime.ExportKubeconfig,
		"wait-ready":             runtime.WaitReady,
	} {
		if err := run(ctx); err != nil {
			t.Fatalf("%s error = %v", name, err)
		}
	}

	if len(fake.appliedConfigs) != 2 {
		t.Fatalf("applied config count = %d, want 2", len(fake.appliedConfigs))
	}
	if _, ok := fake.appliedConfigs["10.2.128.11:50000"]; !ok {
		t.Fatalf("applied config missing control-plane management endpoint: %v", mapKeys(fake.appliedConfigs))
	}
	if _, ok := fake.appliedConfigs["10.2.128.21:50000"]; !ok {
		t.Fatalf("applied config missing worker management endpoint: %v", mapKeys(fake.appliedConfigs))
	}
	if got, want := strings.Join(factoryEndpoints, ","), "10.2.128.11:50000"; got != want {
		t.Fatalf("client factory endpoints = %q, want %q", got, want)
	}
	if len(fake.bootstrapped) != 1 || fake.bootstrapped[0] != "10.2.128.11:50000" {
		t.Fatalf("bootstrapped nodes = %v, want [10.2.128.11:50000]", fake.bootstrapped)
	}
	if len(fake.healthNodes) != 1 || strings.Join(fake.healthNodes[0], ",") != "10.2.128.11:50000,10.2.128.21:50000" {
		t.Fatalf("health nodes = %v", fake.healthNodes)
	}

	artifactPaths := ResolveArtifactPaths(clusterPaths, "demo")
	for _, path := range []string{artifactPaths.MachineSecretsPath, artifactPaths.TalosConfigPath, artifactPaths.KubeconfigPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}

	err = filepath.WalkDir(clusterPaths.ClusterDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.Contains(base, "machine-config") {
			t.Fatalf("machine config was persisted at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk cluster dir: %v", err)
	}
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func writeRuntimeInventory(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(`cluster:
  name: demo
  endpoint: https://10.2.128.5:443
  talos_api_port: 50000
control_plane:
  - name: demo-cp-1
    talos_api_ip: 10.2.128.11
    internal_ip: 10.2.128.11
    install_disk: /dev/vda
workers:
  - name: demo-wn-1
    talos_api_ip: 10.2.128.21
    internal_ip: 10.2.128.21
    install_disk: /dev/vda
patch_inputs:
  pod_subnet: 10.42.0.0/16
  service_subnet: 10.43.0.0/16
`), 0o600); err != nil {
		t.Fatalf("write inventory: %v", err)
	}
}
