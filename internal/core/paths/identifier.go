package paths

import (
	"fmt"
	"strings"
)

// ParseClusterIdentifier parses a cluster identifier in the format "organization/cluster"
// or a bare "cluster" name. If no organization is specified, "opencenter" is used as default.
func ParseClusterIdentifier(identifier string, validateClusterName func(string) error) (organization string, clusterName string, err error) {
	if identifier == "" {
		return "", "", fmt.Errorf("cluster identifier cannot be empty")
	}

	if strings.Contains(identifier, "/") {
		parts := strings.SplitN(identifier, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid cluster identifier format: expected 'organization/cluster'")
		}
		organization = parts[0]
		clusterName = parts[1]
		if organization == "" {
			return "", "", fmt.Errorf("organization name cannot be empty")
		}
		if err := validateClusterName(clusterName); err != nil {
			return "", "", err
		}
		return organization, clusterName, nil
	}

	if err := validateClusterName(identifier); err != nil {
		return "", "", err
	}
	return "opencenter", identifier, nil
}
