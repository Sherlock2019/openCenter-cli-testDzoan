package flux

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/localdev"
	"github.com/opencenter-cloud/opencenter-cli/internal/localdev/gitea"
	"github.com/opencenter-cloud/opencenter-cli/internal/localdev/gitops"
)

// BootstrapResult reports the inputs used for Flux bootstrap.
type BootstrapResult struct {
	GitDir         string
	RepoURL        string
	Branch         string
	KubeconfigPath string
}

// Service manages Flux bootstrap orchestration for local Kind clusters.
type Service struct {
	executor localdev.Executor
	resolver *localdev.ClusterResolver
	gitops   *gitops.Service
	stateDir string
}

// NewService returns a Flux orchestration helper.
func NewService(executor localdev.Executor, stateDir string) (*Service, error) {
	if executor == nil {
		executor = localdev.NewExecutor()
	}
	resolver, err := localdev.NewClusterResolver()
	if err != nil {
		return nil, err
	}
	gitopsService, err := gitops.NewService(executor, stateDir)
	if err != nil {
		return nil, err
	}
	return &Service{
		executor: executor,
		resolver: resolver,
		gitops:   gitopsService,
		stateDir: stateDir,
	}, nil
}

// Bootstrap runs `flux bootstrap git` against the attached local Gitea repo.
func (s *Service) Bootstrap(ctx context.Context, clusterIdentifier string) (*BootstrapResult, error) {
	cluster, err := s.resolver.Resolve(ctx, clusterIdentifier)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(cluster.Config.OpenCenter.Infrastructure.Provider, "kind") {
		return nil, fmt.Errorf("cluster %q is not a kind cluster", clusterIdentifier)
	}

	giteaService, err := gitea.NewService(s.executor, s.stateDir, gitea.DefaultSettings(""))
	if err != nil {
		return nil, err
	}
	status, err := giteaService.Status(ctx)
	if err != nil {
		return nil, err
	}
	if !status.Running {
		return nil, fmt.Errorf("local gitea is not running")
	}
	if status.KindIP == "" {
		return nil, fmt.Errorf("local gitea is not attached to the kind network; run `opencenter local gitea attach-kind --cluster %s` first", cluster.ClusterName)
	}
	if !status.UserTokenExists {
		return nil, fmt.Errorf("missing Gitea user token at %s", status.UserTokenPath)
	}

	gitDir := strings.TrimSpace(cluster.Config.GitOps().GitDir)
	if gitDir == "" {
		gitDir = cluster.Paths.GitOpsDir
	}
	if gitDir == "" {
		return nil, fmt.Errorf("cluster %q does not define a git_dir", clusterIdentifier)
	}
	if _, err := os.Stat(cluster.Paths.KubeconfigPath); err != nil {
		return nil, fmt.Errorf("cluster kubeconfig %s: %w", cluster.Paths.KubeconfigPath, err)
	}

	tokenBytes, err := os.ReadFile(status.UserTokenPath)
	if err != nil {
		return nil, fmt.Errorf("read Gitea user token: %w", err)
	}
	branch, err := s.gitops.CurrentBranch(ctx, gitDir)
	if err != nil {
		return nil, err
	}

	// Use localhost for the initial bootstrap clone (runs on the host).
	// The Kind network IP is unreachable from the host; localhost is
	// port-forwarded to the Gitea container.
	hostURL := fmt.Sprintf("https://localhost:%d/%s/%s.git", status.Metadata.HTTPSPort, status.Metadata.RepoOwner, status.Metadata.RepoName)
	// The in-cluster URL uses the Kind network IP so the Flux
	// source-controller inside the cluster can reach Gitea.
	inClusterURL := fmt.Sprintf("https://%s:%d/%s/%s.git", status.KindIP, status.Metadata.HTTPSPort, status.Metadata.RepoOwner, status.Metadata.RepoName)
	bootstrapPath := filepathForFlux(cluster.ClusterName)
	kubeconfigEnv := map[string]string{"KUBECONFIG": cluster.Paths.KubeconfigPath}

	if _, err := s.executor.Run(ctx, localdev.RunOptions{
		Name: "flux",
		Dir:  gitDir,
		Env:  kubeconfigEnv,
		Args: []string{
			"bootstrap", "git",
			"--url=" + hostURL,
			"--branch=" + branch,
			"--path=" + bootstrapPath,
			"--token-auth",
			"--username=" + status.Metadata.RepoOwner,
			"--password=" + strings.TrimSpace(string(tokenBytes)),
			"--ca-file=" + status.CAPath,
		},
	}); err != nil {
		return nil, fmt.Errorf("flux bootstrap git: %w", err)
	}

	// Patch the in-cluster GitRepository URL to the Kind network IP so
	// the source-controller can reach Gitea from inside the cluster.
	patchJSON := fmt.Sprintf(`{"spec":{"url":"%s"}}`, inClusterURL)
	if _, err := s.executor.Run(ctx, localdev.RunOptions{
		Name: "kubectl",
		Env:  kubeconfigEnv,
		Args: []string{
			"patch", "gitrepository", "flux-system",
			"-n", "flux-system",
			"--type=merge",
			"-p", patchJSON,
		},
	}); err != nil {
		return nil, fmt.Errorf("patch flux-system GitRepository URL to in-cluster address: %w", err)
	}

	if _, err := s.gitops.PullRebase(ctx, gitDir); err != nil {
		return nil, fmt.Errorf("sync local checkout after bootstrap: %w", err)
	}
	if _, err := s.executor.Run(ctx, localdev.RunOptions{
		Name: "flux",
		Env:  kubeconfigEnv,
		Args: []string{"reconcile", "source", "git", "flux-system", "-n", "flux-system"},
	}); err != nil {
		return nil, fmt.Errorf("flux reconcile source git flux-system: %w", err)
	}

	return &BootstrapResult{
		GitDir:         gitDir,
		RepoURL:        inClusterURL,
		Branch:         branch,
		KubeconfigPath: cluster.Paths.KubeconfigPath,
	}, nil
}

func filepathForFlux(clusterName string) string {
	return "applications/overlays/" + clusterName
}
