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
)

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

// GetCertManagerAWSCredentials returns cert-manager AWS credentials from legacy flat fields.
// Deprecated: Use EnabledCertManagerAWSCredentials() for multi-credential support.
func (c Config) GetCertManagerAWSCredentials() (accessKey, secretKey string) {
	if c.Secrets.CertManager.AWSAccessKey != "" && c.Secrets.CertManager.AWSSecretAccessKey != "" {
		return c.Secrets.CertManager.AWSAccessKey, c.Secrets.CertManager.AWSSecretAccessKey
	}
	return "", ""
}

// EnabledCertManagerAWSCredentials returns all enabled AWS credentials for cert-manager.
func (c Config) EnabledCertManagerAWSCredentials() map[string]CertManagerAWSCredential {
	result := make(map[string]CertManagerAWSCredential)
	for name, cred := range c.Secrets.CertManager.AWS {
		if cred.Enabled {
			result[name] = cred
		}
	}
	return result
}

// EnabledCertManagerCloudflareCredentials returns all enabled Cloudflare credentials for cert-manager.
func (c Config) EnabledCertManagerCloudflareCredentials() map[string]CertManagerCloudflareCredential {
	result := make(map[string]CertManagerCloudflareCredential)
	for name, cred := range c.Secrets.CertManager.Cloudflare {
		if cred.Enabled {
			result[name] = cred
		}
	}
	return result
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

// GetCertManagerAWSAccessKey returns cert-manager AWS access key from legacy flat fields.
// Deprecated: Use EnabledCertManagerAWSCredentials() for multi-credential support.
func (c Config) GetCertManagerAWSAccessKey() string {
	accessKey, _ := c.GetCertManagerAWSCredentials()
	return accessKey
}

// GetCertManagerAWSSecretKey returns cert-manager AWS secret key from legacy flat fields.
// Deprecated: Use EnabledCertManagerAWSCredentials() for multi-credential support.
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
