package services

import (
	"github.com/rackerlabs/opencenter-cli/internal/config/registry"
)

// VSphereCSIConfig extends BaseConfig with vSphere CSI configuration
type VSphereCSIConfig struct {
	BaseConfig `yaml:",inline"`

	// Datastore configuration for storage classes
	DataStore              string `yaml:"datastore_name" json:"datastore_name,omitempty" jsonschema:"description=Default datastore name for delete storage class"`
	DataStoreURL           string `yaml:"datastoreurl" json:"datastoreurl,omitempty" jsonschema:"description=Datastore name"`
	DeleteDataStoreUUID    string `yaml:"delete_datastore_uuid" json:"delete_datastore_uuid,omitempty" jsonschema:"description=Datastore UUID for delete storage class"`
	RetainDataStoreName    string `yaml:"retain_datastore_name" json:"retain_datastore_name,omitempty" jsonschema:"description=Datastore name for retain storage class"`
	RetainDataStoreUUID    string `yaml:"retain_datastore_uuid" json:"retain_datastore_uuid,omitempty" jsonschema:"description=Datastore UUID for retain storage class"`
}

func init() {
	registry.RegisterServiceConfig("vsphere-csi", VSphereCSIConfig{})
}
