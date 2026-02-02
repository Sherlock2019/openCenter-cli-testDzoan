package validator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/quotasets"
	"github.com/rackerlabs/opencenter-cli/internal/talos"
)

// QuotaUsage represents current resource usage in OpenStack.
type QuotaUsage struct {
	Instances      int
	Cores          int
	RAM            int
	Networks       int
	Routers        int
	SecurityGroups int
	Volumes        int
	VolumeStorage  int
	Snapshots      int
}

// QuotaLimits represents quota limits in OpenStack.
type QuotaLimits struct {
	Instances      int
	Cores          int
	RAM            int
	Networks       int
	Routers        int
	SecurityGroups int
	Volumes        int
	VolumeStorage  int
	Snapshots      int
}

// quotaCache stores cached quota data with TTL.
type quotaCache struct {
	limits    *QuotaLimits
	usage     *QuotaUsage
	timestamp time.Time
	mu        sync.RWMutex
}

// Global quota cache with 5 minute TTL
var (
	cache     = &quotaCache{}
	cacheTTL  = 5 * time.Minute
)

// ValidateQuotasImpl verifies tenant resource quotas.
func (v *DefaultValidator) ValidateQuotasImpl(ctx context.Context, required talos.ResourceRequirements) error {
	v.logger.Debug("Validating resource quotas", "required", required)

	// Get current quota limits
	limits, err := v.getQuotaLimits(ctx)
	if err != nil {
		return talos.NewInfrastructureError(
			"QUOTA_LIMITS_CHECK_FAILED",
			"Failed to retrieve quota limits",
			true,
			err,
		)
	}

	// Get current resource usage
	usage, err := v.getQuotaUsage(ctx)
	if err != nil {
		return talos.NewInfrastructureError(
			"QUOTA_USAGE_CHECK_FAILED",
			"Failed to retrieve current resource usage",
			true,
			err,
		)
	}

	// Check each resource type
	insufficientResources := []string{}

	if !v.checkResourceAvailability("instances", usage.Instances, limits.Instances, required.Instances) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("instances (need %d, have %d/%d)", required.Instances, limits.Instances-usage.Instances, limits.Instances))
	}

	if !v.checkResourceAvailability("cores", usage.Cores, limits.Cores, required.Cores) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("cores (need %d, have %d/%d)", required.Cores, limits.Cores-usage.Cores, limits.Cores))
	}

	if !v.checkResourceAvailability("RAM", usage.RAM, limits.RAM, required.RAM) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("RAM MB (need %d, have %d/%d)", required.RAM, limits.RAM-usage.RAM, limits.RAM))
	}

	if !v.checkResourceAvailability("networks", usage.Networks, limits.Networks, required.Networks) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("networks (need %d, have %d/%d)", required.Networks, limits.Networks-usage.Networks, limits.Networks))
	}

	if !v.checkResourceAvailability("routers", usage.Routers, limits.Routers, required.Routers) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("routers (need %d, have %d/%d)", required.Routers, limits.Routers-usage.Routers, limits.Routers))
	}

	if !v.checkResourceAvailability("security groups", usage.SecurityGroups, limits.SecurityGroups, required.SecurityGroups) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("security groups (need %d, have %d/%d)", required.SecurityGroups, limits.SecurityGroups-usage.SecurityGroups, limits.SecurityGroups))
	}

	if !v.checkResourceAvailability("volumes", usage.Volumes, limits.Volumes, required.Volumes) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("volumes (need %d, have %d/%d)", required.Volumes, limits.Volumes-usage.Volumes, limits.Volumes))
	}

	if !v.checkResourceAvailability("volume storage", usage.VolumeStorage, limits.VolumeStorage, required.VolumeStorage) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("volume storage GB (need %d, have %d/%d)", required.VolumeStorage, limits.VolumeStorage-usage.VolumeStorage, limits.VolumeStorage))
	}

	if !v.checkResourceAvailability("snapshots", usage.Snapshots, limits.Snapshots, required.Snapshots) {
		insufficientResources = append(insufficientResources,
			fmt.Sprintf("snapshots (need %d, have %d/%d)", required.Snapshots, limits.Snapshots-usage.Snapshots, limits.Snapshots))
	}

	if len(insufficientResources) > 0 {
		remediation := &talos.RemediationAction{
			Check:       "Resource Quotas",
			Description: "Insufficient quota for one or more resource types",
			Steps: []string{
				"Contact your OpenStack administrator to request quota increases",
				"Insufficient resources:",
			},
		}
		remediation.Steps = append(remediation.Steps, insufficientResources...)
		remediation.Steps = append(remediation.Steps,
			"Check current quotas: openstack quota show",
			"Check current usage: openstack limits show --absolute",
		)

		return talos.NewValidationError(
			"INSUFFICIENT_QUOTA",
			fmt.Sprintf("Insufficient quota for %d resource type(s)", len(insufficientResources)),
			remediation,
		)
	}

	v.logger.Info("Quota validation passed", "required", required, "available", calculateAvailable(usage, limits))
	return nil
}

