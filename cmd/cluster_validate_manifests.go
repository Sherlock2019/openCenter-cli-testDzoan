// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rackerlabs/opencenter-cli/internal/config"
	"github.com/rackerlabs/opencenter-cli/internal/gitops"
)

func newClusterValidateManifestsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-manifests [cluster-name]",
		Short: "Validate generated GitOps manifests for common issues",
		Long: `Validate generated GitOps manifests for common issues documented in lessons-learned.

This command checks for:
- Proper YAML indentation (2 spaces)
- Correct interval values (5m for Kustomizations, 15m for GitRepositories)
- No hardcoded cluster names (dev-cluster, stage-cluster)
- Proper repository URL capitalization (openCenter not opencenter)
- Branch-based refs (not tag-based)
- Base64 encoded secrets
- Correct domain names and hostname formats
- Proper snapshotter versions and registries
- Valid IP address ranges

Examples:
  # Validate manifests for current cluster
  opencenter cluster validate-manifests

  # Validate manifests for specific cluster
  opencenter cluster validate-manifests my-cluster
`,
		RunE: runClusterValidateManifests,
	}

	return cmd
}

func runClusterValidateManifests(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfigForCommand(cmd, args)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get git directory
	gitDir := cfg.GitOps().GitDir
	if gitDir == "" {
		return fmt.Errorf("git_dir not configured")
	}

	fmt.Printf("Validating GitOps manifests in: %s\n\n", gitDir)

	// Create validator
	validator := gitops.NewManifestValidator(gitDir)

	// Run validation
	if err := validator.Validate(); err != nil {
		fmt.Printf("❌ Validation failed:\n\n%v\n", err)
		return fmt.Errorf("manifest validation failed")
	}

	fmt.Printf("✅ All manifests validated successfully\n")
	return nil
}

// loadConfigForCommand loads the configuration for a cluster command
func loadConfigForCommand(cmd *cobra.Command, args []string) (config.Config, error) {
	var clusterName string
	if len(args) > 0 {
		clusterName = args[0]
	}

	// Get config manager
	cm, err := config.NewConfigManager("")
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to create config manager: %w", err)
	}

	// If no cluster name provided, use current cluster
	if clusterName == "" {
		currentCluster, err := cm.GetCurrentCluster()
		if err != nil {
			return config.Config{}, fmt.Errorf("no cluster specified and no current cluster set: %w", err)
		}
		clusterName = currentCluster
	}

	// Load cluster configuration
	cfg, err := cm.LoadClusterConfig(clusterName)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to load cluster config: %w", err)
	}

	return cfg, nil
}
