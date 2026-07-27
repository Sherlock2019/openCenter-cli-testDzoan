package cluster

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

const (
	baseRepoSecretStepID = "reconcile-base-repo-secret"
	baseRepoSecretName   = "opencenter-base"
	fluxSystemNamespace  = "flux-system"
)

type baseRepoSecretMaterial struct {
	AuthMethod string
	Data       map[string][]byte
}

// newBaseRepoSecretStep creates the live Flux Secret after bootstrap. The
// Secret payload is assembled from local credentials and never written into
// the GitOps repository.
func newBaseRepoSecretStep(cfg *v2.Config, kubeconfigPath string, runner lifecycleCommandRunner) bootstrapStep {
	return bootstrapStep{
		ID:          baseRepoSecretStepID,
		Description: "Reconcile base Git repository credentials for Flux",
		Plan: BootstrapPlanStep{
			ID:         baseRepoSecretStepID,
			Action:     "Reconcile the opencenter-base Secret in flux-system",
			WorkingDir: cfg.GitDir(),
			Commands:   []BootstrapPlanCommand{commandPlan("kubectl", "--kubeconfig", kubeconfigPath, "-n", fluxSystemNamespace, "create", "secret", "generic", baseRepoSecretName, "--dry-run=client", "-o", "yaml"), commandPlan("kubectl", "--kubeconfig", kubeconfigPath, "apply", "-f", "<temporary-secret-manifest>")},
			Reads:      []string{"local GitOps token or SSH key files", kubeconfigPath},
			Writes:     []string{fmt.Sprintf("Secret/%s in namespace %s", baseRepoSecretName, fluxSystemNamespace)},
			Notes:      []string{"Credential data is passed from local files to the Kubernetes API and is not written to the GitOps repository or dry-run plan."},
		},
		Run: func(ctx context.Context) error {
			return reconcileBaseRepoSecret(ctx, cfg, kubeconfigPath, runner)
		},
	}
}

func reconcileBaseRepoSecret(ctx context.Context, cfg *v2.Config, kubeconfigPath string, runner lifecycleCommandRunner) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	authMethod, err := generatedBaseRepoAuthMethod(cfg)
	if err != nil {
		return err
	}

	knownHosts := ""
	if authMethod == config.GitopsAuthMethodSSH {
		host, err := gitRepositoryHost(cfg.OpenCenter.GitOps.BaseRepo.URL)
		if err != nil {
			return fmt.Errorf("resolve base repository host: %w", err)
		}
		output, err := runner.Run(ctx, "", nil, "ssh-keyscan", "-H", host)
		if err != nil {
			return fmt.Errorf("collect SSH host key for %s: %w", host, err)
		}
		knownHosts = strings.TrimSpace(string(output))
		if knownHosts == "" {
			return fmt.Errorf("ssh-keyscan returned no host keys for %s", host)
		}
	}

	material, err := buildBaseRepoSecretMaterial(cfg, authMethod, knownHosts)
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "opencenter-base-secret-")
	if err != nil {
		return fmt.Errorf("create temporary secret directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	if err := os.Chmod(tempDir, 0o700); err != nil {
		return fmt.Errorf("secure temporary secret directory: %w", err)
	}

	args := []string{"--kubeconfig", kubeconfigPath, "-n", fluxSystemNamespace, "create", "secret", "generic", baseRepoSecretName}
	for _, key := range sortedSecretKeys(material.Data) {
		path := filepath.Join(tempDir, key)
		if err := os.WriteFile(path, material.Data[key], 0o600); err != nil {
			return fmt.Errorf("write temporary %s credential: %w", key, err)
		}
		args = append(args, "--from-file="+key+"="+path)
	}
	args = append(args, "--dry-run=client", "-o", "yaml")

	manifest, err := runner.Run(ctx, "", nil, "kubectl", args...)
	if err != nil {
		return fmt.Errorf("build %s Secret manifest: %w", baseRepoSecretName, err)
	}
	manifestPath := filepath.Join(tempDir, "secret.yaml")
	if err := os.WriteFile(manifestPath, manifest, 0o600); err != nil {
		return fmt.Errorf("write temporary Secret manifest: %w", err)
	}

	if _, err := runner.Run(ctx, "", nil, "kubectl", "--kubeconfig", kubeconfigPath, "apply", "-f", manifestPath); err != nil {
		return fmt.Errorf("apply %s Secret: %w", baseRepoSecretName, err)
	}
	return nil
}

