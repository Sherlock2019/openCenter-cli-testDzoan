package config

// Networking groups network settings and options around VRRP and service networks.
type Networking struct {
	SubnetNodes          string   `yaml:"subnet_nodes" json:"subnet_nodes"`
	AllocationPoolStart  string   `yaml:"allocation_pool_start" json:"allocation_pool_start"`
	AllocationPoolEnd    string   `yaml:"allocation_pool_end" json:"allocation_pool_end"`
	VRRPEnabled          bool     `yaml:"vrrp_enabled" json:"vrrp_enabled"`
	VRRPIP               string   `yaml:"vrrp_ip" json:"vrrp_ip"`
	SubnetServices       string   `yaml:"subnet_services" json:"subnet_services"`
	SubnetPods           string   `yaml:"subnet_pods" json:"subnet_pods"`
	UseOctavia           bool     `yaml:"use_octavia" json:"use_octavia"`
	LoadbalancerProvider string   `yaml:"loadbalancer_provider" json:"loadbalancer_provider"`
	UseDesignate         bool     `yaml:"use_designate" json:"use_designate"`
	DNSZoneName          string   `yaml:"dns_zone_name" json:"dns_zone_name"`
	DNSNameservers       []string `yaml:"dns_nameservers" json:"dns_nameservers"`
	VLAN                 VLAN     `yaml:"vlan" json:"vlan"`
}

// VLAN describes VLAN settings for the cluster.
type VLAN struct {
	ID       string `yaml:"id" json:"id"`
	MTU      int    `yaml:"mtu" json:"mtu"`
	Provider string `yaml:"provider" json:"provider"`
}
