package talos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
)

func TestLoadMachineSecretsCorruptFileIncludesRecoveryGuidance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "machine-secrets.yaml")
	if err := os.WriteFile(path, []byte("not: [valid"), 0o600); err != nil {
		t.Fatalf("write corrupt secrets: %v", err)
	}

	_, err := LoadMachineSecrets(path)
	if err == nil {
		t.Fatal("LoadMachineSecrets() expected error")
	}
	msg := err.Error()
	for _, want := range []string{path, "machine secrets", "delete"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}

func TestTalosConfigRoundTripAndCorruptFileGuidance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "talosconfig.yaml")
	cfg := &clientconfig.Config{
		Context: "demo",
		Contexts: map[string]*clientconfig.Context{
			"demo": {
				Endpoints: []string{"10.2.128.11"},
				CA:        "ca",
				Crt:       "crt",
				Key:       "key",
			},
		},
	}

	if err := WriteTalosConfig(path, cfg); err != nil {
		t.Fatalf("WriteTalosConfig() error = %v", err)
	}
	loaded, err := LoadTalosConfig(path)
	if err != nil {
		t.Fatalf("LoadTalosConfig() error = %v", err)
	}
	if loaded.Context != "demo" {
		t.Fatalf("loaded context = %q, want demo", loaded.Context)
	}

	if err := os.WriteFile(path, []byte("not: [valid"), 0o600); err != nil {
		t.Fatalf("write corrupt talosconfig: %v", err)
	}
	_, err = LoadTalosConfig(path)
	if err == nil {
		t.Fatal("LoadTalosConfig() expected error")
	}
	msg := err.Error()
	for _, want := range []string{path, "talosconfig", "regenerate"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}
