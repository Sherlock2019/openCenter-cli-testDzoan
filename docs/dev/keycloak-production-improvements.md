# Keycloak Production Improvements

Implementation of production-ready Keycloak configuration with enhanced security, performance, and operational features.

## Changes Summary

### Configuration Enhancements

**New fields in `KeycloakConfig`** (`internal/config/services/keycloak.go`):

1. **Production Mode**
   - `start_optimized`: Enable production-optimized startup (default: true)
   - `cache_enabled`: Enable distributed caching (default: true)
   - `cache_stack`: Cache implementation (kubernetes/ispn, default: ispn)

2. **Resource Management**
   - `resource_requests_cpu`: CPU requests (default: "2")
   - `resource_requests_memory`: Memory requests (default: "1250M")
   - `resource_limits_cpu`: CPU limits (default: "6")
   - `resource_limits_memory`: Memory limits (default: "2250M")

3. **High Availability**
   - `instances`: Static replica count (default: 3)
   - `min_replicas`: Autoscaling minimum (default: 3)
   - `max_replicas`: Autoscaling maximum (default: 10)

4. **Database Tuning**
   - `db_pool_min_size`: Minimum connections (default: 30)
   - `db_pool_initial_size`: Initial connections (default: 30)
   - `db_pool_max_size`: Maximum connections (default: 30)

5. **Monitoring**
   - `metrics_enabled`: Prometheus metrics (default: true)
   - `event_metrics_enabled`: User event metrics (default: true)
   - `health_enabled`: Health endpoints (default: true)
   - `log_level`: Logging level (default: INFO)
   - `log_format`: Log format (default: json)

6. **Security**
   - `tls_enabled`: Enable TLS (default: true)
   - `tls_secret_name`: Certificate secret (default: keycloak-tls-secret)

7. **Realm Management**
   - `realm_import_enabled`: Enable realm import (default: true)
   - `realm_groups`: Additional realm groups
   - `realm_admin_email`: Admin user email

8. **Backup**
   - `backup_enabled`: Enable automated backups (default: true)
   - `backup_schedule`: Cron schedule (default: "0 2 * * *")

### Template Updates

**`keycloak-cr-patch.yaml.tpl`** - Enhanced Keycloak CR:
- Production mode enabled by default
- Proper resource limits (2-6 CPU, 1250M-2250M memory)
- Database connection pool configuration
- Comprehensive monitoring options
- Pod topology spread constraints for multi-AZ distribution
- Distributed caching with Infinispan

**`keycloak-hpa.yaml.tpl`** - New HorizontalPodAutoscaler:
- CPU-based scaling (80% threshold)
- Memory-based scaling (85% threshold)
- Configurable min/max replicas
- Stabilization windows to prevent flapping
- Smart scale-up/scale-down policies

**`keycloak-backup-cronjob.yaml.tpl`** - New backup automation:
- Daily realm configuration backups
- RBAC for backup service account
- Exports KeycloakRealmImport and secrets
- Extensible for object storage integration

**`kustomization.yaml.tpl`** - New kustomization file:
- Conditionally includes HPA when autoscaling configured
- Conditionally includes backup CronJob when enabled
- Proper resource ordering

### Security Improvements

**`opencenter-realm.yaml`** - Hardened realm configuration:
- Stronger password policy (14 chars, 2 upper, 2 lower, 2 digits, 2 special)
- Password history enforcement (5 previous passwords)
- Enhanced brute force protection settings
- PKCE enforcement with S256 (SHA-256)
- Cluster-specific redirect URIs (no wildcards)
- Cluster-specific web origins
- Dynamic realm groups from configuration
- Configurable admin email

### Validation Enhancements

**`internal/services/plugins/keycloak.go`** - Enhanced validation:
- Production mode requires >= 2 instances for HA
- Autoscaling min/max validation
- Database pool size validation
- HTTPS enforcement in production mode
- Log level validation (INFO, DEBUG, WARN, ERROR, TRACE)
- Log format validation (default, json)
- Cache stack validation (kubernetes, ispn)

### Testing Updates

**`internal/services/plugins/plugins_test.go`** - New test cases:
- Production mode HA requirement validation
- Autoscaling configuration validation
- Database pool configuration validation
- Existing tests updated with required fields

**`internal/config/service_rendering_test.go`** - Updated expectations:
- Field count increased from 14 to 35+
- All new fields included in validation

## Migration Guide

### Existing Configurations

Existing configurations continue to work with secure defaults:

```yaml
services:
  keycloak:
    enabled: true
    # All new fields use production-ready defaults
```

### Recommended Production Configuration

```yaml
services:
  keycloak:
    enabled: true
    hostname: auth.prod.example.com
    
    # Production mode (defaults are secure)
    start_optimized: true
    cache_enabled: true
    
    # High availability with autoscaling
    instances: 3
    min_replicas: 3
    max_replicas: 10
    
    # Realm customization
    realm_groups:
      - developers
      - qa-team
    realm_admin_email: admin@example.com
    
    # Monitoring
    log_level: INFO
    log_format: json
```

