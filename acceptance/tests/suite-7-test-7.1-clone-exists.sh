#!/bin/bash
# Test 7.1: Clone Already Exists
# Objective: Verify behavior when cloning to existing directory

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git"
mkdir -p existing-dir
set +e  # Temporarily disable exit on error since we expect this to fail
timber-git clone "$TEST_ROOT/standard-git/repos/test-repo-simple" existing-dir 2>&1 | tee results/test-7.1-tg-error.txt
EXIT_CODE=${PIPESTATUS[0]}
set -e  # Re-enable exit on error
echo "Exit code: $EXIT_CODE" >> results/test-7.1-tg-error.txt

# Expected:
# - Should error cleanly
# - Should not overwrite existing directory

# Verify expectations
if [ $EXIT_CODE -ne 0 ]; then
    echo "✓ Test 7.1 PASSED"
    exit 0
else
    echo "✗ Test 7.1 FAILED"
    exit 1
fi
