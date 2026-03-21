package cloud

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CompareInfrastructureState compares desired and actual infrastructure state
// using a shared severity and reconcilability model across providers.
func CompareInfrastructureState(desired, actual *InfrastructureState) *DriftReport {
	if desired == nil {
		desired = &InfrastructureState{}
	}
	if actual == nil {
		actual = &InfrastructureState{}
	}

	report := &DriftReport{
		Drifts:  []DriftItem{},
		Summary: DriftSummary{},
	}

	detectServerDrift(desired.Servers, actual.Servers, report)
	detectNetworkDrift(desired.Networks, actual.Networks, report)
	detectSecurityGroupDrift(desired.SecurityGroups, actual.SecurityGroups, report)
	detectLoadBalancerDrift(desired.LoadBalancers, actual.LoadBalancers, report)
	detectVolumeDrift(desired.Volumes, actual.Volumes, report)
	detectFloatingIPDrift(desired.FloatingIPs, actual.FloatingIPs, report)
	CalculateSummary(report)

	return report
}

// CalculateSummary updates summary counters and aggregate flags for a drift report.
func CalculateSummary(report *DriftReport) {
	report.Summary.TotalDrifts = len(report.Drifts)
	report.Reconcilable = true
	report.OverallSeverity = SeverityInfo

	for _, drift := range report.Drifts {
		switch drift.Severity {
		case SeverityCritical:
			report.Summary.CriticalCount++
			report.OverallSeverity = SeverityCritical
		case SeverityWarning:
			report.Summary.WarningCount++
			if report.OverallSeverity < SeverityWarning {
				report.OverallSeverity = SeverityWarning
			}
		case SeverityInfo:
			report.Summary.InfoCount++
		}

		if drift.Reconcilable {
			report.Summary.ReconcilableCount++
		} else {
			report.Reconcilable = false
		}
	}
}

func detectServerDrift(desired, actual []Server, report *DriftReport) {
	desiredMap := mapByKey(desired, func(s Server) string { return serverKey(s) })
	actualMap := mapByKey(actual, func(s Server) string { return serverKey(s) })

	for key, desiredServer := range desiredMap {
		actualServer, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "server",
				ResourceID:   desiredServer.ID,
				ResourceName: key,
				Field:        "existence",
				Expected:     "exists",
				Actual:       "missing",
				Severity:     missingServerSeverity(desiredServer),
				Reconcilable: false,
				Message:      fmt.Sprintf("Server %s is missing from infrastructure", key),
			})
			continue
		}

		if desiredServer.Flavor != "" && actualServer.Flavor != "" && desiredServer.Flavor != actualServer.Flavor {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "server",
				ResourceID:   actualServer.ID,
				ResourceName: key,
				Field:        "flavor",
				Expected:     desiredServer.Flavor,
				Actual:       actualServer.Flavor,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Server %s has the wrong flavor", key),
			})
		}

		if desiredServer.Image != "" && actualServer.Image != "" && desiredServer.Image != actualServer.Image {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "server",
				ResourceID:   actualServer.ID,
				ResourceName: key,
				Field:        "image",
				Expected:     desiredServer.Image,
				Actual:       actualServer.Image,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Server %s has the wrong image", key),
			})
		}

		if actualServer.Status != "" && !strings.EqualFold(actualServer.Status, "ACTIVE") {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "server",
				ResourceID:   actualServer.ID,
				ResourceName: key,
				Field:        "status",
				Expected:     "ACTIVE",
				Actual:       actualServer.Status,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Server %s is not active", key),
			})
		}

		for tagKey, expectedValue := range desiredServer.Tags {
			actualValue := actualServer.Tags[tagKey]
			if actualValue != expectedValue {
				report.Drifts = append(report.Drifts, DriftItem{
					ResourceType: "server",
					ResourceID:   actualServer.ID,
					ResourceName: key,
					Field:        fmt.Sprintf("tags.%s", tagKey),
					Expected:     expectedValue,
					Actual:       actualValue,
					Severity:     SeverityInfo,
					Reconcilable: true,
					Message:      fmt.Sprintf("Server %s has an unexpected tag value for %s", key, tagKey),
				})
			}
		}

		if len(desiredServer.Networks) > 0 && !equalStringSets(desiredServer.Networks, actualServer.Networks) {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "server",
				ResourceID:   actualServer.ID,
				ResourceName: key,
				Field:        "networks",
				Expected:     sortedCopy(desiredServer.Networks),
				Actual:       sortedCopy(actualServer.Networks),
				Severity:     SeverityWarning,
				Reconcilable: false,
				Message:      fmt.Sprintf("Server %s is attached to unexpected networks", key),
			})
		}
	}

	for key, actualServer := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "server",
			ResourceID:   actualServer.ID,
			ResourceName: key,
			Field:        "existence",
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected server %s found in infrastructure", key),
		})
	}
}

