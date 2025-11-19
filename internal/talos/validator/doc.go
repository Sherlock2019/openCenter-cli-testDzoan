// Package validator provides pre-flight validation checks for OpenStack environments
// before Talos cluster deployment.
//
// The validator verifies that all required OpenStack services are available and properly
// configured, including Keystone (with MFA), Barbican, Octavia, Glance (with image
// signature verification), and sufficient resource quotas.
//
// All validation checks produce structured reports with actionable remediation steps
// for any failures.
package validator
