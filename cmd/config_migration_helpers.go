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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rackerlabs/opencenter-cli/internal/config"
	"github.com/rackerlabs/opencenter-cli/internal/di"
)

var (
	// Global configuration manager instance (lazy-initialized)
	globalConfigManager *config.ConfigurationManager
	configManagerOnce   sync.Once
	configManagerErr    error
)

// getConfigManager returns the global ConfigurationManager instance.
// It initializes the manager on first call and reuses it for subsequent calls.
func getConfigManager() (*config.ConfigurationManager, error) {
	configManagerOnce.Do(func() {
		globalConfigManager, configManagerErr = config.NewConfigurationManager()
	})
	return globalConfigManager, configManagerErr
}

// loadConfig loads a cluster configuration using the new ConfigurationManager.
func loadConfig(ctx context.Context, name string) (config.Config, error) {
	manager, err := getConfigManager()
	if err != nil {
		return config.Config{}, err
	}

	cfg, err := manager.Load(ctx, name)
	if err != nil {
		return config.Config{}, err
	}

	return *cfg, nil
}

// saveConfig saves a cluster configuration using the new ConfigurationManager.
func saveConfig(ctx context.Context, cfg config.Config) error {
	manager, err := getConfigManager()
	if err != nil {
		return err
	}

	return manager.Save(ctx, &cfg)
}

// listClusters lists all cluster configurations using the new ConfigurationManager.
// This is a migration helper to replace config.List() calls.
func listClusters(ctx context.Context) ([]string, error) {
	manager, err := getConfigManager()
	if err != nil {
		return nil, err
	}

	return manager.List(ctx)
}

// getActiveCluster returns the active cluster name.
// This continues to use the legacy config.GetActive() for now since it's
// a simple file read operation that doesn't need the full manager.
func getActiveCluster() (string, error) {
	return config.GetActive()
}

// getConfigPath returns the configuration file path for a cluster.
// This replaces the deprecated config.ConfigPath() function.
func getConfigPath(ctx context.Context, name, organization string) (string, error) {
	manager, err := getConfigManager()
	if err != nil {
		return "", err
	}
	
	// Load the config to get the organization if not provided
	if organization == "" {
		cfg, err := manager.Load(ctx, name)
		if err != nil {
			return "", err
		}
		organization = cfg.OpenCenter.Meta.Organization
	}
	
	// Get base directory (same logic as root.go)
	baseDir := os.Getenv("OPENCENTER_CONFIG_DIR")
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".config", "opencenter", "clusters")
	}
	
	// Get the path resolver
	pathResolver, err := di.ProvidePathResolver(baseDir)
	if err != nil {
		return "", err
	}
	
	clusterPaths, err := pathResolver.Resolve(ctx, name, organization)
	if err != nil {
		return "", err
	}
	
	return clusterPaths.ConfigPath, nil
}
