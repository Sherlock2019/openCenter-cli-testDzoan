# Confirmation Prompts

## Overview

The opencenter CLI provides a testable confirmation prompt system through the `ConfirmationPrompter` interface in the `internal/ui` package. This allows commands to prompt users for confirmation while remaining fully testable.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Usage](#usage)
  - [Interactive Mode](#interactive-mode)
  - [Test Mode](#test-mode)
  - [Automatic Mode Selection](#automatic-mode-selection)
- [Context Support](#context-support)
- [Testing](#testing)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Architecture

The confirmation prompt system consists of three main components:

1. **ConfirmationPrompter Interface**: Defines the contract for confirmation prompts
2. **InteractivePrompter**: Implementation for interactive user prompts
3. **TestPrompter**: Implementation for automated testing

```go
type ConfirmationPrompter interface {
    Confirm(ctx context.Context, message string) (bool, error)
}
```

## Usage

### Interactive Mode

In interactive mode, the `InteractivePrompter` reads from stdin and writes to stdout:

```go
import (
    "context"
    "os"
    "github.com/rackerlabs/opencenter-cli/internal/ui"
)

func main() {
    prompter := ui.NewInteractivePrompter(os.Stdin, os.Stdout)
    ctx := context.Background()
    
    confirmed, err := prompter.Confirm(ctx, "Are you sure you want to proceed?")
    if err != nil {
        log.Fatal(err)
    }
    
    if confirmed {
        fmt.Println("Proceeding with operation...")
    } else {
        fmt.Println("Operation cancelled")
    }
}
```

The prompter accepts the following responses as confirmation:
- `y` or `Y`
- `yes` or `YES` or `Yes`

Any other response (including `n`, `no`, or empty input) is treated as denial.

### Test Mode

In test mode, the `TestPrompter` returns a predetermined response:

```go
import (
    "context"
    "github.com/rackerlabs/opencenter-cli/internal/ui"
)

func TestDestroyCommand(t *testing.T) {
    // Create a prompter that always confirms
    prompter := ui.NewTestPrompter(true)
    ctx := context.Background()
    
    confirmed, err := prompter.Confirm(ctx, "Destroy cluster?")
    require.NoError(t, err)
    assert.True(t, confirmed)
}

func TestDestroyCommandCancelled(t *testing.T) {
    // Create a prompter that always denies
    prompter := ui.NewTestPrompter(false)
    ctx := context.Background()
    
    confirmed, err := prompter.Confirm(ctx, "Destroy cluster?")
    require.NoError(t, err)
    assert.False(t, confirmed)
}
```

### Automatic Mode Selection

The `GetPrompter` function automatically selects the appropriate prompter:

```go
import (
    "context"
    "os"
    "github.com/rackerlabs/opencenter-cli/internal/ui"
)

func runCommand(cmd *cobra.Command, args []string) error {
    // Automatically select prompter based on environment
    testMode := os.Getenv("OPENCENTER_TEST_MODE") != ""
    prompter := ui.GetPrompter(os.Stdin, cmd.OutOrStdout(), testMode)
    
    ctx := context.Background()
    confirmed, err := prompter.Confirm(ctx, "Continue with operation?")
    if err != nil {
        return err
    }
    
    if !confirmed {
        fmt.Fprintln(cmd.OutOrStdout(), "Operation cancelled")
        return nil
    }
    
    // Proceed with operation
    return performOperation()
}
```

## Context Support

All prompters support context cancellation and timeouts:

### Timeout Example

```go
import (
    "context"
    "time"
    "github.com/rackerlabs/opencenter-cli/internal/ui"
)

func main() {
    // Create a context with 30-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    prompter := ui.NewInteractivePrompter(os.Stdin, os.Stdout)
    confirmed, err := prompter.Confirm(ctx, "Are you sure?")
    
    if err == context.DeadlineExceeded {
        fmt.Println("Confirmation timed out")
        return
    }
    if err != nil {
        log.Fatal(err)
    }
    
    // Process confirmation
}
```

### Cancellation Example

```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "github.com/rackerlabs/opencenter-cli/internal/ui"
)

func main() {
    // Create a context that cancels on interrupt signal
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigChan
        cancel()
    }()
    
    prompter := ui.NewInteractivePrompter(os.Stdin, os.Stdout)
    confirmed, err := prompter.Confirm(ctx, "Are you sure?")
    
    if err == context.Canceled {
        fmt.Println("Operation interrupted")
        return
    }
    // Process confirmation
}
```

## Testing

### Unit Testing Commands

When testing commands that use confirmation prompts:

```go
func TestClusterDestroy_Confirmed(t *testing.T) {
    // Set test mode environment variable
    os.Setenv("OPENCENTER_TEST_MODE", "1")
    defer os.Unsetenv("OPENCENTER_TEST_MODE")
    
    // Create command
    cmd := newClusterDestroyCmd()
    
    // Execute command (will use TestPrompter automatically)
    err := cmd.Execute()
    require.NoError(t, err)
    
    // Verify cluster was destroyed
    _, err = config.Load("test-cluster")
    assert.Error(t, err) // Should not exist
}

func TestClusterDestroy_WithForceFlag(t *testing.T) {
    cmd := newClusterDestroyCmd()
    cmd.SetArgs([]string{"test-cluster", "--force"})
    
    // With --force, no prompt is shown
    err := cmd.Execute()
    require.NoError(t, err)
}
```

### Integration Testing

For integration tests that need to simulate user input:

```go
func TestInteractiveDestroy(t *testing.T) {
    // Create a pipe to simulate user input
    reader, writer := io.Pipe()
    
    // Write user response in a goroutine
    go func() {
        writer.Write([]byte("yes\n"))
        writer.Close()
    }()
    
    // Create prompter with simulated input
    prompter := ui.NewInteractivePrompter(reader, os.Stdout)
    ctx := context.Background()
    
    confirmed, err := prompter.Confirm(ctx, "Destroy cluster?")
    require.NoError(t, err)
    assert.True(t, confirmed)
}
```

## Examples

### Cluster Destroy Command

The `cluster destroy` command uses the confirmation prompter:

```go
func runClusterDestroy(cmd *cobra.Command, args []string) error {
    name := args[0]
    
    // Skip confirmation if --force flag is used
    if !force {
        testMode := os.Getenv("OPENCENTER_TEST_MODE") != ""
        prompter := ui.GetPrompter(os.Stdin, cmd.OutOrStdout(), testMode)
        
        message := fmt.Sprintf(
            "WARNING: This will permanently destroy cluster %q. Are you sure?",
            name,
        )
        
        ctx := context.Background()
        confirmed, err := prompter.Confirm(ctx, message)
        if err != nil {
            return fmt.Errorf("confirmation prompt failed: %w", err)
        }
        
        if !confirmed {
            fmt.Fprintln(cmd.OutOrStdout(), "Destroy operation cancelled.")
            return nil
        }
    }
    
    // Proceed with destroy operation
    return destroyCluster(name)
}
```

### Custom Prompter for Specific Tests

```go
func TestCustomPrompter(t *testing.T) {
    // Create a custom prompter for specific test scenarios
    type CustomPrompter struct {
        responses []bool
        index     int
    }
    
    func (p *CustomPrompter) Confirm(ctx context.Context, msg string) (bool, error) {
        if p.index >= len(p.responses) {
            return false, fmt.Errorf("no more responses")
        }
        response := p.responses[p.index]
        p.index++
        return response, nil
    }
    
    // Use custom prompter
    prompter := &CustomPrompter{
        responses: []bool{true, false, true},
    }
    
    // Test multiple confirmations
    confirmed1, _ := prompter.Confirm(context.Background(), "First?")
    assert.True(t, confirmed1)
    
    confirmed2, _ := prompter.Confirm(context.Background(), "Second?")
    assert.False(t, confirmed2)
    
    confirmed3, _ := prompter.Confirm(context.Background(), "Third?")
    assert.True(t, confirmed3)
}
```

## Best Practices

### 1. Always Use Context

Always pass a context to `Confirm()` to support cancellation and timeouts:

```go
// Good
ctx := context.Background()
confirmed, err := prompter.Confirm(ctx, "Continue?")

// Bad - but still works
confirmed, err := prompter.Confirm(context.Background(), "Continue?")
```

### 2. Provide Clear Messages

Make confirmation messages clear and specific:

```go
// Good
message := fmt.Sprintf(
    "WARNING: This will permanently destroy cluster %q in organization %q. "+
    "This action cannot be undone. Are you sure?",
    clusterName, organization,
)

// Bad
message := "Are you sure?"
```

### 3. Handle Errors Properly

Always check for errors from `Confirm()`:

```go
confirmed, err := prompter.Confirm(ctx, message)
if err != nil {
    if err == context.Canceled {
        return fmt.Errorf("operation cancelled by user")
    }
    return fmt.Errorf("confirmation prompt failed: %w", err)
}
```

### 4. Respect the --force Flag

Always provide a `--force` flag to skip confirmation:

```go
if !force {
    // Show confirmation prompt
    confirmed, err := prompter.Confirm(ctx, message)
    // ...
}
```

### 5. Use Test Mode in Tests

Set the `OPENCENTER_TEST_MODE` environment variable in tests:

```go
func TestCommand(t *testing.T) {
    os.Setenv("OPENCENTER_TEST_MODE", "1")
    defer os.Unsetenv("OPENCENTER_TEST_MODE")
    
    // Test code here
}
```

### 6. Write Output to Command's Writer

Use the command's output writer for consistency:

```go
// Good
fmt.Fprintln(cmd.OutOrStdout(), "Operation cancelled")

// Bad
fmt.Println("Operation cancelled")
```

## Related Documentation

- [Error Formatting](error-formatter.md)
- [Testing Guide](testing.md)
- [Command Development](command-development.md)
