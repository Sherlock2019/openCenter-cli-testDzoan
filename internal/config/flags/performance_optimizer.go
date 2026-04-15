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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const (
	// MaxConfigSize is the maximum configuration size in bytes (10MB)
	MaxConfigSize = 10 * 1024 * 1024

	// MaxJSONSize is the maximum JSON flag size in bytes (1MB)
	MaxJSONSize = 1024 * 1024

	// MaxYAMLSize is the maximum YAML flag size in bytes (1MB)
	MaxYAMLSize = 1024 * 1024

	// StreamingThreshold is the size threshold for streaming processing (100KB)
	StreamingThreshold = 100 * 1024

	// ProgressUpdateInterval is how often to update progress indicators
	ProgressUpdateInterval = 100 * time.Millisecond
)

// PerformanceOptimizer handles performance optimizations for large configurations
type PerformanceOptimizer struct {
	progressEnabled  bool
	sizeLimit        int64
	streamingEnabled bool
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		progressEnabled:  true,
		sizeLimit:        MaxConfigSize,
		streamingEnabled: true,
	}
}

// SetProgressEnabled enables or disables progress indicators
func (p *PerformanceOptimizer) SetProgressEnabled(enabled bool) {
	p.progressEnabled = enabled
}

// SetSizeLimit sets the maximum configuration size limit
func (p *PerformanceOptimizer) SetSizeLimit(limit int64) {
	p.sizeLimit = limit
}

// SetStreamingEnabled enables or disables streaming processing
func (p *PerformanceOptimizer) SetStreamingEnabled(enabled bool) {
	p.streamingEnabled = enabled
}

// ValidateSize validates that content size is within limits
func (p *PerformanceOptimizer) ValidateSize(content []byte, contentType string) error {
	size := int64(len(content))

	var limit int64
	switch contentType {
	case "json":
		limit = MaxJSONSize
	case "yaml":
		limit = MaxYAMLSize
	default:
		limit = p.sizeLimit
	}

	if size > limit {
		return fmt.Errorf("%s content size (%d bytes) exceeds limit (%d bytes). Consider using streaming or file-based configuration",
			contentType, size, limit)
	}

	return nil
}

// ValidateFileSize validates that a file size is within limits
func (p *PerformanceOptimizer) ValidateFileSize(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	size := fileInfo.Size()
	if size > p.sizeLimit {
		return fmt.Errorf("file %s size (%d bytes) exceeds limit (%d bytes). Consider breaking it into smaller files",
			filePath, size, p.sizeLimit)
	}

	return nil
}

// ProcessLargeContent processes large content with streaming if needed
func (p *PerformanceOptimizer) ProcessLargeContent(content []byte, processor func([]byte) error) error {
	size := int64(len(content))

	if !p.streamingEnabled || size < StreamingThreshold {
		// Process normally for small content
		return processor(content)
	}

	// Use streaming processing for large content
	return p.processWithStreaming(content, processor)
}

// ProcessLargeFile processes a large file with streaming
func (p *PerformanceOptimizer) ProcessLargeFile(filePath string, processor func(io.Reader) error) error {
	// Validate file size first
	if err := p.ValidateFileSize(filePath); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file size for progress tracking
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()

	if p.progressEnabled && fileSize > StreamingThreshold {
		// Process with progress indicator
		return p.processFileWithProgress(file, fileSize, processor)
	}

	// Process normally
	return processor(file)
}

// processWithStreaming processes content using streaming approach
func (p *PerformanceOptimizer) processWithStreaming(content []byte, processor func([]byte) error) error {
	if p.progressEnabled {
		return p.processWithProgress(content, processor)
	}

	// Process in chunks to avoid memory issues
	chunkSize := StreamingThreshold
	totalSize := len(content)

	for i := 0; i < totalSize; i += int(chunkSize) {
		end := i + int(chunkSize)
		if end > totalSize {
			end = totalSize
		}

		chunk := content[i:end]
		if err := processor(chunk); err != nil {
			return fmt.Errorf("failed to process chunk at offset %d: %w", i, err)
		}
	}

	return nil
}

// processWithProgress processes content with progress indicator
func (p *PerformanceOptimizer) processWithProgress(content []byte, processor func([]byte) error) error {
	totalSize := len(content)
	chunkSize := StreamingThreshold

	progress := NewProgressIndicator("Processing configuration", int64(totalSize))
	progress.Start()
	defer progress.Stop()

	for i := 0; i < totalSize; i += int(chunkSize) {
		end := i + int(chunkSize)
		if end > totalSize {
			end = totalSize
		}

		chunk := content[i:end]
		if err := processor(chunk); err != nil {
			return fmt.Errorf("failed to process chunk at offset %d: %w", i, err)
		}

		progress.Update(int64(end))
	}

	return nil
}

