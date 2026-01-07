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
	"time"
)

// EnhancedFlagProcessor provides high-performance flag processing with optimizations
type EnhancedFlagProcessor struct {
	securityHandler    *SecurityFlagHandler
	streamingProcessor *LargeConfigurationProcessor
	performanceOpt     *PerformanceOptimizer
	sizeAnalyzer       *ConfigurationSizeAnalyzer
	memoryOptimizer    *MemoryOptimizer
	progressEnabled    bool
	securityEnabled    bool
}

// NewEnhancedFlagProcessor creates a new enhanced flag processor
func NewEnhancedFlagProcessor() *EnhancedFlagProcessor {
	return &EnhancedFlagProcessor{
		securityHandler:    NewSecurityFlagHandler(),
		streamingProcessor: NewLargeConfigurationProcessor(),
		performanceOpt:     NewPerformanceOptimizer(),
		sizeAnalyzer:       NewConfigurationSizeAnalyzer(),
		memoryOptimizer:    NewMemoryOptimizer(),
		progressEnabled:    true,
		securityEnabled:    true,
	}
}

// ProcessFlags processes all flags with performance and security optimizations
func (p *EnhancedFlagProcessor) ProcessFlags(flags *ParsedFlags) (*ProcessedConfiguration, error) {
	startTime := time.Now()

	// Analyze configuration complexity
	p.analyzeFlags(flags)
	analysis := p.sizeAnalyzer.GetAnalysis()

	// Get optimization recommendations
	recommendations := p.getOptimizationRecommendations(analysis)

	// Show progress indicator for complex configurations
	var progress *ProgressIndicator
	if p.progressEnabled && (analysis.TotalFlags > 100 || analysis.TotalSize > StreamingThreshold) {
		progress = NewProgressIndicator("Processing configuration flags", int64(analysis.TotalFlags))
		progress.Start()
		defer progress.Stop()
	}

	// Process flags with memory optimization
	var result *ProcessedConfiguration
	err := p.memoryOptimizer.OptimizeMemoryUsage(func() error {
		processed, err := p.processAllFlags(flags, progress)
		if err != nil {
			return err
		}
		result = processed
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process flags: %w", err)
	}

	// Add processing metadata
	result.ProcessingTime = time.Since(startTime)
	result.Analysis = analysis
	result.Recommendations = recommendations

	return result, nil
}

// processAllFlags processes all flag types
func (p *EnhancedFlagProcessor) processAllFlags(flags *ParsedFlags, progress *ProgressIndicator) (*ProcessedConfiguration, error) {
	config := &ProcessedConfiguration{
		Configuration: make(map[string]interface{}),
		SecurityFlags: make(map[string]interface{}),
		Warnings:      []string{},
		Errors:        []string{},
	}

	processed := 0

	// Process dot notation flags
	for key, value := range flags.DotNotation {
		if err := p.processDotNotationFlag(config, key, value); err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("Dot notation flag %s: %s", key, err.Error()))
		}
		processed++
		if progress != nil {
			progress.Update(int64(processed))
		}
	}

	// Process JSON flags with streaming for large ones
	for _, jsonFlag := range flags.JSONFlags {
		if err := p.processJSONFlag(config, jsonFlag); err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("JSON flag %s: %s", jsonFlag.Path, err.Error()))
		}
		processed++
		if progress != nil {
			progress.Update(int64(processed))
		}
	}

	// Process YAML flags with streaming for large ones
	for _, yamlFlag := range flags.YAMLFlags {
		if err := p.processYAMLFlag(config, yamlFlag); err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("YAML flag %s: %s", yamlFlag.Path, err.Error()))
		}
		processed++
		if progress != nil {
			progress.Update(int64(processed))
		}
	}

	// Process security flags
	for _, securityFlag := range flags.SecurityFlags {
		if err := p.processSecurityFlag(config, securityFlag); err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("Security flag: %s", err.Error()))
		}
		processed++
		if progress != nil {
			progress.Update(int64(processed))
		}
	}

	// Process array flags
	for _, arrayFlag := range flags.ArrayFlags {
		if err := p.processArrayFlag(config, arrayFlag); err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("Array flag %s: %s", arrayFlag.Type, err.Error()))
		}
		processed++
		if progress != nil {
			progress.Update(int64(processed))
		}
	}

	// Apply security validation if enabled
	if p.securityEnabled {
		warnings := p.securityHandler.ValidateSecurityConfiguration(config.Configuration)
		for _, warning := range warnings {
			config.Warnings = append(config.Warnings, fmt.Sprintf("[%s] %s: %s", warning.Severity, warning.Type, warning.Message))
		}
	}

	return config, nil
}