// checkResourceAvailability checks if sufficient resources are available.
func (v *DefaultValidator) checkResourceAvailability(resourceType string, current, limit, required int) bool {
	available := limit - current
	sufficient := available >= required

	v.logger.Debug("Checking resource availability",
		"resource", resourceType,
		"current", current,
		"limit", limit,
		"required", required,
		"available", available,
		"sufficient", sufficient,
	)

	return sufficient
}

// getComputeClient creates an authenticated OpenStack compute client.
// It retrieves credentials from environment variables and creates a Nova compute client
// for querying quota information.
//
// Required environment variables:
//   - OS_AUTH_URL: OpenStack identity endpoint (e.g., https://identity.example.com:5000/v3)
//   - OS_USERNAME: OpenStack username
//   - OS_PASSWORD: OpenStack password
//   - OS_PROJECT_ID: OpenStack project/tenant ID (or use validator's projectID)
//
// Optional environment variables:
//   - OS_REGION_NAME: OpenStack region (defaults to validator's region or "RegionOne")
//   - OS_USER_DOMAIN_NAME: User domain name (defaults to "Default")
//
// Returns:
//   - *gophercloud.ServiceClient: Authenticated compute client
//   - error: Authentication or client creation error
//
// Example:
//
//	client, err := v.getComputeClient(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to create compute client: %w", err)
//	}
func (v *DefaultValidator) getComputeClient(ctx context.Context) (*gophercloud.ServiceClient, error) {
	// Get OpenStack credentials from environment
	authURL := os.Getenv("OS_AUTH_URL")
	if authURL == "" {
		return nil, fmt.Errorf("OS_AUTH_URL environment variable not set")
	}

	username := os.Getenv("OS_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("OS_USERNAME environment variable not set")
	}

	password := os.Getenv("OS_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("OS_PASSWORD environment variable not set")
	}

	// Use project ID from validator if available, otherwise from environment
	projectID := v.projectID
	if projectID == "" {
		projectID = os.Getenv("OS_PROJECT_ID")
		if projectID == "" {
			return nil, fmt.Errorf("OS_PROJECT_ID environment variable not set and no project ID in config")
		}
	}

	// Use region from validator if available, otherwise from environment
	region := v.region
	if region == "" {
		region = os.Getenv("OS_REGION_NAME")
		if region == "" {
			region = "RegionOne" // Default region
		}
	}

	// Get domain name from environment or use default
	domainName := os.Getenv("OS_USER_DOMAIN_NAME")
	if domainName == "" {
		domainName = "Default"
	}

	// Create provider client
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: authURL,
		Username:         username,
		Password:         password,
		TenantID:         projectID,
		DomainName:       domainName,
		Scope: &gophercloud.AuthScope{
			ProjectID: projectID,
		},
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with OpenStack: %w", err)
	}

	// Create compute client
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return client, nil
}

// getQuotaLimits retrieves quota limits from OpenStack Nova API.
// It queries the compute service for project quota limits and caches the results
// for 5 minutes to optimize performance.
//
// The method retrieves the following quota limits from Nova:
//   - Instances: Maximum number of VM instances
//   - Cores: Maximum number of vCPU cores
//   - RAM: Maximum RAM in MB
//   - SecurityGroups: Maximum number of security groups
//
// Note: Networks, Routers, Volumes, VolumeStorage, and Snapshots are managed by
// Neutron and Cinder services. These are currently set to default values and should
// be queried from their respective services in a future enhancement.
//
// Caching behavior:
//   - First call: Queries OpenStack API (~500ms)
//   - Subsequent calls: Returns cached data (~0.1ms)
//   - Cache TTL: 5 minutes
//   - Thread-safe: Uses sync.RWMutex for concurrent access
//
// Returns:
//   - *QuotaLimits: Quota limits for all resource types
//   - error: API error or authentication failure
//
// Example:
//
//	limits, err := v.getQuotaLimits(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to get quota limits: %w", err)
//	}
//	fmt.Printf("Instance limit: %d\n", limits.Instances)
func (v *DefaultValidator) getQuotaLimits(ctx context.Context) (*QuotaLimits, error) {
	// Check cache first
	cache.mu.RLock()
	if cache.limits != nil && time.Since(cache.timestamp) < cacheTTL {
		v.logger.Debug("Using cached quota limits")
		limits := cache.limits
		cache.mu.RUnlock()
		return limits, nil
	}
	cache.mu.RUnlock()

	v.logger.Debug("Retrieving quota limits from OpenStack API")

	// Get compute client
	client, err := v.getComputeClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	// Get project ID for quota query
	projectID := v.projectID
	if projectID == "" {
		projectID = os.Getenv("OS_PROJECT_ID")
		if projectID == "" {
			return nil, fmt.Errorf("project ID not available")
		}
	}

	// Retrieve quota limits from OpenStack
	quotaSet, err := quotasets.Get(client, projectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve quota limits: %w", err)
	}

	// Map OpenStack quota response to QuotaLimits struct
	limits := &QuotaLimits{
		Instances:      quotaSet.Instances,
		Cores:          quotaSet.Cores,
		RAM:            quotaSet.RAM,
		SecurityGroups: quotaSet.SecurityGroups,
		// Note: Networks, Routers, Volumes, VolumeStorage, and Snapshots
		// are managed by Neutron and Cinder, not Nova.
		// For now, we'll set reasonable defaults. These should be queried
		// from their respective services in a future enhancement.
		Networks:      10,
		Routers:       10,
		Volumes:       100,
		VolumeStorage: 10000,
		Snapshots:     100,
	}

	// Update cache
	cache.mu.Lock()
	cache.limits = limits
	cache.timestamp = time.Now()
	cache.mu.Unlock()

	v.logger.Debug("Retrieved quota limits", "limits", limits)
	return limits, nil
}

