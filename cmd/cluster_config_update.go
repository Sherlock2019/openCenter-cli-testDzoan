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
	"os"
	"time"

	"github.com/rackerlabs/opencenter-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newClusterConfigUpdateCmd creates the "cluster config update" command.
// This command adds missing keys to an existing cluster configuration by merging
// with the default configuration template. It creates a backup before modifying.
func newClusterConfigUpdateCmd() *cobra.Command {
	var noBackup bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Add missing keys to existing cluster configuration",
		Long: `Add missing keys to an existing cluster configuration.

This command loads the current cluster configuration, merges it with the default
configuration template to add any missing keys, and writes the updated configuration
back to the file.

A timestamped backup is automatically created before modification:
  <config-file>.backup.<timestamp>

The backup allows you to review changes and revert if needed. Delete the backup
once you're satisfied with the updated configuration.

Missing keys are added with their default values based on:
  • Provider-specific defaults (if provider is configured)
  • Global schema defaults
  • Empty/zero values for required fields

Existing values are preserved - only missing keys are added.

If no cluster name is provided, updates the currently active cluster.`,
		Example: `  # Update active cluster configuration
  opencenter cluster config update

  # Update specific cluster
  opencenter cluster config update my-cluster

  # Update with organization prefix
  opencenter cluster config update myorg/my-cluster

  # Dry run to preview changes
  opencenter cluster config update my-cluster --dry-run

  # Update without creating backup (not recommended)
  opencenter cluster config update my-cluster --no-backup`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve cluster name from args or active cluster
			name, err := resolveClusterName(args, true)
			if err != nil {
				return err
			}

			// Get configuration file path
			configPath, err := config.ConfigPath(name)
			if err != nil {
				return fmt.Errorf("failed to resolve configuration path: %w", err)
			}

			// Check if configuration file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return fmt.Errorf("configuration file does not exist: %s", configPath)
			}

			// Load existing configuration
			existingConfig, err := config.Load(name)
			if err != nil {
				return fmt.Errorf("failed to load existing configuration: %w", err)
			}

			// Generate complete configuration with all defaults
			completeConfig, err := config.GenerateCompleteConfig(name)
			if err != nil {
				return fmt.Errorf("failed to generate complete configuration: %w", err)
			}

			// Merge configurations: existing values take precedence, missing keys are added
			mergedConfig := mergeConfigurations(&existingConfig, &completeConfig)

			// Marshal both configs for comparison
			existingData, err := yaml.Marshal(&existingConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal existing configuration: %w", err)
			}

			mergedData, err := yaml.Marshal(mergedConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal merged configuration: %w", err)
			}

			// Check if there are any changes
			if string(existingData) == string(mergedData) {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Configuration is already up to date\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  No missing keys found\n")
				return nil
			}

			// Count added keys (approximate by line difference)
			existingLines := len(existingData)
			mergedLines := len(mergedData)
			addedBytes := mergedLines - existingLines

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run - no changes will be made\n\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Would update configuration:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  File: %s\n", configPath)
				fmt.Fprintf(cmd.OutOrStdout(), "  Current size: %d bytes\n", existingLines)
				fmt.Fprintf(cmd.OutOrStdout(), "  Updated size: %d bytes\n", mergedLines)
				fmt.Fprintf(cmd.OutOrStdout(), "  Added: ~%d bytes\n", addedBytes)
				if !noBackup {
					backupPath := generateBackupPath(configPath)
					fmt.Fprintf(cmd.OutOrStdout(), "  Backup would be created: %s\n", backupPath)
				}
				return nil
			}

			// Create backup unless disabled
			var backupPath string
			if !noBackup {
				backupPath = generateBackupPath(configPath)
				if err := os.WriteFile(backupPath, existingData, 0600); err != nil {
					return fmt.Errorf("failed to create backup: %w", err)
				}
			}

			// Write updated configuration
			if err := os.WriteFile(configPath, mergedData, 0600); err != nil {
				return fmt.Errorf("failed to write updated configuration: %w", err)
			}

			// Success output
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Configuration updated successfully\n\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Updated configuration:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  File: %s\n", configPath)
			fmt.Fprintf(cmd.OutOrStdout(), "  Added: ~%d bytes of missing keys\n", addedBytes)

			if !noBackup {
				fmt.Fprintf(cmd.OutOrStdout(), "\nBackup created:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", backupPath)
				fmt.Fprintf(cmd.OutOrStdout(), "\nReview the changes and delete the backup if satisfied:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  rm %s\n", backupPath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip creating backup before updating")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying files")

	return cmd
}

// generateBackupPath generates a timestamped backup file path.
func generateBackupPath(originalPath string) string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s.backup.%s", originalPath, timestamp)
}

// mergeConfigurations merges two configurations using YAML deep merge.
// Existing values are preserved, missing keys from complete config are added.
func mergeConfigurations(existing, complete *config.Config) *config.Config {
	// Marshal both configs to YAML nodes for deep merge
	existingNode := &yaml.Node{}
	completeNode := &yaml.Node{}

	existingData, _ := yaml.Marshal(existing)
	completeData, _ := yaml.Marshal(complete)

	yaml.Unmarshal(existingData, existingNode)
	yaml.Unmarshal(completeData, completeNode)

	// Perform deep merge: complete provides defaults, existing overrides
	mergedNode := deepMergeYAML(completeNode, existingNode)

	// Unmarshal back to Config struct
	merged := &config.Config{}
	mergedData, _ := yaml.Marshal(mergedNode)
	yaml.Unmarshal(mergedData, merged)

	return merged
}

// deepMergeYAML performs a deep merge of two YAML nodes.
// Values from override take precedence over base.
func deepMergeYAML(base, override *yaml.Node) *yaml.Node {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// If override is not a mapping, use it directly
	if override.Kind != yaml.MappingNode {
		return override
	}

	// If base is not a mapping, use override
	if base.Kind != yaml.MappingNode {
		return override
	}

	// Create result node
	result := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: make([]*yaml.Node, 0),
	}

	// Build map of override keys for quick lookup
	overrideMap := make(map[string]*yaml.Node)
	for i := 0; i < len(override.Content); i += 2 {
		key := override.Content[i].Value
		value := override.Content[i+1]
		overrideMap[key] = value
	}

	// Merge base keys with override
	for i := 0; i < len(base.Content); i += 2 {
		keyNode := base.Content[i]
		baseValue := base.Content[i+1]
		key := keyNode.Value

		if overrideValue, exists := overrideMap[key]; exists {
			// Key exists in both - recursively merge if both are mappings
			if baseValue.Kind == yaml.MappingNode && overrideValue.Kind == yaml.MappingNode {
				result.Content = append(result.Content, keyNode, deepMergeYAML(baseValue, overrideValue))
			} else {
				// Use override value
				result.Content = append(result.Content, keyNode, overrideValue)
			}
			delete(overrideMap, key)
		} else {
			// Key only in base - add it
			result.Content = append(result.Content, keyNode, baseValue)
		}
	}

	// Add remaining keys from override that weren't in base
	for i := 0; i < len(override.Content); i += 2 {
		key := override.Content[i].Value
		if _, processed := overrideMap[key]; processed {
			result.Content = append(result.Content, override.Content[i], override.Content[i+1])
		}
	}

	return result
}
