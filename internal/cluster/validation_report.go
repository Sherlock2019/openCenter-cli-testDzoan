package cluster

import (
	"fmt"
	"strings"
)

const (
	ValidationModeOffline = "offline"
	ValidationModeOnline  = "online"
)

type CheckStatus string

const (
	CheckStatusPass CheckStatus = "pass"
	CheckStatusFail CheckStatus = "fail"
	CheckStatusWarn CheckStatus = "warn"
	CheckStatusSkip CheckStatus = "skip"
)

type ValidationTarget struct {
	Cluster      string `json:"cluster,omitempty"`
	Organization string `json:"organization,omitempty"`
	Provider     string `json:"provider,omitempty"`
	ConfigPath   string `json:"config_path,omitempty"`
}

type ValidationCheckSummary struct {
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Warnings int `json:"warnings"`
	Skipped  int `json:"skipped"`
}

type ValidationCheck struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message,omitempty"`
	Detail  string      `json:"detail,omitempty"`
}

type ValidationServiceReport struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message,omitempty"`
	Missing []string    `json:"missing,omitempty"`
}

type ValidationGitOpsReport struct {
	RepositoryURL string            `json:"repository_url,omitempty"`
	LocalPath     string            `json:"local_path,omitempty"`
	Branch        string            `json:"branch,omitempty"`
	Checks        []ValidationCheck `json:"checks"`
}

type ValidationMissing struct {
	Path    string `json:"path"`
	Message string `json:"message,omitempty"`
}

func (s ValidationCheckSummary) Total() int {
	return s.Passed + s.Failed + s.Warnings + s.Skipped
}

func NormalizeValidationMode(mode, field string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		normalized = ValidationModeOffline
	}
	switch normalized {
	case ValidationModeOffline, ValidationModeOnline:
		return normalized, nil
	default:
		if field == "" {
			field = "validation mode"
		}
		return "", fmt.Errorf("invalid %s %q; expected offline or online", field, normalized)
	}
}
