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
	"fmt"
	"math/rand"
	"time"

	"github.com/rackerlabs/openCenter-cli/internal/config"
)

// ConfigGenerator provides methods for generating realistic test configurations
type ConfigGenerator struct {
	rand *rand.Rand
}

// NewConfigGenerator creates a new configuration generator with a seeded random source
func NewConfigGenerator(seed int64) *ConfigGenerator {
	return &ConfigGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateConfig creates a realistic cluster configuration with random but valid values
func (g *ConfigGenerator) GenerateConfig(provider string) config.Config {
	cfg := config.Config{
		OpenCenter: g.generateOpenCenter(provider),
		OpenTofu:   g.generateOpenTofu(provider),
		Secrets:    g.generateSecrets(),
		Networking: g.generateNetworking(provider),
		Security:   g.generateSecurity(),
		Deployment: g.generateDeployment(),
		Overrides:  make(map[string]any),
	}
	return cfg
}

// generateOpenCenter creates a realistic OpenCenter configuration section
func (g *ConfigGenerator) generateOpenCenter(provider string) config.SimplifiedOpenCenter {
	clusterName := g.randomClusterName()
	region := g.randomRegion()
	
	return config.SimplifiedOpenCenter{
		Meta: config.ClusterMeta{
			Name:         clusterName,
			Env:          g.randomEnvironment(),
			Region:       region,
			Status:       "pending",
			Organization: g.randomOrganization(),
		},
		Infrastructure: config.Infrastructure{
			Provider: provider,
			SSHUser:  "ubuntu",
			Cloud:    g.generateCloudConfig(provider),
		},
		Cluster: config.ClusterConfig{
			ClusterName: clusterName,
			Kubernetes: config.KubernetesConfig{
				Version:        g.randomKubernetesVersion(),
				SubnetPods:     g.randomCIDR("10.42.0.0/16"),
				SubnetServices: g.randomCIDR("10.43.0.0/16"),
			},
		},
		Services: g.generateServices(),
	}
}

// generateCloudConfig creates cloud-specific configuration
func (g *ConfigGenerator) generateCloudConfig(provider string) config.CloudConfig {
	cloudConfig := config.CloudConfig{}

	switch provider {
	case "openstack":
		cloudConfig.OpenStack = config.SimplifiedOpenStackCloud{
			AuthURL:    "https://identity.example.com/v3",
			Region:     g.randomRegion(),
			TenantName: g.randomTenantName(),
		}
	case "aws":
		cloudConfig.AWS = config.SimplifiedAWSCloud{
			Region: g.randomAWSRegion(),
		}
	}

	return cloudConfig
}

// generateOpenTofu creates a realistic OpenTofu configuration section
func (g *ConfigGenerator) generateOpenTofu(provider string) config.SimplifiedOpenTofu {
	return config.SimplifiedOpenTofu{
		Backend: config.BackendConfig{
			Type: g.randomBackendType(),
		},
	}
}

// generateSecrets creates a realistic secrets configuration
func (g *ConfigGenerator) generateSecrets() config.Secrets {
	return config.Secrets{
		SopsAgeKeyFile: "/path/to/age/key.txt",
		SSHKey: config.SSHKey{
			Private: "/path/to/ssh/private",
			Public:  g.randomSSHKey(),
			Cypher:  "rsa",
		},
	}
}

// generateNetworking creates a realistic networking configuration
func (g *ConfigGenerator) generateNetworking(provider string) config.Networking {
	networking := config.Networking{
		SubnetNodes:          g.randomCIDR("10.0.1.0/24"),
		SubnetPods:           g.randomCIDR("10.42.0.0/16"),
		SubnetServices:       g.randomCIDR("10.43.0.0/16"),
		DNSNameservers:       g.randomDNSServers(),
		LoadbalancerProvider: "ovn",
		VRRPEnabled:          false,
	}

	return networking
}

// generateSecurity creates a realistic security configuration
func (g *ConfigGenerator) generateSecurity() config.Security {
	return config.Security{
		K8sHardening:       true,
		OSHardening:        true,
		KubeletRotateCerts: true,
	}
}

// generateDeployment creates a realistic deployment configuration
func (g *ConfigGenerator) generateDeployment() config.Deployment {
	return config.Deployment{
		AutoDeploy: g.randomBool(),
	}
}

// generateServices creates a realistic services configuration
func (g *ConfigGenerator) generateServices() config.ServiceMap {
	services := make(config.ServiceMap)

	// Randomly enable some services
	if g.randomBool() {
		services["cert-manager"] = map[string]interface{}{
			"enabled": true,
		}
	}

	if g.randomBool() {
		services["prometheus-stack"] = map[string]interface{}{
			"enabled": true,
		}
	}

	if g.randomBool() {
		services["loki"] = map[string]interface{}{
			"enabled": true,
		}
	}

	if g.randomBool() {
		services["velero"] = map[string]interface{}{
			"enabled": true,
		}
	}

	return services
}

// Random value generators

func (g *ConfigGenerator) randomOrganization() string {
	orgs := []string{"acme-corp", "example-inc", "test-org", "demo-company", "sample-enterprise"}
	return orgs[g.rand.Intn(len(orgs))]
}

func (g *ConfigGenerator) randomClusterName() string {
	prefixes := []string{"prod", "staging", "dev", "test", "qa"}
	suffixes := []string{"cluster", "k8s", "kube", "platform"}
	return fmt.Sprintf("%s-%s-%d", prefixes[g.rand.Intn(len(prefixes))], suffixes[g.rand.Intn(len(suffixes))], g.rand.Intn(100))
}

func (g *ConfigGenerator) randomKubernetesVersion() string {
	versions := []string{"1.28.0", "1.29.0", "1.30.0", "1.31.0"}
	return versions[g.rand.Intn(len(versions))]
}

func (g *ConfigGenerator) randomInt(min, max int) int {
	return min + g.rand.Intn(max-min+1)
}

func (g *ConfigGenerator) randomBackendType() string {
	backends := []string{"local", "s3", "gcs", "azurerm"}
	return backends[g.rand.Intn(len(backends))]
}

func (g *ConfigGenerator) randomRegion() string {
	regions := []string{"RegionOne", "RegionTwo", "us-east", "us-west", "eu-central"}
	return regions[g.rand.Intn(len(regions))]
}

func (g *ConfigGenerator) randomAWSRegion() string {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	return regions[g.rand.Intn(len(regions))]
}

func (g *ConfigGenerator) randomTenantName() string {
	return fmt.Sprintf("tenant-%d", g.rand.Intn(1000))
}

func (g *ConfigGenerator) randomAgeRecipient() string {
	// Generate a realistic-looking Age public key
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	key := make([]byte, 58)
	for i := range key {
		key[i] = chars[g.rand.Intn(len(chars))]
	}
	return fmt.Sprintf("age1%s", string(key))
}

func (g *ConfigGenerator) randomCIDR(defaultCIDR string) string {
	// For simplicity, return default or slightly modified CIDR
	if g.randomBool() {
		return defaultCIDR
	}
	// Generate a random /16 or /12 CIDR
	return fmt.Sprintf("10.%d.0.0/%d", g.rand.Intn(256), 12+g.rand.Intn(5)*4)
}

func (g *ConfigGenerator) randomDNSServers() []string {
	servers := [][]string{
		{"8.8.8.8", "8.8.4.4"},
		{"1.1.1.1", "1.0.0.1"},
		{"9.9.9.9", "149.112.112.112"},
	}
	return servers[g.rand.Intn(len(servers))]
}

func (g *ConfigGenerator) randomSSHKey() string {
	return fmt.Sprintf("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC%s user@host", g.randomString("", 64))
}

func (g *ConfigGenerator) randomGitRepository() string {
	return fmt.Sprintf("https://github.com/%s/%s-gitops.git", g.randomOrganization(), g.randomClusterName())
}

func (g *ConfigGenerator) randomBranch() string {
	branches := []string{"main", "master", "develop", "staging"}
	return branches[g.rand.Intn(len(branches))]
}

func (g *ConfigGenerator) randomBool() bool {
	return g.rand.Intn(2) == 1
}

func (g *ConfigGenerator) randomString(prefix string, length int) string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[g.rand.Intn(len(chars))]
	}
	return prefix + string(result)
}

