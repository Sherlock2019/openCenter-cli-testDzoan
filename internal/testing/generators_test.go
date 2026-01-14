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

package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGenerator_GenerateConfig(t *testing.T) {
	gen := NewConfigGenerator(42)

	t.Run("generates valid OpenStack config", func(t *testing.T) {
		cfg := gen.GenerateConfig("openstack")

		assert.Equal(t, "openstack", cfg.OpenCenter.Provider)
		assert.NotEmpty(t, cfg.OpenCenter.Organization)
		assert.NotEmpty(t, cfg.OpenCenter.ClusterName)
		assert.NotEmpty(t, cfg.OpenCenter.Cluster.KubernetesVersion)
		assert.Greater(t, cfg.OpenCenter.Cluster.NodeCounts.Masters, 0)
		assert.Greater(t, cfg.OpenCenter.Cluster.NodeCounts.Workers, 0)

		// Verify OpenStack-specific fields
		require.NotNil(t, cfg.OpenTofu.Cloud.OpenStack)
		assert.NotEmpty(t, cfg.OpenTofu.Cloud.OpenStack.AuthURL)
		assert.NotEmpty(t, cfg.OpenTofu.Cloud.OpenStack.Region)
		assert.NotEmpty(t, cfg.OpenTofu.Cloud.OpenStack.TenantName)
	})

	t.Run("generates valid AWS config", func(t *testing.T) {
		cfg := gen.GenerateConfig("aws")

		assert.Equal(t, "aws", cfg.OpenCenter.Provider)
		assert.NotEmpty(t, cfg.OpenCenter.Organization)
		assert.NotEmpty(t, cfg.OpenCenter.ClusterName)

		// Verify AWS-specific fields
		require.NotNil(t, cfg.OpenTofu.Cloud.AWS)
		assert.NotEmpty(t, cfg.OpenTofu.Cloud.AWS.Region)
	})

	t.Run("generates valid networking config", func(t *testing.T) {
		cfg := gen.GenerateConfig("openstack")

		assert.NotEmpty(t, cfg.Networking.PodCIDR)
		assert.NotEmpty(t, cfg.Networking.ServiceCIDR)
		assert.NotEmpty(t, cfg.Networking.DNSServers)
		assert.Greater(t, len(cfg.Networking.DNSServers), 0)
	})

	t.Run("generates valid secrets config", func(t *testing.T) {
		cfg := gen.GenerateConfig("openstack")

		assert.True(t, cfg.Secrets.SOPS.Enabled)
		assert.NotEmpty(t, cfg.Secrets.SOPS.Age.Recipient)
		assert.Contains(t, cfg.Secrets.SOPS.Age.Recipient, "age1")
	})

	t.Run("generates valid security config", func(t *testing.T) {
		cfg := gen.GenerateConfig("openstack")

		assert.NotEmpty(t, cfg.Security.SSHKeys)
		assert.Contains(t, cfg.Security.SSHKeys[0], "ssh-rsa")
	})

	t.Run("generates valid deployment config", func(t *testing.T) {
		cfg := gen.GenerateConfig("openstack")

		assert.True(t, cfg.Deployment.GitOps.Enabled)
		assert.NotEmpty(t, cfg.Deployment.GitOps.Repository)
		assert.NotEmpty(t, cfg.Deployment.GitOps.Branch)
	})
}

func TestConfigGenerator_Deterministic(t *testing.T) {
	// Same seed should produce same results
	gen1 := NewConfigGenerator(123)
	gen2 := NewConfigGenerator(123)

	cfg1 := gen1.GenerateConfig("openstack")
	cfg2 := gen2.GenerateConfig("openstack")

	assert.Equal(t, cfg1.OpenCenter.Organization, cfg2.OpenCenter.Organization)
	assert.Equal(t, cfg1.OpenCenter.ClusterName, cfg2.OpenCenter.ClusterName)
	assert.Equal(t, cfg1.OpenCenter.Cluster.KubernetesVersion, cfg2.OpenCenter.Cluster.KubernetesVersion)
}

func TestConfigGenerator_Variety(t *testing.T) {
	// Different seeds should produce different results
	gen1 := NewConfigGenerator(123)
	gen2 := NewConfigGenerator(456)

	cfg1 := gen1.GenerateConfig("openstack")
	cfg2 := gen2.GenerateConfig("openstack")

	// At least some fields should be different
	different := cfg1.OpenCenter.Organization != cfg2.OpenCenter.Organization ||
		cfg1.OpenCenter.ClusterName != cfg2.OpenCenter.ClusterName ||
		cfg1.OpenCenter.Cluster.KubernetesVersion != cfg2.OpenCenter.Cluster.KubernetesVersion

	assert.True(t, different, "Different seeds should produce different configurations")
}

