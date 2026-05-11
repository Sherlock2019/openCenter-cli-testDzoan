package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	configservices "github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

func TestRenderClusterAppsCertManagerCloudflare(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("cloudflare-guided")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Cluster.ClusterFQDN = "cloudflare-guided.sjc3.k8s.opencenter.cloud"

	// Use new map-based credentials
	cfg.Secrets.CertManager.Cloudflare = map[string]v2.CertManagerCloudflareCredential{
		"prod": {
			Enabled:  true,
			APIToken: "cf-token",
			DNSZones: []string{"cloudflare-guided.sjc3.k8s.opencenter.cloud"},
		},
	}

	certManager := cfg.OpenCenter.Services["cert-manager"].(*configservices.CertManagerConfig)
	certManager.DNSProvider = "cloudflare"
	certManager.Email = "ops@example.com"

	if err := RenderClusterApps(cfg); err != nil {
		t.Fatalf("RenderClusterApps() error = %v", err)
	}

	base := filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "cert-manager")

	kustomization := mustReadFile(t, filepath.Join(base, "kustomization.yaml"))
	if !strings.Contains(kustomization, "./opencenter-cloudflare-credentials-secret-prod.yaml") {
		t.Fatalf("expected Cloudflare secret in kustomization:\n%s", kustomization)
	}
	if strings.Contains(kustomization, "./opencenter-aws-credentials-secret") {
		t.Fatalf("did not expect Route53 AWS secret in Cloudflare kustomization:\n%s", kustomization)
	}

	issuer := mustReadFile(t, filepath.Join(base, "letsencrypt-prod-issuer.yaml"))
	if !strings.Contains(issuer, "cloudflare:") {
		t.Fatalf("expected Cloudflare solver in issuer:\n%s", issuer)
	}
	if !strings.Contains(issuer, "apiTokenSecretRef") {
		t.Fatalf("expected Cloudflare token secret ref in issuer:\n%s", issuer)
	}

	secret := mustReadFile(t, filepath.Join(base, "opencenter-cloudflare-credentials-secret-prod.yaml"))
	if !strings.Contains(secret, "api-token: cf-token") {
		t.Fatalf("expected Cloudflare API token in secret:\n%s", secret)
	}
}

func TestRenderClusterAppsCertManagerDesignate(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("designate-guided")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Cluster.ClusterFQDN = "designate-guided.sjc3.k8s.opencenter.cloud"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL = "https://identity.api.example.com/v3"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region = "sjc3"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ProjectID = "project-123"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ProjectName = "project-name"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Domain = "rackspace"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialID = "app-cred-id"
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialSecret = "app-cred-secret"

	certManager := cfg.OpenCenter.Services["cert-manager"].(*configservices.CertManagerConfig)
	certManager.DNSProvider = "designate"
	certManager.Email = "ops@example.com"

	if err := RenderClusterApps(cfg); err != nil {
		t.Fatalf("RenderClusterApps() error = %v", err)
	}

	base := filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "cert-manager")

	kustomization := mustReadFile(t, filepath.Join(base, "kustomization.yaml"))
	if !strings.Contains(kustomization, "./opencenter-openstack-designate-credentials-secret.yaml") {
		t.Fatalf("expected Designate secret in kustomization:\n%s", kustomization)
	}

	issuer := mustReadFile(t, filepath.Join(base, "letsencrypt-designate-issuer.yaml"))
	if !strings.Contains(issuer, "webhook:") {
		t.Fatalf("expected webhook solver in issuer:\n%s", issuer)
	}
	if !strings.Contains(issuer, "groupName: acme.syseleven.de") {
		t.Fatalf("expected Designate webhook group name in issuer:\n%s", issuer)
	}
	if !strings.Contains(issuer, "solverName: designatedns") {
		t.Fatalf("expected Designate webhook solver name in issuer:\n%s", issuer)
	}

	secret := mustReadFile(t, filepath.Join(base, "opencenter-openstack-designate-credentials-secret.yaml"))
	if !strings.Contains(secret, "OS_APPLICATION_CREDENTIAL_ID: app-cred-id") {
		t.Fatalf("expected Designate application credential id in secret:\n%s", secret)
	}
	if !strings.Contains(secret, "OS_APPLICATION_CREDENTIAL_SECRET: app-cred-secret") {
		t.Fatalf("expected Designate application credential secret in secret:\n%s", secret)
	}
}

