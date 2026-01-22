// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics_test

import (
	"fmt"
	"time"

	"github.com/rackerlabs/opencenter-cli/internal/util/metrics"
)

// ExampleMetricsCollector demonstrates basic metrics collection usage.
func ExampleMetricsCollector() {
	// Create a new metrics collector
	collector := metrics.NewMetricsCollector()

	// Record some template rendering metrics
	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeTemplateRender,
		Name:     "cluster-config.yaml.tmpl",
		Duration: 50 * time.Millisecond,
		Success:  true,
	})

	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeTemplateRender,
		Name:     "service-manifest.yaml.tmpl",
		Duration: 75 * time.Millisecond,
		Success:  true,
	})

	// Record a configuration build metric
	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeConfigBuild,
		Name:     "production-cluster",
		Duration: 200 * time.Millisecond,
		Success:  true,
	})

	// Get summary of collected metrics
	summary := collector.GetSummary()
	fmt.Printf("Total metrics collected: %d\n", summary.TotalMetrics)
	fmt.Printf("Template render count: %d\n", summary.ByType[metrics.MetricTypeTemplateRender].Count)
	fmt.Printf("Config build count: %d\n", summary.ByType[metrics.MetricTypeConfigBuild].Count)

	// Output:
	// Total metrics collected: 3
	// Template render count: 2
	// Config build count: 1
}

// ExampleTimer demonstrates using a timer for automatic metric recording.
func ExampleTimer() {
	collector := metrics.NewMetricsCollector()

	// Start a timer for a template rendering operation
	timer := collector.NewTimer(metrics.MetricTypeTemplateRender, "example-template.tmpl")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Stop the timer (automatically records the metric)
	duration := timer.Stop()

	fmt.Printf("Operation took: %v\n", duration >= 10*time.Millisecond)

	// Check that metric was recorded
	metrics := collector.GetMetrics()
	fmt.Printf("Metrics recorded: %d\n", len(metrics))

	// Output:
	// Operation took: true
	// Metrics recorded: 1
}

// ExampleTimer_WithMetadata demonstrates adding metadata to metrics.
func ExampleTimer_WithMetadata() {
	collector := metrics.NewMetricsCollector()

	// Start a timer with metadata
	timer := collector.NewTimer(metrics.MetricTypeGitOpsGeneration, "production-cluster").
		WithMetadata("provider", "openstack").
		WithMetadata("region", "us-east-1").
		WithMetadata("files_generated", 42)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Stop the timer
	timer.Stop()

	// Retrieve the metric and check metadata
	metrics := collector.GetMetrics()
	if len(metrics) > 0 {
		fmt.Printf("Provider: %v\n", metrics[0].Metadata["provider"])
		fmt.Printf("Files generated: %v\n", metrics[0].Metadata["files_generated"])
	}

	// Output:
	// Provider: openstack
	// Files generated: 42
}

// ExampleMetricsCollector_GetSummary demonstrates getting performance statistics.
func ExampleMetricsCollector_GetSummary() {
	collector := metrics.NewMetricsCollector()

	// Record multiple metrics with varying durations
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		150 * time.Millisecond,
		300 * time.Millisecond,
		250 * time.Millisecond,
	}

	for i, duration := range durations {
		collector.RecordMetric(metrics.Metric{
			Type:     metrics.MetricTypeTemplateRender,
			Name:     fmt.Sprintf("template-%d", i),
			Duration: duration,
			Success:  true,
		})
	}

	// Get summary statistics
	summary := collector.GetSummary()
	templateSummary := summary.ByType[metrics.MetricTypeTemplateRender]

	fmt.Printf("Count: %d\n", templateSummary.Count)
	fmt.Printf("Average: %v\n", templateSummary.AverageDuration)
	fmt.Printf("Min: %v\n", templateSummary.MinDuration)
	fmt.Printf("Max: %v\n", templateSummary.MaxDuration)

	// Output:
	// Count: 5
	// Average: 200ms
	// Min: 100ms
	// Max: 300ms
}

// ExampleRecordTemplateRender demonstrates the convenience function for template metrics.
func ExampleRecordTemplateRender() {
	// Clear global collector for clean example
	metrics.GetGlobalCollector().Clear()

	// Record a successful template render
	metrics.RecordTemplateRender(
		"cluster-config.yaml.tmpl",
		150*time.Millisecond,
		true,
		nil,
	)

	// Record a failed template render
	metrics.RecordTemplateRender(
		"invalid-template.tmpl",
		50*time.Millisecond,
		false,
		fmt.Errorf("template syntax error"),
	)

	// Get metrics from global collector
	allMetrics := metrics.GetGlobalCollector().GetMetrics()
	fmt.Printf("Total metrics: %d\n", len(allMetrics))

	// Get only template render metrics
	templateMetrics := metrics.GetGlobalCollector().GetMetricsByType(metrics.MetricTypeTemplateRender)
	fmt.Printf("Template metrics: %d\n", len(templateMetrics))

	// Check success/failure counts
	successCount := 0
	failureCount := 0
	for _, m := range templateMetrics {
		if m.Success {
			successCount++
		} else {
			failureCount++
		}
	}
	fmt.Printf("Successful: %d, Failed: %d\n", successCount, failureCount)

	// Output:
	// Total metrics: 2
	// Template metrics: 2
	// Successful: 1, Failed: 1
}

// ExampleRecordGitOpsGeneration demonstrates recording GitOps generation metrics.
func ExampleRecordGitOpsGeneration() {
	// Clear global collector for clean example
	metrics.GetGlobalCollector().Clear()

	// Record a successful GitOps generation
	metrics.RecordGitOpsGeneration(
		"production-cluster",
		5*time.Second,
		142, // files generated
		true,
		nil,
	)

	// Get the metric
	allMetrics := metrics.GetGlobalCollector().GetMetrics()
	if len(allMetrics) > 0 {
		metric := allMetrics[0]
		fmt.Printf("Cluster: %v\n", metric.Metadata["cluster_name"])
		fmt.Printf("Files: %v\n", metric.Metadata["files_generated"])
		fmt.Printf("Duration: %v\n", metric.Duration)
		fmt.Printf("Success: %v\n", metric.Success)
	}

	// Output:
	// Cluster: production-cluster
	// Files: 142
	// Duration: 5s
	// Success: true
}

// ExampleMetricsCollector_SetEnabled demonstrates enabling/disabling metrics collection.
func ExampleMetricsCollector_SetEnabled() {
	collector := metrics.NewMetricsCollector()

	// Record a metric while enabled
	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeTemplateRender,
		Name:     "template1",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})

	fmt.Printf("Metrics with collection enabled: %d\n", len(collector.GetMetrics()))

	// Disable metrics collection
	collector.SetEnabled(false)

	// Try to record another metric (won't be recorded)
	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeTemplateRender,
		Name:     "template2",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})

	fmt.Printf("Metrics after disabling: %d\n", len(collector.GetMetrics()))

	// Re-enable and record
	collector.SetEnabled(true)
	collector.RecordMetric(metrics.Metric{
		Type:     metrics.MetricTypeTemplateRender,
		Name:     "template3",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})

	fmt.Printf("Metrics after re-enabling: %d\n", len(collector.GetMetrics()))

	// Output:
	// Metrics with collection enabled: 1
	// Metrics after disabling: 1
	// Metrics after re-enabling: 2
}
