package services

// ServiceConfig is the interface that all service configurations must implement
type ServiceConfig interface {
	IsEnabled() bool
	GetStatus() string
}