func TestRenderClusterAppsLokiSwift(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("loki-swift-guided")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Services["loki"] = &configservices.LokiConfig{
		BaseConfig:                   configservices.BaseConfig{Enabled: true, Namespace: "observability", SourceName: "opencenter-observability", OverrideValuesRendererKey: "loki"},
		StorageType:                  "swift",
		BucketName:                   "loki-container",
		SwiftAuthURL:                 "https://identity.api.example.com/v3",
		SwiftRegion:                  "SJC3",
		SwiftAuthVersion:             3,
		SwiftApplicationCredentialID: "app-cred-id",
		SwiftContainerName:           "loki-container",
		SwiftUserDomainName:          "rackspace",
		SwiftDomainName:              "rackspace",
	}
	cfg.Secrets.Loki.SwiftApplicationCredentialSecret = "swift-secret"

	if err := RenderClusterApps(cfg); err != nil {
		t.Fatalf("RenderClusterApps() error = %v", err)
	}

	overrideValues := mustReadFile(t, filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "loki", "helm-values", "override-values.yaml"))
	if !strings.Contains(overrideValues, "object_store: swift") {
		t.Fatalf("expected swift object store in Loki values:\n%s", overrideValues)
	}
	if !strings.Contains(overrideValues, "type: swift") {
		t.Fatalf("expected swift storage type in Loki values:\n%s", overrideValues)
	}
	if !strings.Contains(overrideValues, "application_credential_secret: swift-secret") {
		t.Fatalf("expected swift application credential secret in Loki values:\n%s", overrideValues)
	}
	if !strings.Contains(overrideValues, "container_name: loki-container") {
		t.Fatalf("expected swift container name in Loki values:\n%s", overrideValues)
	}
}

