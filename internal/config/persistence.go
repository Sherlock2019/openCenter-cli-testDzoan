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

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/opencenter-cloud/opencenter-cli/internal/core/validation/validators"
)

// globalManager is a singleton ConfigurationManager for backward compatibility
var (
	globalManager     *ConfigurationManager
	globalManagerOnce sync.Once
	globalManagerErr  error
)

// getGlobalManager returns the singleton ConfigurationManager instance
func getGlobalManager() (*ConfigurationManager, error) {
	globalManagerOnce.Do(func() {
		globalManager, globalManagerErr = NewConfigurationManager()
	})
	return globalManager, globalManagerErr
}




// ResolveConfigDir resolves the configuration directory based on the OPENCENTER_CONFIG_DIR
// environment variable. If the variable is not set, it falls back to the user's
// standard config directory (e.g., ~/.config/opencenter on Linux).
// The directory is created if it does not exist.
func ResolveConfigDir() (string, error) {
	var err error
	dir := os.Getenv("OPENCENTER_CONFIG_DIR")
	if dir == "" {
		// Determine OS-specific config directory
		switch runtime.GOOS {
		case "windows":
			base := os.Getenv("APPDATA")
			if base == "" {
				base = os.Getenv("LOCALAPPDATA")
			}
			if base == "" {
				base = os.Getenv("USERPROFILE")
			}
			dir = filepath.Join(base, "opencenter")
		default:
			home, herr := os.UserHomeDir()
			if herr != nil {
				err = herr
				return "", err
			}
			dir = filepath.Join(home, ".config", "opencenter")
		}
	}
	// Ensure absolute path
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return "", err
		}
	}
	// Create directory if not exists
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		err = mkErr
		return "", err
	}
	return dir, err
}

// ParseClusterIdentifier parses a cluster identifier which can be in one of two formats:
// 1. "cluster" - just the cluster name (uses default "opencenter" organization)
// 2. "organization/cluster" - organization and cluster name
//
// Inputs:
//   - identifier: The cluster identifier to parse.
//
// Outputs:
//   - organization: The organization name (or "opencenter" if not specified).
//   - clusterName: The cluster name.
//   - error: An error if the identifier is invalid.
func ParseClusterIdentifier(identifier string) (organization string, clusterName string, err error) {
	if identifier == "" {
		return "", "", errors.New("cluster identifier cannot be empty")
	}

	// Check for organization/cluster format
	if strings.Contains(identifier, "/") {
		parts := strings.SplitN(identifier, "/", 2)
		if len(parts) != 2 {
			return "", "", errors.New("invalid cluster identifier format: expected 'organization/cluster'")
		}
		organization = parts[0]
		clusterName = parts[1]

		// Basic validation for organization (allow more flexible format than cluster names)
		if organization == "" {
			return "", "", errors.New("organization name cannot be empty")
		}

		// Validate cluster name using ValidationEngine
		ctx := context.Background()
		validator := validators.NewClusterNameValidator()

		result, err := validator.Validate(ctx, clusterName)
		if err != nil {
			return "", "", fmt.Errorf("cluster name validation failed: %w", err)
		}
		if !result.Valid {
			return "", "", fmt.Errorf("invalid cluster name: %s", result.Errors[0].Message)
		}

		return organization, clusterName, nil
	}

	// Just cluster name, use default organization
	ctx := context.Background()
	validator := validators.NewClusterNameValidator()

	result, err := validator.Validate(ctx, identifier)
	if err != nil {
		return "", "", fmt.Errorf("cluster name validation failed: %w", err)
	}
	if !result.Valid {
		return "", "", fmt.Errorf("invalid cluster name: %s", result.Errors[0].Message)
	}

	return "opencenter", identifier, nil
}










// List returns a sorted list of cluster names from the configuration directory.




