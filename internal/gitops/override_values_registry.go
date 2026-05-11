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
	"sync"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

// OverrideValuesRenderer produces dynamic override-values.yaml content from cluster config.
type OverrideValuesRenderer func(cfg v2.Config) (string, error)

// OverlayFilesRenderer produces additional overlay files from cluster config.
// Returns a map of filename → content.
type OverlayFilesRenderer func(cfg v2.Config) (map[string]string, error)

var (
	overrideValuesRenderers   = make(map[string]OverrideValuesRenderer)
	overlayFilesRenderers     = make(map[string]OverlayFilesRenderer)
	overrideValuesRenderersMu sync.RWMutex
)

// RegisterOverrideValuesRenderer registers a named renderer for override-values generation.
func RegisterOverrideValuesRenderer(key string, renderer OverrideValuesRenderer) {
	overrideValuesRenderersMu.Lock()
	defer overrideValuesRenderersMu.Unlock()
	overrideValuesRenderers[key] = renderer
}

// RegisterOverlayFilesRenderer registers a named renderer for overlay file generation.
func RegisterOverlayFilesRenderer(key string, renderer OverlayFilesRenderer) {
	overrideValuesRenderersMu.Lock()
	defer overrideValuesRenderersMu.Unlock()
	overlayFilesRenderers[key] = renderer
}

// getOverrideValuesRenderer looks up a registered renderer by key.
func getOverrideValuesRenderer(key string) (OverrideValuesRenderer, error) {
	overrideValuesRenderersMu.RLock()
	defer overrideValuesRenderersMu.RUnlock()
	renderer, ok := overrideValuesRenderers[key]
	if !ok {
		return nil, fmt.Errorf("override-values renderer %q not registered", key)
	}
	return renderer, nil
}

// getOverlayFilesRenderer looks up a registered overlay files renderer by key.
func getOverlayFilesRenderer(key string) (OverlayFilesRenderer, error) {
	overrideValuesRenderersMu.RLock()
	defer overrideValuesRenderersMu.RUnlock()
	renderer, ok := overlayFilesRenderers[key]
	if !ok {
		return nil, fmt.Errorf("overlay-files renderer %q not registered", key)
	}
	return renderer, nil
}
