package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestClusterUpdateKindDisableDefaultCNIFlagSetsTrue(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	cfg, _ := saveKindConfigForCommandTest(t, dir, "update-kind-cni", "opencenter")
	if cfg.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind infrastructure config to be present")
	}

	cmd := newClusterUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"update-kind-cni", "--kind-disable-default-cni"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster update failed: %v\nstderr: %s", err, stderr.String())
	}

	resetCommandStateForTests()

	updated, err := loadCanonicalConfig("update-kind-cni")
	if err != nil {
		t.Fatalf("load canonical config: %v", err)
	}
	if updated.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind infrastructure config to be present")
	}
	if !updated.OpenCenter.Infrastructure.Kind.DisableDefaultCNI {
		t.Fatal("expected disable_default_cni to be true after cluster update")
	}
}

func TestClusterUpdateKindDisableDefaultCNIFlagSetsFalse(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	cfg, _ := saveKindConfigForCommandTest(t, dir, "update-kind-cni-false", "opencenter")
	if cfg.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind infrastructure config to be present")
	}
	cfg.OpenCenter.Infrastructure.Kind.DisableDefaultCNI = true
	if err := saveConfig(context.Background(), cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	cmd := newClusterUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"update-kind-cni-false", "--kind-disable-default-cni=false"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cluster update failed: %v\nstderr: %s", err, stderr.String())
	}

	resetCommandStateForTests()

	updated, err := loadCanonicalConfig("update-kind-cni-false")
	if err != nil {
		t.Fatalf("load canonical config: %v", err)
	}
	if updated.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind infrastructure config to be present")
	}
	if updated.OpenCenter.Infrastructure.Kind.DisableDefaultCNI {
		t.Fatal("expected disable_default_cni to be false after cluster update")
	}
}

func TestClusterUpdateRejectsKindDisableDefaultCNIForNonKind(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	saveOpenStackStatusConfig(t, dir, "update-openstack-cni", "opencenter")

	cmd := newClusterUpdateCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"update-openstack-cni", "--kind-disable-default-cni"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected cluster update to reject --kind-disable-default-cni for non-kind clusters")
	}
	if !strings.Contains(err.Error(), "--kind-disable-default-cni is only valid for kind clusters") {
		t.Fatalf("expected kind-only error, got: %v", err)
	}
}

func TestClusterUpdateKindDisableDefaultCNIFromNativeV2Init(t *testing.T) {
	dir := t.TempDir()
	prepareCommandTestEnv(t, dir)

	initCmd := newClusterInitCmd()
	var initStdout, initStderr bytes.Buffer
	initCmd.SetOut(&initStdout)
	initCmd.SetErr(&initStderr)
	initCmd.SetArgs([]string{"update-kind-native-v2", "--type", "kind", "--org", "opencenter", "--no-keygen"})

	if err := initCmd.Execute(); err != nil {
		t.Fatalf("cluster init failed: %v\nstderr: %s", err, initStderr.String())
	}

	resetCommandStateForTests()

	updateCmd := newClusterUpdateCmd()
	var updateStdout, updateStderr bytes.Buffer
	updateCmd.SetOut(&updateStdout)
	updateCmd.SetErr(&updateStderr)
	updateCmd.SetArgs([]string{"update-kind-native-v2", "--kind-disable-default-cni"})

	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("cluster update failed: %v\nstderr: %s", err, updateStderr.String())
	}

	resetCommandStateForTests()

	updated, err := loadCanonicalConfig("update-kind-native-v2")
	if err != nil {
		t.Fatalf("load canonical config: %v", err)
	}
	if updated.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind infrastructure config to be present")
	}
	if !updated.OpenCenter.Infrastructure.Kind.DisableDefaultCNI {
		t.Fatal("expected disable_default_cni to be true after updating a native v2 kind cluster")
	}
}