// TemplateDataGenerator provides methods for generating realistic template test data
type TemplateDataGenerator struct {
	rand *rand.Rand
}

// NewTemplateDataGenerator creates a new template data generator
func NewTemplateDataGenerator(seed int64) *TemplateDataGenerator {
	return &TemplateDataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateTemplateData creates realistic data for template rendering tests
func (g *TemplateDataGenerator) GenerateTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"ClusterName":  g.randomClusterName(),
		"Namespace":    g.randomNamespace(),
		"Version":      g.randomVersion(),
		"Replicas":     g.randomReplicas(),
		"Image":        g.randomImage(),
		"Port":         g.randomPort(),
		"Environment":  g.randomEnvironment(),
		"Labels":       g.randomLabels(),
		"Annotations":  g.randomAnnotations(),
		"Resources":    g.randomResources(),
	}
}

func (g *TemplateDataGenerator) randomClusterName() string {
	return fmt.Sprintf("cluster-%d", g.rand.Intn(100))
}

func (g *TemplateDataGenerator) randomNamespace() string {
	namespaces := []string{"default", "kube-system", "monitoring", "logging", "ingress"}
	return namespaces[g.rand.Intn(len(namespaces))]
}

func (g *TemplateDataGenerator) randomVersion() string {
	return fmt.Sprintf("v%d.%d.%d", g.rand.Intn(3), g.rand.Intn(20), g.rand.Intn(10))
}

