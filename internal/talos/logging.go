package talos

import (
	"github.com/rackerlabs/openCenter-cli/internal/config"
	"github.com/sirupsen/logrus"
)

// Logger returns a logger instance for the Talos provider.
// It uses the global logger from the config package with Talos-specific fields.
func Logger() *logrus.Entry {
	return config.GetGlobalLogger().WithField("component", "talos")
}

// LoggerWithFields returns a logger with additional fields.
func LoggerWithFields(fields logrus.Fields) *logrus.Entry {
	fields["component"] = "talos"
	return config.GetGlobalLogger().WithFields(fields)
}

// ValidatorLogger returns a logger for the validator component.
func ValidatorLogger() *logrus.Entry {
	return config.GetGlobalLogger().WithFields(logrus.Fields{
		"component": "talos",
		"module":    "validator",
	})
}

// GeneratorLogger returns a logger for the generator component.
func GeneratorLogger() *logrus.Entry {
	return config.GetGlobalLogger().WithFields(logrus.Fields{
		"component": "talos",
		"module":    "generator",
	})
}

// PulumiLogger returns a logger for the Pulumi integration component.
func PulumiLogger() *logrus.Entry {
	return config.GetGlobalLogger().WithFields(logrus.Fields{
		"component": "talos",
		"module":    "pulumi",
	})
}
