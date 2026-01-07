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

package flags

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestPerformance_ProcessingTimeWithHundredsOfFlags tests processing time with hundreds of flags
// Validates: Requirements 12.1
func TestPerformance_ProcessingTimeWithHundredsOfFlags(t *testing.T) {
	processor := NewEnhancedFlagProcessor()
	processor.SetProgressEnabled(false) // Disable progress for testing

	// Create test cases with different flag counts
	testCases := []struct {
		name      string
		flagCount int
		maxTime   time.Duration
	}{
		{"100 flags", 100, 2 * time.Second},
		{"200 flags", 200, 3 * time.Second},
		{"500 flags", 500, 5 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate test flags
			flags := generateTestFlags(tc.flagCount)

			// Measure processing time
			startTime := time.Now()
			result, err := processor.ProcessFlags(flags)
			processingTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Failed to process flags: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Verify processing time is within acceptable limits
			if processingTime > tc.maxTime {
				t.Errorf("Processing time %v exceeds maximum allowed time %v for %d flags",
					processingTime, tc.maxTime, tc.flagCount)
			}

			t.Logf("Processed %d flags in %v", tc.flagCount, processingTime)
		})
	}
}

// TestPerformance_MemoryUsageWithLargeConfigurations tests memory usage with large configurations
// Validates: Requirements 12.2
func TestPerformance_MemoryUsageWithLargeConfigurations(t *testing.T) {
	processor := NewEnhancedFlagProcessor()
	processor.SetProgressEnabled(false) // Disable progress for testing

	// Test cases with different configuration sizes
	testCases := []struct {
		name        string
		configSize  int
		maxMemoryMB int64
	}{
		{"Small config (1KB)", 1024, 10},
		{"Medium config (100KB)", 100 * 1024, 50},
		{"Large config (1MB)", 1024 * 1024, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Measure memory before processing
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Generate large configuration
			flags := generateLargeConfigurationFlags(tc.configSize)

			// Process configuration
			result, err := processor.ProcessFlags(flags)
			if err != nil {
				t.Fatalf("Failed to process large configuration: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Measure memory after processing
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Calculate memory usage in MB
			memoryUsedMB := int64(memAfter.Alloc-memBefore.Alloc) / (1024 * 1024)

			// Verify memory usage is within acceptable limits
			if memoryUsedMB > tc.maxMemoryMB {
				t.Errorf("Memory usage %d MB exceeds maximum allowed %d MB for config size %d bytes",
					memoryUsedMB, tc.maxMemoryMB, tc.configSize)
			}

			t.Logf("Processed %d byte configuration using %d MB memory", tc.configSize, memoryUsedMB)
		})
	}
}

// TestPerformance_ProgressIndicatorFunctionality tests progress indicator functionality
// Validates: Requirements 12.3
func TestPerformance_ProgressIndicatorFunctionality(t *testing.T) {
	// Test progress indicator with different total values
	testCases := []struct {
		name       string
		total      int64
		updates    int
		updateStep int64
	}{
		{"Small progress", 100, 10, 10},
		{"Medium progress", 1000, 20, 50},
		{"Large progress", 10000, 50, 200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			progress := NewProgressIndicator(fmt.Sprintf("Test progress %s", tc.name), tc.total)

			// Start progress indicator
			startTime := time.Now()
			progress.Start()

			// Simulate progress updates
			for i := 0; i < tc.updates; i++ {
				current := int64(i+1) * tc.updateStep
				if current > tc.total {
					current = tc.total
				}
				progress.Update(current)

				// Small delay to simulate work
				time.Sleep(10 * time.Millisecond)
			}

			// Stop progress indicator
			progress.Stop()
			totalTime := time.Since(startTime)

			// Verify progress indicator ran for reasonable time
			expectedMinTime := time.Duration(tc.updates) * 10 * time.Millisecond
			if totalTime < expectedMinTime {
				t.Errorf("Progress indicator ran for %v, expected at least %v", totalTime, expectedMinTime)
			}

			t.Logf("Progress indicator for %s completed in %v", tc.name, totalTime)
		})
	}
}

