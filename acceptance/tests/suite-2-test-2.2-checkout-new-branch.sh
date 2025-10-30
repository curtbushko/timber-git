#!/bin/bash
# Test 2.2: Checkout with -b Flag (Create New Branch)
# Objective: Verify `timber-git checkout -b` creates new branch and worktree

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b new-feature

# Verify worktree created
test -d new-feature && echo "WORKTREE_EXISTS" > ../results/test-2.2-tg-worktree-check.txt

cd new-feature
git branch --show-current > ../../results/test-2.2-tg-branch.txt
git log --oneline -1 > ../../results/test-2.2-tg-log.txt

# Verify branch exists in bare repo
cd ../.bare
git branch -a | grep new-feature > ../../results/test-2.2-tg-branch-exists.txt

# Expected:
# - `new-feature/` worktree should exist
# - Current branch should be new-feature
# - Branch should exist in bare repository
# - Should be based on current main branch

# Verify expectations
if [ -f "../../results/test-2.2-tg-worktree-check.txt" ] && \
   grep -q "WORKTREE_EXISTS" "../../results/test-2.2-tg-worktree-check.txt" && \
   [ -f "../../results/test-2.2-tg-branch.txt" ] && \
   grep -q "new-feature" "../../results/test-2.2-tg-branch.txt" && \
   [ -f "../../results/test-2.2-tg-branch-exists.txt" ] && \
   grep -q "new-feature" "../../results/test-2.2-tg-branch-exists.txt"; then
    echo "✓ Test 2.2 PASSED"
    exit 0
else
    echo "✗ Test 2.2 FAILED"
    exit 1
fi
