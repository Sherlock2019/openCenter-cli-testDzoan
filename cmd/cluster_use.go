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

	"github.com/rackerlabs/openCenter-cli/internal/config"
	"github.com/spf13/cobra"
)

func newClusterUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <cluster-name>",
		Short: "Switch to a cluster for the current shell session",
		Long: `Switch the active cluster context for the current shell session.

This command sets the OPENCENTER_CLUSTER environment variable and updates
the session file, affecting only the current terminal session.

Requires shell integration to be enabled:
    eval "$(opencenter shell-init)"

If shell integration is not active, this command will fall back to setting
the persistent cluster selection (same as 'opencenter cluster select').`,
		Example: `  # Switch to a cluster in current session
  opencenter cluster use prod-cluster

  # Switch to a cluster with organization
  opencenter cluster use myorg/prod-cluster`,
		Args: cobra.ExactArgs(1),
		RunE: runClusterUse,
	}

	return cmd
}

func runClusterUse(cmd *cobra.Command, args []string) error {
	clusterName := args[0]

	// Validate cluster exists by trying to load its config
	_, err := config.Load(clusterName)
	if err != nil {
		return fmt.Errorf("cluster '%s' not found: %w", clusterName, err)
	}

	// Check if shell integration is active
	sessionFile := os.Getenv("OPENCENTER_SESSION_FILE")
	if sessionFile == "" {
		// No shell integration - fall back to persistent selection
		fmt.Fprintf(os.Stderr, "⚠️  Shell integration not detected. Setting persistent cluster selection.\n")
		fmt.Fprintf(os.Stderr, "💡 To enable session-scoped selection, run: eval \"$(opencenter shell-init)\"\n\n")

		if err := config.SetActive(clusterName); err != nil {
			return fmt.Errorf("failed to set active cluster: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Switched to cluster: %s (persistent)\n", clusterName)
		return nil
	}

	// Shell integration is active - use session file
	if err := os.WriteFile(sessionFile, []byte(clusterName), 0600); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Output for shell to evaluate (export command)
	fmt.Fprintf(cmd.OutOrStdout(), "export OPENCENTER_CLUSTER=%s\n", clusterName)
	fmt.Fprintf(os.Stderr, "Switched to cluster: %s (session)\n", clusterName)

	return nil
}
