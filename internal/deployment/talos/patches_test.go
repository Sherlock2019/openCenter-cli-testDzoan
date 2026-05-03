package talos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadNamedPatchesRendersTemplatesAndReportsContext(t *testing.T) {
	dir := t.TempDir()
	writePatchFile(t, dir, "static.yaml", `- op: add
  path: /machine/network/hostname
  value: cp-1
`)
	writePatchFile(t, dir, "network.yaml.tmpl", `- op: add
  path: /cluster/network/podSubnets/0
  value: "{{ .PatchInputs.PodSubnet }}"
`)

	inventory := &Inventory{
		PatchInputs: PatchInputs{PodSubnet: "10.42.0.0/16"},
	}
	node := Node{Name: "cp-1", Role: RoleControlPlane}

	patches, err := LoadNamedPatches(dir, []string{"static", "network"}, inventory, node)
	if err != nil {
		t.Fatalf("LoadNamedPatches() error = %v", err)
	}
	if len(patches) != 2 {
		t.Fatalf("patch count = %d, want 2", len(patches))
	}
	if !strings.Contains(patches[1].Contents, "10.42.0.0/16") {
		t.Fatalf("template patch was not rendered: %q", patches[1].Contents)
	}
}

func TestLoadNamedPatchesErrorIncludesPatchNodeAndRole(t *testing.T) {
	dir := t.TempDir()
	writePatchFile(t, dir, "bad.yaml", `not: [valid`)

	_, err := LoadNamedPatches(dir, []string{"bad"}, &Inventory{}, Node{Name: "cp-1", Role: RoleControlPlane})
	if err == nil {
		t.Fatal("LoadNamedPatches() expected error")
	}
	msg := err.Error()
	for _, want := range []string{"bad", "cp-1", string(RoleControlPlane)} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}

func writePatchFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write patch %s: %v", name, err)
	}
}
