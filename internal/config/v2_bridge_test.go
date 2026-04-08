package config

import (
	"context"
	"path/filepath"
	"testing"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/validation"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/errors"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/fs"
)

func TestConfigurationManagerLoad_NativeV2Compatibility(t *testing.T) {
	baseDir := t.TempDir()
	pathResolver := paths.NewPathResolver(filepath.Join(baseDir, "clusters"))
	ctx := context.Background()

	if err := pathResolver.CreateClusterDirectories(ctx, "compat-kind", "test-org"); err != nil {
		t.Fatalf("CreateClusterDirectories() error = %v", err)
	}

	clusterPaths, err := pathResolver.Resolve(ctx, "compat-kind", "test-org")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	cfg, err := v2.NewV2Default("compat-kind", "kind")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}
	cfg.OpenCenter.Meta.Organization = "test-org"
	cfg.OpenCenter.GitOps.GitDir = clusterPaths.GitOpsDir
	cfg.Secrets.SopsAgeKeyFile = clusterPaths.SOPSKeyPath
	cfg.Secrets.SOPSConfig.AgeKeyFile = clusterPaths.SOPSKeyPath
	cfg.Secrets.SSHKey.Private = clusterPaths.SSHKeyPath
	cfg.Secrets.SSHKey.Public = clusterPaths.SSHKeyPath + ".pub"
	cfg.OpenCenter.GitOps.GitSSHKey = clusterPaths.SSHKeyPath
	cfg.OpenCenter.GitOps.GitSSHPub = clusterPaths.SSHKeyPath + ".pub"
	cfg.OpenCenter.Infrastructure.SSH.KeyPath = clusterPaths.SSHKeyPath
	cfg.OpenCenter.Infrastructure.Kind = &v2.KindCompatibilityConfig{DisableDefaultCNI: true}

	if err := defaultLegacyV2Loader().SaveToFile(cfg, clusterPaths.ConfigPath); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	errorHandler := errors.NewDefaultErrorHandlerWithoutMasking()
	fileSystem := fs.NewDefaultFileSystem(errorHandler)
	manager := NewConfigurationManagerWithDeps(
		NewConfigIOHandler(fileSystem),
		validation.NewValidationEngine(),
		NewConfigCache(),
		pathResolver,
		fileSystem,
	)

	legacyCfg, err := manager.Load(ctx, "test-org/compat-kind")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if legacyCfg.OpenCenter.Infrastructure.Provider != "kind" {
		t.Fatalf("provider = %q, want kind", legacyCfg.OpenCenter.Infrastructure.Provider)
	}
	if legacyCfg.OpenCenter.Infrastructure.Kind == nil {
		t.Fatal("expected kind compatibility projection to populate infrastructure.kind")
	}
	if !legacyCfg.OpenCenter.Infrastructure.Kind.DisableDefaultCNI {
		t.Fatal("expected native v2 kind.disable_default_cni to bridge into canonical config")
	}
	if legacyCfg.OpenCenter.Cluster.Kubernetes.MasterCount != cfg.OpenCenter.Infrastructure.Compute.MasterCount {
		t.Fatalf("master_count = %d, want %d", legacyCfg.OpenCenter.Cluster.Kubernetes.MasterCount, cfg.OpenCenter.Infrastructure.Compute.MasterCount)
	}
	if legacyCfg.OpenCenter.Cluster.Kubernetes.FlavorBastion == "" {
		t.Fatal("expected legacy kind projection to retain a non-empty bastion flavor default")
	}
	if legacyCfg.OpenTofu.Enabled {
		t.Fatal("expected opentofu to stay disabled for kind")
	}
}
