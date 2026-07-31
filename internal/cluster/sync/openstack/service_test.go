package openstacksync

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncDryRunDoesNotMutateConfigOrOpenStack(t *testing.T) {
	path := writeSyncConfig(t, `
opencenter:
  cluster:
    cluster_name: prod
  services:
    loki:
      enabled: true
secrets: {}
`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	service, err := NewService(Dependencies{
		DiscoverCore:    func(context.Context, string) (CoreDiscovery, error) { return testDiscovery(), nil },
		EnsureContainer: func(context.Context, ContainerRequest) error { mutated = true; return nil },
		CreateEC2Credentials: func(context.Context, EC2CredentialRequest) (EC2Credentials, error) {
			mutated = true
			return EC2Credentials{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Sync(context.Background(), Options{ConfigPath: path, OSCloud: "prod", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if mutated || string(after) != string(before) {
		t.Fatalf("dry-run mutated OpenStack=%t configChanged=%t", mutated, string(after) != string(before))
	}
	if !strings.Contains(string(result.UpdatedYAML), "CHANGEME") {
		t.Fatalf("dry-run output did not use placeholders: %s", result.UpdatedYAML)
	}
}

func TestSyncPreservesExistingS3Credentials(t *testing.T) {
	path := writeSyncConfig(t, `
opencenter:
  cluster:
    cluster_name: prod
  services:
    loki:
      enabled: true
      loki_bucket_name: existing-loki
secrets:
  loki:
    s3_access_key_id: old-access
    s3_secret_access_key: old-secret
`)
	containers := 0
	service, err := NewService(Dependencies{
		DiscoverCore:    func(context.Context, string) (CoreDiscovery, error) { return testDiscovery(), nil },
		EnsureContainer: func(context.Context, ContainerRequest) error { containers++; return nil },
		CreateEC2Credentials: func(context.Context, EC2CredentialRequest) (EC2Credentials, error) {
			t.Fatal("existing credentials must be preserved without --rotate-creds")
			return EC2Credentials{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Sync(context.Background(), Options{ConfigPath: path, OSCloud: "prod"})
	if err != nil {
		t.Fatal(err)
	}
	if containers != 1 {
		t.Fatalf("EnsureContainer calls = %d, want 1", containers)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "old-access") || !strings.Contains(string(updated), "old-secret") {
		t.Fatalf("existing credentials were not retained: %s", updated)
	}
	if got := result.CredentialActions; len(got) != 2 || got[1] != "reused existing s3 credentials for loki" {
		t.Fatalf("credential actions = %v", got)
	}
}

func TestBuildContainerAccessRulesKeepsPOSTAtContainerRoot(t *testing.T) {
	rules := buildContainerAccessRules("project", "logs")
	for _, rule := range rules {
		if rule.Method == "POST" && rule.Path != "/v1/AUTH_project/logs/**" {
			t.Fatalf("POST rule must be limited to the container path: %#v", rule)
		}
	}
}

func writeSyncConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cluster.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func testDiscovery() CoreDiscovery {
	return CoreDiscovery{
		AuthURL: "https://identity.example/v3", Region: "RegionOne", ProjectID: "project-id", ProjectName: "project",
		ApplicationCredentialID: "app-id", ApplicationCredentialSecret: "app-secret", UserDomainName: "Default", ProjectDomainName: "Default",
		ImageID: "linux", ImageIDWindows: "windows", ExternalNetworkID: "external", FloatingIPPool: "external", AvailabilityZones: []string{"nova"},
		InternalSubnets: []Subnet{{ID: "subnet", Name: "internal"}}, SwiftEndpoint: "https://swift.example/v1/AUTH_project",
	}
}
