package localdev

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveLayout_DefaultsToDotOpenCenterLocal(t *testing.T) {
	wd := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	layout, err := ResolveLayout("")
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	resolvedWD, err := filepath.EvalSymlinks(wd)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error = %v", wd, err)
	}
	resolvedParent, err := filepath.EvalSymlinks(filepath.Dir(layout.Root))
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error = %v", filepath.Dir(layout.Root), err)
	}
	if resolvedParent != resolvedWD {
		t.Fatalf("layout.Root parent = %q, want %q", resolvedParent, resolvedWD)
	}
	if filepath.Base(layout.Root) != ".opencenter-local" {
		t.Fatalf("layout.Root base = %q, want .opencenter-local", filepath.Base(layout.Root))
	}
	if layout.AdminTokenPath != filepath.Join(layout.Root, "tokens", "gitea-admin.token") {
		t.Fatalf("unexpected admin token path: %s", layout.AdminTokenPath)
	}
	if layout.CACertPath != filepath.Join(layout.Root, "gitea", "gitea", "certs", "ca.pem") {
		t.Fatalf("unexpected ca path: %s", layout.CACertPath)
	}
}

func TestLayoutEnsureCreatesExpectedDirectories(t *testing.T) {
	layout, err := ResolveLayout(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	if err := layout.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	for _, dir := range []string{layout.Root, layout.GiteaConfDir, layout.GiteaCertDir, layout.TokensDir} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist, err=%v", dir, err)
		}
	}
}