// processDotNotationFlag processes a dot notation flag
func (p *EnhancedFlagProcessor) processDotNotationFlag(config *ProcessedConfiguration, key, value string) error {
	// Mask sensitive values if security is enabled
	if p.securityEnabled {
		maskedValue := p.securityHandler.MaskFlagValue(key, value)
		if maskedValue != value {
			// Store original in secure section, masked in regular config for logging
			config.SecurityFlags[key] = value
			config.Configuration[key] = maskedValue
		} else {
			config.Configuration[key] = value
		}
	} else {
		config.Configuration[key] = value
	}

	return nil
}

// processJSONFlag processes a JSON flag with streaming support
func (p *EnhancedFlagProcessor) processJSONFlag(config *ProcessedConfiguration, jsonFlag JSONFlag) error {
	// Convert value to string for processing
	jsonStr, ok := jsonFlag.Value.(string)
	if !ok {
		return fmt.Errorf("JSON flag value is not a string")
	}

	// Use streaming processor for large JSON
	if len(jsonStr) > int(StreamingThreshold) {
		result, err := p.streamingProcessor.ProcessLargeJSONFlag(jsonFlag.Path, jsonStr)
		if err != nil {
			return err
		}
		config.Configuration[jsonFlag.Path] = result
	} else {
		// Process normally for small JSON
		config.Configuration[jsonFlag.Path] = jsonFlag.Value
	}

	return nil
}

// processYAMLFlag processes a YAML flag with streaming support
func (p *EnhancedFlagProcessor) processYAMLFlag(config *ProcessedConfiguration, yamlFlag YAMLFlag) error {
	// Convert value to string for processing
	yamlStr, ok := yamlFlag.Value.(string)
	if !ok {
		return fmt.Errorf("YAML flag value is not a string")
	}

	// Use streaming processor for large YAML
	if len(yamlStr) > int(StreamingThreshold) {
		result, err := p.streamingProcessor.ProcessLargeYAMLFlag(yamlFlag.Path, yamlStr)
		if err != nil {
			return err
		}
		config.Configuration[yamlFlag.Path] = result
	} else {
		// Process normally for small YAML
		config.Configuration[yamlFlag.Path] = yamlFlag.Value
	}

	return nil
}

// processSecurityFlag processes a security flag
func (p *EnhancedFlagProcessor) processSecurityFlag(config *ProcessedConfiguration, securityFlag SecurityFlag) error {
	switch flag := securityFlag.(type) {
	case *SecureTemplateVarFlag:
		// Store in security section
		config.SecurityFlags[flag.Key] = flag.Value

		// Add warning if not from file
		if !flag.IsFile {
			config.Warnings = append(config.Warnings, fmt.Sprintf("Template variable '%s' provided via command line may be exposed in history", flag.Key))
		}

	case *MaskSensitiveFlag:
		config.SecurityFlags["mask_sensitive"] = flag.Enabled

	case *SecurityWarningsFlag:
		config.SecurityFlags["security_warnings"] = flag.Enabled
		p.securityEnabled = flag.Enabled

	case *SOPSConfigFlag:
		config.SecurityFlags["sops_config"] = flag.ConfigPath

	case *EncryptedConfigFlag:
		config.SecurityFlags["encrypted_config"] = flag.ConfigPath

	default:
		return fmt.Errorf("unknown security flag type")
	}

	return nil
}

// processArrayFlag processes an array flag
func (p *EnhancedFlagProcessor) processArrayFlag(config *ProcessedConfiguration, arrayFlag ArrayFlag) error {
	// Process array configuration
	if arrayFlag.Config != nil {
		config.Configuration[arrayFlag.Config.Path] = arrayFlag.Config.Fields
	}

	return nil
}