func (g *TemplateDataGenerator) randomReplicas() int {
	return 1 + g.rand.Intn(5)
}

func (g *TemplateDataGenerator) randomImage() string {
	images := []string{
		"nginx:latest",
		"redis:alpine",
		"postgres:14",
		"mysql:8.0",
		"mongodb:latest",
	}
	return images[g.rand.Intn(len(images))]
}

func (g *TemplateDataGenerator) randomPort() int {
	return 8000 + g.rand.Intn(1000)
}

func (g *TemplateDataGenerator) randomEnvironment() string {
	envs := []string{"development", "staging", "production", "test"}
	return envs[g.rand.Intn(len(envs))]
}

func (g *TemplateDataGenerator) randomLabels() map[string]string {
	return map[string]string{
		"app":         fmt.Sprintf("app-%d", g.rand.Intn(100)),
		"environment": g.randomEnvironment(),
		"version":     g.randomVersion(),
	}
}

func (g *TemplateDataGenerator) randomAnnotations() map[string]string {
	return map[string]string{
		"description": fmt.Sprintf("Generated test annotation %d", g.rand.Intn(1000)),
		"timestamp":   time.Now().Format(time.RFC3339),
	}
}

func (g *TemplateDataGenerator) randomResources() map[string]interface{} {
	return map[string]interface{}{
		"requests": map[string]string{
			"cpu":    fmt.Sprintf("%dm", 100+g.rand.Intn(900)),
			"memory": fmt.Sprintf("%dMi", 128+g.rand.Intn(896)),
		},
		"limits": map[string]string{
			"cpu":    fmt.Sprintf("%dm", 500+g.rand.Intn(1500)),
			"memory": fmt.Sprintf("%dMi", 256+g.rand.Intn(1792)),
		},
	}
}

// ServiceDataGenerator provides methods for generating realistic service test data
type ServiceDataGenerator struct {
	rand *rand.Rand
}

