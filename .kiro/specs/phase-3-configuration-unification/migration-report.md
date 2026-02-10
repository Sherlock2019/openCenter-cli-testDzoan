# Configuration Migration Report

## Summary

- **Total files to migrate**: 19
- **Files using config.Load**: 15
- **Files using config.Save**: 5
- **Files using config.Validate**: 3

## Files Using config.Load

- [ ] `cmd/cluster_status.go`
- [ ] `internal/cluster/setup_service.go`
- [ ] `cmd/cluster_lock.go`
- [ ] `cmd/cluster_select.go`
- [ ] `internal/config/persistence.go`
- [ ] `cmd/cluster_preflight.go`
- [ ] `cmd/config_helpers.go`
- [ ] `cmd/secrets.go`
- [ ] `internal/cluster/bootstrap_service.go`
- [ ] `tests/features/steps/helpers.go`
- [ ] `cmd/cluster_config.go`
- [ ] `cmd/cluster_validate_manifests.go`
- [ ] `cmd/cluster_credentials_export.go`
- [ ] `cmd/cluster_env.go`
- [ ] `cmd/cluster_info.go`

## Files Using config.Save

- [ ] `cmd/cluster_destroy.go`
- [ ] `cmd/cluster_service.go`
- [ ] `cmd/cluster_update.go`
- [ ] `internal/config/persistence.go`
- [ ] `tests/features/steps/helpers.go`

## Files Using config.Validate

- [ ] `cmd/cluster_info.go`
- [ ] `cmd/cluster_update.go`
- [ ] `internal/cluster/init_service.go`

## Migration Checklist

### Command Layer (cmd/)

- [ ] cmd/cluster_init.go
- [ ] cmd/cluster_validate.go
- [ ] cmd/cluster_setup.go
- [ ] cmd/cluster_bootstrap.go
- [ ] cmd/cluster_list.go
- [ ] cmd/config_*.go files

### Service Layer (internal/cluster/)

- [ ] internal/cluster/init_service.go
- [ ] internal/cluster/validate_service.go
- [ ] internal/cluster/setup_service.go
- [ ] internal/cluster/bootstrap_service.go

### GitOps Layer (internal/gitops/)

- [ ] internal/gitops/generator.go
- [ ] internal/gitops/workspace.go
- [ ] internal/gitops/pipeline.go

### SOPS Layer (internal/sops/)

- [ ] internal/sops/manager.go
- [ ] internal/sops/git.go

## Migration Instructions

### Replace config.Load

```go
// Before
config, err := config.Load(clusterName)

// After
config, err := manager.Load(ctx, clusterName)
```

### Replace config.Save

```go
// Before
err := config.Save(cfg)

// After
err := manager.Save(ctx, cfg)
```

### Replace config.Validate

```go
// Before
err := config.Validate(cfg)

// After
err := manager.Validate(ctx, cfg)
```

## Notes

- All operations now require a `context.Context` parameter
- ConfigurationManager must be injected via dependency injection
- Test files will be updated alongside their source files
- Run tests after each file migration to ensure correctness
