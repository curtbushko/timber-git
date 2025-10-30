#!/bin/bash
# Test 2.1: Checkout Existing Branch
# Objective: Verify `timber-git checkout` creates worktree for existing branch

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Get expected state from git
cd "$TEST_ROOT/standard-git/repos/test-repo-simple"
git checkout feature-a
find . -type f -not -path './.git/*' | sort > ../../results/test-2.1-git-files.txt
git log --oneline > ../../results/test-2.1-git-log.txt
git checkout main

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout feature-a

# Verify worktree created
test -d feature-a && echo "WORKTREE_EXISTS" > ../results/test-2.1-tg-worktree-check.txt
ls -la > ../results/test-2.1-tg-worktrees.txt

# Verify files in worktree
cd feature-a
find . -type f -not -path './.git' | sort > ../../results/test-2.1-tg-files.txt
git log --oneline > ../../results/test-2.1-tg-log.txt
git branch --show-current > ../../results/test-2.1-tg-branch.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Checkout Comparison:" > test-2.1-comparison.txt
echo -e "\nFiles:" >> test-2.1-comparison.txt
diff ../standard-git/results/test-2.1-git-files.txt \
     ../timber-git/results/test-2.1-tg-files.txt >> test-2.1-comparison.txt || true

echo -e "\nLog:" >> test-2.1-comparison.txt
diff ../standard-git/results/test-2.1-git-log.txt \
     ../timber-git/results/test-2.1-tg-log.txt >> test-2.1-comparison.txt || true

# Expected:
# - `feature-a/` worktree should exist
# - Files should match git checkout of feature-a
# - Current branch should be feature-a

# Verify expectations
if [ -f "../timber-git/results/test-2.1-tg-worktree-check.txt" ] && \
   grep -q "WORKTREE_EXISTS" "../timber-git/results/test-2.1-tg-worktree-check.txt" ] && \
   [ -f "../timber-git/results/test-2.1-tg-branch.txt" ] && \
   grep -q "feature-a" "../timber-git/results/test-2.1-tg-branch.txt"; then
    echo "✓ Test 2.1 PASSED"
    exit 0
else
    echo "✗ Test 2.1 FAILED"
    exit 1
fi
