package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openstackprovider "github.com/opencenter-cloud/opencenter-cli/internal/cloud/openstack"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
	"github.com/opencenter-cloud/opencenter-cli/internal/credentials"
)

type openstackBootstrapProvider struct {
	runner lifecycleCommandRunner
}

func newOpenStackBootstrapProvider(runner lifecycleCommandRunner) lifecycleBootstrapProvider {
	return &openstackBootstrapProvider{runner: runner}
}

func (p *openstackBootstrapProvider) BuildSteps(cfg *v2.Config, _ *paths.ClusterPaths, opts *BootstrapOptions) ([]bootstrapStep, error) {
	clusterDir, err := infrastructureClusterDir(cfg)
	if err != nil {
		return nil, err
	}
	if !cfg.OpenTofu.Enabled {
		return nil, fmt.Errorf("opentofu must be enabled for openstack bootstrap")
	}

	openTofuPath := strings.TrimSpace(cfg.OpenTofu.Path)
	if openTofuPath == "" {
		openTofuPath = "opentofu"
	}
	planEnv := openStackPlanEnv(opts.KubeconfigPath)

	return []bootstrapStep{
		{
			ID:          "openstack-preflight",
			Description: "Validate OpenStack credentials and bootstrap prerequisites",
			Plan: BootstrapPlanStep{
				ID:         "openstack-preflight",
				Action:     "Validate OpenStack credentials and bootstrap prerequisites",
				WorkingDir: clusterDir,
				Reads:      []string{clusterDir},
				Notes:      []string{"Plan only; OpenStack credentials, infrastructure directory, and OpenTofu availability were not checked."},
			},
			Run: func(ctx context.Context) error {
				if _, err := os.Stat(clusterDir); err != nil {
					return fmt.Errorf("cluster infrastructure directory not found in GitOps repository: %s", clusterDir)
				}
				creds, err := extractOpenStackBootstrapCredentials(cfg)
				if err != nil {
					return err
				}
				return validateOpenStackBootstrap(creds)
			},
		},
		{
			ID:          "opentofu-init",
			Description: "Initialize OpenTofu",
			Plan: BootstrapPlanStep{
				ID:          "opentofu-init",
				Action:      "Initialize OpenTofu",
				WorkingDir:  clusterDir,
				Commands:    []BootstrapPlanCommand{commandPlan(openTofuPath, "init")},
				Environment: planEnv,
				Reads:       []string{clusterDir},
				Writes:      []string{filepath.Join(clusterDir, ".terraform")},
				Notes:       []string{"Plan only; OpenTofu binary, backend access, and provider initialization were not checked."},
			},
			Run: func(ctx context.Context) error {
				env, err := buildOpenStackBootstrapEnvironment(cfg, opts.KubeconfigPath)
				if err != nil {
					return err
				}
				_, runErr := p.runner.Run(ctx, clusterDir, env, openTofuPath, "init")
				return runErr
			},
		},
		{
			ID:          "opentofu-apply",
			Description: "Apply OpenTofu infrastructure",
			Plan: BootstrapPlanStep{
				ID:          "opentofu-apply",
				Action:      "Apply OpenTofu infrastructure",
				WorkingDir:  clusterDir,
				Commands:    []BootstrapPlanCommand{commandPlan(openTofuPath, "apply", "-auto-approve")},
				Environment: planEnv,
				Reads:       []string{clusterDir},
				Writes:      []string{"OpenStack infrastructure resources", filepath.Join(clusterDir, "terraform.tfstate")},
				Notes:       []string{"Plan only; OpenStack API access and infrastructure changes were not simulated."},
			},
			Run: func(ctx context.Context) error {
				env, err := buildOpenStackBootstrapEnvironment(cfg, opts.KubeconfigPath)
				if err != nil {
					return err
				}
				_, runErr := p.runner.Run(ctx, clusterDir, env, openTofuPath, "apply", "-auto-approve")
				return runErr
			},
		},
		{
			ID:          "openstack-normalize-kubeconfig",
			Description: "Normalize kubeconfig into the cluster-owned path",
			Plan: BootstrapPlanStep{
				ID:         "openstack-normalize-kubeconfig",
				Action:     "Normalize kubeconfig into the cluster-owned path",
				WorkingDir: clusterDir,
				Reads:      kubeconfigCandidatePaths(clusterDir, opts.KubeconfigPath),
				Writes:     []string{opts.KubeconfigPath},
				Notes:      []string{"Plan only; kubeconfig candidates were not checked."},
			},
			Run: func(ctx context.Context) error {
				return normalizeOpenStackKubeconfig(clusterDir, opts.KubeconfigPath)
			},
		},
	}, nil
}

