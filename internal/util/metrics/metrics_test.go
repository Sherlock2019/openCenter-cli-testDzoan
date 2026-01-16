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

package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	assert.NotNil(t, collector)
	assert.True(t, collector.IsEnabled())
	assert.Empty(t, collector.GetMetrics())
}

func TestMetricsCollector_SetEnabled(t *testing.T) {
	collector := NewMetricsCollector()

	// Initially enabled
	assert.True(t, collector.IsEnabled())

	// Disable
	collector.SetEnabled(false)
	assert.False(t, collector.IsEnabled())

	// Re-enable
	collector.SetEnabled(true)
	assert.True(t, collector.IsEnabled())
}

func TestMetricsCollector_RecordMetric(t *testing.T) {
	collector := NewMetricsCollector()

	metric := Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "test-template",
		Duration: 100 * time.Millisecond,
		Success:  true,
	}

	collector.RecordMetric(metric)

	metrics := collector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeTemplateRender, metrics[0].Type)
	assert.Equal(t, "test-template", metrics[0].Name)
	assert.Equal(t, 100*time.Millisecond, metrics[0].Duration)
	assert.True(t, metrics[0].Success)
	assert.False(t, metrics[0].Timestamp.IsZero())
}

func TestMetricsCollector_RecordMetric_WhenDisabled(t *testing.T) {
	collector := NewMetricsCollector()
	collector.SetEnabled(false)

	metric := Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "test-template",
		Duration: 100 * time.Millisecond,
		Success:  true,
	}

	collector.RecordMetric(metric)

	// Metric should not be recorded when disabled
	metrics := collector.GetMetrics()
	assert.Empty(t, metrics)
}

func TestMetricsCollector_GetMetricsByType(t *testing.T) {
	collector := NewMetricsCollector()

	// Record metrics of different types
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template1",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})
	collector.RecordMetric(Metric{
		Type:     MetricTypeConfigBuild,
		Name:     "config1",
		Duration: 200 * time.Millisecond,
		Success:  true,
	})
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template2",
		Duration: 150 * time.Millisecond,
		Success:  true,
	})

	// Get template render metrics
	templateMetrics := collector.GetMetricsByType(MetricTypeTemplateRender)
	assert.Len(t, templateMetrics, 2)
	assert.Equal(t, "template1", templateMetrics[0].Name)
	assert.Equal(t, "template2", templateMetrics[1].Name)

	// Get config build metrics
	configMetrics := collector.GetMetricsByType(MetricTypeConfigBuild)
	assert.Len(t, configMetrics, 1)
	assert.Equal(t, "config1", configMetrics[0].Name)
}

func TestMetricsCollector_Clear(t *testing.T) {
	collector := NewMetricsCollector()

	// Record some metrics
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template1",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})
	collector.RecordMetric(Metric{
		Type:     MetricTypeConfigBuild,
		Name:     "config1",
		Duration: 200 * time.Millisecond,
		Success:  true,
	})

	assert.Len(t, collector.GetMetrics(), 2)

	// Clear metrics
	collector.Clear()
	assert.Empty(t, collector.GetMetrics())
}

func TestMetricsCollector_GetSummary(t *testing.T) {
	collector := NewMetricsCollector()

	// Record metrics with varying durations
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template1",
		Duration: 100 * time.Millisecond,
		Success:  true,
	})
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template2",
		Duration: 200 * time.Millisecond,
		Success:  true,
	})
	collector.RecordMetric(Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "template3",
		Duration: 300 * time.Millisecond,
		Success:  false,
		Error:    errors.New("test error"),
	})

	summary := collector.GetSummary()

	assert.Equal(t, 3, summary.TotalMetrics)
	assert.Contains(t, summary.ByType, MetricTypeTemplateRender)

	templateSummary := summary.ByType[MetricTypeTemplateRender]
	assert.Equal(t, 3, templateSummary.Count)
	assert.Equal(t, 2, templateSummary.SuccessCount)
	assert.Equal(t, 1, templateSummary.FailureCount)
	assert.Equal(t, 600*time.Millisecond, templateSummary.TotalDuration)
	assert.Equal(t, 200*time.Millisecond, templateSummary.AverageDuration)
	assert.Equal(t, 100*time.Millisecond, templateSummary.MinDuration)
	assert.Equal(t, 300*time.Millisecond, templateSummary.MaxDuration)
}