### Development Configuration

For development/testing environments:

```yaml
services:
  keycloak:
    enabled: true
    start_optimized: false  # Development mode
    instances: 1            # Single instance
    resource_requests_cpu: "500m"
    resource_requests_memory: "1Gi"
```

## Deployment

### Build and Validate

```bash
# Build with new configuration
mise run build

# Validate configuration
./bin/opencenter cluster validate my-cluster

# Initialize new cluster
./bin/opencenter cluster init my-cluster

# Setup GitOps repository
./bin/opencenter cluster setup my-cluster
```

### Verify Generated Manifests

```bash
# Check Keycloak CR
cat ~/.config/opencenter/gitops/my-cluster/applications/overlays/my-cluster/services/keycloak/20-keycloak/keycloak-cr-patch.yaml

# Check HPA (if autoscaling enabled)
cat ~/.config/opencenter/gitops/my-cluster/applications/overlays/my-cluster/services/keycloak/20-keycloak/keycloak-hpa.yaml

# Check backup CronJob
cat ~/.config/opencenter/gitops/my-cluster/applications/overlays/my-cluster/services/keycloak/20-keycloak/keycloak-backup-cronjob.yaml

# Check realm configuration
cat ~/.config/opencenter/gitops/my-cluster/applications/overlays/my-cluster/services/keycloak/20-keycloak/opencenter-realm.yaml
```

## Monitoring

### Metrics

Access Keycloak metrics:

```bash
kubectl port-forward -n keycloak svc/keycloak 8080:8080
curl http://localhost:8080/metrics | grep keycloak
```

### Health Checks

```bash
kubectl get keycloak -n keycloak
kubectl get pods -n keycloak -l app=keycloak
kubectl logs -n keycloak -l app=keycloak --tail=100
```

### Autoscaling Status

```bash
kubectl get hpa -n keycloak keycloak-hpa
kubectl describe hpa -n keycloak keycloak-hpa
```

### Backup Status

```bash
kubectl get cronjob -n keycloak keycloak-realm-backup
kubectl get jobs -n keycloak -l app.kubernetes.io/name=keycloak-backup
```

## Performance Considerations

### Resource Sizing

Based on Keycloak documentation and production experience:

- **Small deployments** (< 1000 users): 2 CPU / 1250M memory
- **Medium deployments** (1000-10000 users): 4 CPU / 2250M memory  
- **Large deployments** (> 10000 users): 6+ CPU / 4Gi+ memory

### Database Connection Pools

Default pool size (30 connections) is suitable for most deployments:

- 3 instances × 10 connections = 30 total connections
- Adjust based on concurrent authentication load
- Monitor connection usage in PostgreSQL

### Caching

Infinispan distributed cache (default) provides:
- Session replication across instances
- Better performance than Kubernetes cache
- Required for production HA deployments

## Security Considerations

### Password Policy

New default policy enforces:
- Minimum 14 characters
- 2 uppercase, 2 lowercase, 2 digits, 2 special characters
- Cannot contain username
- 5 password history

### PKCE Enforcement

All OAuth clients now require PKCE with SHA-256:
- Prevents authorization code interception
- Required for mobile and SPA applications
- Compliant with OAuth 2.1 draft

### Redirect URI Restrictions

Wildcard redirects removed:
- Only cluster-specific URIs allowed
- Prevents open redirect vulnerabilities
- Follows OAuth security best practices

### TLS

HTTPS enforced in production mode:
- TLS certificate from cert-manager
- HTTP disabled by default
- Proxy headers configured for edge termination

## Troubleshooting

### Common Issues

**Pods not starting**:
- Check resource availability: `kubectl describe pod -n keycloak`
- Reduce resource requests if needed

**Database connection errors**:
- Verify PostgreSQL max_connections setting
- Check connection pool configuration
- Review Keycloak logs for connection errors

**Realm import failures**:
- Check KeycloakRealmImport status
- Verify realm YAML syntax
- Review operator logs

**Backup failures**:
- Check CronJob and Job status
- Verify RBAC permissions
- Review backup pod logs

## Related Documentation

- [Keycloak Production Configuration Guide](.kiro/steering/keycloak-production-configuration.md)
- [Service Registry Patterns](.kiro/steering/service-registry-patterns.md)
- [GitOps Manifest Standards](.kiro/steering/gitops-manifest-standards.md)
- [Keycloak Operator Documentation](https://www.keycloak.org/operator/installation)

## References

- Keycloak High Availability Guide
- Keycloak Performance Tuning Guide
- OAuth 2.0 Security Best Practices (RFC 8252)
- PKCE for OAuth Public Clients (RFC 7636)
