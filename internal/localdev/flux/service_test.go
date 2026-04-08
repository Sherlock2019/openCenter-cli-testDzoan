package flux

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	"github.com/opencenter-cloud/opencenter-cli/internal/localdev"
	"github.com/opencenter-cloud/opencenter-cli/internal/localdev/gitea"
	"gopkg.in/yaml.v3"
)

type fakeExecutor struct {
	t        *testing.T
	handlers map[string]func(opts localdev.RunOptions) ([]byte, error)
}

func (f *fakeExecutor) Run(ctx context.Context, opts localdev.RunOptions) ([]byte, error) {
	key := opts.Name + " " + strings.Join(opts.Args, " ")
	handler, ok := f.handlers[key]
	if !ok {
		f.t.Fatalf("unexpected command: %s", key)
	}
	return handler(opts)
}

func (f *fakeExecutor) RunStreaming(ctx context.Context, opts localdev.RunOptions) error {
	_, err := f.Run(ctx, opts)
	return err
}

func TestBootstrapUsesKindIPURLAndReconciles(t *testing.T) {
	configDir := t.TempDir()
	stateDir := t.TempDir()
	gitDir := filepath.Join(t.TempDir(), "gitops")
	clusterName := "dev-cluster"
	org := "local"
	t.Setenv("OPENCENTER_CONFIG_DIR", configDir)

	clusterCtx := writeClusterFixture(t, configDir, clusterName, org, gitDir)
	status := writeGiteaState(t, stateDir)

	executor := &fakeExecutor{
		t: t,
		handlers: map[string]func(opts localdev.RunOptions) ([]byte, error){
			"podman inspect --format {{.State.Running}} gitea": func(opts localdev.RunOptions) ([]byte, error) {
				return []byte("true"), nil
			},
			"podman inspect --format {{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{\"\\n\"}}{{end}} gitea": func(opts localdev.RunOptions) ([]byte, error) {
				return []byte("podman\nkind\n"), nil
			},
			"podman inspect --format {{with index .NetworkSettings.Networks \"kind\"}}{{.IPAddress}}{{end}} gitea": func(opts localdev.RunOptions) ([]byte, error) {
				return []byte(status.KindIP), nil
			},
			"git branch --show-current": func(opts localdev.RunOptions) ([]byte, error) {
				return []byte("main"), nil
			},
			fmt.Sprintf("flux bootstrap git --url=https://%s:3001/newuser/test-repo.git --branch=main --path=applications/overlays/%s --token-auth --username=newuser --password=user-token --ca-file=%s", status.KindIP, clusterName, status.CAPath): func(opts localdev.RunOptions) ([]byte, error) {
				if opts.Env["KUBECONFIG"] != clusterCtx.Paths.KubeconfigPath {
					t.Fatalf("bootstrap KUBECONFIG = %q, want %q", opts.Env["KUBECONFIG"], clusterCtx.Paths.KubeconfigPath)
				}
				return nil, nil
			},
			fmt.Sprintf("git -c http.sslCAInfo=%s -c http.extraHeader=Authorization: Basic bmV3dXNlcjp1c2VyLXRva2Vu pull --rebase origin main", status.CAPath): func(opts localdev.RunOptions) ([]byte, error) {
				return nil, nil
			},
			"flux reconcile source git flux-system -n flux-system": func(opts localdev.RunOptions) ([]byte, error) {
				if opts.Env["KUBECONFIG"] != clusterCtx.Paths.KubeconfigPath {
					t.Fatalf("reconcile KUBECONFIG = %q, want %q", opts.Env["KUBECONFIG"], clusterCtx.Paths.KubeconfigPath)
				}
				return nil, nil
			},
		},
	}

	service, err := NewService(executor, stateDir)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	result, err := service.Bootstrap(context.Background(), org+"/"+clusterName)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	wantURL := "https://10.89.0.11:3001/newuser/test-repo.git"
	if result.RepoURL != wantURL {
		t.Fatalf("RepoURL = %q, want %q", result.RepoURL, wantURL)
	}
	if result.KubeconfigPath != clusterCtx.Paths.KubeconfigPath {
		t.Fatalf("KubeconfigPath = %q, want %q", result.KubeconfigPath, clusterCtx.Paths.KubeconfigPath)
	}
}

func writeClusterFixture(t *testing.T, configDir, clusterName, org, gitDir string) *localdev.ClusterContext {
	t.Helper()

	cfg, err := config.NewProviderDefault(clusterName, "kind")
	if err != nil {
		t.Fatalf("NewProviderDefault() error = %v", err)
	}
	cfg.OpenCenter.Meta.Organization = org
	cfg.OpenCenter.GitOps.GitDir = gitDir
	cfg.OpenCenter.Infrastructure.Kind.Runtime = "podman"

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	orgDir := filepath.Join(configDir, "clusters", org)
	clusterDir := filepath.Join(orgDir, "infrastructure", "clusters", clusterName)
	if err := os.MkdirAll(clusterDir, 0o755); err != nil {
		t.Fatalf("mkdir cluster dir: %v", err)
	}
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clusterDir, "kubeconfig.yaml"), []byte("apiVersion: v1\n"), 0o644); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	if err := os.WriteFile(filepath.Join(orgDir, "."+clusterName+"-config.yaml"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	resolver, err := localdev.NewClusterResolver()
	if err != nil {
		t.Fatalf("NewClusterResolver() error = %v", err)
	}
	clusterCtx, err := resolver.Resolve(context.Background(), org+"/"+clusterName)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	return clusterCtx
}

func writeGiteaState(t *testing.T, stateDir string) *gitea.Status {
	t.Helper()

	layout, err := localdev.ResolveLayout(stateDir)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	if err := layout.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	metadata := gitea.Metadata{
		Runtime:       "podman",
		Image:         "docker.gitea.com/gitea:1.24.5",
		ContainerName: "gitea",
		HTTPPort:      3000,
		HTTPSPort:     3001,
		SSHPort:       2222,
		AdminUser:     "admin",
		RepoOwner:     "newuser",
		RepoName:      "test-repo",
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(layout.MetadataPath, data, 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	if err := os.WriteFile(layout.UserTokenPath, []byte("user-token"), 0o600); err != nil {
		t.Fatalf("write user token: %v", err)
	}
	if err := os.WriteFile(layout.CACertPath, []byte("ca"), 0o644); err != nil {
		t.Fatalf("write ca: %v", err)
	}

	return &gitea.Status{
		Metadata:     metadata,
		BaseURL:      "https://localhost:3001",
		LocalRepoURL: "https://localhost:3001/newuser/test-repo.git",
		KindIP:       "10.89.0.11",
		Running:      true,
		CAPath:       layout.CACertPath,
	}
}
