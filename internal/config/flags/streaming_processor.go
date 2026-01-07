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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// StreamingJSONProcessor handles streaming JSON processing for large JSON flags
type StreamingJSONProcessor struct {
	maxSize         int64
	chunkSize       int64
	progressEnabled bool
}

// NewStreamingJSONProcessor creates a new streaming JSON processor
func NewStreamingJSONProcessor() *StreamingJSONProcessor {
	return &StreamingJSONProcessor{
		maxSize:         MaxJSONSize,
		chunkSize:       StreamingThreshold,
		progressEnabled: true,
	}
}

// ProcessJSONStream processes a JSON stream and returns parsed data
func (p *StreamingJSONProcessor) ProcessJSONStream(reader io.Reader) (interface{}, error) {
	// Read the entire stream into memory first (for JSON we need the complete document)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON stream: %w", err)
	}

	// Validate size
	if int64(len(data)) > p.maxSize {
		return nil, fmt.Errorf("JSON data size (%d bytes) exceeds maximum allowed size (%d bytes)", len(data), p.maxSize)
	}

	// Parse JSON
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// ProcessJSONString processes a JSON string with size validation
func (p *StreamingJSONProcessor) ProcessJSONString(jsonStr string) (interface{}, error) {
	// Validate size
	if int64(len(jsonStr)) > p.maxSize {
		return nil, fmt.Errorf("JSON string size (%d bytes) exceeds maximum allowed size (%d bytes)", len(jsonStr), p.maxSize)
	}

	// Parse JSON
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// ValidateJSONSize validates JSON size before processing
func (p *StreamingJSONProcessor) ValidateJSONSize(data []byte) error {
	if int64(len(data)) > p.maxSize {
		return fmt.Errorf("JSON data size (%d bytes) exceeds maximum allowed size (%d bytes). Consider using file-based configuration or breaking into smaller pieces", len(data), p.maxSize)
	}
	return nil
}

// StreamingYAMLProcessor handles streaming YAML processing for large YAML flags
type StreamingYAMLProcessor struct {
	maxSize         int64
	chunkSize       int64
	progressEnabled bool
}

// NewStreamingYAMLProcessor creates a new streaming YAML processor
func NewStreamingYAMLProcessor() *StreamingYAMLProcessor {
	return &StreamingYAMLProcessor{
		maxSize:         MaxYAMLSize,
		chunkSize:       StreamingThreshold,
		progressEnabled: true,
	}
}

// ProcessYAMLStream processes a YAML stream and returns parsed data
func (p *StreamingYAMLProcessor) ProcessYAMLStream(reader io.Reader) (interface{}, error) {
	// Read the entire stream into memory first (for YAML we need the complete document)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML stream: %w", err)
	}

	// Validate size
	if int64(len(data)) > p.maxSize {
		return nil, fmt.Errorf("YAML data size (%d bytes) exceeds maximum allowed size (%d bytes)", len(data), p.maxSize)
	}

	// Parse YAML
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return result, nil
}

// ProcessYAMLString processes a YAML string with size validation
func (p *StreamingYAMLProcessor) ProcessYAMLString(yamlStr string) (interface{}, error) {
	// Validate size
	if int64(len(yamlStr)) > p.maxSize {
		return nil, fmt.Errorf("YAML string size (%d bytes) exceeds maximum allowed size (%d bytes)", len(yamlStr), p.maxSize)
	}

	// Parse YAML
	var result interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return result, nil
}

