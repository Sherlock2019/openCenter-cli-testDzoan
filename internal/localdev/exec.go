package localdev

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

	"github.com/opencenter-cloud/opencenter-cli/internal/security"
)

// RunOptions describes a single external command invocation.
type RunOptions struct {
	Dir    string
	Env    map[string]string
	Name   string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Executor runs sanitized external commands.
type Executor interface {
	Run(ctx context.Context, opts RunOptions) ([]byte, error)
	RunStreaming(ctx context.Context, opts RunOptions) error
}

// DefaultExecutor delegates to the shared command runner.
type DefaultExecutor struct {
	runner security.CommandRunner
}

// NewExecutor returns an executor backed by the process-wide command runner.
func NewExecutor() *DefaultExecutor {
	return &DefaultExecutor{runner: security.GetDefaultCommandRunner()}
}

func (e *DefaultExecutor) prepare(ctx context.Context, opts RunOptions) (*exec.Cmd, error) {
	cmd, err := e.runner.PrepareCommandContext(ctx, opts.Name, opts.Args...)
	if err != nil {
		return nil, fmt.Errorf("prepare %s %v: %w", opts.Name, opts.Args, err)
	}

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if len(opts.Env) > 0 {
		env := os.Environ()
		keys := make([]string, 0, len(opts.Env))
		for key := range opts.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			env = append(env, fmt.Sprintf("%s=%s", key, opts.Env[key]))
		}
		cmd.Env = env
	}
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	return cmd, nil
}

// Run executes a command and returns captured stdout/stderr.
func (e *DefaultExecutor) Run(ctx context.Context, opts RunOptions) ([]byte, error) {
	cmd, err := e.prepare(ctx, opts)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if opts.Stdout != nil {
		cmd.Stdout = io.MultiWriter(&stdout, opts.Stdout)
	} else {
		cmd.Stdout = &stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = io.MultiWriter(&stderr, opts.Stderr)
	} else {
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		output := bytes.TrimSpace(append(stdout.Bytes(), stderr.Bytes()...))
		if len(output) == 0 {
			return nil, fmt.Errorf("command failed: %s %v: %w", opts.Name, opts.Args, err)
		}
		return output, fmt.Errorf("command failed: %s %v: %w\n%s", opts.Name, opts.Args, err, string(output))
	}

	return bytes.TrimSpace(append(stdout.Bytes(), stderr.Bytes()...)), nil
}

// RunStreaming executes a command while streaming stdout/stderr to the caller.
func (e *DefaultExecutor) RunStreaming(ctx context.Context, opts RunOptions) error {
	cmd, err := e.prepare(ctx, opts)
	if err != nil {
		return err
	}

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v: %w", opts.Name, opts.Args, err)
	}

	return nil
}