func TestTemplateDataGenerator_GenerateTemplateData(t *testing.T) {
	gen := NewTemplateDataGenerator(42)

	t.Run("generates complete template data", func(t *testing.T) {
		data := gen.GenerateTemplateData()

		assert.NotEmpty(t, data["ClusterName"])
		assert.NotEmpty(t, data["Namespace"])
		assert.NotEmpty(t, data["Version"])
		assert.NotEmpty(t, data["Image"])
		assert.NotEmpty(t, data["Environment"])

		// Verify numeric fields
		replicas, ok := data["Replicas"].(int)
		assert.True(t, ok)
		assert.Greater(t, replicas, 0)

		port, ok := data["Port"].(int)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, port, 8000)
		assert.Less(t, port, 9000)
	})

	t.Run("generates valid labels", func(t *testing.T) {
		data := gen.GenerateTemplateData()

		labels, ok := data["Labels"].(map[string]string)
		require.True(t, ok)
		assert.NotEmpty(t, labels["app"])
		assert.NotEmpty(t, labels["environment"])
		assert.NotEmpty(t, labels["version"])
	})

	t.Run("generates valid annotations", func(t *testing.T) {
		data := gen.GenerateTemplateData()

		annotations, ok := data["Annotations"].(map[string]string)
		require.True(t, ok)
		assert.NotEmpty(t, annotations["description"])
		assert.NotEmpty(t, annotations["timestamp"])
	})

	t.Run("generates valid resources", func(t *testing.T) {
		data := gen.GenerateTemplateData()

		resources, ok := data["Resources"].(map[string]interface{})
		require.True(t, ok)

		requests, ok := resources["requests"].(map[string]string)
		require.True(t, ok)
		assert.NotEmpty(t, requests["cpu"])
		assert.NotEmpty(t, requests["memory"])

		limits, ok := resources["limits"].(map[string]string)
		require.True(t, ok)
		assert.NotEmpty(t, limits["cpu"])
		assert.NotEmpty(t, limits["memory"])
	})
}

func TestServiceDataGenerator_GenerateServiceDefinition(t *testing.T) {
	gen := NewServiceDataGenerator(42)

	t.Run("generates complete service definition", func(t *testing.T) {
		svc := gen.GenerateServiceDefinition()

		assert.NotEmpty(t, svc["name"])
		assert.NotEmpty(t, svc["type"])
		assert.NotEmpty(t, svc["version"])

		enabled, ok := svc["enabled"].(bool)
		assert.True(t, ok)
		_ = enabled // enabled can be true or false

		dependencies, ok := svc["dependencies"].([]string)
		assert.True(t, ok)
		_ = dependencies // dependencies can be empty or have items
	})

	t.Run("generates valid service config", func(t *testing.T) {
		svc := gen.GenerateServiceDefinition()

		config, ok := svc["config"].(map[string]interface{})
		require.True(t, ok)

		replicas, ok := config["replicas"].(int)
		assert.True(t, ok)
		assert.Greater(t, replicas, 0)

		assert.NotEmpty(t, config["namespace"])

		resources, ok := config["resources"].(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, resources["cpu"])
		assert.NotEmpty(t, resources["memory"])
	})
}

func TestGitOpsDataGenerator_GenerateGitOpsConfig(t *testing.T) {
	gen := NewGitOpsDataGenerator(42)

	t.Run("generates complete GitOps config", func(t *testing.T) {
		cfg := gen.GenerateGitOpsConfig()

		enabled, ok := cfg["enabled"].(bool)
		assert.True(t, ok)
		assert.True(t, enabled)

		assert.NotEmpty(t, cfg["repository"])
		assert.NotEmpty(t, cfg["branch"])
		assert.NotEmpty(t, cfg["path"])

		sync, ok := cfg["sync"].(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, sync["interval"])
		assert.NotEmpty(t, sync["timeout"])
	})

	t.Run("generates valid repository URL", func(t *testing.T) {
		cfg := gen.GenerateGitOpsConfig()

		repo, ok := cfg["repository"].(string)
		require.True(t, ok)
		assert.Contains(t, repo, "github.com")
		assert.Contains(t, repo, ".git")
	})
}

func TestGenerators_Realistic(t *testing.T) {
	t.Run("config generator produces realistic values", func(t *testing.T) {
		gen := NewConfigGenerator(42)
		cfg := gen.GenerateConfig("openstack")

		// Kubernetes version should be realistic
		assert.Regexp(t, `^1\.\d+\.\d+$`, cfg.OpenCenter.Cluster.KubernetesVersion)

		// Node counts should be reasonable
		assert.LessOrEqual(t, cfg.OpenCenter.Cluster.NodeCounts.Masters, 5)
		assert.LessOrEqual(t, cfg.OpenCenter.Cluster.NodeCounts.Workers, 10)

		// DNS servers should be valid IPs
		for _, dns := range cfg.Networking.DNSServers {
			assert.Regexp(t, `^\d+\.\d+\.\d+\.\d+$`, dns)
		}

		// Age recipient should have correct format
		assert.Regexp(t, `^age1[a-z0-9]{58}$`, cfg.Secrets.SOPS.Age.Recipient)
	})

	t.Run("template generator produces realistic values", func(t *testing.T) {
		gen := NewTemplateDataGenerator(42)
		data := gen.GenerateTemplateData()

		// Version should follow semver pattern
		version, ok := data["Version"].(string)
		require.True(t, ok)
		assert.Regexp(t, `^v\d+\.\d+\.\d+$`, version)

		// Port should be in valid range
		port, ok := data["Port"].(int)
		require.True(t, ok)
		assert.GreaterOrEqual(t, port, 8000)
		assert.Less(t, port, 9000)

		// Image should have valid format
		image, ok := data["Image"].(string)
		require.True(t, ok)
		assert.Contains(t, image, ":")
	})
}

func BenchmarkConfigGenerator(b *testing.B) {
	gen := NewConfigGenerator(42)

	b.Run("GenerateConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = gen.GenerateConfig("openstack")
		}
	})
}

func BenchmarkTemplateDataGenerator(b *testing.B) {
	gen := NewTemplateDataGenerator(42)

	b.Run("GenerateTemplateData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = gen.GenerateTemplateData()
		}
	})
}

func BenchmarkServiceDataGenerator(b *testing.B) {
	gen := NewServiceDataGenerator(42)

	b.Run("GenerateServiceDefinition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = gen.GenerateServiceDefinition()
		}
	})
}

func BenchmarkGitOpsDataGenerator(b *testing.B) {
	gen := NewGitOpsDataGenerator(42)

	b.Run("GenerateGitOpsConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = gen.GenerateGitOpsConfig()
		}
	})
}