// ProcessMultiDocumentYAML processes multi-document YAML with streaming
func (p *StreamingYAMLProcessor) ProcessMultiDocumentYAML(reader io.Reader) ([]interface{}, error) {
	// Read the entire stream
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML stream: %w", err)
	}

	// Validate size
	if int64(len(data)) > p.maxSize {
		return nil, fmt.Errorf("YAML data size (%d bytes) exceeds maximum allowed size (%d bytes)", len(data), p.maxSize)
	}

	// Split documents by "---" separator
	documents := strings.Split(string(data), "---")
	var results []interface{}

	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue // Skip empty documents
		}

		var result interface{}
		if err := yaml.Unmarshal([]byte(doc), &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML document %d: %w", i+1, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// ValidateYAMLSize validates YAML size before processing
func (p *StreamingYAMLProcessor) ValidateYAMLSize(data []byte) error {
	if int64(len(data)) > p.maxSize {
		return fmt.Errorf("YAML data size (%d bytes) exceeds maximum allowed size (%d bytes). Consider using file-based configuration or breaking into smaller pieces", len(data), p.maxSize)
	}
	return nil
}

// LargeConfigurationProcessor handles processing of large configuration files
type LargeConfigurationProcessor struct {
	jsonProcessor *StreamingJSONProcessor
	yamlProcessor *StreamingYAMLProcessor
	optimizer     *PerformanceOptimizer
}

// NewLargeConfigurationProcessor creates a new large configuration processor
func NewLargeConfigurationProcessor() *LargeConfigurationProcessor {
	return &LargeConfigurationProcessor{
		jsonProcessor: NewStreamingJSONProcessor(),
		yamlProcessor: NewStreamingYAMLProcessor(),
		optimizer:     NewPerformanceOptimizer(),
	}
}

// ProcessLargeJSONFlag processes a large JSON flag with streaming and progress
func (p *LargeConfigurationProcessor) ProcessLargeJSONFlag(flagName, jsonValue string) (interface{}, error) {
	// Validate size first
	if err := p.jsonProcessor.ValidateJSONSize([]byte(jsonValue)); err != nil {
		return nil, fmt.Errorf("JSON flag %s: %w", flagName, err)
	}

	// Process with progress if enabled and size warrants it
	if len(jsonValue) > int(StreamingThreshold) {
		progress := NewProgressIndicator(fmt.Sprintf("Processing large JSON flag %s", flagName), int64(len(jsonValue)))
		progress.Start()
		defer progress.Stop()
	}

	// Parse JSON
	result, err := p.jsonProcessor.ProcessJSONString(jsonValue)
	if err != nil {
		return nil, fmt.Errorf("failed to process JSON flag %s: %w", flagName, err)
	}

	return result, nil
}

// ProcessLargeYAMLFlag processes a large YAML flag with streaming and progress
func (p *LargeConfigurationProcessor) ProcessLargeYAMLFlag(flagName, yamlValue string) (interface{}, error) {
	// Validate size first
	if err := p.yamlProcessor.ValidateYAMLSize([]byte(yamlValue)); err != nil {
		return nil, fmt.Errorf("YAML flag %s: %w", flagName, err)
	}

	// Process with progress if enabled and size warrants it
	if len(yamlValue) > int(StreamingThreshold) {
		progress := NewProgressIndicator(fmt.Sprintf("Processing large YAML flag %s", flagName), int64(len(yamlValue)))
		progress.Start()
		defer progress.Stop()
	}

	// Parse YAML
	result, err := p.yamlProcessor.ProcessYAMLString(yamlValue)
	if err != nil {
		return nil, fmt.Errorf("failed to process YAML flag %s: %w", flagName, err)
	}

	return result, nil
}

// ProcessLargeConfigurationFile processes a large configuration file with streaming
func (p *LargeConfigurationProcessor) ProcessLargeConfigurationFile(filePath string) (interface{}, error) {
	// Validate file size first
	if err := p.optimizer.ValidateFileSize(filePath); err != nil {
		return nil, err
	}

	// Determine file type and process accordingly
	if strings.HasSuffix(strings.ToLower(filePath), ".json") {
		return p.processLargeJSONFile(filePath)
	} else if strings.HasSuffix(strings.ToLower(filePath), ".yaml") || strings.HasSuffix(strings.ToLower(filePath), ".yml") {
		return p.processLargeYAMLFile(filePath)
	}

	return nil, fmt.Errorf("unsupported file type for %s", filePath)
}

// processLargeJSONFile processes a large JSON file with streaming
func (p *LargeConfigurationProcessor) processLargeJSONFile(filePath string) (interface{}, error) {
	var result interface{}

	err := p.optimizer.ProcessLargeFile(filePath, func(reader io.Reader) error {
		parsed, err := p.jsonProcessor.ProcessJSONStream(reader)
		if err != nil {
			return err
		}
		result = parsed
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process large JSON file %s: %w", filePath, err)
	}

	return result, nil
}

// processLargeYAMLFile processes a large YAML file with streaming
func (p *LargeConfigurationProcessor) processLargeYAMLFile(filePath string) (interface{}, error) {
	var result interface{}

	err := p.optimizer.ProcessLargeFile(filePath, func(reader io.Reader) error {
		parsed, err := p.yamlProcessor.ProcessYAMLStream(reader)
		if err != nil {
			return err
		}
		result = parsed
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process large YAML file %s: %w", filePath, err)
	}

	return result, nil
}

// GetProcessingRecommendations returns recommendations for processing large configurations
func (p *LargeConfigurationProcessor) GetProcessingRecommendations(flagCount int, totalSize int64) []ProcessingRecommendation {
	var recommendations []ProcessingRecommendation

	// Size-based recommendations
	if totalSize > MaxConfigSize/2 {
		recommendations = append(recommendations, ProcessingRecommendation{
			Type:       "size",
			Severity:   "medium",
			Message:    "Large total configuration size detected",
			Suggestion: "Consider using file-based configuration instead of command-line flags",
			Impact:     "Improved performance and reduced memory usage",
		})
	}

	if totalSize > MaxConfigSize {
		recommendations = append(recommendations, ProcessingRecommendation{
			Type:       "size",
			Severity:   "high",
			Message:    "Configuration size exceeds recommended limits",
			Suggestion: "Break configuration into multiple smaller files",
			Impact:     "Prevents memory issues and improves processing speed",
		})
	}

	// Flag count recommendations
	if flagCount > 50 {
		recommendations = append(recommendations, ProcessingRecommendation{
			Type:       "complexity",
			Severity:   "medium",
			Message:    "High number of configuration flags",
			Suggestion: "Consider using JSON/YAML flags to group related settings",
			Impact:     "Simplified command-line usage and better organization",
		})
	}

	if flagCount > 200 {
		recommendations = append(recommendations, ProcessingRecommendation{
			Type:       "complexity",
			Severity:   "high",
			Message:    "Excessive number of configuration flags",
			Suggestion: "Use configuration files instead of individual flags",
			Impact:     "Significant performance improvement and reduced complexity",
		})
	}

	return recommendations
}

// ProcessingRecommendation represents a processing optimization recommendation
type ProcessingRecommendation struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Impact     string `json:"impact"`
}

// ConfigurationSizeAnalyzer analyzes configuration size and complexity
type ConfigurationSizeAnalyzer struct {
	totalFlags   int
	totalSize    int64
	jsonFlags    int
	yamlFlags    int
	fileFlags    int
	complexFlags int
}

// NewConfigurationSizeAnalyzer creates a new configuration size analyzer
func NewConfigurationSizeAnalyzer() *ConfigurationSizeAnalyzer {
	return &ConfigurationSizeAnalyzer{}
}

// AnalyzeFlag analyzes a single flag and updates statistics
func (a *ConfigurationSizeAnalyzer) AnalyzeFlag(flagName, flagValue string, flagType FlagType) {
	a.totalFlags++
	a.totalSize += int64(len(flagValue))

	switch flagType {
	case FlagTypeJSON:
		a.jsonFlags++
		if len(flagValue) > 1000 {
			a.complexFlags++
		}
	case FlagTypeYAML:
		a.yamlFlags++
		if len(flagValue) > 1000 {
			a.complexFlags++
		}
	case FlagTypeFile:
		a.fileFlags++
	default:
		if len(flagValue) > 500 {
			a.complexFlags++
		}
	}
}

// GetAnalysis returns the configuration analysis
func (a *ConfigurationSizeAnalyzer) GetAnalysis() ConfigurationAnalysis {
	return ConfigurationAnalysis{
		TotalFlags:   a.totalFlags,
		TotalSize:    a.totalSize,
		JSONFlags:    a.jsonFlags,
		YAMLFlags:    a.yamlFlags,
		FileFlags:    a.fileFlags,
		ComplexFlags: a.complexFlags,
		AverageSize:  a.getAverageSize(),
		Complexity:   a.getComplexityLevel(),
	}
}

// getAverageSize calculates the average flag size
func (a *ConfigurationSizeAnalyzer) getAverageSize() int64 {
	if a.totalFlags == 0 {
		return 0
	}
	return a.totalSize / int64(a.totalFlags)
}

// getComplexityLevel determines the configuration complexity level
func (a *ConfigurationSizeAnalyzer) getComplexityLevel() string {
	if a.totalFlags > 200 || a.totalSize > MaxConfigSize {
		return "high"
	} else if a.totalFlags > 50 || a.totalSize > MaxConfigSize/2 {
		return "medium"
	}
	return "low"
}

// ConfigurationAnalysis represents the analysis of configuration complexity
type ConfigurationAnalysis struct {
	TotalFlags   int    `json:"total_flags"`
	TotalSize    int64  `json:"total_size"`
	JSONFlags    int    `json:"json_flags"`
	YAMLFlags    int    `json:"yaml_flags"`
	FileFlags    int    `json:"file_flags"`
	ComplexFlags int    `json:"complex_flags"`
	AverageSize  int64  `json:"average_size"`
	Complexity   string `json:"complexity"`
}

// MemoryOptimizer provides memory optimization for large configurations
type MemoryOptimizer struct {
	maxMemoryUsage int64
	gcThreshold    int64
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		maxMemoryUsage: 100 * 1024 * 1024, // 100MB
		gcThreshold:    50 * 1024 * 1024,  // 50MB
	}
}

// OptimizeMemoryUsage optimizes memory usage during configuration processing
func (m *MemoryOptimizer) OptimizeMemoryUsage(processor func() error) error {
	// This would implement memory optimization strategies
	// For now, just call the processor
	return processor()
}

// CheckMemoryUsage checks current memory usage and suggests optimizations
func (m *MemoryOptimizer) CheckMemoryUsage() MemoryUsageInfo {
	// This would implement actual memory usage checking
	// For now, return placeholder data
	return MemoryUsageInfo{
		CurrentUsage: 0,
		MaxUsage:     m.maxMemoryUsage,
		Recommendations: []string{
			"Use streaming processing for large files",
			"Process configuration in chunks",
			"Enable garbage collection optimization",
		},
	}
}

// MemoryUsageInfo represents memory usage information
type MemoryUsageInfo struct {
	CurrentUsage    int64    `json:"current_usage"`
	MaxUsage        int64    `json:"max_usage"`
	Recommendations []string `json:"recommendations"`
}
