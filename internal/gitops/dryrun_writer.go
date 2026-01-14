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

package gitops

import (
	"fmt"
	"os"
	"time"
)

// DryRunAtomicWriter is an atomic writer that records operations without making filesystem changes.
// It provides the same interface as AtomicWriter but operates in dry-run mode.
type DryRunAtomicWriter struct {
	workspace    *DryRunWorkspace
	currentStage string
}

// NewDryRunAtomicWriter creates a new dry-run atomic writer.
func NewDryRunAtomicWriter(workspace *DryRunWorkspace, stageName string) *DryRunAtomicWriter {
	return &DryRunAtomicWriter{
		workspace:    workspace,
		currentStage: stageName,
	}
}

// WriteFile records a file write operation without actually writing to disk.
func (w *DryRunAtomicWriter) WriteFile(relativePath string, content []byte, mode os.FileMode) error {
	w.workspace.RecordOperation(DryRunOperation{
		Type:      OpWriteFile,
		Path:      relativePath,
		Content:   string(content),
		Mode:      mode,
		Timestamp: time.Now(),
		Stage:     w.currentStage,
	})
	return nil
}

// WriteFileString records a file write operation without actually writing to disk.
func (w *DryRunAtomicWriter) WriteFileString(relativePath string, content string, mode os.FileMode) error {
	w.workspace.RecordOperation(DryRunOperation{
		Type:      OpWriteFile,
		Path:      relativePath,
		Content:   content,
		Mode:      mode,
		Timestamp: time.Now(),
		Stage:     w.currentStage,
	})
	return nil
}

// MkdirAll records a directory creation operation without actually creating directories.
func (w *DryRunAtomicWriter) MkdirAll(relativePath string, mode os.FileMode) error {
	w.workspace.RecordOperation(DryRunOperation{
		Type:      OpCreateDir,
		Path:      relativePath,
		Mode:      mode,
		Timestamp: time.Now(),
		Stage:     w.currentStage,
	})
	return nil
}

// Remove records a file removal operation without actually removing files.
func (w *DryRunAtomicWriter) Remove(relativePath string) error {
	w.workspace.RecordOperation(DryRunOperation{
		Type:      OpRemoveFile,
		Path:      relativePath,
		Timestamp: time.Now(),
		Stage:     w.currentStage,
	})
	return nil
}

// RemoveAll records a directory removal operation without actually removing directories.
func (w *DryRunAtomicWriter) RemoveAll(relativePath string) error {
	w.workspace.RecordOperation(DryRunOperation{
		Type:      OpRemoveDir,
		Path:      relativePath,
		Timestamp: time.Now(),
		Stage:     w.currentStage,
	})
	return nil
}

// Commit is a no-op in dry-run mode since no actual operations are performed.
func (w *DryRunAtomicWriter) Commit() error {
	return nil
}

// Rollback is a no-op in dry-run mode since no actual operations are performed.
func (w *DryRunAtomicWriter) Rollback() error {
	return nil
}

// GetWorkspace returns the dry-run workspace.
func (w *DryRunAtomicWriter) GetWorkspace() *DryRunWorkspace {
	return w.workspace
}

// SetStage sets the current stage name for operation tracking.
func (w *DryRunAtomicWriter) SetStage(stageName string) {
	w.currentStage = stageName
}

// DryRunAtomicWriterAdapter adapts a DryRunAtomicWriter to work with code expecting AtomicWriter.
// This allows stages to work in both normal and dry-run modes without modification.
type DryRunAtomicWriterAdapter struct {
	dryRunWriter *DryRunAtomicWriter
}

// NewDryRunAtomicWriterAdapter creates a new adapter.
func NewDryRunAtomicWriterAdapter(workspace *DryRunWorkspace, stageName string) *DryRunAtomicWriterAdapter {
	return &DryRunAtomicWriterAdapter{
		dryRunWriter: NewDryRunAtomicWriter(workspace, stageName),
	}
}

// WriteFile delegates to the dry-run writer.
func (a *DryRunAtomicWriterAdapter) WriteFile(relativePath string, content []byte, mode os.FileMode) error {
	return a.dryRunWriter.WriteFile(relativePath, content, mode)
}

// WriteFileString delegates to the dry-run writer.
func (a *DryRunAtomicWriterAdapter) WriteFileString(relativePath string, content string, mode os.FileMode) error {
	return a.dryRunWriter.WriteFileString(relativePath, content, mode)
}

// MkdirAll delegates to the dry-run writer.
func (a *DryRunAtomicWriterAdapter) MkdirAll(relativePath string, mode os.FileMode) error {
	return a.dryRunWriter.MkdirAll(relativePath, mode)
}

// Remove delegates to the dry-run writer.
func (a *DryRunAtomicWriterAdapter) Remove(relativePath string) error {
	return a.dryRunWriter.Remove(relativePath)
}

// RemoveAll delegates to the dry-run writer.
func (a *DryRunAtomicWriterAdapter) RemoveAll(relativePath string) error {
	return a.dryRunWriter.RemoveAll(relativePath)
}

// Commit is a no-op in dry-run mode.
func (a *DryRunAtomicWriterAdapter) Commit() error {
	return a.dryRunWriter.Commit()
}

// Rollback is a no-op in dry-run mode.
func (a *DryRunAtomicWriterAdapter) Rollback() error {
	return a.dryRunWriter.Rollback()
}

// GetDryRunWorkspace returns the underlying dry-run workspace.
func (a *DryRunAtomicWriterAdapter) GetDryRunWorkspace() *DryRunWorkspace {
	return a.dryRunWriter.GetWorkspace()
}

// IsDryRun returns true to indicate this is a dry-run writer.
func (a *DryRunAtomicWriterAdapter) IsDryRun() bool {
	return true
}

// String returns a string representation of the adapter.
func (a *DryRunAtomicWriterAdapter) String() string {
	return fmt.Sprintf("DryRunAtomicWriterAdapter{stage: %s, operations: %d}",
		a.dryRunWriter.currentStage,
		a.dryRunWriter.workspace.GetOperationCount())
}
