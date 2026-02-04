// Package migration provides tools for migrating from legacy configuration
// patterns to the unified ConfigurationManager.
//
// The migration scanner identifies files using legacy config.Load, config.Save,
// and config.Validate function calls, and generates reports to track migration
// progress across the codebase.
package migration
