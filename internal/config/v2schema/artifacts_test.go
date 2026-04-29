package v2schema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/defaults"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

func TestCheckedInSchemaIsCurrent(t *testing.T) {
	root := repoRoot(t)
	if err := CheckFile(filepath.Join(root, "schema", "opencenter-v2.schema.json"), Options{}); err != nil {
		t.Fatalf("checked-in v2 schema is not current: %v", err)
	}
}

func TestV2ExampleFixturesLoad(t *testing.T) {
	root := repoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, "testdata", "config", "v2", "*.yaml"))
	if err != nil {
		t.Fatalf("glob v2 examples: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("expected at least one v2 example fixture in testdata/config/v2")
	}

	loader := v2.NewConfigLoader(defaults.NewRegistry())
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if _, err := loader.LoadFromFile(path); err != nil {
				t.Fatalf("LoadFromFile(%s) error = %v", path, err)
			}
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repo root")
		}
		dir = parent
	}
}
