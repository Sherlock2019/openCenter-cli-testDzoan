// Package talos provides the Talos OpenStack provider implementation for the opencenter CLI.
//
// The Talos provider enables secure, immutable Kubernetes cluster deployment on OpenStack
// infrastructure using Talos Linux. It integrates with the opencenter CLI to deliver a
// declarative, GitOps-friendly lifecycle powered by Pulumi Go bindings.
//
// The provider enforces Zero Trust networking, cryptographic attestation, and defense-in-depth
// policies by default, eliminating SSH access and traditional mutable management patterns.
//
// Package Structure:
//   - validator: Pre-flight validation of OpenStack environment prerequisites
//   - generator: Generation of declarative artifacts (Talos configs, Pulumi stacks, etc.)
//   - pulumi: Pulumi integration for infrastructure lifecycle management
//   - errors: Structured error handling with categorization and remediation
package talos
