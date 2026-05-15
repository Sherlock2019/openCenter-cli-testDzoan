// Copyright 2025 opencenter-cloud
// Licensed under the Apache License, Version 2.0

package gitops

import (
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	"github.com/stretchr/testify/assert"
)

func TestGetAdoptionMode(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected AdoptionMode
	}{
		{
			name:     "nil input returns managed",
			input:    nil,
			expected: AdoptionModeManaged,
		},
		{
			name:     "empty struct returns managed",
			input:    struct{}{},
			expected: AdoptionModeManaged,
		},
		{
			name: "struct without AdoptionMode field returns managed",
			input: struct {
				Enabled bool
			}{Enabled: true},
			expected: AdoptionModeManaged,
		},
		{
			name: "struct with empty AdoptionMode returns managed",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: ""},
			expected: AdoptionModeManaged,
		},
		{
			name: "struct with managed mode",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeManaged},
			expected: AdoptionModeManaged,
		},
		{
			name: "struct with external mode",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeExternal},
			expected: AdoptionModeExternal,
		},
		{
			name: "struct with sync mode",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeSync},
			expected: AdoptionModeSync,
		},
		{
			name: "struct with deferred mode",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeDeferred},
			expected: AdoptionModeDeferred,
		},
		{
			name: "struct with takeover mode",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeTakeover},
			expected: AdoptionModeTakeover,
		},
		{
			name: "pointer to struct with external mode",
			input: &struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeExternal},
			expected: AdoptionModeExternal,
		},
		{
			name: "BaseServiceCfg with sync mode",
			input: services.BaseConfig{
				Enabled:      true,
				AdoptionMode: services.AdoptionModeSync,
			},
			expected: AdoptionModeSync,
		},
		{
			name: "ServiceCfg with deferred mode",
			input: services.BaseConfig{
				Enabled:      true,
				AdoptionMode: services.AdoptionModeDeferred,
			},
			expected: AdoptionModeDeferred,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdoptionMode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsServiceExternal(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "nil is not external",
			input:    nil,
			expected: false,
		},
		{
			name: "managed mode is not external",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeManaged},
			expected: false,
		},
		{
			name: "external mode is external",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeExternal},
			expected: true,
		},
		{
			name: "sync mode is not external",
			input: struct {
				AdoptionMode services.AdoptionMode
			}{AdoptionMode: services.AdoptionModeSync},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceExternal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldRenderService(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name: "enabled + managed = render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: true, AdoptionMode: services.AdoptionModeManaged},
			expected: true,
		},
		{
			name: "enabled + external = no render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: true, AdoptionMode: services.AdoptionModeExternal},
			expected: false,
		},
		{
			name: "disabled + managed = no render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: false, AdoptionMode: services.AdoptionModeManaged},
			expected: false,
		},
		{
			name: "enabled + sync = render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: true, AdoptionMode: services.AdoptionModeSync},
			expected: true,
		},
		{
			name: "enabled + deferred = render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: true, AdoptionMode: services.AdoptionModeDeferred},
			expected: true,
		},
		{
			name: "enabled + takeover = render",
			input: struct {
				Enabled      bool
				AdoptionMode services.AdoptionMode
			}{Enabled: true, AdoptionMode: services.AdoptionModeTakeover},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRenderService(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdoptionSettings(t *testing.T) {
	tests := []struct {
		name     string
		mode     AdoptionMode
		expected AdoptionSettings
	}{
		{
			name:     "managed mode",
			mode:     AdoptionModeManaged,
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
		{
			name:     "external mode (should not be called, but handle gracefully)",
			mode:     AdoptionModeExternal,
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
		{
			name:     "sync mode",
			mode:     AdoptionModeSync,
			expected: AdoptionSettings{Force: false, Suspend: false},
		},
		{
			name:     "deferred mode",
			mode:     AdoptionModeDeferred,
			expected: AdoptionSettings{Force: true, Suspend: true},
		},
		{
			name:     "takeover mode",
			mode:     AdoptionModeTakeover,
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
		{
			name:     "empty mode defaults to managed",
			mode:     "",
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
		{
			name:     "unknown mode defaults to managed",
			mode:     "unknown",
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdoptionSettings(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetServiceAdoptionSettings(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected AdoptionSettings
	}{
		{
			name: "service with sync mode",
			input: services.BaseConfig{
				Enabled:      true,
				AdoptionMode: services.AdoptionModeSync,
			},
			expected: AdoptionSettings{Force: false, Suspend: false},
		},
		{
			name: "service with deferred mode",
			input: services.BaseConfig{
				Enabled:      true,
				AdoptionMode: services.AdoptionModeDeferred,
			},
			expected: AdoptionSettings{Force: true, Suspend: true},
		},
		{
			name:     "nil service defaults to managed settings",
			input:    nil,
			expected: AdoptionSettings{Force: true, Suspend: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceAdoptionSettings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
