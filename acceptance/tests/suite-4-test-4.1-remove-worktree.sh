#!/bin/bash
# Test 4.1: Remove Worktree
# Objective: Verify `timber-git remove` deletes worktree correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b temp-branch 2>/dev/null || true
test -d temp-branch && echo "SETUP: WORKTREE_EXISTS" > ../results/test-4.1-tg-setup.txt

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git remove temp-branch

# Verify worktree removed
test -d temp-branch && echo "ERROR: STILL_EXISTS" > ../results/test-4.1-tg-removed.txt || echo "REMOVED" > ../results/test-4.1-tg-removed.txt

# Verify branch still exists (or doesn't, depending on flags)
cd .bare
git branch -a > ../../results/test-4.1-tg-branches.txt

# Expected:
# - `temp-branch/` directory should be removed
# - Git worktree should be removed
# - Branch handling depends on implementation (may or may not delete branch)

# Verify expectations
if [ -f "../../results/test-4.1-tg-removed.txt" ] && \
   grep -q "REMOVED" "../../results/test-4.1-tg-removed.txt"; then
    echo "✓ Test 4.1 PASSED"
    exit 0
else
    echo "✗ Test 4.1 FAILED"
    exit 1
fi
