// Copyrigho 2025 Victor Palma <victor.palma@rackspace.com>
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
	"encoding/json"

	"github.com/rackerlabs/opencenter-cli/internal/config/services"
)

// Types are now defined in types.go

// Default functions are now in defaults.go

// I/O functions are now in persistence.go

// validateServiceSecretsSimple validates service-specific secrets configuration.
// This function checks that required secrets are present when corresponding services are enabled.
//
// Deprecated: This function will be migrated to use internal/core/validation.ValidationEngine in v2.0.0.
// For now, it remains as an internal helper for the main Validate function.
func validateServiceSecretsSimple(cfg Config) []string {
	var errs []string

	isEnabled := func(name string) bool {
		svc, exists := cfg.OpenCenter.Services[name]
		if !exists {
			return false
		}
		if svcConf, ok := svc.(services.ServiceConfig); ok {
			return svcConf.IsEnabled()
		}
		return false
	}

	// Validate cert-manager secrets
	if isEnabled("cert-manager") {
		accessKey, secretKey := cfg.GetCertManagerAWSCredentials()
		if accessKey == "" {
			errs = append(errs, "AWS credentials required for cert-manager: either set secrets.cert_manager.aws_access_key or secrets.global.aws.application.access_key or secrets.global.aws.infrastructure.access_key")
		}
		if secretKey == "" {
			errs = append(errs, "AWS credentials required for cert-manager: either set secrets.cert_manager.aws_secret_access_key or secrets.global.aws.application.secret_access_key or secrets.global.aws.infrastructure.secret_access_key")
		}
	}

	// Validate loki secrets
	if isEnabled("loki") {
		// Check for Swift credentials (legacy)
		if cfg.Secrets.Loki.SwiftPassword == "" {
			// If no Swift password, check for S3 credentials (with fallback)
			accessKey, secretKey := cfg.GetLokiS3Credentials()
			if accessKey == "" || secretKey == "" {
				errs = append(errs, "Loki requires either Swift password (secrets.loki.swift_password) or S3 credentials (secrets.loki.s3_access_key_id/secrets.loki.s3_secret_access_key or secrets.global.aws.application.access_key/secret_access_key or secrets.global.aws.infrastructure.access_key/secret_access_key)")
			}
		}
	}

	// Validate tempo secrets
	if isEnabled("tempo") {
		accessKey, secretKey := cfg.GetTempoS3Credentials()
		if accessKey == "" {
			errs = append(errs, "S3 credentials required for Tempo: either set secrets.tempo.access_key or secrets.global.aws.application.access_key or secrets.global.aws.infrastructure.access_key")
		}
		if secretKey == "" {
			errs = append(errs, "S3 credentials required for Tempo: either set secrets.tempo.secret_key or secrets.global.aws.application.secret_access_key or secrets.global.aws.infrastructure.secret_access_key")
		}
	}

	// Validate keycloak secrets
	if isEnabled("keycloak") {
		if cfg.Secrets.Keycloak.AdminPassword == "" {
			errs = append(errs, "secrets.keycloak.admin_password is required when keycloak is enabled")
		}
	}

	return errs
}

