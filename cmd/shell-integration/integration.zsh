#!/usr/bin/env zsh
# openCenter Shell Integration for Zsh
# This script enables session-scoped cluster selection

# Generate unique session ID for this shell
export OPENCENTER_SESSION_ID="${OPENCENTER_SESSION_ID:-$$-$(date +%s)}"
export OPENCENTER_SESSION_FILE="${HOME}/.config/openCenter/.session-${OPENCENTER_SESSION_ID}"

# Ensure config directory exists
mkdir -p "${HOME}/.config/openCenter"

# Load existing session cluster if present
if [[ -f "$OPENCENTER_SESSION_FILE" ]]; then
    export OPENCENTER_CLUSTER=$(cat "$OPENCENTER_SESSION_FILE" 2>/dev/null | tr -d '\n')
fi

# Cleanup session file on shell exit
zshexit() {
    rm -f "$OPENCENTER_SESSION_FILE" 2>/dev/null
}

# Function to get current cluster (respects precedence)
opencenter_current_cluster() {
    if [[ -n "$OPENCENTER_CLUSTER" ]]; then
        echo "$OPENCENTER_CLUSTER"
    elif [[ -f "$OPENCENTER_SESSION_FILE" ]]; then
        cat "$OPENCENTER_SESSION_FILE" 2>/dev/null | tr -d '\n'
    else
        opencenter cluster active --quiet 2>/dev/null || echo ""
    fi
}

# Function to get short cluster name (without organization prefix)
opencenter_current_cluster_short() {
    local cluster=$(opencenter_current_cluster)
    if [[ -n "$cluster" ]]; then
        echo "${cluster##*/}"  # Remove everything before last /
    fi
}

# Wrapper for 'opencenter cluster use' that evaluates the output
opencenter() {
    if [[ "$1" == "cluster" && "$2" == "use" && -n "$OPENCENTER_SESSION_FILE" ]]; then
        # Capture the output and evaluate it to set environment variable
        local output
        output=$(command opencenter "$@" 2>&1)
        local exit_code=$?
        
        # Evaluate shell environment commands
        if [[ $exit_code -eq 0 ]]; then
            echo "$output" | grep -E '^(export|unset) ' | while read -r line; do
                eval "$line"
            done
            # Print non-shell lines to stderr
            echo "$output" | grep -vE '^(export|unset) ' >&2
        else
            echo "$output" >&2
            return $exit_code
        fi
    else
        command opencenter "$@"
    fi
}

# Optional: Add cluster to prompt
# Uncomment one of these to enable prompt integration:

# Option 1: Simple prefix (requires PROMPT_SUBST)
# setopt PROMPT_SUBST
# PROMPT='%F{cyan}$(opencenter_current_cluster_short)%f${PROMPT:+ }$PROMPT'

# Option 2: Bracketed format
# setopt PROMPT_SUBST
# PROMPT='%F{cyan}[$(opencenter_current_cluster_short)]%f${PROMPT:+ }$PROMPT'

# Option 3: Right prompt
# setopt PROMPT_SUBST
# RPROMPT='%F{cyan}[$(opencenter_current_cluster_short)]%f'

# Option 4: Only show if cluster is set
# setopt PROMPT_SUBST
# opencenter_prompt() {
#     local cluster=$(opencenter_current_cluster_short)
#     if [[ -n "$cluster" ]]; then
#         echo "%F{cyan}[$cluster]%f "
#     fi
# }
# PROMPT='$(opencenter_prompt)$PROMPT'
