package talos

// NetworkTopology defines the three-zone network architecture.
type NetworkTopology struct {
	ManagementNetwork Network `json:"management_network"`
	ControlNetwork    Network `json:"control_network"`
	DataNetwork       Network `json:"data_network"`
	Router            Router  `json:"router"`
	Routes            []Route `json:"routes"`
}

// Network represents a Neutron network and subnet.
type Network struct {
	Name           string         `json:"name"`
	CIDR           string         `json:"cidr"`
	Gateway        string         `json:"gateway"`
	DNSServers     []string       `json:"dns_servers"`
	AllocationPool AllocationPool `json:"allocation_pool"`
}

// AllocationPool defines IP allocation range.
type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Router represents a Neutron router.
type Router struct {
	Name              string `json:"name"`
	ExternalNetworkID string `json:"external_network_id"`
}

// Route represents a network route.
type Route struct {
	Destination string `json:"destination"`
	NextHop     string `json:"next_hop"`
}

// SecurityGroup defines firewall rules.
type SecurityGroup struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Rules       []SecurityRule `json:"rules"`
}

// SecurityRule defines a single firewall rule.
type SecurityRule struct {
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   int    `json:"port_range_min"`
	PortRangeMax   int    `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	RemoteGroupID  string `json:"remote_group_id"`
	Description    string `json:"description"`
}

// WireGuardConfig holds VPN configuration.
type WireGuardConfig struct {
	ServerPublicKey  string          `json:"server_public_key"`
	ServerPrivateKey string          `json:"server_private_key"`
	ServerAddress    string          `json:"server_address"`
	ServerPort       int             `json:"server_port"`
	Peers            []WireGuardPeer `json:"peers"`
}

// WireGuardPeer represents a VPN peer.
type WireGuardPeer struct {
	PublicKey    string   `json:"public_key"`
	AllowedIPs   []string `json:"allowed_ips"`
	PresharedKey string   `json:"preshared_key"`
}

// TalosNode represents a Talos Linux instance.
type TalosNode struct {
	Name           string           `json:"name"`
	Type           NodeType         `json:"type"`
	FlavorID       string           `json:"flavor_id"`
	ImageID        string           `json:"image_id"`
	NetworkID      string           `json:"network_id"`
	SecurityGroups []string         `json:"security_groups"`
	VTPMEnabled    bool             `json:"vtpm_enabled"`
	BootVolume     BootVolumeConfig `json:"boot_volume"`
	MachineConfig  []byte           `json:"machine_config"`
}

// BootVolumeConfig defines encrypted boot volume.
type BootVolumeConfig struct {
	Size      int    `json:"size"`
	Type      string `json:"type"`
	Encrypted bool   `json:"encrypted"`
	KeyID     string `json:"key_id"`
}