func TestTimer_Stop(t *testing.T) {
	collector := NewMetricsCollector()

	timer := collector.NewTimer(MetricTypeTemplateRender, "test-template")
	time.Sleep(50 * time.Millisecond)
	duration := timer.Stop()

	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)

	metrics := collector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeTemplateRender, metrics[0].Type)
	assert.Equal(t, "test-template", metrics[0].Name)
	assert.True(t, metrics[0].Success)
	assert.GreaterOrEqual(t, metrics[0].Duration, 50*time.Millisecond)
}

func TestTimer_StopWithError(t *testing.T) {
	collector := NewMetricsCollector()

	testErr := errors.New("test error")
	timer := collector.NewTimer(MetricTypeConfigBuild, "test-config")
	time.Sleep(50 * time.Millisecond)
	duration := timer.StopWithError(testErr)

	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)

	metrics := collector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeConfigBuild, metrics[0].Type)
	assert.Equal(t, "test-config", metrics[0].Name)
	assert.False(t, metrics[0].Success)
	assert.Equal(t, testErr, metrics[0].Error)
	assert.GreaterOrEqual(t, metrics[0].Duration, 50*time.Millisecond)
}

func TestTimer_WithMetadata(t *testing.T) {
	collector := NewMetricsCollector()

	timer := collector.NewTimer(MetricTypeGitOpsGeneration, "test-cluster").
		WithMetadata("files_generated", 42).
		WithMetadata("provider", "openstack")

	timer.Stop()

	metrics := collector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, 42, metrics[0].Metadata["files_generated"])
	assert.Equal(t, "openstack", metrics[0].Metadata["provider"])
}

func TestRecordTemplateRender(t *testing.T) {
	// Clear global collector
	globalCollector.Clear()

	RecordTemplateRender("test-template.tmpl", 100*time.Millisecond, true, nil)

	metrics := globalCollector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeTemplateRender, metrics[0].Type)
	assert.Equal(t, "test-template.tmpl", metrics[0].Name)
	assert.Equal(t, 100*time.Millisecond, metrics[0].Duration)
	assert.True(t, metrics[0].Success)
	assert.Equal(t, "test-template.tmpl", metrics[0].Metadata["template_path"])
}

func TestRecordConfigBuild(t *testing.T) {
	// Clear global collector
	globalCollector.Clear()

	RecordConfigBuild("test-cluster", 200*time.Millisecond, true, nil)

	metrics := globalCollector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeConfigBuild, metrics[0].Type)
	assert.Equal(t, "test-cluster", metrics[0].Name)
	assert.Equal(t, 200*time.Millisecond, metrics[0].Duration)
	assert.True(t, metrics[0].Success)
	assert.Equal(t, "test-cluster", metrics[0].Metadata["cluster_name"])
}

func TestRecordGitOpsGeneration(t *testing.T) {
	// Clear global collector
	globalCollector.Clear()

	RecordGitOpsGeneration("test-cluster", 5*time.Second, 42, true, nil)

	metrics := globalCollector.GetMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, MetricTypeGitOpsGeneration, metrics[0].Type)
	assert.Equal(t, "test-cluster", metrics[0].Name)
	assert.Equal(t, 5*time.Second, metrics[0].Duration)
	assert.True(t, metrics[0].Success)
	assert.Equal(t, "test-cluster", metrics[0].Metadata["cluster_name"])
	assert.Equal(t, 42, metrics[0].Metadata["files_generated"])
}

func TestCalculateTypeSummary_EmptyMetrics(t *testing.T) {
	summary := calculateTypeSummary([]Metric{})
	assert.Equal(t, 0, summary.Count)
	assert.Equal(t, time.Duration(0), summary.TotalDuration)
}

func TestCalculateTypeSummary_SingleMetric(t *testing.T) {
	metrics := []Metric{
		{
			Type:     MetricTypeTemplateRender,
			Duration: 100 * time.Millisecond,
			Success:  true,
		},
	}

	summary := calculateTypeSummary(metrics)
	assert.Equal(t, 1, summary.Count)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, 0, summary.FailureCount)
	assert.Equal(t, 100*time.Millisecond, summary.TotalDuration)
	assert.Equal(t, 100*time.Millisecond, summary.AverageDuration)
	assert.Equal(t, 100*time.Millisecond, summary.MinDuration)
	assert.Equal(t, 100*time.Millisecond, summary.MaxDuration)
}

