// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
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
	"github.com/opencenter-cloud/opencenter-cli/internal/logging"
	"github.com/sirupsen/logrus"
)

// LoggerManager manages the logging configuration and setup.
// Deprecated: Use logging.LoggerManager directly.
type LoggerManager = logging.LoggerManager

// YAMLFormatter is a custom formatter that outputs logs in YAML format.
// Deprecated: Use logging.YAMLFormatter directly.
type YAMLFormatter = logging.YAMLFormatter

// NewLoggerManager creates a new logger manager with the given configuration.
// Deprecated: Use logging.NewLoggerManager directly.
func NewLoggerManager(config *LoggingConfig) (*LoggerManager, error) {
	return logging.NewLoggerManager(config)
}

// InitializeLogging initializes the global logging system.
// Deprecated: Use logging.Initialize directly.
func InitializeLogging(config *LoggingConfig) error {
	return logging.Initialize(config)
}

// GetGlobalLogger returns the global logger instance.
// Deprecated: Use logging.GetGlobalLogger directly.
func GetGlobalLogger() *logrus.Logger {
	return logging.GetGlobalLogger()
}

// ValidateLoggingConfig validates the logging configuration.
// Deprecated: Use logging.ValidateLoggingConfig directly.
func ValidateLoggingConfig(config *LoggingConfig) error {
	return logging.ValidateLoggingConfig(config)
}

// SetLogLevel sets the log level for the global logger.
// Deprecated: Use logging.SetLogLevel directly.
func SetLogLevel(level string) error {
	return logging.SetLogLevel(level)
}

// SetLogFormat sets the log format for the global logger.
// Deprecated: Use logging.SetLogFormat directly.
func SetLogFormat(format string) error {
	return logging.SetLogFormat(format)
}

// Debug logs a debug message.
// Deprecated: Use logging.Debug directly.
func Debug(args ...interface{}) { logging.Debug(args...) }

// Debugf logs a formatted debug message.
// Deprecated: Use logging.Debugf directly.
func Debugf(format string, args ...interface{}) { logging.Debugf(format, args...) }

// Info logs an info message.
// Deprecated: Use logging.Info directly.
func Info(args ...interface{}) { logging.Info(args...) }

// Infof logs a formatted info message.
// Deprecated: Use logging.Infof directly.
func Infof(format string, args ...interface{}) { logging.Infof(format, args...) }

// Warn logs a warning message.
// Deprecated: Use logging.Warn directly.
func Warn(args ...interface{}) { logging.Warn(args...) }

// Warnf logs a formatted warning message.
// Deprecated: Use logging.Warnf directly.
func Warnf(format string, args ...interface{}) { logging.Warnf(format, args...) }

// Error logs an error message.
// Deprecated: Use logging.Error directly.
func Error(args ...interface{}) { logging.Error(args...) }

// Errorf logs a formatted error message.
// Deprecated: Use logging.Errorf directly.
func Errorf(format string, args ...interface{}) { logging.Errorf(format, args...) }

// Fatal logs a fatal message and exits.
// Deprecated: Use logging.Fatal directly.
func Fatal(args ...interface{}) { logging.Fatal(args...) }

// Fatalf logs a formatted fatal message and exits.
// Deprecated: Use logging.Fatalf directly.
func Fatalf(format string, args ...interface{}) { logging.Fatalf(format, args...) }

// WithField creates a logger with a single field.
// Deprecated: Use logging.WithField directly.
func WithField(key string, value interface{}) *logrus.Entry {
	return logging.WithField(key, value)
}

// WithFields creates a logger with multiple fields.
// Deprecated: Use logging.WithFields directly.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logging.WithFields(fields)
}
