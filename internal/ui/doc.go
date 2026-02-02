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

// Package ui provides user interface utilities for the opencenter CLI.
//
// # Confirmation Prompts
//
// The package provides a testable confirmation prompt system through the
// ConfirmationPrompter interface. This allows commands to prompt users for
// confirmation while remaining testable.
//
// # Interactive Mode
//
// In interactive mode, the InteractivePrompter reads from stdin and writes
// to stdout, prompting the user for yes/no confirmation:
//
//	prompter := ui.NewInteractivePrompter(os.Stdin, os.Stdout)
//	confirmed, err := prompter.Confirm(ctx, "Are you sure?")
//	if err != nil {
//	    return err
//	}
//	if !confirmed {
//	    fmt.Println("Operation cancelled")
//	    return nil
//	}
//
// # Test Mode
//
// In test mode, the TestPrompter returns a predetermined response without
// prompting the user. This is useful for automated testing:
//
//	// Create a prompter that always confirms
//	prompter := ui.NewTestPrompter(true)
//	confirmed, err := prompter.Confirm(ctx, "Are you sure?")
//	// confirmed will always be true
//
//	// Create a prompter that always denies
//	prompter := ui.NewTestPrompter(false)
//	confirmed, err := prompter.Confirm(ctx, "Are you sure?")
//	// confirmed will always be false
//
// # Automatic Mode Selection
//
// The GetPrompter function automatically selects the appropriate prompter
// based on the environment:
//
//	testMode := os.Getenv("OPENCENTER_TEST_MODE") != ""
//	prompter := ui.GetPrompter(os.Stdin, os.Stdout, testMode)
//	confirmed, err := prompter.Confirm(ctx, "Are you sure?")
//
// # Context Support
//
// All prompters support context cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	prompter := ui.NewInteractivePrompter(os.Stdin, os.Stdout)
//	confirmed, err := prompter.Confirm(ctx, "Are you sure?")
//	if err == context.DeadlineExceeded {
//	    fmt.Println("Confirmation timed out")
//	    return err
//	}
//
// # Usage in Commands
//
// Commands should use the prompter interface to allow for testing:
//
//	func runDestroyCommand(cmd *cobra.Command, args []string) error {
//	    if !force {
//	        testMode := os.Getenv("OPENCENTER_TEST_MODE") != ""
//	        prompter := ui.GetPrompter(os.Stdin, cmd.OutOrStdout(), testMode)
//
//	        confirmed, err := prompter.Confirm(ctx, "Destroy cluster?")
//	        if err != nil {
//	            return err
//	        }
//	        if !confirmed {
//	            fmt.Fprintln(cmd.OutOrStdout(), "Operation cancelled")
//	            return nil
//	        }
//	    }
//	    // Proceed with destroy operation
//	    return nil
//	}
//
// # Error Formatting
//
// The package also provides error formatting utilities for consistent
// error messages across the CLI. See error_formatter.go for details.
package ui
