package config

// Storage describes block volume storage sizes and types for various node roles.
type Storage struct {
	MasterNodeBFV        BFV `yaml:"master_node_bfv" json:"master_node_bfv"`
	WorkerNodeBFV        BFV `yaml:"worker_node_bfv" json:"worker_node_bfv"`
	WorkerNodeBFVWindows BFV `yaml:"worker_node_bfv_windows" json:"worker_node_bfv_windows"`
}

// BFV describes the block volume size and type.
type BFV struct {
	Size                        int              `yaml:"size" json:"size"`
	Type                        string           `yaml:"type" json:"type"`
	WorkerVolumeSize            int              `yaml:"worker_volume_size,omitempty" json:"worker_volume_size,omitempty" jsonschema:"description=Worker node boot volume size in GB,default=100"`
	WorkerVolumeDestinationType string           `yaml:"worker_volume_destination_type,omitempty" json:"worker_volume_destination_type,omitempty" jsonschema:"description=Worker node boot volume destination type,default=volume"`
	WorkerVolumeSourceType      string           `yaml:"worker_volume_source_type,omitempty" json:"worker_volume_source_type,omitempty" jsonschema:"description=Worker node boot volume source type,default=image"`
	WorkerVolumeType            string           `yaml:"worker_volume_type,omitempty" json:"worker_volume_type,omitempty" jsonschema:"description=Worker node boot volume type,default=HA-Performance"`
	AdditionalBlockDevices      []map[string]any `yaml:"additional_block_devices,omitempty" json:"additional_block_devices,omitempty" jsonschema:"description=Additional block devices for worker nodes"`
}

// Windows holds Windows-specific settings.
type Windows struct {
	User          string `yaml:"user" json:"user"`
	AdminPassword string `yaml:"admin_password" json:"admin_password"`
}

// StorageConfig represents the storage configuration for the cluster
type StorageConfig struct {
	DefaultStorageClass             string           `yaml:"default_storage_class,omitempty" json:"default_storage_class,omitempty" jsonschema:"description=Default storage class for persistent volumes,default=csi-cinder-sc-delete" validate:"required"`
	WorkerVolumeSize                int              `yaml:"worker_volume_size,omitempty" json:"worker_volume_size,omitempty" jsonschema:"description=Worker node boot volume size in GB,default=40" validate:"required,min=10,max=10000"`
	WorkerVolumeDestinationType     string           `yaml:"worker_volume_destination_type,omitempty" json:"worker_volume_destination_type,omitempty" jsonschema:"description=Worker node boot volume destination type,default=volume" validate:"required,oneof=volume local"`
	WorkerVolumeSourceType          string           `yaml:"worker_volume_source_type,omitempty" json:"worker_volume_source_type,omitempty" jsonschema:"description=Worker node boot volume source type,default=image" validate:"required,oneof=image volume snapshot"`
	WorkerVolumeType                string           `yaml:"worker_volume_type,omitempty" json:"worker_volume_type,omitempty" jsonschema:"description=Worker node boot volume type,default=HA-Standard" validate:"required"`
	WorkerVolumeDeleteOnTermination bool             `yaml:"worker_volume_delete_on_termination,omitempty" json:"worker_volume_delete_on_termination,omitempty" jsonschema:"description=Delete worker boot volume on instance termination,default=false"`
	MasterVolumeSize                int              `yaml:"master_volume_size,omitempty" json:"master_volume_size,omitempty" jsonschema:"description=Master node boot volume size in GB,default=40" validate:"min=0,max=10000"`
	MasterVolumeDestinationType     string           `yaml:"master_volume_destination_type,omitempty" json:"master_volume_destination_type,omitempty" jsonschema:"description=Master node boot volume destination type,default=volume"`
	MasterVolumeSourceType          string           `yaml:"master_volume_source_type,omitempty" json:"master_volume_source_type,omitempty" jsonschema:"description=Master node boot volume source type,default=image"`
	MasterVolumeType                string           `yaml:"master_volume_type,omitempty" json:"master_volume_type,omitempty" jsonschema:"description=Master node boot volume type,default=HA-Standard"`
	MasterVolumeDeleteOnTermination bool             `yaml:"master_volume_delete_on_termination,omitempty" json:"master_volume_delete_on_termination,omitempty" jsonschema:"description=Delete master boot volume on instance termination,default=false"`
	AdditionalBlockDevices          []map[string]any `yaml:"additional_block_devices,omitempty" json:"additional_block_devices,omitempty" jsonschema:"description=Additional block devices for worker nodes"`
}
