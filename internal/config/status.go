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
	"fmt"
	"strings"
	"time"
)

// UpdateStatus updates the cluster's stage and status in the configuration file.
// It loads the configuration, updates the values, and saves it back.
func UpdateStatus(clusterName, stage, status string) error {
	// Use ConfigurationManager for load/save
	mgr, err := NewConfigurationManager()
	if err != nil {
		return fmt.Errorf("failed to create configuration manager: %w", err)
	}

	ctx := context.Background()
	configPath, err := resolveClusterConfigPath(ctx, mgr, clusterName)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster configuration path for status update: %w", err)
	}

	data, err := mgr.fileSystem.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read cluster configuration for status update: %w", err)
	}

	if isNativeV2ConfigData(data) {
		loader := defaultLegacyV2Loader()
		cfg, err := loader.LoadFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to load native v2 cluster configuration for status update: %w", err)
		}

		cfg.OpenCenter.Meta.Stage = stage
		cfg.OpenCenter.Meta.Status = status
		cfg.Metadata.UpdatedAt = time.Now().Format(time.RFC3339Nano)

		if err := loader.SaveToFile(cfg, configPath); err != nil {
			return fmt.Errorf("failed to save native v2 cluster configuration with new status: %w", err)
		}

		return nil
	}

	cfg, err := mgr.Load(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration for status update: %w", err)
	}

	cfg.OpenCenter.Meta.Stage = stage
	cfg.OpenCenter.Meta.Status = status

	if err := mgr.Save(context.Background(), cfg); err != nil {
		return fmt.Errorf("failed to save cluster configuration with new status: %w", err)
	}

	return nil
}

func resolveClusterConfigPath(ctx context.Context, mgr *ConfigurationManager, clusterName string) (string, error) {
	if strings.Contains(clusterName, "/") {
		parts := strings.SplitN(clusterName, "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid cluster identifier format: expected 'organization/cluster'")
		}

		clusterPaths, err := mgr.pathResolver.Resolve(ctx, parts[1], parts[0])
		if err != nil {
			return "", err
		}
		return clusterPaths.ConfigPath, nil
	}

	clusterPaths, err := mgr.pathResolver.ResolveWithFallback(ctx, clusterName)
	if err != nil {
		return "", err
	}

	return clusterPaths.ConfigPath, nil
}