func detectNetworkDrift(desired, actual []Network, report *DriftReport) {
	desiredMap := mapByKey(desired, func(n Network) string { return networkKey(n) })
	actualMap := mapByKey(actual, func(n Network) string { return networkKey(n) })

	for key, desiredNetwork := range desiredMap {
		actualNetwork, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "network",
				ResourceID:   desiredNetwork.ID,
				ResourceName: key,
				Field:        "existence",
				Expected:     "exists",
				Actual:       "missing",
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Network %s is missing from infrastructure", key),
			})
			continue
		}

		if desiredNetwork.CIDR != "" && actualNetwork.CIDR != "" && desiredNetwork.CIDR != actualNetwork.CIDR {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "network",
				ResourceID:   actualNetwork.ID,
				ResourceName: key,
				Field:        "cidr",
				Expected:     desiredNetwork.CIDR,
				Actual:       actualNetwork.CIDR,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Network %s has the wrong CIDR", key),
			})
		}

		detectSubnetDrift(key, actualNetwork.ID, desiredNetwork.Subnets, actualNetwork.Subnets, report)
	}

	for key, actualNetwork := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "network",
			ResourceID:   actualNetwork.ID,
			ResourceName: key,
			Field:        "existence",
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected network %s found in infrastructure", key),
		})
	}
}

func detectSubnetDrift(networkName, networkID string, desired, actual []Subnet, report *DriftReport) {
	desiredMap := mapByKey(desired, func(s Subnet) string { return subnetKey(s) })
	actualMap := mapByKey(actual, func(s Subnet) string { return subnetKey(s) })

	for key, desiredSubnet := range desiredMap {
		actualSubnet, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "network",
				ResourceID:   networkID,
				ResourceName: networkName,
				Field:        fmt.Sprintf("subnets.%s.existence", key),
				Expected:     "exists",
				Actual:       "missing",
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Network %s is missing subnet %s", networkName, key),
			})
			continue
		}

		if desiredSubnet.CIDR != "" && actualSubnet.CIDR != "" && desiredSubnet.CIDR != actualSubnet.CIDR {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "network",
				ResourceID:   networkID,
				ResourceName: networkName,
				Field:        fmt.Sprintf("subnets.%s.cidr", key),
				Expected:     desiredSubnet.CIDR,
				Actual:       actualSubnet.CIDR,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Subnet %s in network %s has the wrong CIDR", key, networkName),
			})
		}
	}

	for key := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "network",
			ResourceID:   networkID,
			ResourceName: networkName,
			Field:        fmt.Sprintf("subnets.%s.existence", key),
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected subnet %s found in network %s", key, networkName),
		})
	}
}

func detectSecurityGroupDrift(desired, actual []SecurityGroup, report *DriftReport) {
	desiredMap := mapByKey(desired, func(s SecurityGroup) string { return securityGroupKey(s) })
	actualMap := mapByKey(actual, func(s SecurityGroup) string { return securityGroupKey(s) })

	for key, desiredGroup := range desiredMap {
		actualGroup, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "security_group",
				ResourceID:   desiredGroup.ID,
				ResourceName: key,
				Field:        "existence",
				Expected:     "exists",
				Actual:       "missing",
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Security group %s is missing from infrastructure", key),
			})
			continue
		}

		if len(desiredGroup.Rules) == 0 {
			continue
		}

		missingRules, riskyExtraRules := compareSecurityRules(desiredGroup.Rules, actualGroup.Rules)
		if len(missingRules) == 0 && len(riskyExtraRules) == 0 {
			continue
		}

		severity := SeverityWarning
		if len(missingRules) > 0 && hasHighRiskRules(missingRules) || hasHighRiskRules(riskyExtraRules) {
			severity = SeverityCritical
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "security_group",
			ResourceID:   actualGroup.ID,
			ResourceName: key,
			Field:        "rules",
			Expected:     desiredGroup.Rules,
			Actual:       actualGroup.Rules,
			Severity:     severity,
			Reconcilable: true,
			Message:      fmt.Sprintf("Security group %s rules differ from the desired configuration", key),
		})
	}

	for key, actualGroup := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "security_group",
			ResourceID:   actualGroup.ID,
			ResourceName: key,
			Field:        "existence",
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected security group %s found in infrastructure", key),
		})
	}
}