func TestCalculateTypeSummary_MultipleMetrics(t *testing.T) {
	metrics := []Metric{
		{Duration: 100 * time.Millisecond, Success: true},
		{Duration: 200 * time.Millisecond, Success: true},
		{Duration: 300 * time.Millisecond, Success: false},
		{Duration: 400 * time.Millisecond, Success: true},
		{Duration: 500 * time.Millisecond, Success: true},
	}

	summary := calculateTypeSummary(metrics)
	assert.Equal(t, 5, summary.Count)
	assert.Equal(t, 4, summary.SuccessCount)
	assert.Equal(t, 1, summary.FailureCount)
	assert.Equal(t, 1500*time.Millisecond, summary.TotalDuration)
	assert.Equal(t, 300*time.Millisecond, summary.AverageDuration)
	assert.Equal(t, 100*time.Millisecond, summary.MinDuration)
	assert.Equal(t, 500*time.Millisecond, summary.MaxDuration)
	assert.Equal(t, 300*time.Millisecond, summary.P50Duration)
	assert.Equal(t, 500*time.Millisecond, summary.P95Duration)
	assert.Equal(t, 500*time.Millisecond, summary.P99Duration)
}

func TestPercentile(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
	}

	tests := []struct {
		name       string
		percentile int
		expected   time.Duration
	}{
		{"P0", 0, 100 * time.Millisecond},
		{"P50", 50, 300 * time.Millisecond},
		{"P95", 95, 500 * time.Millisecond},
		{"P99", 99, 500 * time.Millisecond},
		{"P100", 100, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentile(durations, tt.percentile)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	result := percentile([]time.Duration{}, 50)
	assert.Equal(t, time.Duration(0), result)
}

func TestSortDurations(t *testing.T) {
	durations := []time.Duration{
		500 * time.Millisecond,
		100 * time.Millisecond,
		300 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}

	sortDurations(durations)

	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
	}

	assert.Equal(t, expected, durations)
}

func TestMetricsSummary_String(t *testing.T) {
	summary := MetricsSummary{
		TotalMetrics: 3,
		ByType: map[MetricType]TypeSummary{
			MetricTypeTemplateRender: {
				Count:           3,
				SuccessCount:    2,
				FailureCount:    1,
				TotalDuration:   600 * time.Millisecond,
				AverageDuration: 200 * time.Millisecond,
				MinDuration:     100 * time.Millisecond,
				MaxDuration:     300 * time.Millisecond,
				P50Duration:     200 * time.Millisecond,
				P95Duration:     300 * time.Millisecond,
				P99Duration:     300 * time.Millisecond,
			},
		},
	}

	str := summary.String()
	assert.Contains(t, str, "Metrics Summary")
	assert.Contains(t, str, "Total: 3")
	assert.Contains(t, str, "template_render")
	assert.Contains(t, str, "Count: 3")
	assert.Contains(t, str, "Success: 2")
	assert.Contains(t, str, "Failure: 1")
	assert.Contains(t, str, "Average Duration")
}

func TestGetGlobalCollector(t *testing.T) {
	collector := GetGlobalCollector()
	assert.NotNil(t, collector)
	assert.Same(t, globalCollector, collector)
}

// Benchmark tests
func BenchmarkMetricsCollector_RecordMetric(b *testing.B) {
	collector := NewMetricsCollector()
	metric := Metric{
		Type:     MetricTypeTemplateRender,
		Name:     "test-template",
		Duration: 100 * time.Millisecond,
		Success:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordMetric(metric)
	}
}

func BenchmarkMetricsCollector_GetSummary(b *testing.B) {
	collector := NewMetricsCollector()

	// Pre-populate with metrics
	for i := 0; i < 1000; i++ {
		collector.RecordMetric(Metric{
			Type:     MetricTypeTemplateRender,
			Name:     "test-template",
			Duration: time.Duration(i) * time.Millisecond,
			Success:  true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.GetSummary()
	}
}

func BenchmarkTimer_Stop(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := collector.NewTimer(MetricTypeTemplateRender, "test-template")
		timer.Stop()
	}
}
