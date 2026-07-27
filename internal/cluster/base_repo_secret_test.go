package cluster

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBaseRepoSecretMaterialToken(t *testing.T) {
	cfg := &v2.Config{}
	cfg.OpenCenter.GitOps.Auth.Token = &v2.GitOpsTokenAuth{Provider: "github", Token: "test-token"}

	material, err := buildBaseRepoSecretMaterial(cfg, config.GitopsAuthMethodToken, "")
	require.NoError(t, err)
	assert.Equal(t, config.GitopsAuthMethodToken, material.AuthMethod)
	assert.Equal(t, map[string][]byte{"username": []byte("git"), "password": []byte("test-token")}, material.Data)
}

func TestBuildBaseRepoSecretMaterialSSH(t *testing.T) {
	dir := t.TempDir()
	privatePath := filepath.Join(dir, "id_ed25519")
	publicPath := privatePath + ".pub"
	require.NoError(t, os.WriteFile(privatePath, []byte("private-key"), 0o600))
	require.NoError(t, os.WriteFile(publicPath, []byte("public-key"), 0o600))

	cfg := &v2.Config{}
	cfg.OpenCenter.GitOps.Auth.SSH = &v2.GitOpsSSHAuth{PrivateKey: privatePath, PublicKey: publicPath}

	material, err := buildBaseRepoSecretMaterial(cfg, config.GitopsAuthMethodSSH, "github.com ssh-ed25519 AAAA")
	require.NoError(t, err)
	assert.Equal(t, []byte("private-key"), material.Data["identity"])
	assert.Equal(t, []byte("public-key"), material.Data["identity.pub"])
	assert.Equal(t, []byte("github.com ssh-ed25519 AAAA\n"), material.Data["known_hosts"])
}

func TestBuildBaseRepoSecretMaterialRejectsMissingCredentials(t *testing.T) {
	cfg := &v2.Config{}
	_, err := buildBaseRepoSecretMaterial(cfg, config.GitopsAuthMethodSSH, "known-host")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gitops.auth.ssh")

	cfg.OpenCenter.GitOps.Auth.Token = &v2.GitOpsTokenAuth{Provider: "github", Token: v2.PlaceholderSecret}
	_, err = buildBaseRepoSecretMaterial(cfg, config.GitopsAuthMethodToken, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "placeholder")
}

func TestGeneratedBaseRepoAuthMethodReadsActiveMarker(t *testing.T) {
	gitDir := t.TempDir()
	cfg := &v2.Config{}
	cfg.OpenCenter.Cluster.ClusterName = "demo"
	cfg.OpenCenter.GitOps.Repository.LocalDir = gitDir
	sourceDir := filepath.Join(gitDir, "applications", "overlays", "demo", "services", "sources")
	require.NoError(t, os.MkdirAll(sourceDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "opencenter-cert-manager.yaml"), []byte("secretRef:\n  name: opencenter-base\n# --- ssh auth (active) ---\n"), 0o600))

	method, err := generatedBaseRepoAuthMethod(cfg)
	require.NoError(t, err)
	assert.Equal(t, config.GitopsAuthMethodSSH, method)
}

func TestReconcileBaseRepoSecretUsesTemporaryFilesAndDoesNotExposeToken(t *testing.T) {
	gitDir := t.TempDir()
	cfg := &v2.Config{}
	cfg.OpenCenter.Cluster.ClusterName = "demo"
	cfg.OpenCenter.GitOps.Repository.LocalDir = gitDir
	cfg.OpenCenter.GitOps.ResolvedAuthMethod = config.GitopsAuthMethodToken
	cfg.OpenCenter.GitOps.Auth.Token = &v2.GitOpsTokenAuth{Provider: "github", Token: "super-secret-token"}

	runner := &fakeLifecycleRunner{onRun: func(dir string, env map[string]string, name string, args ...string) ([]byte, error) {
		if name == "kubectl" && strings.Contains(strings.Join(args, " "), "create secret generic") {
			return []byte("apiVersion: v1\nkind: Secret\n"), nil
		}
		return nil, nil
	}}

	require.NoError(t, reconcileBaseRepoSecret(context.Background(), cfg, "/tmp/kubeconfig", runner))
	require.Len(t, runner.calls, 2)
	assert.Equal(t, "kubectl", runner.calls[0].name)
	assert.Contains(t, strings.Join(runner.calls[0].args, " "), "create secret generic opencenter-base")
	assert.NotContains(t, strings.Join(runner.calls[0].args, " "), "super-secret-token")
	assert.Equal(t, "kubectl", runner.calls[1].name)
	assert.Contains(t, strings.Join(runner.calls[1].args, " "), "apply -f")
}