func detectLoadBalancerDrift(desired, actual []LoadBalancer, report *DriftReport) {
	desiredMap := mapByKey(desired, func(lb LoadBalancer) string { return loadBalancerKey(lb) })
	actualMap := mapByKey(actual, func(lb LoadBalancer) string { return loadBalancerKey(lb) })

	for key, desiredLB := range desiredMap {
		actualLB, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "load_balancer",
				ResourceID:   desiredLB.ID,
				ResourceName: key,
				Field:        "existence",
				Expected:     "exists",
				Actual:       "missing",
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Load balancer %s is missing from infrastructure", key),
			})
			continue
		}

		if desiredLB.Protocol != "" && actualLB.Protocol != "" && desiredLB.Protocol != actualLB.Protocol {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "load_balancer",
				ResourceID:   actualLB.ID,
				ResourceName: key,
				Field:        "protocol",
				Expected:     desiredLB.Protocol,
				Actual:       actualLB.Protocol,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Load balancer %s has the wrong protocol", key),
			})
		}

		if desiredLB.Port != 0 && actualLB.Port != 0 && desiredLB.Port != actualLB.Port {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "load_balancer",
				ResourceID:   actualLB.ID,
				ResourceName: key,
				Field:        "port",
				Expected:     desiredLB.Port,
				Actual:       actualLB.Port,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Load balancer %s listens on the wrong port", key),
			})
		}

		if desiredLB.VIP != "" && actualLB.VIP != "" && desiredLB.VIP != actualLB.VIP {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "load_balancer",
				ResourceID:   actualLB.ID,
				ResourceName: key,
				Field:        "vip",
				Expected:     desiredLB.VIP,
				Actual:       actualLB.VIP,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Load balancer %s has the wrong VIP", key),
			})
		}

		if len(desiredLB.Members) > 0 && !equalStringSets(desiredLB.Members, actualLB.Members) {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "load_balancer",
				ResourceID:   actualLB.ID,
				ResourceName: key,
				Field:        "members",
				Expected:     sortedCopy(desiredLB.Members),
				Actual:       sortedCopy(actualLB.Members),
				Severity:     SeverityWarning,
				Reconcilable: true,
				Message:      fmt.Sprintf("Load balancer %s has an unexpected backend member set", key),
			})
		}
	}

	for key, actualLB := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "load_balancer",
			ResourceID:   actualLB.ID,
			ResourceName: key,
			Field:        "existence",
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected load balancer %s found in infrastructure", key),
		})
	}
}