// TestPerformance_StreamingProcessorWithLargeJSON tests streaming processor with large JSON
func TestPerformance_StreamingProcessorWithLargeJSON(t *testing.T) {
	processor := NewStreamingJSONProcessor()

	// Test cases with different JSON sizes
	testCases := []struct {
		name     string
		jsonSize int
		maxTime  time.Duration
	}{
		{"Small JSON (1KB)", 1024, 100 * time.Millisecond},
		{"Medium JSON (100KB)", 100 * 1024, 500 * time.Millisecond},
		{"Large JSON (500KB)", 500 * 1024, 2 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate large JSON string
			jsonStr := generateLargeJSONString(tc.jsonSize)

			// Measure processing time
			startTime := time.Now()
			result, err := processor.ProcessJSONString(jsonStr)
			processingTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Failed to process JSON: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Verify processing time is within acceptable limits
			if processingTime > tc.maxTime {
				t.Errorf("JSON processing time %v exceeds maximum allowed time %v for size %d bytes",
					processingTime, tc.maxTime, tc.jsonSize)
			}

			t.Logf("Processed %d byte JSON in %v", tc.jsonSize, processingTime)
		})
	}
}

// TestPerformance_StreamingProcessorWithLargeYAML tests streaming processor with large YAML
func TestPerformance_StreamingProcessorWithLargeYAML(t *testing.T) {
	processor := NewStreamingYAMLProcessor()

	// Test cases with different YAML sizes
	testCases := []struct {
		name     string
		yamlSize int
		maxTime  time.Duration
	}{
		{"Small YAML (1KB)", 1024, 100 * time.Millisecond},
		{"Medium YAML (100KB)", 100 * 1024, 500 * time.Millisecond},
		{"Large YAML (500KB)", 500 * 1024, 2 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate large YAML string
			yamlStr := generateLargeYAMLString(tc.yamlSize)

			// Measure processing time
			startTime := time.Now()
			result, err := processor.ProcessYAMLString(yamlStr)
			processingTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Failed to process YAML: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Verify processing time is within acceptable limits
			if processingTime > tc.maxTime {
				t.Errorf("YAML processing time %v exceeds maximum allowed time %v for size %d bytes",
					processingTime, tc.maxTime, tc.yamlSize)
			}

			t.Logf("Processed %d byte YAML in %v", tc.yamlSize, processingTime)
		})
	}
}

// TestPerformance_ConfigurationSizeAnalyzer tests configuration size analyzer
func TestPerformance_ConfigurationSizeAnalyzer(t *testing.T) {
	analyzer := NewConfigurationSizeAnalyzer()

	// Analyze different types of flags
	testFlags := []struct {
		name     string
		value    string
		flagType FlagType
	}{
		{"small-flag", "value", FlagTypeDotNotation},
		{"json-flag", generateLargeJSONString(1024), FlagTypeJSON},
		{"yaml-flag", generateLargeYAMLString(2048), FlagTypeYAML},
		{"large-flag", strings.Repeat("data", 500), FlagTypeDotNotation},
	}

	// Analyze all flags
	startTime := time.Now()
	for _, flag := range testFlags {
		analyzer.AnalyzeFlag(flag.name, flag.value, flag.flagType)
	}
	analysisTime := time.Since(startTime)

	// Get analysis results
	analysis := analyzer.GetAnalysis()

	// Verify analysis results
	if analysis.TotalFlags != len(testFlags) {
		t.Errorf("Expected %d total flags, got %d", len(testFlags), analysis.TotalFlags)
	}

	if analysis.JSONFlags != 1 {
		t.Errorf("Expected 1 JSON flag, got %d", analysis.JSONFlags)
	}

	if analysis.YAMLFlags != 1 {
		t.Errorf("Expected 1 YAML flag, got %d", analysis.YAMLFlags)
	}

	if analysis.TotalSize == 0 {
		t.Error("Total size should be greater than 0")
	}

	if analysis.AverageSize == 0 {
		t.Error("Average size should be greater than 0")
	}

	// Verify analysis time is reasonable
	maxAnalysisTime := 100 * time.Millisecond
	if analysisTime > maxAnalysisTime {
		t.Errorf("Analysis time %v exceeds maximum allowed time %v", analysisTime, maxAnalysisTime)
	}

	t.Logf("Analyzed %d flags in %v: %+v", len(testFlags), analysisTime, analysis)
}