func extractOpenStackBootstrapCredentials(cfg *v2.Config) (*credentials.OpenStackCredentials, error) {
	extractor := credentials.NewExtractor(*cfg)
	creds, err := extractor.ExtractOpenStack()
	if err != nil {
		return nil, fmt.Errorf("extract openstack credentials: %w", err)
	}
	return creds, nil
}

func buildOpenStackBootstrapEnvironment(cfg *v2.Config, kubeconfigPath string) (map[string]string, error) {
	creds, err := extractOpenStackBootstrapCredentials(cfg)
	if err != nil {
		return nil, err
	}
	env := buildBootstrapEnvironment(kubeconfigPath)
	mergeBootstrapEnvironment(env, creds.ToEnvMap())
	return env, nil
}

func validateOpenStackBootstrap(creds *credentials.OpenStackCredentials) error {
	if creds == nil || creds.IsEmpty() {
		return fmt.Errorf("openstack credentials are incomplete; set auth_url and application credentials or username/password before bootstrap")
	}

	// Reject placeholder credentials
	if creds.ApplicationCredentialID == "CHANGEME" || creds.ApplicationCredentialSecret == "CHANGEME" {
		return fmt.Errorf("openstack credentials are incomplete; application_credential_id and application_credential_secret must be replaced before bootstrap")
	}

	for _, warning := range openstackprovider.PreflightOpenStack(creds.AuthURL) {
		if strings.Contains(warning, "auth_url is empty") {
			return fmt.Errorf("%s", warning)
		}
	}

	return nil
}

func buildBootstrapEnvironment(kubeconfigPath string) map[string]string {
	env := make(map[string]string)

	if strings.TrimSpace(kubeconfigPath) != "" {
		env["KUBECONFIG"] = kubeconfigPath
	}
	if path := os.Getenv("PATH"); path != "" {
		env["PATH"] = path
	}

	return env
}

func mergeBootstrapEnvironment(target, extra map[string]string) {
	for key, value := range extra {
		if strings.TrimSpace(value) == "" {
			continue
		}
		target[key] = value
	}
}

func infrastructureClusterDir(cfg *v2.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration is nil")
	}

	gitDir := strings.TrimSpace(cfg.GitDir())
	if gitDir == "" {
		return "", fmt.Errorf("gitops.git_dir must be configured for provider %q", cfg.Provider())
	}

	clusterName := strings.TrimSpace(cfg.ClusterName())
	if clusterName == "" {
		return "", fmt.Errorf("cluster name must be set")
	}

	return filepath.Join(gitDir, "infrastructure", "clusters", clusterName), nil
}

func normalizeOpenStackKubeconfig(clusterDir, targetPath string) error {
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("kubeconfig path must be set")
	}

	candidates := []string{
		targetPath,
		filepath.Join(clusterDir, "kubeconfig.yaml"),
		filepath.Join(clusterDir, "kubeconfig"),
		filepath.Join(clusterDir, "kube_config_cluster.yml"),
	}

	var sourcePath string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			sourcePath = candidate
			break
		}
	}

	if sourcePath == "" {
		return fmt.Errorf("kubeconfig not found after bootstrap in %s", clusterDir)
	}

	if sameFilePath(sourcePath, targetPath) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create kubeconfig directory: %w", err)
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read kubeconfig %s: %w", sourcePath, err)
	}

	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		return fmt.Errorf("write kubeconfig %s: %w", targetPath, err)
	}

	return nil
}

func sameFilePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}

	absA, err := filepath.Abs(a)
	if err != nil {
		return a == b
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		return a == b
	}

	return absA == absB
}
