package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoadExternalPlugins_RegistersLocalPluginName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows for shell script exec")
	}

	dir := t.TempDir()
	script := filepath.Join(dir, "opencenter-local")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write plugin: %v", err)
	}
	t.Setenv("OPENCENTER_PLUGINS_DIR", dir)

	root := &cobra.Command{Use: "opencenter-test"}
	LoadExternalPlugins(root)

	if root.Commands()[0].Name() != "local" {
		t.Fatalf("expected plugin command name local, got %q", root.Commands()[0].Name())
	}
}