// TestPerformance_MemoryOptimizer tests memory optimizer functionality
func TestPerformance_MemoryOptimizer(t *testing.T) {
	optimizer := NewMemoryOptimizer()

	// Test memory optimization with different workloads
	testCases := []struct {
		name     string
		workload func() error
		maxTime  time.Duration
	}{
		{
			name: "Light workload",
			workload: func() error {
				// Simulate light processing
				data := make([]byte, 1024)
				for i := range data {
					data[i] = byte(i % 256)
				}
				return nil
			},
			maxTime: 100 * time.Millisecond,
		},
		{
			name: "Medium workload",
			workload: func() error {
				// Simulate medium processing
				data := make([]byte, 100*1024)
				for i := range data {
					data[i] = byte(i % 256)
				}
				return nil
			},
			maxTime: 500 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Measure memory before optimization
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Run workload with memory optimization
			startTime := time.Now()
			err := optimizer.OptimizeMemoryUsage(tc.workload)
			processingTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Memory optimization failed: %v", err)
			}

			// Measure memory after optimization
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Verify processing time is reasonable
			if processingTime > tc.maxTime {
				t.Errorf("Processing time %v exceeds maximum allowed time %v", processingTime, tc.maxTime)
			}

			// Get memory usage info
			memInfo := optimizer.CheckMemoryUsage()
			if len(memInfo.Recommendations) == 0 {
				t.Error("Memory optimizer should provide recommendations")
			}

			t.Logf("Memory optimization for %s completed in %v", tc.name, processingTime)
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkEnhancedFlagProcessor_ProcessFlags benchmarks flag processing
func BenchmarkEnhancedFlagProcessor_ProcessFlags(b *testing.B) {
	processor := NewEnhancedFlagProcessor()
	processor.SetProgressEnabled(false) // Disable progress for benchmarking

	// Generate test flags
	flags := generateTestFlags(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessFlags(flags)
		if err != nil {
			b.Fatalf("Failed to process flags: %v", err)
		}
	}
}

// BenchmarkStreamingJSONProcessor_ProcessJSONString benchmarks JSON processing
func BenchmarkStreamingJSONProcessor_ProcessJSONString(b *testing.B) {
	processor := NewStreamingJSONProcessor()
	jsonStr := generateLargeJSONString(10 * 1024) // 10KB JSON

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessJSONString(jsonStr)
		if err != nil {
			b.Fatalf("Failed to process JSON: %v", err)
		}
	}
}

// BenchmarkStreamingYAMLProcessor_ProcessYAMLString benchmarks YAML processing
func BenchmarkStreamingYAMLProcessor_ProcessYAMLString(b *testing.B) {
	processor := NewStreamingYAMLProcessor()
	yamlStr := generateLargeYAMLString(10 * 1024) // 10KB YAML

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessYAMLString(yamlStr)
		if err != nil {
			b.Fatalf("Failed to process YAML: %v", err)
		}
	}
}

// BenchmarkConfigurationSizeAnalyzer_AnalyzeFlag benchmarks flag analysis
func BenchmarkConfigurationSizeAnalyzer_AnalyzeFlag(b *testing.B) {
	analyzer := NewConfigurationSizeAnalyzer()
	flagValue := strings.Repeat("test-data", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeFlag(fmt.Sprintf("flag-%d", i), flagValue, FlagTypeDotNotation)
	}
}

// Helper functions for generating test data

// generateTestFlags generates test flags for performance testing
func generateTestFlags(count int) *ParsedFlags {
	flags := &ParsedFlags{
		DotNotation:   make(map[string]string),
		ArrayFlags:    []ArrayFlag{},
		JSONFlags:     []JSONFlag{},
		YAMLFlags:     []YAMLFlag{},
		TemplateVars:  make(map[string]string),
		ConfigFiles:   []ConfigFile{},
		SecurityFlags: []SecurityFlag{},
	}

	// Generate dot notation flags
	for i := 0; i < count/2; i++ {
		key := fmt.Sprintf("config.section%d.key%d", i/10, i%10)
		value := fmt.Sprintf("value-%d", i)
		flags.DotNotation[key] = value
	}

	// Generate JSON flags
	for i := 0; i < count/4; i++ {
		jsonFlag := JSONFlag{
			Path:  fmt.Sprintf("json.config%d", i),
			Value: fmt.Sprintf(`{"key%d": "value%d", "number": %d}`, i, i, i),
		}
		flags.JSONFlags = append(flags.JSONFlags, jsonFlag)
	}

	// Generate YAML flags
	for i := 0; i < count/4; i++ {
		yamlFlag := YAMLFlag{
			Path:  fmt.Sprintf("yaml.config%d", i),
			Value: fmt.Sprintf("key%d: value%d\nnumber: %d\n", i, i, i),
		}
		flags.YAMLFlags = append(flags.YAMLFlags, yamlFlag)
	}

	return flags
}