// processFileWithProgress processes a file with progress indicator
func (p *PerformanceOptimizer) processFileWithProgress(file *os.File, fileSize int64, processor func(io.Reader) error) error {
	progress := NewProgressIndicator("Processing file", fileSize)
	progress.Start()
	defer progress.Stop()

	// Create a progress reader that updates progress as data is read
	progressReader := &ProgressReader{
		reader:   file,
		progress: progress,
	}

	return processor(progressReader)
}

// GetOptimizationRecommendations returns recommendations for optimizing configuration processing
func (p *PerformanceOptimizer) GetOptimizationRecommendations(configSize int64, flagCount int) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation

	// Size-based recommendations
	if configSize > MaxConfigSize/2 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:       "size",
			Severity:   "medium",
			Message:    "Configuration size is large",
			Suggestion: "Consider breaking configuration into multiple files",
			Impact:     "Improved memory usage and processing speed",
		})
	}

	if configSize > MaxConfigSize {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:       "size",
			Severity:   "high",
			Message:    "Configuration size exceeds recommended limit",
			Suggestion: "Use file-based configuration instead of command-line flags",
			Impact:     "Prevents memory issues and improves performance",
		})
	}

	// Flag count-based recommendations
	if flagCount > 100 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:       "complexity",
			Severity:   "medium",
			Message:    "Large number of command-line flags",
			Suggestion: "Consider using configuration files or YAML/JSON flags",
			Impact:     "Simplified command-line usage and better maintainability",
		})
	}

	if flagCount > 500 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:       "complexity",
			Severity:   "high",
			Message:    "Excessive number of command-line flags",
			Suggestion: "Use configuration files instead of individual flags",
			Impact:     "Significant performance improvement and reduced complexity",
		})
	}

	return recommendations
}

// OptimizationRecommendation represents a performance optimization recommendation
type OptimizationRecommendation struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Impact     string `json:"impact"`
}

// ProgressIndicator provides progress indication for long operations
type ProgressIndicator struct {
	message   string
	total     int64
	current   atomic.Int64
	startTime time.Time
	ticker    *time.Ticker
	done      chan bool
}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(message string, total int64) *ProgressIndicator {
	return &ProgressIndicator{
		message: message,
		total:   total,
		done:    make(chan bool),
	}
}

// Start starts the progress indicator
func (p *ProgressIndicator) Start() {
	p.startTime = time.Now()
	p.ticker = time.NewTicker(ProgressUpdateInterval)

	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.display()
			case <-p.done:
				return
			}
		}
	}()
}

// Update updates the progress
func (p *ProgressIndicator) Update(current int64) {
	p.current.Store(current)
}

// Stop stops the progress indicator
func (p *ProgressIndicator) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.done)
	p.displayFinal()
}

// display displays the current progress
func (p *ProgressIndicator) display() {
	if p.total == 0 {
		return
	}

	current := p.current.Load()
	percentage := float64(current) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)

	// Create progress bar
	barWidth := 40
	filled := int(percentage / 100 * float64(barWidth))
	bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)

	fmt.Fprintf(os.Stderr, "\r%s [%s] %.1f%% (%s)", p.message, bar, percentage, elapsed.Truncate(time.Second))
}

// displayFinal displays the final progress message
func (p *ProgressIndicator) displayFinal() {
	elapsed := time.Since(p.startTime)
	fmt.Fprintf(os.Stderr, "\r%s [%s] 100.0%% (%s) - Complete\n",
		p.message, strings.Repeat("=", 40), elapsed.Truncate(time.Second))
}

// ProgressReader wraps an io.Reader to provide progress updates
type ProgressReader struct {
	reader    io.Reader
	progress  *ProgressIndicator
	bytesRead int64
}

// Read implements io.Reader interface with progress tracking
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.bytesRead += int64(n)
	pr.progress.Update(pr.bytesRead)
	return n, err
}

// StreamingProcessor provides streaming processing capabilities
type StreamingProcessor struct {
	chunkSize int64
	processor func([]byte) error
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(chunkSize int64, processor func([]byte) error) *StreamingProcessor {
	return &StreamingProcessor{
		chunkSize: chunkSize,
		processor: processor,
	}
}

// ProcessStream processes a stream of data in chunks
func (sp *StreamingProcessor) ProcessStream(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	// Set a large buffer for the scanner
	buf := make([]byte, sp.chunkSize)
	scanner.Buffer(buf, int(sp.chunkSize))

	for scanner.Scan() {
		chunk := scanner.Bytes()
		if err := sp.processor(chunk); err != nil {
			return fmt.Errorf("failed to process stream chunk: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
