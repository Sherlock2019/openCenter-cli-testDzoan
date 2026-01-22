# Metrics Package

The metrics package provides performance tracking and metrics collection for opencenter operations. It tracks template rendering times, configuration building times, and GitOps generation times to help monitor and optimize system performance.

## Features

- **Automatic Metrics Collection**: Integrated into template engine, config builder, and GitOps generator
- **Performance Statistics**: Calculate min, max, average, and percentile durations
- **Success/Failure Tracking**: Monitor operation success rates
- **Metadata Support**: Attach custom metadata to metrics for detailed analysis
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Global Collector**: Convenient global instance for application-wide metrics

## Metric Types

The package tracks the following metric types:

- `MetricTypeTemplateRender`: Template rendering operations
- `MetricTypeConfigBuild`: Configuration building operations
- `MetricTypeGitOpsGeneration`: GitOps repository generation operations
- `MetricTypeValidation`: Validation operations
- `MetricTypeMigration`: Configuration migration operations

## Usage

### Basic Metrics Collection

```go
import "github.com/rackerlabs/opencenter-cli/internal/util/metrics"

// Create a metrics collector
collector := metrics.NewMetricsCollector()

// Record a metric manually
collector.RecordMetric(metrics.Metric{
    Type:     metrics.MetricTypeTemplateRender,
    Name:     "cluster-config.yaml.tmpl",
    Duration: 50 * time.Millisecond,
    Success:  true,
})

// Get all metrics
allMetrics := collector.GetMetrics()

// Get metrics by type
templateMetrics := collector.GetMetricsByType(metrics.MetricTypeTemplateRender)
```

### Using Timers

Timers provide a convenient way to measure operation duration:

```go
// Start a timer
timer := collector.NewTimer(metrics.MetricTypeTemplateRender, "example-template.tmpl")

// Perform your operation
result, err := renderTemplate(...)

// Stop the timer (automatically records the metric)
if err != nil {
    timer.StopWithError(err)
} else {
    timer.Stop()
}
```

### Adding Metadata

Attach custom metadata to metrics for detailed analysis:

```go
timer := collector.NewTimer(metrics.MetricTypeGitOpsGeneration, "production-cluster").
    WithMetadata("provider", "openstack").
    WithMetadata("region", "us-east-1").
    WithMetadata("files_generated", 42)

timer.Stop()
```

### Getting Performance Statistics

```go
// Get summary with statistics
summary := collector.GetSummary()

// Access statistics by metric type
templateSummary := summary.ByType[metrics.MetricTypeTemplateRender]

fmt.Printf("Count: %d\n", templateSummary.Count)
fmt.Printf("Success Rate: %.2f%%\n", 
    float64(templateSummary.SuccessCount) / float64(templateSummary.Count) * 100)
fmt.Printf("Average Duration: %v\n", templateSummary.AverageDuration)
fmt.Printf("P95 Duration: %v\n", templateSummary.P95Duration)
fmt.Printf("P99 Duration: %v\n", templateSummary.P99Duration)
```

### Using the Global Collector

For convenience, a global metrics collector is available:

```go
// Record metrics using convenience functions
metrics.RecordTemplateRender("template.tmpl", 100*time.Millisecond, true, nil)
metrics.RecordConfigBuild("cluster-name", 200*time.Millisecond, true, nil)
metrics.RecordGitOpsGeneration("cluster-name", 5*time.Second, 142, true, nil)

// Access the global collector
globalCollector := metrics.GetGlobalCollector()
summary := globalCollector.GetSummary()
```

### Enabling/Disabling Metrics

Metrics collection can be enabled or disabled at runtime:

```go
collector := metrics.NewMetricsCollector()

// Disable metrics collection (e.g., for production)
collector.SetEnabled(false)

// Re-enable when needed
collector.SetEnabled(true)
```

## Integration

The metrics package is automatically integrated into:

### Template Engine

Template rendering operations are automatically tracked:

```go
engine := template.NewGoTemplateEngine()
result, err := engine.Render(ctx, "template.tmpl", data)
// Metric is automatically recorded with duration and success status
```

### Configuration Builder

Configuration building operations are automatically tracked:

```go
builder := config.NewConfigBuilder("cluster-name")
cfg, err := builder.
    WithProvider("openstack").
    WithOrganization("my-org").
    Build()
// Metric is automatically recorded with duration and success status
```

### GitOps Generator

GitOps generation operations are automatically tracked:

```go
generator := gitops.NewPipelineGenerator(workspaceManager, stages)
err := generator.Generate(ctx, cfg)
// Metric is automatically recorded with duration, files generated, and success status
```

## Performance Statistics

The package calculates the following statistics for each metric type:

- **Count**: Total number of operations
- **Success Count**: Number of successful operations
- **Failure Count**: Number of failed operations
- **Total Duration**: Sum of all operation durations
- **Average Duration**: Mean operation duration
- **Min Duration**: Fastest operation
- **Max Duration**: Slowest operation
- **P50 Duration**: Median operation duration (50th percentile)
- **P95 Duration**: 95th percentile duration
- **P99 Duration**: 99th percentile duration

## Thread Safety

All operations on the metrics collector are thread-safe and can be called concurrently from multiple goroutines.

## Best Practices

1. **Use Timers**: Prefer using timers over manual metric recording for automatic duration tracking
2. **Add Metadata**: Include relevant metadata for better analysis and debugging
3. **Monitor P95/P99**: Focus on percentile metrics rather than averages for performance monitoring
4. **Clear Periodically**: Clear metrics periodically in long-running applications to prevent memory growth
5. **Disable in Production**: Consider disabling metrics collection in production if not needed

## Example Output

```
Metrics Summary (Total: 150)
================================

template_render:
  Count: 100 (Success: 98, Failure: 2)
  Total Duration: 15s
  Average Duration: 150ms
  Min Duration: 50ms
  Max Duration: 500ms
  P50 Duration: 140ms
  P95 Duration: 300ms
  P99 Duration: 450ms

config_build:
  Count: 25 (Success: 25, Failure: 0)
  Total Duration: 5s
  Average Duration: 200ms
  Min Duration: 150ms
  Max Duration: 350ms
  P50 Duration: 190ms
  P95 Duration: 320ms
  P99 Duration: 340ms

gitops_generation:
  Count: 25 (Success: 24, Failure: 1)
  Total Duration: 125s
  Average Duration: 5s
  Min Duration: 3s
  Max Duration: 12s
  P50 Duration: 4.8s
  P95 Duration: 8.5s
  P99 Duration: 11s
```

## Requirements

This package satisfies **Requirement 9.6** from the configuration system refactor specification:

> THE System SHALL provide performance metrics for optimization analysis

The metrics package provides comprehensive performance tracking for:
- Template rendering times
- Configuration building times
- GitOps generation times

These metrics enable performance analysis, optimization, and monitoring of the opencenter system.
