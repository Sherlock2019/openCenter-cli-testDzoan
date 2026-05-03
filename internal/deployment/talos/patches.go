package talos

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/siderolabs/talos/pkg/machinery/config/configpatcher"
)

type NamedPatch struct {
	Name     string
	Path     string
	Contents string
}

func LoadNamedPatches(patchesDir string, names []string, inventory *Inventory, node Node) ([]NamedPatch, error) {
	if len(names) == 0 {
		return nil, nil
	}

	patches := make([]NamedPatch, 0, len(names))
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}

		path, err := resolvePatchPath(patchesDir, name)
		if err != nil {
			return nil, contextualPatchError(name, node, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, contextualPatchError(name, node, fmt.Errorf("read %s: %w", path, err))
		}
		rendered := string(content)
		if strings.HasSuffix(path, ".tmpl") {
			rendered, err = renderPatchTemplate(name, rendered, inventory, node)
			if err != nil {
				return nil, contextualPatchError(name, node, err)
			}
		}
		if _, err := configpatcher.LoadPatch([]byte(rendered)); err != nil {
			return nil, contextualPatchError(name, node, err)
		}

		patches = append(patches, NamedPatch{Name: name, Path: path, Contents: rendered})
	}
	return patches, nil
}

func ApplyNamedPatches(configBytes []byte, patches []NamedPatch, node Node) ([]byte, error) {
	if len(patches) == 0 {
		return configBytes, nil
	}

	values := make([]string, 0, len(patches))
	for _, patch := range patches {
		values = append(values, patch.Contents)
	}
	loaded, err := configpatcher.LoadPatches(values)
	if err != nil {
		return nil, contextualPatchError(joinPatchNames(patches), node, err)
	}

	patched, err := configpatcher.Apply(configpatcher.WithBytes(configBytes), loaded)
	if err != nil {
		return nil, contextualPatchError(joinPatchNames(patches), node, err)
	}
	return patched.Bytes()
}

func resolvePatchPath(patchesDir, name string) (string, error) {
	candidates := []string{name}
	if filepath.Ext(name) == "" {
		candidates = []string{name + ".yaml", name + ".yaml.tmpl", name + ".yml", name + ".yml.tmpl"}
	}
	for _, candidate := range candidates {
		path := candidate
		if !filepath.IsAbs(path) {
			path = filepath.Join(patchesDir, candidate)
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("patch file not found in %s", patchesDir)
}

func renderPatchTemplate(name, content string, inventory *Inventory, node Node) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var out bytes.Buffer
	data := struct {
		Inventory   *Inventory
		PatchInputs PatchInputs
		Node        Node
	}{
		Inventory: inventory,
		Node:      node,
	}
	if inventory != nil {
		data.PatchInputs = inventory.PatchInputs
	}
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return out.String(), nil
}

func contextualPatchError(name string, node Node, err error) error {
	nodeName := strings.TrimSpace(node.Name)
	if nodeName == "" {
		nodeName = "<unknown>"
	}
	role := string(node.Role)
	if role == "" {
		role = "<unknown>"
	}
	return fmt.Errorf("patch %s for node %s (%s): %w", name, nodeName, role, err)
}

func joinPatchNames(patches []NamedPatch) string {
	names := make([]string, 0, len(patches))
	for _, patch := range patches {
		names = append(names, patch.Name)
	}
	return strings.Join(names, ",")
}
