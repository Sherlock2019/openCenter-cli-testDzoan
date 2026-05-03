package talos

import (
	"fmt"
	"os"
	"path/filepath"

	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"github.com/siderolabs/talos/pkg/machinery/config/generate/secrets"
	"gopkg.in/yaml.v3"
)

func WriteMachineSecrets(path string, bundle *secrets.Bundle) error {
	if bundle == nil {
		return fmt.Errorf("writing machine secrets %s: bundle is nil", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating machine secrets directory %s: %w", filepath.Dir(path), err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening machine secrets %s: %w", path, err)
	}
	defer file.Close() //nolint:errcheck

	if err := yaml.NewEncoder(file).Encode(bundle); err != nil {
		return fmt.Errorf("writing machine secrets %s: %w", path, err)
	}
	return nil
}

func LoadMachineSecrets(path string) (*secrets.Bundle, error) {
	bundle, err := secrets.LoadBundle(path)
	if err != nil {
		return nil, fmt.Errorf("loading machine secrets %s: %w; delete the file and rerun deploy to regenerate if it is corrupted", path, err)
	}
	return bundle, nil
}

func WriteTalosConfig(path string, cfg *clientconfig.Config) error {
	if cfg == nil {
		return fmt.Errorf("writing talosconfig %s: config is nil", path)
	}
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("writing talosconfig %s: %w", path, err)
	}
	return nil
}

func LoadTalosConfig(path string) (*clientconfig.Config, error) {
	cfg, err := clientconfig.Open(path)
	if err != nil {
		return nil, fmt.Errorf("loading talosconfig %s: %w; regenerate it by rerunning deploy", path, err)
	}
	return cfg, nil
}
