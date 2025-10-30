#!/bin/bash
# Test 3.1: List All Worktrees
# Objective: Verify `timber-git list` shows all worktrees

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Ensure multiple worktrees exist
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout feature-a 2>/dev/null || true
timber-git checkout feature-b 2>/dev/null || true

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git list > ../results/test-3.1-tg-list.txt

# Also get actual directory listing
ls -la > ../results/test-3.1-tg-dirs.txt

# Compare with git worktree list
cd .bare
git worktree list > ../../results/test-3.1-git-worktree-list.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Worktree Listing:" > test-3.1-comparison.txt
echo -e "\ntimber-git list output:" >> test-3.1-comparison.txt
cat ../timber-git/results/test-3.1-tg-list.txt >> test-3.1-comparison.txt
echo -e "\ngit worktree list output:" >> test-3.1-comparison.txt
cat ../timber-git/results/test-3.1-git-worktree-list.txt >> test-3.1-comparison.txt

# Expected:
# - Should list all worktrees (main, feature-a, feature-b)
# - Should show branch names and paths
# - Output should align with `git worktree list`

# Verify expectations
if [ -f "../timber-git/results/test-3.1-tg-list.txt" ] && \
   grep -q "main" "../timber-git/results/test-3.1-tg-list.txt" && \
   grep -q "feature-a" "../timber-git/results/test-3.1-tg-list.txt" && \
   grep -q "feature-b" "../timber-git/results/test-3.1-tg-list.txt"; then
    echo "✓ Test 3.1 PASSED"
    exit 0
else
    echo "✗ Test 3.1 FAILED"
    exit 1
fi
