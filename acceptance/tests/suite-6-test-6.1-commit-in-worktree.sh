#!/bin/bash
# Test 6.1: Commit in Worktree
# Objective: Verify git commits work correctly in timber-git worktrees

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple/main"
echo "// new code" >> src/main.go
git add src/main.go
git commit -m "Test commit in worktree"
git log -1 --oneline > ../../results/test-6.1-tg-commit.txt
git show HEAD --stat > ../../results/test-6.1-tg-show.txt

# Verify commit appears in bare repo
cd ../.bare
git log main -1 --oneline > ../../results/test-6.1-tg-bare-commit.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Commit Comparison:" > test-6.1-comparison.txt
diff ../timber-git/results/test-6.1-tg-commit.txt \
     ../timber-git/results/test-6.1-tg-bare-commit.txt >> test-6.1-comparison.txt || true

# Expected:
# - Commit should succeed in worktree
# - Commit should appear in bare repository
# - Both should show same commit

# Verify expectations
if [ -f "../timber-git/results/test-6.1-tg-commit.txt" ] && \
   [ -f "../timber-git/results/test-6.1-tg-bare-commit.txt" ]; then
    # Check if commits match
    WORKTREE_COMMIT=$(cat ../timber-git/results/test-6.1-tg-commit.txt | awk '{print $1}')
    BARE_COMMIT=$(cat ../timber-git/results/test-6.1-tg-bare-commit.txt | awk '{print $1}')
    if [ "$WORKTREE_COMMIT" = "$BARE_COMMIT" ]; then
        echo "✓ Test 6.1 PASSED"
        exit 0
    fi
fi

echo "✗ Test 6.1 FAILED"
exit 1
