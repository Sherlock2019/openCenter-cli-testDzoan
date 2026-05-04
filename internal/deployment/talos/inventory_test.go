package talos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadInventoryValidatesTalosInventoryContract(t *testing.T) {
	path := writeTestInventory(t, `cluster:
  name: demo
  endpoint: https://10.2.128.5:443
  talos_api_port: 50000
control_plane:
  - name: demo-cp-1
    talos_api_ip: 10.2.128.11
    internal_ip: 10.2.128.11
    install_disk: /dev/vda
    cert_sans:
      - 10.2.128.5
workers:
  - name: demo-wn-1
    talos_api_ip: 10.2.128.21
    internal_ip: 10.2.128.21
    install_disk: /dev/vda
patch_inputs:
  pod_subnet: 10.42.0.0/16
  service_subnet: 10.43.0.0/16
`)

	inventory, err := LoadInventory(path)
	if err != nil {
		t.Fatalf("LoadInventory() error = %v", err)
	}

	if inventory.Cluster.Name != "demo" {
		t.Fatalf("cluster name = %q, want demo", inventory.Cluster.Name)
	}
	if got := inventory.AllNodes(); len(got) != 2 {
		t.Fatalf("AllNodes() length = %d, want 2", len(got))
	}
	if got := inventory.ControlPlane[0].Role; got != RoleControlPlane {
		t.Fatalf("control plane role = %q, want %q", got, RoleControlPlane)
	}
	if got := inventory.Workers[0].Role; got != RoleWorker {
		t.Fatalf("worker role = %q, want %q", got, RoleWorker)
	}
	if got, want := inventory.ControlPlaneEndpoints(), []string{"10.2.128.11:50000"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ControlPlaneEndpoints() = %v, want %v", got, want)
	}
	if got, want := inventory.AllNodeEndpoints(), []string{"10.2.128.11:50000", "10.2.128.21:50000"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("AllNodeEndpoints() = %v, want %v", got, want)
	}
}

func TestLoadInventoryErrorsIncludePathAndField(t *testing.T) {
	path := writeTestInventory(t, `cluster:
  name: demo
  endpoint: https://10.2.128.5:443
  talos_api_port: 50000
control_plane:
  - name: demo-cp-1
    internal_ip: 10.2.128.11
    install_disk: /dev/vda
`)

	_, err := LoadInventory(path)
	if err == nil {
		t.Fatal("LoadInventory() expected error")
	}
	msg := err.Error()
	for _, want := range []string{path, "control_plane[0].talos_api_ip"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}

func TestLoadInventoryRejectsInvalidClusterEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "http",
			endpoint: "http://10.2.128.5:443",
			want:     "must use https scheme",
		},
		{
			name:     "kubernetes default port",
			endpoint: "https://10.2.128.5:6443",
			want:     "must use port 443",
		},
		{
			name:     "missing explicit port",
			endpoint: "https://10.2.128.5",
			want:     "must include explicit port 443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTestInventory(t, `cluster:
  name: demo
  endpoint: `+tt.endpoint+`
  talos_api_port: 50000
control_plane:
  - name: demo-cp-1
    talos_api_ip: 10.2.128.11
    internal_ip: 10.2.128.11
    install_disk: /dev/vda
`)

			_, err := LoadInventory(path)
			if err == nil {
				t.Fatal("LoadInventory() expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestTalosEndpointAddsPortWhenMissing(t *testing.T) {
	if got, want := talosEndpoint("198.51.100.11", 50000), "198.51.100.11:50000"; got != want {
		t.Fatalf("talosEndpoint() = %q, want %q", got, want)
	}
	if got, want := talosEndpoint("198.51.100.11:50001", 50000), "198.51.100.11:50001"; got != want {
		t.Fatalf("talosEndpoint() = %q, want %q", got, want)
	}
}

func writeTestInventory(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "inventory.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write inventory: %v", err)
	}
	return path
}
