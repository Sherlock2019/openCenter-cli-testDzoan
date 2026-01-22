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
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rackerlabs/openCenter-cli/internal/config"
)

func newClusterTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "template",
		Short:  "Generate a complete cluster configuration template with all fields",
		Hidden: true, // Hidden from help - development/documentation command
		Long: `Generate a complete cluster configuration template showing all available fields.

This command creates a comprehensive YAML configuration file that includes:
- All configuration sections (opencenter, opentofu, secrets, metadata)
- All service configurations with their specific fields
- All provider options (OpenStack, AWS, Talos)
- Inline documentation and examples
- Default values for all fields

The generated template is useful for:
- Understanding the complete configuration schema
- Documentation and examples
- IDE autocomplete reference
- Migration from older schema versions
- Testing configuration validation

Unlike 'cluster init', which creates a minimal working configuration,
'cluster template' generates a comprehensive reference showing every
available option.`,
		Example: `  # Generate template to stdout
  openCenter cluster template

  # Save template to file
  openCenter cluster template --out complete-config.yaml

  # Generate template for specific provider
  openCenter cluster template --provider openstack --out openstack-template.yaml

  # Generate with comments explaining each field
  openCenter cluster template --comments --out documented-config.yaml

  # Generate minimal template (only required fields)
  openCenter cluster template --minimal --out minimal-config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outPath, _ := cmd.Flags().GetString("out")
			provider, _ := cmd.Flags().GetString("provider")
			withComments, _ := cmd.Flags().GetBool("comments")
			minimal, _ := cmd.Flags().GetBool("minimal")

			// Validate provider
			validProviders := []string{"openstack", "aws", "talos", "kind", "baremetal", "all"}
			if provider != "" && provider != "all" {
				valid := false
				for _, p := range validProviders {
					if p == provider {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("invalid provider '%s', must be one of: openstack, aws, talos, kind, baremetal, all", provider)
				}
			}

			// Generate template configuration
			var cfg config.Config
			if minimal {
				cfg = generateMinimalTemplate(provider)
			} else {
				cfg = generateCompleteTemplate(provider)
			}

			// Marshal to YAML
			data, err := yaml.Marshal(&cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal configuration: %w", err)
			}

			// Add comments if requested
			var output []byte
			if withComments {
				output = addConfigComments(data, provider)
			} else {
				output = data
			}

			// Write to file or stdout
			if outPath == "" {
				fmt.Fprint(cmd.OutOrStdout(), string(output))
				return nil
			}

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			if err := os.WriteFile(outPath, output, 0o644); err != nil {
				return fmt.Errorf("failed to write template file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Complete cluster configuration template written to %s\n", outPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Schema version: %s\n", config.GetSchemaVersion())
			return nil
		},
	}

	cmd.Flags().String("out", "", "output file path (default stdout)")
	cmd.Flags().String("provider", "all", "generate template for specific provider (openstack, aws, talos, kind, baremetal, all)")
	cmd.Flags().Bool("comments", false, "include inline comments explaining each field")
	cmd.Flags().Bool("minimal", false, "generate minimal template with only required fields")

	return cmd
}

// generateCompleteTemplate creates a configuration with all fields populated
func generateCompleteTemplate(provider string) config.Config {
	// Start with default config
	cfg := config.NewDefault("example-cluster")

	// Populate all optional fields with example values
	cfg.Metadata.CreatedBy = "admin@example.com"
	cfg.Metadata.Tags = map[string]string{
		"environment": "production",
		"cost-center": "engineering",
		"team":        "platform",
	}
	cfg.Metadata.Annotations = map[string]string{
		"description": "Complete cluster configuration template",
		"owner":       "Platform Team",
	}

	// Populate cluster fields
	cfg.OpenCenter.Cluster.BaseDomain = "k8s.opencenter.cloud"
	cfg.OpenCenter.Cluster.ClusterFQDN = "example-cluster.sjc3.k8s.opencenter.cloud"
	cfg.OpenCenter.Cluster.AdminEmail = "admin@example.com"
	cfg.OpenCenter.Cluster.SSHAuthorizedKeys = []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExamplePublicKeyDataHere user@example.com",
	}

	// Populate Kubernetes configuration
	cfg.OpenCenter.Cluster.Kubernetes.Version = "1.33.5"
	cfg.OpenCenter.Cluster.Kubernetes.MasterCount = 3
	cfg.OpenCenter.Cluster.Kubernetes.WorkerCount = 3

	// Enable Talos if requested or if provider is "all"
	if provider == "talos" || provider == "all" {
		cfg.OpenCenter.Talos = config.DefaultTalosConfig("example-cluster")
	}

	// Configure provider-specific settings
	switch provider {
	case "openstack":
		cfg.OpenCenter.Infrastructure.Provider = "openstack"
		populateOpenStackConfig(&cfg)
	case "aws":
		cfg.OpenCenter.Infrastructure.Provider = "aws"
		populateAWSConfig(&cfg)
	case "kind":
		cfg.OpenCenter.Infrastructure.Provider = "kind"
	case "baremetal":
		cfg.OpenCenter.Infrastructure.Provider = "baremetal"
		populateBaremetalConfig(&cfg)
	case "all", "":
		// Include all provider configurations
		populateOpenStackConfig(&cfg)
		populateAWSConfig(&cfg)
	}

	return cfg
}

// generateMinimalTemplate creates a configuration with only required fields
func generateMinimalTemplate(provider string) config.Config {
	cfg := config.Config{
		SchemaVersion: config.SchemaVersion,
		OpenCenter: config.SimplifiedOpenCenter{
			Meta: config.ClusterMeta{
				Name:         "example-cluster",
				Organization: "opencenter",
			},
			Infrastructure: config.Infrastructure{
				Provider: "openstack",
			},
			Cluster: config.ClusterConfig{
				ClusterName: "example-cluster",
				Kubernetes: config.KubernetesConfig{
					Version:     "1.33.5",
					MasterCount: 3,
					WorkerCount: 2,
				},
			},
			GitOps: config.GitOpsConfig{
				GitDir: "./gitops-repo",
			},
		},
		OpenTofu: config.SimplifiedOpenTofu{
			Enabled: true,
		},
		Secrets: config.Secrets{},
	}

	if provider != "" && provider != "all" {
		cfg.OpenCenter.Infrastructure.Provider = provider
	}

	return cfg
}

// populateOpenStackConfig adds OpenStack-specific configuration
func populateOpenStackConfig(cfg *config.Config) {
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL = "https://identity.api.sjc3.rackspacecloud.com/v3"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region = "sjc3"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Domain = "Default"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialID = ""
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialSecret = ""
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ImageID = "799dcf97-3656-4361-8187-13ab1b295e33"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Networking.FloatingIPPool = "PUBLICNET"
}

// populateAWSConfig adds AWS-specific configuration
func populateAWSConfig(cfg *config.Config) {
	cfg.OpenCenter.Infrastructure.Cloud.AWS.Region = "us-east-1"
	cfg.OpenCenter.Infrastructure.Cloud.AWS.Profile = ""
	cfg.OpenCenter.Infrastructure.Cloud.AWS.VPCID = ""
	cfg.OpenCenter.Infrastructure.Cloud.AWS.PrivateSubnets = []string{}
	cfg.OpenCenter.Infrastructure.Cloud.AWS.PublicSubnets = []string{}
}

// populateBaremetalConfig adds baremetal-specific configuration
func populateBaremetalConfig(cfg *config.Config) {
	cfg.OpenCenter.Infrastructure.Bastion.Address = "bastion.example.com"
	cfg.OpenCenter.Cluster.Kubernetes.MasterNodes = []config.NodeConfig{
		{
			ID:         "master-1",
			Name:       "master-1.example.com",
			AccessIPv4: "192.168.1.10",
		},
	}
	cfg.OpenCenter.Cluster.Kubernetes.WorkerNodes = []config.NodeConfig{
		{
			ID:         "worker-1",
			Name:       "worker-1.example.com",
			AccessIPv4: "192.168.1.20",
		},
	}
}

// addConfigComments adds inline comments to the YAML output
func addConfigComments(data []byte, provider string) []byte {
	// For now, return the data as-is
	// TODO: Implement comment injection using yaml.v3 Node API
	// This would parse the YAML, add comments to nodes, and re-marshal
	header := []byte(`# Complete openCenter Cluster Configuration Template
# Schema Version: ` + config.GetSchemaVersion() + `
# Generated by: openCenter cluster template
#
# This template shows all available configuration options.
# Remove or comment out sections you don't need.
# See https://docs.opencenter.cloud for detailed documentation.

`)
	return append(header, data...)
}
