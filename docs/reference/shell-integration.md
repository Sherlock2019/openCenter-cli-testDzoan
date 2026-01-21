# Shell Integration

openCenter provides shell integration for session-scoped cluster selection, allowing multiple terminal sessions to work with different clusters independently.

## Overview

By default, `opencenter cluster select` sets a persistent cluster selection that affects all terminal sessions. With shell integration enabled, you can use `opencenter cluster use` to switch clusters only in the current terminal session.

## Features

- **Session Isolation**: Each terminal has its own active cluster context
- **Visual Feedback**: Optional prompt integration shows the active cluster
- **Automatic Cleanup**: Session files are cleaned up when the shell exits
- **Backward Compatible**: Falls back to persistent selection if shell integration is not active

## Installation

### Quick Install

```bash
# Install shell integration automatically
opencenter config ide --shell-integration
```

This will:
1. Detect your shell (bash, zsh, or fish)
2. Add the integration line to your shell RC file
3. Provide instructions for activation

### Manual Install

Add this line to your shell configuration file:

**Bash** (`~/.bashrc`):
```bash
eval "$(opencenter shell-init)"
```

**Zsh** (`~/.zshrc`):
```bash
eval "$(opencenter shell-init)"
```

**Fish** (`~/.config/fish/config.fish`):
```fish
opencenter shell-init --shell fish | source
```

Then reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc
```

## Usage

### Basic Commands

```bash
# Switch to a cluster in the current session
opencenter cluster use prod-cluster

# Switch to a cluster with organization
opencenter cluster use myorg/prod-cluster

# Show current cluster and its source
opencenter cluster current
# Output: prod-cluster (session)

# Show only cluster name (for scripting)
opencenter cluster current --quiet
# Output: prod-cluster
```

### Cluster Selection Precedence

The active cluster is determined by this precedence order:

1. **OPENCENTER_CLUSTER** environment variable (highest priority)
2. **Session file** (if shell integration is active)
3. **Persistent selection** (fallback)

### Example Workflow

```bash
# Terminal 1: Work on production cluster
opencenter cluster use prod-cluster
opencenter cluster status
# All commands operate on prod-cluster

# Terminal 2: Work on development cluster (independent)
opencenter cluster use dev-cluster
opencenter cluster validate
# All commands operate on dev-cluster

# Both terminals maintain their own cluster context
```

## Prompt Integration

You can add the active cluster to your shell prompt for visual feedback.

### Bash

Add to your `~/.bashrc` after the shell integration line:

```bash
# Option 1: Simple prefix
PS1="\[\033[36m\]\$(opencenter_current_cluster_short)\[\033[0m\] $PS1"

# Option 2: Bracketed format
PS1="\[\033[36m\][\$(opencenter_current_cluster_short)]\[\033[0m\] $PS1"

# Option 3: Only show if cluster is set
opencenter_prompt() {
    local cluster=$(opencenter_current_cluster_short)
    if [[ -n "$cluster" ]]; then
        echo -e "\033[36m[$cluster]\033[0m "
    fi
}
PS1="\$(opencenter_prompt)$PS1"
```

### Zsh

Add to your `~/.zshrc` after the shell integration line:

```zsh
# Enable prompt substitution
setopt PROMPT_SUBST

# Option 1: Left prompt
PROMPT='%F{cyan}$(opencenter_current_cluster_short)%f $PROMPT'

# Option 2: Right prompt
RPROMPT='%F{cyan}[$(opencenter_current_cluster_short)]%f'

# Option 3: Only show if cluster is set
opencenter_prompt() {
    local cluster=$(opencenter_current_cluster_short)
    if [[ -n "$cluster" ]]; then
        echo "%F{cyan}[$cluster]%f "
    fi
}
PROMPT='$(opencenter_prompt)$PROMPT'
```

### Fish

Add to your `~/.config/fish/config.fish`:

```fish
# Option 1: Left prompt
function fish_prompt
    set -l cluster (opencenter_current_cluster_short)
    if test -n "$cluster"
        set_color cyan
        echo -n "[$cluster] "
        set_color normal
    end
    # Add your existing prompt here
    echo -n "> "
end

# Option 2: Right prompt
function fish_right_prompt
    set -l cluster (opencenter_current_cluster_short)
    if test -n "$cluster"
        set_color cyan
        echo -n "[$cluster]"
        set_color normal
    end
end
```

### Starship Integration

If you use [Starship](https://starship.rs/), add this to your `~/.config/starship.toml`:

```toml
[env_var.OPENCENTER_CLUSTER]
variable = "OPENCENTER_CLUSTER"
format = "[$env_value]($style) "
style = "cyan bold"
disabled = false
```

## Environment Variables

- **OPENCENTER_CLUSTER**: Current cluster name (set by `opencenter cluster use`)
- **OPENCENTER_SESSION_ID**: Unique session identifier
- **OPENCENTER_SESSION_FILE**: Path to session file storing cluster selection

## Helper Functions

The shell integration provides these helper functions:

- `opencenter_current_cluster()`: Get the current cluster name (full path with organization)
- `opencenter_current_cluster_short()`: Get the short cluster name (without organization prefix)

## Troubleshooting

### Shell integration not detected

If you see this warning:
```
⚠️  Shell integration not detected. Setting persistent cluster selection.
💡 To enable session-scoped selection, run: eval "$(opencenter shell-init)"
```

Make sure you've:
1. Added the integration line to your shell RC file
2. Reloaded your shell or started a new terminal session

### Session file not cleaned up

Session files are automatically cleaned up when the shell exits. If you find stale session files in `~/.config/openCenter/.session-*`, you can safely delete them.

### Cluster not switching

Make sure you're using `opencenter cluster use` (not `opencenter cluster select`) for session-scoped selection.

## Comparison with Persistent Selection

| Feature | `cluster select` | `cluster use` (with shell integration) |
|---------|------------------|----------------------------------------|
| Scope | Global (all terminals) | Session (current terminal only) |
| Persistence | Survives shell restart | Lost on shell exit |
| Use case | Default cluster for all work | Temporary cluster switching |
| Requires setup | No | Yes (shell integration) |

## See Also

- [CLI Commands Reference](cli-commands.md)
- [Configuration](configuration.md)
- [IDE Integration](../how-to/ide-integration.md)
