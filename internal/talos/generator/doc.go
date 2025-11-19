// Package generator creates declarative artifacts for Talos cluster deployment.
//
// The generator produces all necessary configuration files, network definitions,
// security policies, and GitOps structures required for a secure Talos cluster
// on OpenStack infrastructure.
//
// Key responsibilities:
//   - Talos machine configuration generation with security hardening
//   - Pulumi stack configuration file creation
//   - WireGuard VPN configuration generation
//   - Network topology definitions (management, control, data plane subnets)
//   - Security group rule generation
//   - SOPS encryption policy creation
//   - GitOps directory structure generation
package generator
