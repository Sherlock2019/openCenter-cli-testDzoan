package talos

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Role string

const (
	RoleControlPlane Role = "control-plane"
	RoleWorker       Role = "worker"
)

type Inventory struct {
	Cluster      InventoryCluster `yaml:"cluster"`
	ControlPlane []Node           `yaml:"control_plane"`
	Workers      []Node           `yaml:"workers,omitempty"`
	PatchInputs  PatchInputs      `yaml:"patch_inputs,omitempty"`
}

type InventoryCluster struct {
	Name         string `yaml:"name"`
	Endpoint     string `yaml:"endpoint"`
	TalosAPIPort int    `yaml:"talos_api_port"`
}

type Node struct {
	Name        string            `yaml:"name"`
	Role        Role              `yaml:"-"`
	TalosAPIIP  string            `yaml:"talos_api_ip"`
	InternalIP  string            `yaml:"internal_ip"`
	InstallDisk string            `yaml:"install_disk"`
	CertSANs    []string          `yaml:"cert_sans,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

type PatchInputs struct {
	DNSServers    []string `yaml:"dns_servers,omitempty"`
	NTPServers    []string `yaml:"ntp_servers,omitempty"`
	PodSubnet     string   `yaml:"pod_subnet,omitempty"`
	ServiceSubnet string   `yaml:"service_subnet,omitempty"`
}

func LoadInventory(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading Talos inventory %s: %w", path, err)
	}

	var inventory Inventory
	if err := yaml.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("parsing Talos inventory %s: %w", path, err)
	}
	inventory.assignRoles()

	if err := inventory.Validate(path); err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (i *Inventory) Validate(path string) error {
	if i == nil {
		return fmt.Errorf("validating Talos inventory %s: inventory is nil", path)
	}
	if strings.TrimSpace(i.Cluster.Name) == "" {
		return inventoryFieldError(path, "cluster.name")
	}
	if strings.TrimSpace(i.Cluster.Endpoint) == "" {
		return inventoryFieldError(path, "cluster.endpoint")
	}
	if i.Cluster.TalosAPIPort <= 0 {
		return inventoryFieldError(path, "cluster.talos_api_port")
	}
	if len(i.ControlPlane) == 0 {
		return inventoryFieldError(path, "control_plane")
	}
	for idx := range i.ControlPlane {
		i.ControlPlane[idx].Role = RoleControlPlane
		if err := validateInventoryNode(path, fmt.Sprintf("control_plane[%d]", idx), i.ControlPlane[idx]); err != nil {
			return err
		}
	}
	for idx := range i.Workers {
		i.Workers[idx].Role = RoleWorker
		if err := validateInventoryNode(path, fmt.Sprintf("workers[%d]", idx), i.Workers[idx]); err != nil {
			return err
		}
	}
	return nil
}

func (i *Inventory) AllNodes() []Node {
	if i == nil {
		return nil
	}
	nodes := make([]Node, 0, len(i.ControlPlane)+len(i.Workers))
	for _, node := range i.ControlPlane {
		node.Role = RoleControlPlane
		nodes = append(nodes, node)
	}
	for _, node := range i.Workers {
		node.Role = RoleWorker
		nodes = append(nodes, node)
	}
	return nodes
}

func (i *Inventory) EndpointIPs() []string {
	if i == nil {
		return nil
	}
	endpoints := make([]string, 0, len(i.ControlPlane))
	for _, node := range i.ControlPlane {
		if ip := strings.TrimSpace(node.TalosAPIIP); ip != "" {
			endpoints = append(endpoints, ip)
		}
	}
	return endpoints
}

func (i *Inventory) FirstControlPlane() (Node, error) {
	if i == nil || len(i.ControlPlane) == 0 {
		return Node{}, fmt.Errorf("Talos inventory has no control-plane nodes")
	}
	node := i.ControlPlane[0]
	node.Role = RoleControlPlane
	return node, nil
}

func (i *Inventory) assignRoles() {
	for idx := range i.ControlPlane {
		i.ControlPlane[idx].Role = RoleControlPlane
	}
	for idx := range i.Workers {
		i.Workers[idx].Role = RoleWorker
	}
}

func validateInventoryNode(path, prefix string, node Node) error {
	if strings.TrimSpace(node.Name) == "" {
		return inventoryFieldError(path, prefix+".name")
	}
	if strings.TrimSpace(node.TalosAPIIP) == "" {
		return inventoryFieldError(path, prefix+".talos_api_ip")
	}
	if strings.TrimSpace(node.InternalIP) == "" {
		return inventoryFieldError(path, prefix+".internal_ip")
	}
	if strings.TrimSpace(node.InstallDisk) == "" {
		return inventoryFieldError(path, prefix+".install_disk")
	}
	return nil
}

func inventoryFieldError(path, field string) error {
	return fmt.Errorf("validating Talos inventory %s: missing required field %s", path, field)
}
