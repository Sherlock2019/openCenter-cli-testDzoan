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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rackerlabs/openCenter-cli/internal/barbican"
	"github.com/rackerlabs/openCenter-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage secrets with Barbican",
		Long:  `Provides a Barbican-backed control plane for handling credentials, bootstrap bundles, and opaque data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSecretsLoginCmd())
	cmd.AddCommand(newSecretsListCmd())
	cmd.AddCommand(newSecretsDescribeCmd())
	cmd.AddCommand(newSecretsGetCmd())
	cmd.AddCommand(newSecretsPutCmd())
	cmd.AddCommand(newSecretsDeleteCmd())
	cmd.AddCommand(newSecretsSyncCmd())

	return cmd
}

func newSecretsLoginCmd() *cobra.Command {
	var (
		username   string
		projectID  string
		passwordIn bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Create or refresh a Keystone token",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}

			barbicanCfg := &cfg.OpenCenter.Secrets.Barbican
			if projectID != "" {
				barbicanCfg.ProjectID = projectID
			}

			client, err := barbican.NewClient(barbicanCfg)
			if err != nil {
				return err
			}

			var password string
			if passwordIn {
				// Read password from stdin
				bytePassword, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("could not read password from stdin: %w", err)
				}
				password = strings.TrimSpace(string(bytePassword))
			}

			token, err := client.Login(cmd.Context(), username, password)
			if err != nil {
				return err
			}

			err = barbican.StoreToken(token)
			if err != nil {
				return err
			}

			fmt.Println("Successfully authenticated and token stored.")
			return nil
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "OpenStack username")
	cmd.Flags().StringVar(&projectID, "project-id", "", "OpenStack project ID")
	cmd.Flags().BoolVar(&passwordIn, "password-stdin", false, "Read password from stdin")

	return cmd
}

func newSecretsListCmd() *cobra.Command {
	var (
		labels []string
		format string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets associated with the current cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}
			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}
			labelMap, err := barbican.ParseLabels(labels)
			if err != nil {
				return err
			}
			secrets, err := client.ListSecrets(cmd.Context(), labelMap)
			if err != nil {
				return err
			}

			switch format {
			case "json":
				json.NewEncoder(os.Stdout).Encode(secrets)
			case "yaml":
				yaml.NewEncoder(os.Stdout).Encode(secrets)
			default:
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
				fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tCREATED")
				for _, secret := range secrets {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", secret.Name, secret.SecretType, secret.Status, secret.Created)
				}
				w.Flush()
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&labels, "label", []string{}, "Filter secrets by labels in key=value form")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, or yaml")
	return cmd
}

func newSecretsDescribeCmd() *cobra.Command {
	var (
		format string
	)
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Show metadata for a single secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}
			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}
			secret, err := client.DescribeSecret(cmd.Context(), name)
			if err != nil {
				return err
			}

			switch format {
			case "json":
				json.NewEncoder(os.Stdout).Encode(secret)
			case "yaml":
				yaml.NewEncoder(os.Stdout).Encode(secret)
			default:
				fmt.Printf("Name: %s\n", secret.Name)
				fmt.Printf("Type: %s\n", secret.SecretType)
				fmt.Printf("Status: %s\n", secret.Status)
				fmt.Printf("Created: %s\n", secret.Created)
				fmt.Printf("Content Types: %v\n", secret.ContentTypes)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, or yaml")
	return cmd
}

func newSecretsGetCmd() *cobra.Command {
	var (
		outputFile string
		show       bool
	)
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Download and decrypt a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}
			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}

			if outputFile == "" && !show {
				return fmt.Errorf("use --output-file to save the secret or --show to print it to stdout (warning: printing to stdout is insecure)")
			}

			payload, err := client.GetSecret(cmd.Context(), name)
			if err != nil {
				return err
			}
			if outputFile != "" {
				err := os.WriteFile(outputFile, payload, 0600)
				if err != nil {
					return err
				}
				fmt.Printf("Secret '%s' saved to %s\n", name, outputFile)
			}
			if show {
				if outputFile != "" {
					fmt.Println("--- Secret Content ---")
				} else {
					fmt.Fprintln(os.Stderr, "Warning: Printing secret to stdout is insecure.")
				}
				fmt.Println(string(payload))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outputFile, "output-file", "", "Path to save the secret")
	cmd.Flags().BoolVar(&show, "show", false, "Print secret to stdout (insecure)")
	return cmd
}

func newSecretsPutCmd() *cobra.Command {
	var (
		fromFile               string
		labels                 []string
		secretType             string
		payloadContentEncoding string
	)
	cmd := &cobra.Command{
		Use:   "put <name>",
		Short: "Create or update a Barbican secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			var payload []byte
			var err error
			if fromFile != "" {
				payload, err = os.ReadFile(fromFile)
				if err != nil {
					return err
				}
			} else {
				payload, err = io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				if len(payload) == 0 {
					return fmt.Errorf("secret payload must be provided via --from-file or stdin")
				}
			}

			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}

			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}

			labelMap, err := barbican.ParseLabels(labels)
			if err != nil {
				return err
			}

			encodedPayload := payload
			if payloadContentEncoding == "base64" {
				// If the user specified base64, we assume they are providing raw bytes that need encoding,
				// OR they are providing a string that is already encoded?
				// The Barbican API expects the payload to be consistent with the encoding specified.
				// If we look at the previous implementation:
				// encodedPayload := base64.StdEncoding.EncodeToString(payload)
				// It was always base64 encoding the input.
				// If we allow 'text/plain' type, we might not want to base64 encode.
				// For now, let's keep the behavior that if encoding is base64, we encode it.
				b64 := base64.StdEncoding.EncodeToString(payload)
				encodedPayload = []byte(b64)
			}

			err = client.PutSecret(cmd.Context(), name, encodedPayload, labelMap, secretType, payloadContentEncoding)
			if err != nil {
				return err
			}
			fmt.Printf("Secret '%s' created/updated successfully\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to a file containing the secret")
	cmd.Flags().StringArrayVar(&labels, "label", []string{}, "Additional Barbican labels in key=value form")
	cmd.Flags().StringVar(&secretType, "secret-type", "opaque", "Type of the secret (e.g. opaque, passphrase)")
	cmd.Flags().StringVar(&payloadContentEncoding, "payload-encoding", "base64", "Encoding of the payload (e.g. base64)")
	return cmd
}

func newSecretsDeleteCmd() *cobra.Command {
	var (
		force bool
	)
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}
			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}

			err = client.DeleteSecret(cmd.Context(), name)
			if err != nil {
				return err
			}
			fmt.Printf("Secret '%s' deleted successfully\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force deletion of a secret")
	return cmd
}

func newSecretsSyncCmd() *cobra.Command {
	var (
		directory string
		labels    []string
		format    string
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Materialize a filtered subset of Barbican secrets onto disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName, _ := cmd.Flags().GetString("cluster")
			cfg, err := config.Load(clusterName)
			if err != nil {
				return err
			}
			client, err := barbican.NewClient(&cfg.OpenCenter.Secrets.Barbican)
			if err != nil {
				return err
			}

			labelMap, err := barbican.ParseLabels(labels)
			if err != nil {
				return err
			}

			secrets, err := client.ListSecrets(cmd.Context(), labelMap)
			if err != nil {
				return err
			}

			if directory == "" {
				return fmt.Errorf("--directory is required")
			}
			err = os.MkdirAll(directory, 0755)
			if err != nil {
				return err
			}

			for _, secret := range secrets {
				payload, err := client.GetSecret(cmd.Context(), secret.Name)
				if err != nil {
					return err
				}

				var data []byte
				switch format {
				case "json":
					data, err = json.MarshalIndent(map[string]string{"name": secret.Name, "payload": string(payload)}, "", "  ")
				case "yaml":
					data, err = yaml.Marshal(map[string]string{"name": secret.Name, "payload": string(payload)})
				default:
					data = payload
				}
				if err != nil {
					return err
				}

				fileName := fmt.Sprintf("%s.%s", secret.Name, format)
				filePath := fmt.Sprintf("%s/%s", directory, fileName)
				err = os.WriteFile(filePath, data, 0600)
				if err != nil {
					return err
				}
				fmt.Printf("Synced secret '%s' to %s\n", secret.Name, filePath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&directory, "directory", "", "Directory to sync secrets to")
	cmd.Flags().StringArrayVar(&labels, "label", []string{}, "Filter secrets by labels in key=value form")
	cmd.Flags().StringVar(&format, "format", "yaml", "Output format: yaml or json")
	return cmd
}
