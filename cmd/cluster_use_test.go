package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClusterUseClearRemovesSessionFileAndEmitsShellClear(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	sessionFile := filepath.Join(dir, "session")
	if err := os.WriteFile(sessionFile, []byte("session-cluster"), 0o600); err != nil {
		t.Fatalf("write session file: %v", err)
	}
	t.Setenv("OPENCENTER_SESSION_FILE", sessionFile)

	if err := setActiveCluster("persistent-cluster"); err != nil {
		t.Fatalf("set persistent active cluster: %v", err)
	}

	cmd := newClusterUseCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--clear"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster use --clear failed: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(sessionFile); !os.IsNotExist(err) {
		t.Fatalf("expected session file to be removed, stat err=%v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "unset OPENCENTER_CLUSTER") {
		t.Fatalf("expected shell clear line in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Active cluster cleared") {
		t.Fatalf("expected human clear line in output, got:\n%s", output)
	}

	active, err := getActiveCluster()
	if err != nil {
		t.Fatalf("get active cluster: %v", err)
	}
	if active != "persistent-cluster" {
		t.Fatalf("expected persistent active cluster to remain, got %q", active)
	}
}