func TestRenderClusterAppsTempoSwift(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("tempo-swift-guided")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Services["tempo"] = &configservices.TempoConfig{
		BaseConfig:                   configservices.BaseConfig{Enabled: true, Namespace: "observability", SourceName: "opencenter-observability", OverrideValuesRendererKey: "tempo"},
		StorageType:                  "swift",
		BucketName:                   "tempo-container",
		SwiftAuthURL:                 "https://identity.api.example.com/v3",
		SwiftRegion:                  "SJC3",
		SwiftAuthVersion:             3,
		SwiftApplicationCredentialID: "app-cred-id",
		SwiftContainerName:           "tempo-container",
		SwiftUserDomainName:          "rackspace",
		SwiftDomainName:              "rackspace",
	}
	cfg.Secrets.Tempo.SwiftApplicationCredentialSecret = "swift-secret"

	if err := RenderClusterApps(cfg); err != nil {
		t.Fatalf("RenderClusterApps() error = %v", err)
	}

	overrideValues := mustReadFile(t, filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "tempo", "helm-values", "override-values.yaml"))
	if !strings.Contains(overrideValues, "backend: swift") {
		t.Fatalf("expected swift backend in Tempo values:\n%s", overrideValues)
	}
	if !strings.Contains(overrideValues, "application_credential_secret: swift-secret") {
		t.Fatalf("expected swift application credential secret in Tempo values:\n%s", overrideValues)
	}
	if !strings.Contains(overrideValues, "container_name: tempo-container") {
		t.Fatalf("expected swift container name in Tempo values:\n%s", overrideValues)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestRenderClusterAppsCertManagerMultiCredential(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("multi-cred")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Cluster.ClusterFQDN = "multi-cred.sjc3.k8s.opencenter.cloud"

	// Configure multiple AWS and Cloudflare credentials
	cfg.Secrets.CertManager.AWS = map[string]v2.CertManagerAWSCredential{
		"production": {
			Enabled:            true,
			AWSAccessKey:       "AKIAPROD123",
			AWSSecretAccessKey: "prodSecretKey",
			Region:             "us-east-1",
			DNSZones:           []string{"prod.example.com"},
		},
		"staging": {
			Enabled:            true,
			AWSAccessKey:       "AKIASTAGE456",
			AWSSecretAccessKey: "stageSecretKey",
			Region:             "us-west-2",
			DNSZones:           []string{"staging.example.com"},
		},
		"disabled-cred": {
			Enabled:            false,
			AWSAccessKey:       "SHOULD_NOT_APPEAR",
			AWSSecretAccessKey: "SHOULD_NOT_APPEAR",
		},
	}
	cfg.Secrets.CertManager.Cloudflare = map[string]v2.CertManagerCloudflareCredential{
		"cf-main": {
			Enabled:  true,
			APIToken: "cf-main-token",
			DNSZones: []string{"cf.example.com"},
		},
	}

	certManager := cfg.OpenCenter.Services["cert-manager"].(*configservices.CertManagerConfig)
	certManager.Email = "ops@example.com"
	certManager.Region = "us-east-1"

	if err := RenderClusterApps(cfg); err != nil {
		t.Fatalf("RenderClusterApps() error = %v", err)
	}

	base := filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "cert-manager")

	// Verify kustomization includes all enabled credentials
	kustomization := mustReadFile(t, filepath.Join(base, "kustomization.yaml"))
	for _, expected := range []string{
		"./opencenter-aws-credentials-secret-production.yaml",
		"./opencenter-aws-credentials-secret-staging.yaml",
		"./opencenter-cloudflare-credentials-secret-cf-main.yaml",
		"./letsencrypt-production-issuer.yaml",
		"./letsencrypt-staging-issuer.yaml",
		"./letsencrypt-cf-main-issuer.yaml",
	} {
		if !strings.Contains(kustomization, expected) {
			t.Errorf("expected %q in kustomization:\n%s", expected, kustomization)
		}
	}

	// Verify disabled credential is NOT rendered
	if strings.Contains(kustomization, "disabled-cred") {
		t.Fatalf("disabled credential should not appear in kustomization:\n%s", kustomization)
	}

	// Verify AWS production secret
	prodSecret := mustReadFile(t, filepath.Join(base, "opencenter-aws-credentials-secret-production.yaml"))
	if !strings.Contains(prodSecret, "name: opencenter-aws-credentials-secret-production") {
		t.Fatalf("expected production secret name:\n%s", prodSecret)
	}
	if !strings.Contains(prodSecret, "access-key-id: AKIAPROD123") {
		t.Fatalf("expected production access key:\n%s", prodSecret)
	}

	// Verify AWS staging secret
	stageSecret := mustReadFile(t, filepath.Join(base, "opencenter-aws-credentials-secret-staging.yaml"))
	if !strings.Contains(stageSecret, "name: opencenter-aws-credentials-secret-staging") {
		t.Fatalf("expected staging secret name:\n%s", stageSecret)
	}

	// Verify Cloudflare secret
	cfSecret := mustReadFile(t, filepath.Join(base, "opencenter-cloudflare-credentials-secret-cf-main.yaml"))
	if !strings.Contains(cfSecret, "name: opencenter-cloudflare-credentials-secret-cf-main") {
		t.Fatalf("expected cloudflare secret name:\n%s", cfSecret)
	}
	if !strings.Contains(cfSecret, "api-token: cf-main-token") {
		t.Fatalf("expected cloudflare API token:\n%s", cfSecret)
	}

	// Verify production issuer references correct secret
	prodIssuer := mustReadFile(t, filepath.Join(base, "letsencrypt-production-issuer.yaml"))
	if !strings.Contains(prodIssuer, "name: letsencrypt-production") {
		t.Fatalf("expected production issuer name:\n%s", prodIssuer)
	}
	if !strings.Contains(prodIssuer, `name: "opencenter-aws-credentials-secret-production"`) {
		t.Fatalf("expected production secret ref in issuer:\n%s", prodIssuer)
	}
	if !strings.Contains(prodIssuer, "region: us-east-1") {
		t.Fatalf("expected production region in issuer:\n%s", prodIssuer)
	}

	// Verify staging issuer uses its own region
	stageIssuer := mustReadFile(t, filepath.Join(base, "letsencrypt-staging-issuer.yaml"))
	if !strings.Contains(stageIssuer, "region: us-west-2") {
		t.Fatalf("expected staging region in issuer:\n%s", stageIssuer)
	}

	// Verify cloudflare issuer
	cfIssuer := mustReadFile(t, filepath.Join(base, "letsencrypt-cf-main-issuer.yaml"))
	if !strings.Contains(cfIssuer, "cloudflare:") {
		t.Fatalf("expected cloudflare solver in issuer:\n%s", cfIssuer)
	}
	if !strings.Contains(cfIssuer, `name: "opencenter-cloudflare-credentials-secret-cf-main"`) {
		t.Fatalf("expected cloudflare secret ref in issuer:\n%s", cfIssuer)
	}

	// Verify disabled credential file does NOT exist
	disabledPath := filepath.Join(base, "opencenter-aws-credentials-secret-disabled-cred.yaml")
	if _, err := os.Stat(disabledPath); err == nil {
		t.Fatalf("disabled credential file should not exist: %s", disabledPath)
	}
}

