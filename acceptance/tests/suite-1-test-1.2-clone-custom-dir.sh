#!/bin/bash
# Test 1.2: Clone with Custom Target Directory
# Objective: Verify clone respects custom target directory specification

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git"
timber-git clone "$TEST_ROOT/standard-git/repos/test-repo-simple" custom-name

# Verify custom directory created
test -d custom-name && echo "CUSTOM_DIR_EXISTS" > results/test-1.2-tg-custom-check.txt
test -d custom-name/.bare && echo "BARE_EXISTS" >> results/test-1.2-tg-custom-check.txt
test -d custom-name/main && echo "MAIN_EXISTS" >> results/test-1.2-tg-custom-check.txt

# Expected:
# - Repository should be cloned to `custom-name/` directory
# - Same bare + worktree structure

# Verify expectations
if [ -f "results/test-1.2-tg-custom-check.txt" ] && \
   grep -q "CUSTOM_DIR_EXISTS" "results/test-1.2-tg-custom-check.txt" && \
   grep -q "BARE_EXISTS" "results/test-1.2-tg-custom-check.txt" && \
   grep -q "MAIN_EXISTS" "results/test-1.2-tg-custom-check.txt" ]; then
    echo "✓ Test 1.2 PASSED"
    exit 0
else
    echo "✗ Test 1.2 FAILED"
    exit 1
fi
