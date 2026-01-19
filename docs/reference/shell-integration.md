---
title: Shell Integration Reference
doc_type: reference
category: reference
weight: 40
---

# Shell Integration Reference

Shell integration provides functions, aliases, and prompt customization for displaying the active cluster in your shell. Integration scripts are located in `hack/shell-integration/`.

## Installation

Shell integration files must be manually sourced or copied to your shell configuration directory.

**Files**:
- `hack/shell-integration/shell-integration.sh` - Bash/Zsh integration
- `hack/shell-integration/shell-integration.fish` - Fish shell integration
- `hack/shell-integration/starship-opencenter.toml` - Starship prompt configuration

## Setup by Shell

### Bash

Add to `~/.bashrc`:

```bash
source /path/to/openCenter-cli/hack/shell-integration/shell-integration.sh

# Optional: Add to prompt
PS1="\$(opencenter_prompt)$PS1"
```

### Zsh

Add to `~/.zshrc`:

```bash
source /path/to/openCenter-cli/hack/shell-integration/shell-integration.sh

# Optional: Add to prompt
PROMPT="\$(opencenter_prompt)$PROMPT"
```

### Fish

Copy to Fish config directory:

```fish
cp /path/to/openCenter-cli/hack/shell-integration/shell-integration.fish \
   ~/.config/fish/conf.d/opencenter.fish
```

Add to prompt function in `~/.config/fish/functions/fish_prompt.fish`:

```fish
echo -n (opencenter_prompt)
```

### Starship

Append to `~/.config/starship.toml`:

```bash
cat /path/to/openCenter-cli/hack/shell-integration/starship-opencenter.toml \
    >> ~/.config/starship.toml
```

Or manually add:

```toml
[custom.opencenter]
command = "cat ~/.config/openCenter/.active 2>/dev/null || echo ''"
when = "test -f ~/.config/openCenter/.active"
format = "[$symbol$output]($style) "
symbol = "🚀 "
style = "bold blue"
description = "Show active openCenter cluster"
```

## Functions

### opencenter_active

Returns the active cluster name from `~/.config/openCenter/.active`.

**Output**: Cluster identifier (e.g., `myorg/mycluster` or `mycluster`)

**Example**:
```bash
$ opencenter_active
myorg/production-cluster
```

### opencenter_prompt

Returns formatted prompt string with brackets.

**Output**: `[cluster]` or empty string if no active cluster

**Example**:
```bash
$ opencenter_prompt
[myorg/production-cluster]
```

### opencenter_active_short

Returns short cluster name without organization prefix.

**Output**: Cluster name only (e.g., `mycluster`)

**Example**:
```bash
$ opencenter_active_short
production-cluster
```

### opencenter_update_env

Updates `$OPENCENTER_ACTIVE_CLUSTER` environment variable. Called automatically by shell hooks.

**Behavior**: Reads from cache file, updates environment variable

## Aliases

| Alias | Command | Description |
|-------|---------|-------------|
| `oc-active` | `opencenter_active` | Get active cluster |
| `oc-status` | `openCenter cluster status` | Show cluster status |
| `oc-select` | `openCenter cluster select` | Select active cluster |
| `oc-list` | `openCenter cluster list` | List all clusters |

## Environment Variables

### OPENCENTER_ACTIVE_CLUSTER

Set automatically by `opencenter_update_env`. Contains the current active cluster identifier.

**Type**: Read-only (managed by shell integration)

**Example**:
```bash
$ echo $OPENCENTER_ACTIVE_CLUSTER
myorg/production-cluster
```

## Caching Behavior

Shell integration uses file-based caching for performance.

**Cache file**: `~/.cache/openCenter/active_cluster`

**Update trigger**: When `~/.config/openCenter/.active` is newer than cache

**Performance**: Sub-millisecond reads after initial cache

**Cache invalidation**: Automatic when active cluster changes

## Prompt Examples

### Bash

```bash
# Basic
PS1="\$(opencenter_prompt)$PS1"

# With color
PS1="\[\033[36m\]\$(opencenter_prompt)\[\033[0m\]$PS1"
```

### Zsh

```zsh
# Basic
PROMPT="\$(opencenter_prompt)$PROMPT"

# With color
PROMPT="%F{cyan}\$(opencenter_prompt)%f$PROMPT"
```

### Fish

```fish
function fish_prompt
    echo -n (opencenter_prompt)
    # Existing prompt code
end
```

### Oh My Zsh

```zsh
opencenter_prompt_info() {
    local cluster=$(opencenter_active 2>/dev/null)
    if [[ -n "$cluster" ]]; then
        echo "%{$fg[cyan]%}[$cluster]%{$reset_color%} "
    fi
}

PROMPT='$(opencenter_prompt_info)'$PROMPT
```

## Troubleshooting

### Prompt not showing

Check active cluster:
```bash
openCenter cluster current
```

Test function:
```bash
opencenter_prompt
```

Verify integration loaded:
```bash
type opencenter_active
```

### Function not found

Source the integration script:
```bash
source /path/to/hack/shell-integration/shell-integration.sh
```

For Fish:
```bash
ls ~/.config/fish/conf.d/opencenter.fish
```

### Stale cluster name

Clear cache:
```bash
rm ~/.cache/openCenter/active_cluster
```

## Advanced Patterns

### Conditional display

Show cluster only in specific directories:

```bash
opencenter_conditional_prompt() {
    if [[ "$PWD" == *"/k8s/"* ]] || [[ "$PWD" == *"/clusters/"* ]]; then
        opencenter_prompt
    fi
}

PS1="\$(opencenter_conditional_prompt)$PS1"
```

### Combined indicators

Show both openCenter cluster and kubectl context:

```bash
k8s_prompt() {
    local oc_cluster=$(opencenter_active 2>/dev/null)
    local k8s_context=$(kubectl config current-context 2>/dev/null)
    
    if [[ -n "$oc_cluster" ]]; then
        echo -n "[oc:$oc_cluster]"
    fi
    if [[ -n "$k8s_context" ]]; then
        echo -n "[k8s:$k8s_context]"
    fi
}

PS1="\$(k8s_prompt)$PS1"
```

## Shell Completion

Enable command completion:

```bash
# Bash
openCenter completion bash > /etc/bash_completion.d/openCenter

# Zsh
openCenter completion zsh > "${fpath[1]}/_openCenter"

# Fish
openCenter completion fish > ~/.config/fish/completions/openCenter.fish
```

## Related Documentation

- [CLI Commands Reference](./cli-commands.md)
- [Environment Variables](./environment-variables.md)
- [Cluster Commands](./cluster/README.md)