func TestRenderClusterAppsCertManagerValidationFailsOnMissingSecrets(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("validation-test")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Cluster.ClusterFQDN = "validation-test.example.com"

	// Enable an AWS credential but leave secrets empty
	cfg.Secrets.CertManager.AWS = map[string]v2.CertManagerAWSCredential{
		"broken": {
			Enabled:            true,
			AWSAccessKey:       "",
			AWSSecretAccessKey: "",
		},
	}

	certManager := cfg.OpenCenter.Services["cert-manager"].(*configservices.CertManagerConfig)
	certManager.Email = "ops@example.com"
	certManager.Region = "us-east-1"

	err := RenderClusterApps(cfg)
	if err == nil {
		t.Fatal("expected validation error for missing secrets, got nil")
	}
	if !strings.Contains(err.Error(), "secrets.cert_manager.aws.broken.aws_access_key is required") {
		t.Fatalf("expected access key validation error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "secrets.cert_manager.aws.broken.aws_secret_access_key is required") {
		t.Fatalf("expected secret key validation error, got: %v", err)
	}
}

func TestRenderClusterAppsCertManagerValidationFailsOnMissingCloudflareToken(t *testing.T) {
	dst := t.TempDir()
	cfg := newDefault("cf-validation-test")
	cfg.OpenCenter.GitOps.Repository.LocalDir = dst
	cfg.OpenCenter.Cluster.ClusterFQDN = "cf-validation-test.example.com"

	// Enable a Cloudflare credential but leave token empty
	cfg.Secrets.CertManager.Cloudflare = map[string]v2.CertManagerCloudflareCredential{
		"missing-token": {
			Enabled:  true,
			APIToken: "",
		},
	}

	certManager := cfg.OpenCenter.Services["cert-manager"].(*configservices.CertManagerConfig)
	certManager.DNSProvider = "cloudflare"
	certManager.Email = "ops@example.com"

	err := RenderClusterApps(cfg)
	if err == nil {
		t.Fatal("expected validation error for missing Cloudflare token, got nil")
	}
	if !strings.Contains(err.Error(), "secrets.cert_manager.cloudflare.missing-token.api_token is required") {
		t.Fatalf("expected Cloudflare token validation error, got: %v", err)
	}
}
