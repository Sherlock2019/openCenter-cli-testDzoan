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

package flags

import (
	"fmt"
	"strings"
)

// OutputFlagHandler handles --output-format, --dry-run, and --quiet flags
type OutputFlagHandler struct{}

// NewOutputFlagHandler creates a new output flag handler
func NewOutputFlagHandler() *OutputFlagHandler {
	return &OutputFlagHandler{}
}

// CanHandle returns true if this handler can process the given flag
func (h *OutputFlagHandler) CanHandle(flagName string) bool {
	return strings.HasPrefix(flagName, "output-format") ||
		strings.HasPrefix(flagName, "dry-run") ||
		strings.HasPrefix(flagName, "quiet")
}

// GetFlagType returns the type of flags this handler processes
func (h *OutputFlagHandler) GetFlagType() FlagType {
	return FlagTypeOutput
}

// ParseFlag processes a single flag and returns the parsed result
func (h *OutputFlagHandler) ParseFlag(flagName, value string) (interface{}, error) {
	switch {
	case strings.HasPrefix(flagName, "output-format"):
		return h.parseOutputFormat(value)
	case strings.HasPrefix(flagName, "dry-run"):
		return h.parseDryRun(value)
	case strings.HasPrefix(flagName, "quiet"):
		return h.parseQuiet(value)
	default:
		return nil, fmt.Errorf("unsupported output flag: %s", flagName)
	}
}

// parseOutputFormat parses the --output-format flag
func (h *OutputFlagHandler) parseOutputFormat(value string) (*OutputFormatFlag, error) {
	if value == "" {
		return nil, fmt.Errorf("empty output format value")
	}
	
	format, err := ParseOutputFormat(value)
	if err != nil {
		return nil, fmt.Errorf("invalid output format: %w", err)
	}
	
	return &OutputFormatFlag{
		Format: format,
	}, nil
}

// parseDryRun parses the --dry-run flag
func (h *OutputFlagHandler) parseDryRun(value string) (*OutputModeFlag, error) {
	// --dry-run can be a boolean flag (no value) or have a value
	if value == "" || value == "true" {
		return &OutputModeFlag{
			Mode: OutputModeDryRun,
		}, nil
	}
	
	if value == "false" {
		return &OutputModeFlag{
			Mode: OutputModeNormal,
		}, nil
	}
	
	return nil, fmt.Errorf("invalid dry-run value: %s (expected true/false or no value)", value)
}

// parseQuiet parses the --quiet flag
func (h *OutputFlagHandler) parseQuiet(value string) (*OutputModeFlag, error) {
	// --quiet can be a boolean flag (no value) or have a value
	if value == "" || value == "true" {
		return &OutputModeFlag{
			Mode: OutputModeQuiet,
		}, nil
	}
	
	if value == "false" {
		return &OutputModeFlag{
			Mode: OutputModeNormal,
		}, nil
	}
	
	return nil, fmt.Errorf("invalid quiet value: %s (expected true/false or no value)", value)
}

// OutputFormatFlag represents a parsed output format flag
type OutputFormatFlag struct {
	Format OutputFormat
}

// GetPath returns the configuration path this flag affects
func (f *OutputFormatFlag) GetPath() string {
	return "output.format"
}

// MergeIntoConfiguration applies this flag to the configuration
func (f *OutputFormatFlag) MergeIntoConfiguration(flag ParsedFlag, config map[string]interface{}) error {
	// Output format flags don't modify the configuration data itself
	// They are used by the output formatter
	return nil
}

// OutputModeFlag represents a parsed output mode flag
type OutputModeFlag struct {
	Mode OutputMode
}

// GetPath returns the configuration path this flag affects
func (f *OutputModeFlag) GetPath() string {
	return "output.mode"
}

// MergeIntoConfiguration applies this flag to the configuration
func (f *OutputModeFlag) MergeIntoConfiguration(flag ParsedFlag, config map[string]interface{}) error {
	// Output mode flags don't modify the configuration data itself
	// They are used by the output formatter
	return nil
}