#!/usr/bin/env bash
# Test script for shell integration
# This verifies that the shell integration works correctly

set -e

echo "=== Testing openCenter Shell Integration ==="
echo

# Test 1: Check if shell-init command works
echo "Test 1: Checking shell-init command..."
if ./bin/openCenter shell-init --shell bash > /dev/null 2>&1; then
    echo "✓ shell-init command works"
else
    echo "✗ shell-init command failed"
    exit 1
fi

# Test 2: Check if integration script defines required functions
echo
echo "Test 2: Checking if integration defines helper functions..."
SCRIPT_OUTPUT=$(./bin/openCenter shell-init --shell bash)

if echo "$SCRIPT_OUTPUT" | grep -q "opencenter_current_cluster()"; then
    echo "✓ opencenter_current_cluster() function defined"
else
    echo "✗ opencenter_current_cluster() function not found"
    exit 1
fi

if echo "$SCRIPT_OUTPUT" | grep -q "opencenter_current_cluster_short()"; then
    echo "✓ opencenter_current_cluster_short() function defined"
else
    echo "✗ opencenter_current_cluster_short() function not found"
    exit 1
fi

# Test 3: Check if integration sets up session variables
echo
echo "Test 3: Checking session variable setup..."
if echo "$SCRIPT_OUTPUT" | grep -q "OPENCENTER_SESSION_ID"; then
    echo "✓ OPENCENTER_SESSION_ID variable setup found"
else
    echo "✗ OPENCENTER_SESSION_ID variable setup not found"
    exit 1
fi

if echo "$SCRIPT_OUTPUT" | grep -q "OPENCENTER_SESSION_FILE"; then
    echo "✓ OPENCENTER_SESSION_FILE variable setup found"
else
    echo "✗ OPENCENTER_SESSION_FILE variable setup not found"
    exit 1
fi

# Test 4: Check zsh integration
echo
echo "Test 4: Checking zsh-specific integration..."
ZSH_OUTPUT=$(./bin/openCenter shell-init --shell zsh)

if echo "$ZSH_OUTPUT" | grep -q "setopt PROMPT_SUBST"; then
    echo "⚠ Prompt integration is enabled by default (should be commented)"
else
    echo "✓ Prompt integration is optional (commented out)"
fi

echo
echo "=== All tests passed ==="
echo
echo "To test manually:"
echo "1. eval \"\$(./bin/openCenter shell-init)\""
echo "2. type opencenter_current_cluster_short"
echo "3. ./bin/openCenter cluster use test-cluster"
echo "4. opencenter_current_cluster_short"