func detectVolumeDrift(desired, actual []Volume, report *DriftReport) {
	desiredMap := mapByKey(desired, func(v Volume) string { return volumeKey(v) })
	actualMap := mapByKey(actual, func(v Volume) string { return volumeKey(v) })

	for key, desiredVolume := range desiredMap {
		actualVolume, exists := actualMap[key]
		if !exists {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "volume",
				ResourceID:   desiredVolume.ID,
				ResourceName: key,
				Field:        "existence",
				Expected:     "exists",
				Actual:       "missing",
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Volume %s is missing from infrastructure", key),
			})
			continue
		}

		if desiredVolume.Size > 0 && actualVolume.Size > 0 && desiredVolume.Size != actualVolume.Size {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "volume",
				ResourceID:   actualVolume.ID,
				ResourceName: key,
				Field:        "size",
				Expected:     desiredVolume.Size,
				Actual:       actualVolume.Size,
				Severity:     SeverityWarning,
				Reconcilable: false,
				Message:      fmt.Sprintf("Volume %s has the wrong size", key),
			})
		}

		if desiredVolume.Status != "" && actualVolume.Status != "" && !strings.EqualFold(desiredVolume.Status, actualVolume.Status) {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "volume",
				ResourceID:   actualVolume.ID,
				ResourceName: key,
				Field:        "status",
				Expected:     desiredVolume.Status,
				Actual:       actualVolume.Status,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Volume %s has the wrong status", key),
			})
		}

		if desiredVolume.AttachedTo != "" && actualVolume.AttachedTo != "" && desiredVolume.AttachedTo != actualVolume.AttachedTo {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "volume",
				ResourceID:   actualVolume.ID,
				ResourceName: key,
				Field:        "attached_to",
				Expected:     desiredVolume.AttachedTo,
				Actual:       actualVolume.AttachedTo,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Volume %s is attached to the wrong server", key),
			})
		}
	}

	for key, actualVolume := range actualMap {
		if _, exists := desiredMap[key]; exists {
			continue
		}

		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "volume",
			ResourceID:   actualVolume.ID,
			ResourceName: key,
			Field:        "existence",
			Expected:     "not exists",
			Actual:       "exists",
			Severity:     SeverityWarning,
			Reconcilable: false,
			Message:      fmt.Sprintf("Unexpected volume %s found in infrastructure", key),
		})
	}
}

func detectFloatingIPDrift(desired, actual []FloatingIP, report *DriftReport) {
	if len(desired) == 0 && len(actual) == 0 {
		return
	}

	if len(desired) != len(actual) {
		report.Drifts = append(report.Drifts, DriftItem{
			ResourceType: "floating_ip",
			ResourceName: "cluster-floating-ips",
			Field:        "count",
			Expected:     len(desired),
			Actual:       len(actual),
			Severity:     SeverityCritical,
			Reconcilable: false,
			Message:      "Floating IP count does not match the desired configuration",
		})
	}

	desiredCopy := append([]FloatingIP(nil), desired...)
	actualCopy := append([]FloatingIP(nil), actual...)
	sort.Slice(desiredCopy, func(i, j int) bool { return floatingIPKey(desiredCopy[i]) < floatingIPKey(desiredCopy[j]) })
	sort.Slice(actualCopy, func(i, j int) bool { return floatingIPKey(actualCopy[i]) < floatingIPKey(actualCopy[j]) })

	for i := 0; i < len(desiredCopy) && i < len(actualCopy); i++ {
		key := floatingIPKey(actualCopy[i])
		if key == "" {
			key = fmt.Sprintf("floating-ip-%d", i+1)
		}

		if desiredCopy[i].Address != "" && actualCopy[i].Address != "" && desiredCopy[i].Address != actualCopy[i].Address {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "floating_ip",
				ResourceID:   actualCopy[i].ID,
				ResourceName: key,
				Field:        "address",
				Expected:     desiredCopy[i].Address,
				Actual:       actualCopy[i].Address,
				Severity:     SeverityWarning,
				Reconcilable: false,
				Message:      fmt.Sprintf("Floating IP %s has the wrong address", key),
			})
		}

		if desiredCopy[i].AttachedTo != "" && actualCopy[i].AttachedTo != "" && desiredCopy[i].AttachedTo != actualCopy[i].AttachedTo {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "floating_ip",
				ResourceID:   actualCopy[i].ID,
				ResourceName: key,
				Field:        "attached_to",
				Expected:     desiredCopy[i].AttachedTo,
				Actual:       actualCopy[i].AttachedTo,
				Severity:     SeverityCritical,
				Reconcilable: false,
				Message:      fmt.Sprintf("Floating IP %s is attached to the wrong target", key),
			})
		}

		if desiredCopy[i].Status != "" && actualCopy[i].Status != "" && !strings.EqualFold(desiredCopy[i].Status, actualCopy[i].Status) {
			report.Drifts = append(report.Drifts, DriftItem{
				ResourceType: "floating_ip",
				ResourceID:   actualCopy[i].ID,
				ResourceName: key,
				Field:        "status",
				Expected:     desiredCopy[i].Status,
				Actual:       actualCopy[i].Status,
				Severity:     SeverityWarning,
				Reconcilable: false,
				Message:      fmt.Sprintf("Floating IP %s has the wrong status", key),
			})
		}
	}
}

