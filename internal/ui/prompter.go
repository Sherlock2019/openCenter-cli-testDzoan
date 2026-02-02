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

package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// ConfirmationPrompter provides an interface for prompting users for confirmation.
// This abstraction allows for different implementations in interactive and test modes.
type ConfirmationPrompter interface {
	// Confirm prompts the user with a message and returns true if they confirm.
	// The context can be used to cancel the prompt operation.
	Confirm(ctx context.Context, message string) (bool, error)
}

// InteractivePrompter provides interactive confirmation prompts for users.
// It reads from stdin and writes to the provided output writer.
type InteractivePrompter struct {
	input  io.Reader
	output io.Writer
}

// NewInteractivePrompter creates a new interactive prompter.
// If input or output are nil, they default to os.Stdin and os.Stdout.
func NewInteractivePrompter(input io.Reader, output io.Writer) *InteractivePrompter {
	return &InteractivePrompter{
		input:  input,
		output: output,
	}
}

// Confirm prompts the user with a message and waits for yes/no input.
// Returns true if the user confirms (y/yes), false otherwise.
// The prompt is case-insensitive and accepts y/yes for confirmation.
func (p *InteractivePrompter) Confirm(ctx context.Context, message string) (bool, error) {
	// Write the prompt message
	if _, err := p.output.Write([]byte(message + " (y/n): ")); err != nil {
		return false, err
	}

	// Create a channel to receive the response
	responseCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Read response in a goroutine to support context cancellation
	go func() {
		var response string
		_, err := fmt.Fscanln(p.input, &response)
		if err != nil {
			errCh <- err
			return
		}
		responseCh <- response
	}()

	// Wait for response or context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case err := <-errCh:
		return false, err
	case response := <-responseCh:
		// Normalize response to lowercase
		response = strings.ToLower(strings.TrimSpace(response))
		return response == "y" || response == "yes", nil
	}
}

// TestPrompter provides a non-interactive prompter for testing.
// It returns a predetermined response without prompting the user.
type TestPrompter struct {
	response bool
}

// NewTestPrompter creates a new test prompter with a predetermined response.
// If response is true, Confirm() will always return true.
// If response is false, Confirm() will always return false.
func NewTestPrompter(response bool) *TestPrompter {
	return &TestPrompter{
		response: response,
	}
}

// Confirm returns the predetermined response without prompting.
// This is useful for testing code that requires user confirmation.
func (p *TestPrompter) Confirm(ctx context.Context, message string) (bool, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return p.response, nil
	}
}

// GetPrompter returns an appropriate ConfirmationPrompter based on the environment.
// In test mode (when OPENCENTER_TEST_MODE is set), it returns a TestPrompter.
// Otherwise, it returns an InteractivePrompter using the provided input and output.
func GetPrompter(input io.Reader, output io.Writer, testMode bool) ConfirmationPrompter {
	if testMode {
		// In test mode, default to confirming (true)
		// Tests can override this by creating their own TestPrompter
		return NewTestPrompter(true)
	}
	return NewInteractivePrompter(input, output)
}
