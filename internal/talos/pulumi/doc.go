// Package pulumi provides Pulumi integration for infrastructure lifecycle management.
//
// This package manages the complete infrastructure lifecycle through Pulumi Go SDK,
// including preview, apply, refresh, and destroy operations. It handles Pulumi state
// management in Swift backend with encryption and implements retry logic for OpenStack
// eventual consistency.
//
// Key responsibilities:
//   - Pulumi stack initialization and backend configuration
//   - Infrastructure preview and change planning
//   - Resource provisioning and updates
//   - Configuration drift detection
//   - Resource cleanup and destruction
//   - Swift backend state management with encryption
//   - Secrets provider passphrase management via SOPS/Barbican
//   - Stack isolation per cluster/environment
package pulumi