// ToJSON marshals the configuration to JSON. This is used for generating
// the JSON schema and for other tools that consume JSON.
//
// Outputs:
//   - []byte: The JSON-encoded configuration.
//   - error: An error if the configuration cannot be marshaled.
func (c Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// GetAWSCredentials returns AWS credentials with service-specific override and fallback logic.
// It first tries service-specific credentials, then falls back to global infrastructure credentials.
//
// Parameters:
//   - serviceAccessKey: Service-specific AWS access key
//   - serviceSecretKey: Service-specific AWS secret access key
//
// Returns:
//   - accessKey: The resolved AWS access key
//   - secretKey: The resolved AWS secret access key
func (c Config) GetAWSCredentials(serviceAccessKey, serviceSecretKey string) (accessKey, secretKey string) {
	// Use service-specific credentials if provided
	if serviceAccessKey != "" && serviceSecretKey != "" {
		return serviceAccessKey, serviceSecretKey
	}

	// Fall back to global infrastructure AWS credentials
	return c.Secrets.Global.AWS.Infrastructure.AccessKey, c.Secrets.Global.AWS.Infrastructure.SecretAccessKey
}

// GetCertManagerAWSCredentials returns cert-manager AWS credentials with fallback to global AWS application credentials.
func (c Config) GetCertManagerAWSCredentials() (accessKey, secretKey string) {
	// Use service-specific credentials if provided
	if c.Secrets.CertManager.AWSAccessKey != "" && c.Secrets.CertManager.AWSSecretAccessKey != "" {
		return c.Secrets.CertManager.AWSAccessKey, c.Secrets.CertManager.AWSSecretAccessKey
	}

	// Fall back to global application AWS credentials
	return c.GetAWSApplicationCredentials()
}

// GetLokiS3Credentials returns Loki S3 credentials with fallback to global AWS application credentials.
func (c Config) GetLokiS3Credentials() (accessKey, secretKey string) {
	// Use service-specific credentials if provided
	if c.Secrets.Loki.S3AccessKeyID != "" && c.Secrets.Loki.S3SecretAccessKey != "" {
		return c.Secrets.Loki.S3AccessKeyID, c.Secrets.Loki.S3SecretAccessKey
	}

	// Fall back to global application AWS credentials
	return c.GetAWSApplicationCredentials()
}

// GetTempoS3Credentials returns Tempo S3 credentials with fallback to global AWS application credentials.
func (c Config) GetTempoS3Credentials() (accessKey, secretKey string) {
	// Use service-specific credentials if provided
	if c.Secrets.Tempo.AccessKey != "" && c.Secrets.Tempo.SecretKey != "" {
		return c.Secrets.Tempo.AccessKey, c.Secrets.Tempo.SecretKey
	}

	// Fall back to global application AWS credentials
	return c.GetAWSApplicationCredentials()
}

// GetS3BackendCredentials returns S3 backend credentials with fallback to global AWS credentials.
func (c Config) GetS3BackendCredentials() (accessKey, secretKey string) {
	return c.GetAWSCredentials(c.OpenCenter.Cluster.AWSAccessKey, c.OpenCenter.Cluster.AWSSecretAccessKey)
}

// GetAWSApplicationCredentials returns AWS application credentials with fallback logic.
// It first tries the global application credentials, then falls back to infrastructure credentials.
//
// Returns:
//   - accessKey: The resolved AWS access key
//   - secretKey: The resolved AWS secret access key
func (c Config) GetAWSApplicationCredentials() (accessKey, secretKey string) {
	// Use global application AWS credentials if provided
	if c.Secrets.Global.AWS.Application.AccessKey != "" && c.Secrets.Global.AWS.Application.SecretAccessKey != "" {
		return c.Secrets.Global.AWS.Application.AccessKey, c.Secrets.Global.AWS.Application.SecretAccessKey
	}

	// Fall back to infrastructure credentials
	return c.Secrets.Global.AWS.Infrastructure.AccessKey, c.Secrets.Global.AWS.Infrastructure.SecretAccessKey
}

// Template-friendly functions that return single values for use in Go templates

// GetCertManagerAWSAccessKey returns cert-manager AWS access key with fallback.
func (c Config) GetCertManagerAWSAccessKey() string {
	accessKey, _ := c.GetCertManagerAWSCredentials()
	return accessKey
}

// GetCertManagerAWSSecretKey returns cert-manager AWS secret key with fallback.
func (c Config) GetCertManagerAWSSecretKey() string {
	_, secretKey := c.GetCertManagerAWSCredentials()
	return secretKey
}

// GetLokiS3AccessKey returns Loki S3 access key with fallback.
func (c Config) GetLokiS3AccessKey() string {
	accessKey, _ := c.GetLokiS3Credentials()
	return accessKey
}

// GetLokiS3SecretKey returns Loki S3 secret key with fallback.
func (c Config) GetLokiS3SecretKey() string {
	_, secretKey := c.GetLokiS3Credentials()
	return secretKey
}

// GetTempoS3AccessKey returns Tempo S3 access key with fallback.
func (c Config) GetTempoS3AccessKey() string {
	accessKey, _ := c.GetTempoS3Credentials()
	return accessKey
}

// GetTempoS3SecretKey returns Tempo S3 secret key with fallback.
func (c Config) GetTempoS3SecretKey() string {
	_, secretKey := c.GetTempoS3Credentials()
	return secretKey
}

// GetS3BackendAccessKey returns S3 backend access key with fallback.
func (c Config) GetS3BackendAccessKey() string {
	accessKey, _ := c.GetS3BackendCredentials()
	return accessKey
}

// GetS3BackendSecretKey returns S3 backend secret key with fallback.
func (c Config) GetS3BackendSecretKey() string {
	_, secretKey := c.GetS3BackendCredentials()
	return secretKey
}
