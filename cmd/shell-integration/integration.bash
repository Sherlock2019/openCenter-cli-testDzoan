#!/usr/bin/env bash
# openCenter Shell Integration for Bash
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
trap 'rm -f "$OPENCENTER_SESSION_FILE" 2>/dev/null' EXIT

# Function to get current cluster (respects precedence)
opencenter_current_cluster() {
    if [[ -n "$OPENCENTER_CLUSTER" ]]; then
        echo "$OPENCENTER_CLUSTER"
    elif [[ -f "$OPENCENTER_SESSION_FILE" ]]; then
        cat "$OPENCENTER_SESSION_FILE" 2>/dev/null | tr -d '\n'
    else
        opencenter cluster current --quiet 2>/dev/null || echo ""
    fi
}

# Function to get short cluster name (without organization prefix)
opencenter_current_cluster_short() {
    local cluster=$(opencenter_current_cluster)
    if [[ -n "$cluster" ]]; then
        echo "${cluster##*/}"  # Remove everything before last /
    fi
}

# Wrapper for 'opencenter cluster select' that evaluates the output
opencenter() {
    if [[ "$1" == "cluster" && "$2" == "select" && -n "$OPENCENTER_SESSION_FILE" ]]; then
        # Capture the output and evaluate it to set environment variable
        local output
        output=$(command opencenter "$@" 2>&1)
        local exit_code=$?
        
        # Evaluate any export commands
        if [[ $exit_code -eq 0 ]]; then
            echo "$output" | grep -E '^export ' | while read -r line; do
                eval "$line"
            done
            # Print non-export lines to stderr
            echo "$output" | grep -v '^export ' >&2
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

# Option 1: Simple prefix
# PS1="\[\033[36m\]\$(opencenter_current_cluster_short)\[\033[0m\]${PS1:+ }$PS1"

# Option 2: Bracketed format
# PS1="\[\033[36m\][\$(opencenter_current_cluster_short)]\[\033[0m\]${PS1:+ }$PS1"

# Option 3: Only show if cluster is set
# opencenter_prompt() {
#     local cluster=$(opencenter_current_cluster_short)
#     if [[ -n "$cluster" ]]; then
#         echo -e "\033[36m[$cluster]\033[0m "
#     fi
# }
# PS1="\$(opencenter_prompt)$PS1"
