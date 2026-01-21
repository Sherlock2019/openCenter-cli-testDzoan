#!/usr/bin/env fish
# openCenter Shell Integration for Fish
# This script enables session-scoped cluster selection

# Generate unique session ID for this shell
if not set -q OPENCENTER_SESSION_ID
    set -gx OPENCENTER_SESSION_ID "$fish_pid-"(date +%s)
end

set -gx OPENCENTER_SESSION_FILE "$HOME/.config/openCenter/.session-$OPENCENTER_SESSION_ID"

# Ensure config directory exists
mkdir -p "$HOME/.config/openCenter"

# Load existing session cluster if present
if test -f "$OPENCENTER_SESSION_FILE"
    set -gx OPENCENTER_CLUSTER (cat "$OPENCENTER_SESSION_FILE" 2>/dev/null | tr -d '\n')
end

# Cleanup session file on shell exit
function __opencenter_cleanup --on-event fish_exit
    rm -f "$OPENCENTER_SESSION_FILE" 2>/dev/null
end

# Function to get current cluster (respects precedence)
function opencenter_current_cluster
    if set -q OPENCENTER_CLUSTER; and test -n "$OPENCENTER_CLUSTER"
        echo "$OPENCENTER_CLUSTER"
    else if test -f "$OPENCENTER_SESSION_FILE"
        cat "$OPENCENTER_SESSION_FILE" 2>/dev/null | tr -d '\n'
    else
        opencenter cluster current --quiet 2>/dev/null; or echo ""
    end
end

# Function to get short cluster name (without organization prefix)
function opencenter_current_cluster_short
    set -l cluster (opencenter_current_cluster)
    if test -n "$cluster"
        # Remove everything before last /
        echo $cluster | sed 's/.*\///'
    end
end

# Wrapper for 'opencenter cluster select' that evaluates the output
function opencenter
    if test "$argv[1]" = "cluster"; and test "$argv[2]" = "select"; and test -n "$OPENCENTER_SESSION_FILE"
        # Capture the output
        set -l output (command opencenter $argv 2>&1)
        set -l exit_code $status
        
        if test $exit_code -eq 0
            # Evaluate any export commands
            for line in $output
                if string match -q -r '^export ' -- $line
                    # Parse export command: export VAR=value
                    set -l var_value (string replace 'export ' '' -- $line)
                    set -l var (string split -m 1 '=' -- $var_value)[1]
                    set -l value (string split -m 1 '=' -- $var_value)[2]
                    set -gx $var $value
                else
                    echo $line >&2
                end
            end
        else
            echo $output >&2
            return $exit_code
        end
    else
        command opencenter $argv
    end
end

# Optional: Add cluster to prompt
# Uncomment one of these to enable prompt integration:

# Option 1: Simple prefix
# function fish_prompt
#     set -l cluster (opencenter_current_cluster_short)
#     if test -n "$cluster"
#         set_color cyan
#         echo -n "[$cluster] "
#         set_color normal
#     end
#     # Add your existing prompt here
#     echo -n "> "
# end

# Option 2: Right prompt
# function fish_right_prompt
#     set -l cluster (opencenter_current_cluster_short)
#     if test -n "$cluster"
#         set_color cyan
#         echo -n "[$cluster]"
#         set_color normal
#     end
# end
