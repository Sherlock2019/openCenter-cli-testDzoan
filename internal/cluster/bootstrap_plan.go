package cluster

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
)

// BootstrapPlan describes what a dry-run cluster deploy would do.
type BootstrapPlan struct {
	Cluster         string
	Organization    string
	Provider        string
	ConfigPath      string
	GitOpsDir       string
	ClusterDir      string
	KubeconfigPath  string
	LogPath         string
	ResumeStatePath string
	Filter          string
	Steps           []BootstrapPlanStep
	Notes           []string
}

// BootstrapPlanStep describes a provider bootstrap step without executing it.
type BootstrapPlanStep struct {
	ID          string
	Action      string
	WorkingDir  string
	Commands    []BootstrapPlanCommand
	Reads       []string
	Writes      []string
	Environment []BootstrapPlanEnv
	Notes       []string
}

// BootstrapPlanCommand describes one command a real deploy may run.
type BootstrapPlanCommand struct {
	Name string
	Args []string
}

// BootstrapPlanEnv describes an environment variable used by a planned step.
type BootstrapPlanEnv struct {
	Name     string
	Value    string
	Redacted bool
}

func (s *BootstrapService) buildDryRunPlan(cfg *v2.Config, clusterPaths *paths.ClusterPaths, runtimePaths *bootstrapRuntimePaths, opts *BootstrapOptions, selectedSteps []bootstrapStep, filterText string) *BootstrapPlan {
	steps := make([]BootstrapPlanStep, 0, len(selectedSteps))
	for _, step := range selectedSteps {
		planStep := step.Plan
		if planStep.ID == "" {
			planStep.ID = step.ID
		}
		if planStep.Action == "" {
			planStep.Action = step.Description
		}
		steps = append(steps, planStep)
	}

	logPath := ""
	resumeStatePath := ""
	if runtimePaths != nil {
		logPath = runtimePaths.LogPath
		resumeStatePath = runtimePaths.StatePath
	}

	provider := strings.ToLower(strings.TrimSpace(cfg.Provider()))
	if provider == "" {
		provider = "openstack"
	}

	return &BootstrapPlan{
		Cluster:         cfg.ClusterName(),
		Organization:    cfg.Organization(),
		Provider:        provider,
		ConfigPath:      clusterPaths.ConfigPath,
		GitOpsDir:       clusterPaths.GitOpsDir,
		ClusterDir:      clusterPaths.ClusterDir,
		KubeconfigPath:  opts.KubeconfigPath,
		LogPath:         logPath,
		ResumeStatePath: resumeStatePath,
		Filter:          filterText,
		Steps:           steps,
		Notes: []string{
			"Plan only; command availability, local files, credentials, cluster state, and remote APIs were not fully validated.",
		},
	}
}

func filterDescription(opts *BootstrapOptions, ignoreState bool) string {
	if opts == nil {
		return ""
	}
	if strings.TrimSpace(opts.OnlyStep) != "" {
		return "--step " + strings.TrimSpace(opts.OnlyStep)
	}
	if strings.TrimSpace(opts.FromStep) != "" {
		return "--from-step " + strings.TrimSpace(opts.FromStep)
	}
	if opts.Restart || ignoreState {
		return "--restart"
	}
	return ""
}

func commandPlan(name string, args ...string) BootstrapPlanCommand {
	return BootstrapPlanCommand{Name: name, Args: args}
}

func envPlanFromMap(env map[string]string, redactedKeys map[string]bool) []BootstrapPlanEnv {
	if len(env) == 0 {
		return nil
	}

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	values := make([]BootstrapPlanEnv, 0, len(keys))
	for _, key := range keys {
		value := env[key]
		redacted := redactedKeys[key]
		if redacted {
			value = ""
		}
		values = append(values, BootstrapPlanEnv{
			Name:     key,
			Value:    value,
			Redacted: redacted,
		})
	}
	return values
}

func kindPlanEnv(env map[string]string) []BootstrapPlanEnv {
	safeEnv := make(map[string]string)
	if provider := strings.TrimSpace(env["KIND_EXPERIMENTAL_PROVIDER"]); provider != "" {
		safeEnv["KIND_EXPERIMENTAL_PROVIDER"] = provider
	}
	if _, ok := env["PATH"]; ok {
		safeEnv["PATH"] = "<current PATH>"
	}
	return envPlanFromMap(safeEnv, nil)
}

func openStackPlanEnv(kubeconfigPath string) []BootstrapPlanEnv {
	env := map[string]string{
		"OS_APPLICATION_CREDENTIAL_ID":     "",
		"OS_APPLICATION_CREDENTIAL_SECRET": "",
		"OS_AUTH_URL":                      "",
		"OS_IDENTITY_API_VERSION":          "3",
		"OS_INTERFACE":                     "public",
		"OS_PASSWORD":                      "",
		"OS_PROJECT_DOMAIN_NAME":           "",
		"OS_PROJECT_NAME":                  "",
		"OS_REGION_NAME":                   "",
		"OS_USER_DOMAIN_NAME":              "",
		"OS_USERNAME":                      "",
		"PATH":                             "<current PATH>",
	}
	if strings.TrimSpace(kubeconfigPath) != "" {
		env["KUBECONFIG"] = kubeconfigPath
	}

	redacted := map[string]bool{
		"OS_APPLICATION_CREDENTIAL_ID":     true,
		"OS_APPLICATION_CREDENTIAL_SECRET": true,
		"OS_AUTH_URL":                      true,
		"OS_PASSWORD":                      true,
		"OS_PROJECT_DOMAIN_NAME":           true,
		"OS_PROJECT_NAME":                  true,
		"OS_REGION_NAME":                   true,
		"OS_USER_DOMAIN_NAME":              true,
		"OS_USERNAME":                      true,
	}
	return envPlanFromMap(env, redacted)
}

func normalizePlanPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	expanded := strings.TrimSpace(path)
	if filepath.IsAbs(expanded) {
		return filepath.Clean(expanded)
	}
	if abs, err := filepath.Abs(expanded); err == nil {
		return filepath.Clean(abs)
	}
	return expanded
}

func kubeconfigCandidatePaths(clusterDir, targetPath string) []string {
	candidates := []string{
		targetPath,
		filepath.Join(clusterDir, "kubeconfig.yaml"),
		filepath.Join(clusterDir, "kubeconfig"),
		filepath.Join(clusterDir, "kube_config_cluster.yml"),
	}
	seen := make(map[string]bool, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = normalizePlanPath(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		result = append(result, candidate)
	}
	return result
}

func missingStepError(stepID string, steps []bootstrapStep) error {
	ids := make([]string, 0, len(steps))
	for _, step := range steps {
		ids = append(ids, step.ID)
	}
	return fmt.Errorf("unknown bootstrap step %q (available: %s)", stepID, strings.Join(ids, ", "))
}

func appendPlanNotes(notes []string, err error) []string {
	if err == nil {
		return notes
	}
	return append(notes, "Plan metadata warning: "+err.Error())
}
