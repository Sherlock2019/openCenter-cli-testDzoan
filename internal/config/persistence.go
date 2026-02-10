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
	stderrors "errors"
	"strings"
	"sync"

	corePaths "github.com/rackerlabs/opencenter-cli/internal/core/paths"

	"github.com/rackerlabs/opencenter-cli/internal/core/validation/validators"
	utilErrors "github.com/rackerlabs/opencenter-cli/internal/util/errors"
	utilFs "github.com/rackerlabs/opencenter-cli/internal/util/fs"
)

// globalManager is a singleton ConfigurationManager for backward compatibility
var (
	globalManager     *ConfigurationManager
	globalManagerOnce sync.Once
	globalManagerErr  error
	globalFileSystem  utilFs.FileSystem
	fileSystemOnce    sync.Once
)

// getGlobalFileSystem returns a singleton FileSystem instance
func getGlobalFileSystem() utilFs.FileSystem {
	fileSystemOnce.Do(func() {
		errorHandler := utilErrors.NewDefaultErrorHandlerWithoutMasking()
		globalFileSystem = utilFs.NewDefaultFileSystem(errorHandler)
	})
	return globalFileSystem
}

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

		// Validate both parts using ValidationEngine
		ctx := context.Background()
		validator := validators.NewClusterNameValidator()

		result, err := validator.Validate(ctx, organization)
		if err != nil {
			return "", "", fmt.Errorf("organization name validation failed: %w", err)
		}
		if !result.Valid {
			return "", "", fmt.Errorf("invalid organization name: %s", result.Errors[0].Message)
		}

		result, err = validator.Validate(ctx, clusterName)
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
// It looks for cluster directories within the configured clustersDir.
// It supports both organization-based and legacy directory structures.
//
// Deprecated: Use internal/core/paths.PathResolver for cluster discovery instead.
// This function will be removed in v2.0.0.
// Migration: Use pathResolver methods for discovering clusters.
//
// Outputs:
//   - []string: A list of cluster names.
//   - error: An error if the directory cannot be read.
func List() ([]string, error) {
	logDeprecationWarning(
		"config.List()",
		"internal/core/paths.PathResolver",
		"v2.0.0",
	)
	dir, err := ResolveConfigDir()
	if err != nil {
		Debugf("List: failed to resolve config directory: %v", err)
		return nil, fmt.Errorf("failed to resolve configuration directory: %w", err)
	}
	Debugf("List: resolved config directory: %s", dir)

	// Load CLI configuration to get the configured clustersDir
	configManager, err := NewConfigManager("")
	if err != nil {
		Debugf("List: failed to load CLI config manager: %v", err)
		// Fall back to default behavior if CLI config can't be loaded
	}

	var clustersDir string
	if configManager != nil {
		clustersDir = configManager.GetConfig().Paths.ClustersDir
		Debugf("List: using clustersDir from CLI config: %s", clustersDir)
	} else {
		// Fallback to default
		clustersDir = filepath.Join(dir, "clusters")
		Debugf("List: using default clustersDir: %s", clustersDir)
	}

	// Expand environment variables and tilde in clustersDir
	clustersDir = corePaths.ExpandPath(clustersDir)
	Debugf("List: expanded clustersDir: %s", clustersDir)

	var names []string
	nameSet := make(map[string]bool) // Use set to avoid duplicates

	// First, check for flat YAML files in the config directory (for backward compatibility and tests)
	Debugf("List: checking for flat config files in: %s", dir)
	if flatEntries, flatErr := os.ReadDir(dir); flatErr == nil {
		for _, flatEntry := range flatEntries {
			if !flatEntry.IsDir() && strings.HasSuffix(flatEntry.Name(), ".yaml") {
				// Extract cluster name by removing .yaml extension
				clusterName := strings.TrimSuffix(flatEntry.Name(), ".yaml")
				// Skip the CLI config file itself
				if clusterName != "" && clusterName != "config" && !nameSet[clusterName] {
					Debugf("List: found flat config file: %s (cluster: %s)", flatEntry.Name(), clusterName)
					names = append(names, clusterName)
					nameSet[clusterName] = true
				}
			}
		}
	}
	Debugf("List: found %d flat config clusters", len(names))

	// Check clusters directory for legacy and organization-based structures
	Debugf("List: checking clusters directory: %s", clustersDir)
	entries, readErr := os.ReadDir(clustersDir)
	if readErr != nil {
		// If clusters directory doesn't exist, just return flat config files
		if os.IsNotExist(readErr) {
			Debugf("List: clusters directory does not exist, returning %d flat config clusters", len(names))
			// Sort lexically
			if len(names) > 1 {
				sortStrings(names)
			}
			return names, nil
		}
		Debugf("List: failed to read clusters directory: %v", readErr)
		return nil, fmt.Errorf("failed to read clusters directory: %w", readErr)
	}
	Debugf("List: found %d entries in clusters directory", len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			entryName := entry.Name()
			Debugf("List: processing directory entry: %s", entryName)

			// Check for legacy structure first: clustersDir/clusterName/.clusterName-config.yaml
			// This is for backward compatibility with old flat structure
			legacyConfigFile := filepath.Join(clustersDir, entryName, "."+entryName+"-config.yaml")
			Debugf("List: checking for legacy config file: %s", legacyConfigFile)
			if _, err := os.Stat(legacyConfigFile); err == nil {
				Debugf("List: found legacy config file for: %s", entryName)
				// Check if this is truly legacy (no infrastructure/clusters subdirs OR no applications subdirs)
				infraDir := filepath.Join(clustersDir, entryName, "infrastructure", "clusters")
				appsDir := filepath.Join(clustersDir, entryName, "applications", "overlays")
				hasInfra := false
				hasApps := false
				if _, err := os.Stat(infraDir); err == nil {
					hasInfra = true
					Debugf("List: %s has infrastructure directory", entryName)
				}
				if _, err := os.Stat(appsDir); err == nil {
					hasApps = true
					Debugf("List: %s has applications directory", entryName)
				}

				// If it has neither infrastructure nor applications subdirs, it's legacy flat structure
				if !hasInfra && !hasApps {
					Debugf("List: %s is legacy flat structure (no infra/apps dirs)", entryName)
					if !nameSet[entryName] {
						Debugf("List: adding legacy cluster: %s", entryName)
						names = append(names, entryName)
						nameSet[entryName] = true
					}
					continue // Skip organization check for this entry
				} else {
					Debugf("List: %s has subdirs (infra=%v, apps=%v), treating as organization", entryName, hasInfra, hasApps)
				}
			} else {
				Debugf("List: no legacy config file found for: %s", entryName)
			}

			// Check for organization-based structure
			// Look for clusters in: clustersDir/organization/infrastructure/clusters/<cluster>/.<cluster>-config.yaml
			orgDir := filepath.Join(clustersDir, entryName)
			infraClustersDir := filepath.Join(orgDir, "infrastructure", "clusters")
			Debugf("List: checking organization infrastructure/clusters directory: %s", infraClustersDir)

			if infraEntries, err := os.ReadDir(infraClustersDir); err == nil {
				Debugf("List: found %d entries in infrastructure/clusters directory for org: %s", len(infraEntries), entryName)
				for _, clusterEntry := range infraEntries {
					if clusterEntry.IsDir() {
						clusterName := clusterEntry.Name()
						// Check for config file at cluster directory level
						clusterConfigPath := filepath.Join(infraClustersDir, clusterName, "."+clusterName+"-config.yaml")
						Debugf("List: checking for config file: %s", clusterConfigPath)
						if _, statErr := os.Stat(clusterConfigPath); statErr == nil {
							Debugf("List: found cluster config file for: %s", clusterName)
							// Format as organization/cluster
							fullName := entryName + "/" + clusterName
							if !nameSet[fullName] {
								Debugf("List: adding organization cluster: %s", fullName)
								names = append(names, fullName)
								nameSet[fullName] = true
							} else {
								Debugf("List: skipping duplicate cluster: %s", fullName)
							}
						}
					}
				}
			} else {
				Debugf("List: infrastructure/clusters directory does not exist for org %s: %v", entryName, err)
			}

			// Also check for config files at organization level (alternative location)
			if orgFiles, err := os.ReadDir(orgDir); err == nil {
				Debugf("List: found %d files in organization directory: %s", len(orgFiles), entryName)
				for _, orgFile := range orgFiles {
					if !orgFile.IsDir() && strings.HasPrefix(orgFile.Name(), ".") && strings.HasSuffix(orgFile.Name(), "-config.yaml") {
						Debugf("List: found organization-level config file: %s", orgFile.Name())
						// Extract cluster name from .<cluster>-config.yaml
						clusterName := strings.TrimPrefix(orgFile.Name(), ".")
						clusterName = strings.TrimSuffix(clusterName, "-config.yaml")
						Debugf("List: extracted cluster name: %s from file: %s", clusterName, orgFile.Name())
						if clusterName != "" {
							// Format as organization/cluster
							fullName := entryName + "/" + clusterName
							if !nameSet[fullName] {
								Debugf("List: adding organization cluster: %s", fullName)
								names = append(names, fullName)
								nameSet[fullName] = true
							} else {
								Debugf("List: skipping duplicate cluster: %s", fullName)
							}
						} else {
							Debugf("List: skipping cluster (name is empty)")
						}
					}
				}
			} else {
				Debugf("List: failed to read organization directory %s: %v", orgDir, err)
			}
		} else {
			Debugf("List: skipping non-directory entry: %s", entry.Name())
		}
	}

	// Sort lexically
	Debugf("List: sorting %d cluster names", len(names))
	if len(names) > 1 {
		sortStrings(names)
	}
	Debugf("List: returning %d total clusters", len(names))
	for i, name := range names {
		Debugf("List: final result[%d]: %s", i, name)
	}
	return names, nil
}