// NewServiceDataGenerator creates a new service data generator
func NewServiceDataGenerator(seed int64) *ServiceDataGenerator {
	return &ServiceDataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateServiceDefinition creates a realistic service definition for testing
func (g *ServiceDataGenerator) GenerateServiceDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":         g.randomServiceName(),
		"type":         g.randomServiceType(),
		"enabled":      g.randomBool(),
		"version":      g.randomVersion(),
		"dependencies": g.randomDependencies(),
		"config":       g.randomServiceConfig(),
	}
}

func (g *ServiceDataGenerator) randomServiceName() string {
	services := []string{
		"cert-manager",
		"prometheus-stack",
		"loki",
		"velero",
		"keycloak",
		"headlamp",
		"weave-gitops",
	}
	return services[g.rand.Intn(len(services))]
}

func (g *ServiceDataGenerator) randomServiceType() string {
	types := []string{"monitoring", "security", "backup", "identity", "ui", "gitops"}
	return types[g.rand.Intn(len(types))]
}

func (g *ServiceDataGenerator) randomBool() bool {
	return g.rand.Intn(2) == 1
}

func (g *ServiceDataGenerator) randomVersion() string {
	return fmt.Sprintf("v%d.%d.%d", g.rand.Intn(3), g.rand.Intn(20), g.rand.Intn(10))
}

func (g *ServiceDataGenerator) randomDependencies() []string {
	allDeps := []string{"cert-manager", "prometheus-stack", "loki", "storage-class"}
	count := g.rand.Intn(3)
	deps := make([]string, 0, count)
	for i := 0; i < count; i++ {
		deps = append(deps, allDeps[g.rand.Intn(len(allDeps))])
	}
	return deps
}

func (g *ServiceDataGenerator) randomServiceConfig() map[string]interface{} {
	return map[string]interface{}{
		"replicas":  1 + g.rand.Intn(3),
		"namespace": fmt.Sprintf("service-%d", g.rand.Intn(100)),
		"resources": map[string]interface{}{
			"cpu":    fmt.Sprintf("%dm", 100+g.rand.Intn(400)),
			"memory": fmt.Sprintf("%dMi", 128+g.rand.Intn(384)),
		},
	}
}

// GitOpsDataGenerator provides methods for generating realistic GitOps test data
type GitOpsDataGenerator struct {
	rand *rand.Rand
}

// NewGitOpsDataGenerator creates a new GitOps data generator
func NewGitOpsDataGenerator(seed int64) *GitOpsDataGenerator {
	return &GitOpsDataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateGitOpsConfig creates a realistic GitOps configuration for testing
func (g *GitOpsDataGenerator) GenerateGitOpsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    true,
		"repository": g.randomRepository(),
		"branch":     g.randomBranch(),
		"path":       g.randomPath(),
		"sync": map[string]interface{}{
			"interval": g.randomInterval(),
			"timeout":  g.randomTimeout(),
		},
	}
}

func (g *GitOpsDataGenerator) randomRepository() string {
	orgs := []string{"acme", "example", "test", "demo"}
	repos := []string{"gitops", "k8s-config", "cluster-config", "infrastructure"}
	return fmt.Sprintf("https://github.com/%s/%s.git",
		orgs[g.rand.Intn(len(orgs))],
		repos[g.rand.Intn(len(repos))])
}

func (g *GitOpsDataGenerator) randomBranch() string {
	branches := []string{"main", "master", "develop", "staging", "production"}
	return branches[g.rand.Intn(len(branches))]
}

func (g *GitOpsDataGenerator) randomPath() string {
	paths := []string{"clusters", "infrastructure", "applications", "services"}
	return fmt.Sprintf("%s/cluster-%d", paths[g.rand.Intn(len(paths))], g.rand.Intn(100))
}

func (g *GitOpsDataGenerator) randomInterval() string {
	intervals := []string{"1m", "5m", "10m", "30m", "1h"}
	return intervals[g.rand.Intn(len(intervals))]
}

func (g *GitOpsDataGenerator) randomTimeout() string {
	timeouts := []string{"30s", "1m", "2m", "5m"}
	return timeouts[g.rand.Intn(len(timeouts))]
}
