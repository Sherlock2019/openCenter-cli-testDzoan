package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/cluster"
)

func printClusterDeployPlan(w io.Writer, plan *cluster.BootstrapPlan) {
	fmt.Fprintln(w, "Deploy plan only (dry-run)")
	fmt.Fprintln(w, "No commands will be run, no files will be written, and prerequisites are not fully validated.")
	fmt.Fprintln(w)

	if plan == nil {
		fmt.Fprintln(w, "No deploy plan was produced.")
		return
	}

	clusterName := plan.Cluster
	if strings.TrimSpace(plan.Organization) != "" {
		clusterName = plan.Organization + "/" + plan.Cluster
	}

	fmt.Fprintf(w, "Cluster: %s\n", clusterName)
	fmt.Fprintf(w, "Provider: %s\n", plan.Provider)
	fmt.Fprintf(w, "Config: %s\n", plan.ConfigPath)
	fmt.Fprintf(w, "GitOps dir: %s\n", plan.GitOpsDir)
	fmt.Fprintf(w, "Cluster dir: %s\n", plan.ClusterDir)
	fmt.Fprintf(w, "Kubeconfig: %s\n", plan.KubeconfigPath)
	fmt.Fprintf(w, "Log file would be: %s\n", plan.LogPath)
	fmt.Fprintf(w, "Resume state would be: %s\n", plan.ResumeStatePath)
	if strings.TrimSpace(plan.Filter) != "" {
		fmt.Fprintf(w, "Filter: %s\n", plan.Filter)
	}
	fmt.Fprintln(w)

	if len(plan.Notes) > 0 {
		fmt.Fprintln(w, "Notes:")
		for _, note := range plan.Notes {
			fmt.Fprintf(w, "  - %s\n", note)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Steps that would run:")
	for i, step := range plan.Steps {
		fmt.Fprintf(w, "  %d. %s\n", i+1, step.ID)
		if strings.TrimSpace(step.Action) != "" {
			fmt.Fprintf(w, "     Action: %s\n", step.Action)
		}
		if strings.TrimSpace(step.WorkingDir) != "" {
			fmt.Fprintf(w, "     Working dir: %s\n", step.WorkingDir)
		}
		for _, command := range step.Commands {
			fmt.Fprintf(w, "     Command: %s\n", formatPlanCommand(command))
		}
		printPlanList(w, "Reads", step.Reads)
		printPlanList(w, "Writes", step.Writes)
		if len(step.Environment) > 0 {
			fmt.Fprintln(w, "     Environment:")
			for _, env := range step.Environment {
				fmt.Fprintf(w, "       - %s\n", formatPlanEnv(env))
			}
		}
		printPlanList(w, "Notes", step.Notes)
	}
}

func formatPlanCommand(command cluster.BootstrapPlanCommand) string {
	parts := []string{command.Name}
	parts = append(parts, command.Args...)
	return strings.Join(parts, " ")
}

func formatPlanEnv(env cluster.BootstrapPlanEnv) string {
	if env.Redacted {
		return env.Name + "=<redacted>"
	}
	if strings.TrimSpace(env.Value) == "" {
		return env.Name
	}
	return env.Name + "=" + env.Value
}

func printPlanList(w io.Writer, label string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(w, "     %s:\n", label)
	for _, value := range values {
		fmt.Fprintf(w, "       - %s\n", value)
	}
}