// analyzeFlags analyzes all flags for size and complexity
func (p *EnhancedFlagProcessor) analyzeFlags(flags *ParsedFlags) {
	// Analyze dot notation flags
	for key, value := range flags.DotNotation {
		p.sizeAnalyzer.AnalyzeFlag(key, value, FlagTypeDotNotation)
	}

	// Analyze JSON flags
	for _, jsonFlag := range flags.JSONFlags {
		if jsonStr, ok := jsonFlag.Value.(string); ok {
			p.sizeAnalyzer.AnalyzeFlag(jsonFlag.Path, jsonStr, FlagTypeJSON)
		}
	}

	// Analyze YAML flags
	for _, yamlFlag := range flags.YAMLFlags {
		if yamlStr, ok := yamlFlag.Value.(string); ok {
			p.sizeAnalyzer.AnalyzeFlag(yamlFlag.Path, yamlStr, FlagTypeYAML)
		}
	}

	// Analyze array flags
	for _, arrayFlag := range flags.ArrayFlags {
		p.sizeAnalyzer.AnalyzeFlag(arrayFlag.Type, fmt.Sprintf("%v", arrayFlag.Config), FlagTypeArray)
	}
}

// getOptimizationRecommendations gets optimization recommendations
func (p *EnhancedFlagProcessor) getOptimizationRecommendations(analysis ConfigurationAnalysis) []ProcessingRecommendation {
	return p.streamingProcessor.GetProcessingRecommendations(analysis.TotalFlags, analysis.TotalSize)
}

// SetProgressEnabled enables or disables progress indicators
func (p *EnhancedFlagProcessor) SetProgressEnabled(enabled bool) {
	p.progressEnabled = enabled
	p.performanceOpt.SetProgressEnabled(enabled)
}

// SetSecurityEnabled enables or disables security features
func (p *EnhancedFlagProcessor) SetSecurityEnabled(enabled bool) {
	p.securityEnabled = enabled
}

// GetPerformanceMetrics returns performance metrics
func (p *EnhancedFlagProcessor) GetPerformanceMetrics() PerformanceMetrics {
	memoryInfo := p.memoryOptimizer.CheckMemoryUsage()

	return PerformanceMetrics{
		MemoryUsage:     memoryInfo.CurrentUsage,
		MaxMemoryUsage:  memoryInfo.MaxUsage,
		ProcessingTime:  0, // Will be set during processing
		FlagsProcessed:  p.sizeAnalyzer.totalFlags,
		StreamingUsed:   p.sizeAnalyzer.totalSize > StreamingThreshold,
		SecurityEnabled: p.securityEnabled,
		ProgressEnabled: p.progressEnabled,
	}
}

// ProcessedConfiguration represents the result of flag processing
type ProcessedConfiguration struct {
	Configuration   map[string]interface{}     `json:"configuration"`
	SecurityFlags   map[string]interface{}     `json:"security_flags"`
	Warnings        []string                   `json:"warnings"`
	Errors          []string                   `json:"errors"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	Analysis        ConfigurationAnalysis      `json:"analysis"`
	Recommendations []ProcessingRecommendation `json:"recommendations"`
}

// PerformanceMetrics represents performance metrics for flag processing
type PerformanceMetrics struct {
	MemoryUsage     int64         `json:"memory_usage"`
	MaxMemoryUsage  int64         `json:"max_memory_usage"`
	ProcessingTime  time.Duration `json:"processing_time"`
	FlagsProcessed  int           `json:"flags_processed"`
	StreamingUsed   bool          `json:"streaming_used"`
	SecurityEnabled bool          `json:"security_enabled"`
	ProgressEnabled bool          `json:"progress_enabled"`
}

// ConfigurationProcessor provides a high-level interface for configuration processing
type ConfigurationProcessor struct {
	processor *EnhancedFlagProcessor
}

// NewConfigurationProcessor creates a new configuration processor
func NewConfigurationProcessor() *ConfigurationProcessor {
	return &ConfigurationProcessor{
		processor: NewEnhancedFlagProcessor(),
	}
}

// ProcessConfiguration processes configuration with all optimizations
func (cp *ConfigurationProcessor) ProcessConfiguration(flags *ParsedFlags) (*ProcessedConfiguration, error) {
	return cp.processor.ProcessFlags(flags)
}

// EnableProgress enables progress indicators
func (cp *ConfigurationProcessor) EnableProgress(enabled bool) {
	cp.processor.SetProgressEnabled(enabled)
}

// EnableSecurity enables security features
func (cp *ConfigurationProcessor) EnableSecurity(enabled bool) {
	cp.processor.SetSecurityEnabled(enabled)
}

// GetMetrics returns processing metrics
func (cp *ConfigurationProcessor) GetMetrics() PerformanceMetrics {
	return cp.processor.GetPerformanceMetrics()
}