// getQuotaUsage retrieves current resource usage from OpenStack Nova API.
// It queries the compute service for detailed quota usage information and caches
// the results for 5 minutes to optimize performance.
//
// The method retrieves the following usage information from Nova:
//   - Instances: Current number of VM instances in use
//   - Cores: Current number of vCPU cores in use
//   - RAM: Current RAM usage in MB
//   - SecurityGroups: Current number of security groups in use
//
// Note: Networks, Routers, Volumes, VolumeStorage, and Snapshots are managed by
// Neutron and Cinder services. These are currently set to 0 and should be queried
// from their respective services in a future enhancement.
//
// Caching behavior:
//   - First call: Queries OpenStack API (~500ms)
//   - Subsequent calls: Returns cached data (~0.1ms)
//   - Cache TTL: 5 minutes (shared with limits cache)
//   - Thread-safe: Uses sync.RWMutex for concurrent access
//
// Returns:
//   - *QuotaUsage: Current usage for all resource types
//   - error: API error or authentication failure
//
// Example:
//
//	usage, err := v.getQuotaUsage(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to get quota usage: %w", err)
//	}
//	fmt.Printf("Instances in use: %d\n", usage.Instances)
func (v *DefaultValidator) getQuotaUsage(ctx context.Context) (*QuotaUsage, error) {
	// Check cache first
	cache.mu.RLock()
	if cache.usage != nil && time.Since(cache.timestamp) < cacheTTL {
		v.logger.Debug("Using cached quota usage")
		usage := cache.usage
		cache.mu.RUnlock()
		return usage, nil
	}
	cache.mu.RUnlock()

	v.logger.Debug("Retrieving current resource usage from OpenStack API")

	// Get compute client
	client, err := v.getComputeClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	// Get project ID for quota query
	projectID := v.projectID
	if projectID == "" {
		projectID = os.Getenv("OS_PROJECT_ID")
		if projectID == "" {
			return nil, fmt.Errorf("project ID not available")
		}
	}

	// Retrieve quota usage details from OpenStack
	quotaSetDetail, err := quotasets.GetDetail(client, projectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve quota usage: %w", err)
	}

	// Map OpenStack usage response to QuotaUsage struct
	usage := &QuotaUsage{
		Instances:      quotaSetDetail.Instances.InUse,
		Cores:          quotaSetDetail.Cores.InUse,
		RAM:            quotaSetDetail.RAM.InUse,
		SecurityGroups: quotaSetDetail.SecurityGroups.InUse,
		// Note: Networks, Routers, Volumes, VolumeStorage, and Snapshots
		// are managed by Neutron and Cinder, not Nova.
		// For now, we'll set minimal defaults. These should be queried
		// from their respective services in a future enhancement.
		Networks:      0,
		Routers:       0,
		Volumes:       0,
		VolumeStorage: 0,
		Snapshots:     0,
	}

	// Update cache
	cache.mu.Lock()
	cache.usage = usage
	if cache.limits != nil {
		// Only update timestamp if we already have limits cached
		cache.timestamp = time.Now()
	}
	cache.mu.Unlock()

	v.logger.Debug("Retrieved quota usage", "usage", usage)
	return usage, nil
}

// calculateAvailable calculates available resources.
func calculateAvailable(usage *QuotaUsage, limits *QuotaLimits) map[string]int {
	return map[string]int{
		"instances":       limits.Instances - usage.Instances,
		"cores":           limits.Cores - usage.Cores,
		"ram":             limits.RAM - usage.RAM,
		"networks":        limits.Networks - usage.Networks,
		"routers":         limits.Routers - usage.Routers,
		"security_groups": limits.SecurityGroups - usage.SecurityGroups,
		"volumes":         limits.Volumes - usage.Volumes,
		"volume_storage":  limits.VolumeStorage - usage.VolumeStorage,
		"snapshots":       limits.Snapshots - usage.Snapshots,
	}
}