func buildBaseRepoSecretMaterial(cfg *v2.Config, authMethod, knownHosts string) (baseRepoSecretMaterial, error) {
	authMethod = strings.ToLower(strings.TrimSpace(authMethod))
	if err := config.ValidateGitopsAuthMethod(authMethod); err != nil {
		return baseRepoSecretMaterial{}, err
	}

	material := baseRepoSecretMaterial{AuthMethod: authMethod, Data: map[string][]byte{}}
	switch authMethod {
	case config.GitopsAuthMethodToken:
		token, err := resolveFluxToken(cfg)
		if err != nil {
			return baseRepoSecretMaterial{}, fmt.Errorf("resolve token for %s Secret: %w", baseRepoSecretName, err)
		}
		material.Data["username"] = []byte("git")
		material.Data["password"] = []byte(token)
	case config.GitopsAuthMethodSSH:
		if cfg.OpenCenter.GitOps.Auth.SSH == nil {
			return baseRepoSecretMaterial{}, fmt.Errorf("SSH base repository auth requires opencenter.gitops.auth.ssh")
		}
		privatePath := strings.TrimSpace(cfg.OpenCenter.GitOps.Auth.SSH.PrivateKey)
		publicPath := strings.TrimSpace(cfg.OpenCenter.GitOps.Auth.SSH.PublicKey)
		if privatePath == "" || publicPath == "" {
			return baseRepoSecretMaterial{}, fmt.Errorf("SSH base repository auth requires private_key and public_key paths")
		}
		privateKey, err := os.ReadFile(privatePath)
		if err != nil {
			return baseRepoSecretMaterial{}, fmt.Errorf("read SSH private key: %w", err)
		}
		publicKey, err := os.ReadFile(publicPath)
		if err != nil {
			return baseRepoSecretMaterial{}, fmt.Errorf("read SSH public key: %w", err)
		}
		knownHosts = strings.TrimSpace(knownHosts)
		if knownHosts == "" {
			return baseRepoSecretMaterial{}, fmt.Errorf("SSH base repository auth requires known_hosts data")
		}
		material.Data["identity"] = privateKey
		material.Data["identity.pub"] = publicKey
		material.Data["known_hosts"] = []byte(knownHosts + "\n")
	}
	return material, nil
}

// generatedBaseRepoAuthMethod identifies the active base-source variant. A
// generate override is not persisted, so deploy reads the explicit marker that
// generation wrote into the current source manifests.
func generatedBaseRepoAuthMethod(cfg *v2.Config) (string, error) {
	if method := strings.ToLower(strings.TrimSpace(cfg.OpenCenter.GitOps.ResolvedAuthMethod)); method != "" {
		if err := config.ValidateGitopsAuthMethod(method); err != nil {
			return "", err
		}
		return method, nil
	}

	clusterName := strings.TrimSpace(cfg.ClusterName())
	gitDir := strings.TrimSpace(cfg.GitDir())
	if clusterName == "" || gitDir == "" {
		return "", fmt.Errorf("cannot determine base repository auth: cluster name and GitOps directory are required")
	}

	patterns := []string{
		filepath.Join(gitDir, "applications", "overlays", clusterName, "services", "sources", "*.yaml"),
		filepath.Join(gitDir, "applications", "overlays", clusterName, "managed-services", "sources", "*.yaml"),
	}
	for _, pattern := range patterns {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("find generated source manifests: %w", err)
		}
		for _, path := range paths {
			contents, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read generated source manifest %s: %w", path, err)
			}
			text := string(contents)
			if !strings.Contains(text, "name: "+baseRepoSecretName) {
				continue
			}
			switch {
			case strings.Contains(text, "# --- token auth (active) ---"):
				return config.GitopsAuthMethodToken, nil
			case strings.Contains(text, "# --- ssh auth (active) ---"):
				return config.GitopsAuthMethodSSH, nil
			}
		}
	}

	return "", fmt.Errorf("cannot determine base repository auth from generated sources; run opencenter cluster generate %s before deploy", clusterName)
}

func gitRepositoryHost(repositoryURL string) (string, error) {
	repositoryURL = strings.TrimSpace(repositoryURL)
	if repositoryURL == "" {
		return "", fmt.Errorf("repository URL is required")
	}
	if at := strings.Index(repositoryURL, "@"); at > 0 && !strings.Contains(repositoryURL[:at], "://") {
		parts := strings.SplitN(repositoryURL[at+1:], ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0]), nil
		}
	}
	parsed, err := url.Parse(repositoryURL)
	if err != nil {
		return "", err
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("repository URL %q does not include a host", repositoryURL)
	}
	return parsed.Hostname(), nil
}

func sortedSecretKeys(data map[string][]byte) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