// generateLargeConfigurationFlags generates large configuration flags
func generateLargeConfigurationFlags(totalSize int) *ParsedFlags {
	flags := &ParsedFlags{
		DotNotation:   make(map[string]string),
		ArrayFlags:    []ArrayFlag{},
		JSONFlags:     []JSONFlag{},
		YAMLFlags:     []YAMLFlag{},
		TemplateVars:  make(map[string]string),
		ConfigFiles:   []ConfigFile{},
		SecurityFlags: []SecurityFlag{},
	}

	// Generate large JSON flag
	jsonSize := totalSize / 2
	jsonFlag := JSONFlag{
		Path:  "large.json.config",
		Value: generateLargeJSONString(jsonSize),
	}
	flags.JSONFlags = append(flags.JSONFlags, jsonFlag)

	// Generate large YAML flag
	yamlSize := totalSize / 2
	yamlFlag := YAMLFlag{
		Path:  "large.yaml.config",
		Value: generateLargeYAMLString(yamlSize),
	}
	flags.YAMLFlags = append(flags.YAMLFlags, yamlFlag)

	return flags
}

// generateLargeJSONString generates a large JSON string of approximately the specified size
func generateLargeJSONString(size int) string {
	// Calculate how many entries we need
	entrySize := 50 // Approximate size per entry
	entryCount := size / entrySize

	var builder strings.Builder
	builder.WriteString("{")

	for i := 0; i < entryCount; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`"key%d": "value%d-data-content"`, i, i))
	}

	builder.WriteString("}")
	return builder.String()
}

// generateLargeYAMLString generates a large YAML string of approximately the specified size
func generateLargeYAMLString(size int) string {
	// Calculate how many entries we need
	entrySize := 30 // Approximate size per entry
	entryCount := size / entrySize

	var builder strings.Builder

	for i := 0; i < entryCount; i++ {
		builder.WriteString(fmt.Sprintf("key%d: value%d-data-content\n", i, i))
	}

	return builder.String()
}

// Integration performance tests

// TestPerformance_EndToEndProcessing tests end-to-end processing performance
func TestPerformance_EndToEndProcessing(t *testing.T) {
	processor := NewConfigurationProcessor()
	processor.EnableProgress(false) // Disable progress for testing

	// Test with realistic configuration
	flags := &ParsedFlags{
		DotNotation: map[string]string{
			"cluster.name":       "test-cluster",
			"cluster.region":     "us-east-1",
			"cluster.node_count": "3",
			"network.cidr":       "10.0.0.0/16",
			"network.subnets[0]": "10.0.1.0/24",
			"network.subnets[1]": "10.0.2.0/24",
		},
		JSONFlags: []JSONFlag{
			{
				Path: "services.config",
				Value: `{
					"database": {
						"host": "localhost",
						"port": 5432,
						"name": "testdb"
					},
					"cache": {
						"host": "localhost",
						"port": 6379
					}
				}`,
			},
		},
		YAMLFlags: []YAMLFlag{
			{
				Path: "deployment.config",
				Value: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: app
        image: test-app:latest
        ports:
        - containerPort: 8080
`,
			},
		},
		SecurityFlags: []SecurityFlag{
			&MaskSensitiveFlag{Enabled: true},
			&SecurityWarningsFlag{Enabled: true},
		},
	}

	// Measure end-to-end processing time
	startTime := time.Now()
	result, err := processor.ProcessConfiguration(flags)
	processingTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("End-to-end processing failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Verify processing time is reasonable (should be under 1 second for this config)
	maxTime := 1 * time.Second
	if processingTime > maxTime {
		t.Errorf("End-to-end processing time %v exceeds maximum allowed time %v", processingTime, maxTime)
	}

	// Verify configuration was processed correctly
	if len(result.Configuration) == 0 {
		t.Error("Configuration should not be empty")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Processing should not have errors: %v", result.Errors)
	}

	t.Logf("End-to-end processing completed in %v with %d configuration items",
		processingTime, len(result.Configuration))
}