func compareSecurityRules(desired, actual []SecurityRule) ([]SecurityRule, []SecurityRule) {
	actualSet := make(map[string]struct{}, len(actual))
	for _, rule := range actual {
		actualSet[normalizeSecurityRule(rule)] = struct{}{}
	}

	missing := make([]SecurityRule, 0)
	for _, rule := range desired {
		if _, ok := actualSet[normalizeSecurityRule(rule)]; !ok {
			missing = append(missing, rule)
		}
	}

	desiredSet := make(map[string]struct{}, len(desired))
	for _, rule := range desired {
		desiredSet[normalizeSecurityRule(rule)] = struct{}{}
	}

	riskyExtra := make([]SecurityRule, 0)
	for _, rule := range actual {
		key := normalizeSecurityRule(rule)
		if _, ok := desiredSet[key]; ok {
			continue
		}
		if isHighRiskRule(rule) {
			riskyExtra = append(riskyExtra, rule)
		}
	}

	return missing, riskyExtra
}

func normalizeSecurityRule(rule SecurityRule) string {
	return strings.ToLower(strings.Join([]string{
		strings.TrimSpace(rule.Direction),
		strings.TrimSpace(rule.Protocol),
		strings.TrimSpace(rule.PortRange),
		strings.TrimSpace(rule.RemoteIP),
	}, "|"))
}

func hasHighRiskRules(rules []SecurityRule) bool {
	for _, rule := range rules {
		if isHighRiskRule(rule) {
			return true
		}
	}
	return false
}

func isHighRiskRule(rule SecurityRule) bool {
	if !strings.EqualFold(rule.Direction, "ingress") {
		return false
	}

	remote := strings.TrimSpace(rule.RemoteIP)
	if remote != "0.0.0.0/0" && remote != "::/0" {
		return false
	}

	portRange := strings.TrimSpace(rule.PortRange)
	if portRange == "" || portRange == "*" {
		return true
	}

	if strings.Contains(portRange, "-") {
		parts := strings.SplitN(portRange, "-", 2)
		if len(parts) != 2 {
			return false
		}
		minPort, errMin := strconv.Atoi(parts[0])
		maxPort, errMax := strconv.Atoi(parts[1])
		if errMin != nil || errMax != nil {
			return false
		}
		return minPort <= 6443 && 6443 <= maxPort
	}

	port, err := strconv.Atoi(portRange)
	if err != nil {
		return false
	}
	return port == 6443
}

func missingServerSeverity(server Server) Severity {
	role := strings.ToLower(server.Tags["role"])
	if role == "control-plane" || role == "master" {
		return SeverityCritical
	}
	return SeverityWarning
}

func serverKey(server Server) string {
	if server.Name != "" {
		return server.Name
	}
	return server.ID
}

func networkKey(network Network) string {
	if network.Name != "" {
		return network.Name
	}
	return network.ID
}

func subnetKey(subnet Subnet) string {
	if subnet.Name != "" {
		return subnet.Name
	}
	if subnet.CIDR != "" {
		return subnet.CIDR
	}
	return subnet.ID
}

func securityGroupKey(group SecurityGroup) string {
	if group.Name != "" {
		return group.Name
	}
	return group.ID
}

func loadBalancerKey(lb LoadBalancer) string {
	if lb.Name != "" {
		return lb.Name
	}
	return lb.ID
}

func volumeKey(volume Volume) string {
	if volume.AttachedTo != "" {
		return "attached:" + volume.AttachedTo
	}
	if volume.Name != "" {
		return "name:" + volume.Name
	}
	if volume.ID != "" {
		return "id:" + volume.ID
	}
	return ""
}

func floatingIPKey(fip FloatingIP) string {
	if fip.AttachedTo != "" {
		return "attached:" + fip.AttachedTo
	}
	if fip.Address != "" {
		return "address:" + fip.Address
	}
	return fip.ID
}

func equalStringSets(expected, actual []string) bool {
	return strings.Join(sortedCopy(expected), "\x00") == strings.Join(sortedCopy(actual), "\x00")
}

func sortedCopy(values []string) []string {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return copyValues
}

func mapByKey[T any](items []T, keyFn func(T) string) map[string]T {
	result := make(map[string]T, len(items))
	for _, item := range items {
		key := keyFn(item)
		if key == "" {
			continue
		}
		result[key] = item
	}
	return result
}