// simple string sorter to avoid pulling in a larger dependency.
func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// activeClusterPath returns the absolute path to the file tracking
// the active cluster. This file stores the cluster name as plain
// text.
func activeClusterPath() (string, error) {
	dir, err := ResolveConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".active"), nil
}

// SetActive writes the given cluster name into the active marker file.
// If the name is empty, the marker file is removed.
//
// Deprecated: Use ConfigurationManager for active cluster management.
// This function will be removed in v2.0.0.
// Migration: Use configManager methods for active cluster tracking.
//
// Inputs:
//   - name: The name of the cluster to set as active.
//
// Outputs:
//   - error: An error if the file cannot be written.
func SetActive(name string) error {
	logDeprecationWarning(
		"config.SetActive()",
		"ConfigurationManager",
		"v2.0.0",
	)
	path, err := activeClusterPath()
	if err != nil {
		return err
	}
	if name == "" {
		return os.Remove(path)
	}
	
	// Use FileSystem for writing active cluster marker (atomic write)
	fileSystem := getGlobalFileSystem()
	return fileSystem.WriteFileAtomic(path, []byte(name), 0o600)
}

// GetActive reads the active cluster name with precedence:
// 1. OPENCENTER_CLUSTER environment variable (session-scoped)
// 2. Session file (if shell integration is active)
// 3. Persistent selection from marker file
//
// Deprecated: Use ConfigurationManager for active cluster retrieval.
// This function will be removed in v2.0.0.
// Migration: Use configManager methods for active cluster retrieval.
//
// Outputs:
//   - string: The active cluster name.
//   - error: An error if the file cannot be read.
func GetActive() (string, error) {
	logDeprecationWarning(
		"config.GetActive()",
		"ConfigurationManager",
		"v2.0.0",
	)
	// Priority 1: Check environment variable (highest priority)
	if cluster := os.Getenv("OPENCENTER_CLUSTER"); cluster != "" {
		return strings.TrimSpace(cluster), nil
	}

	// Priority 2: Check session file (shell integration)
	if sessionFile := os.Getenv("OPENCENTER_SESSION_FILE"); sessionFile != "" {
		fileSystem := getGlobalFileSystem()
		if data, err := fileSystem.ReadFile(sessionFile); err == nil && len(data) > 0 {
			return strings.TrimSpace(string(data)), nil
		}
	}

	// Priority 3: Fall back to persistent selection
	path, err := activeClusterPath()
	if err != nil {
		return "", err
	}
	
	fileSystem := getGlobalFileSystem()
	data, readErr := fileSystem.ReadFile(path)
	if readErr != nil {
		if os.IsNotExist(stderrors.Unwrap(readErr)) {
			return "", nil
		}
		return "", readErr
	}
	return strings.TrimSpace(string(data)), nil
}
