#!/bin/bash
# Test 1.1: Basic Clone - Create Bare Repository
# Objective: Verify `timber-git clone` creates a bare repository with proper structure

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup
cd "$TEST_ROOT/standard-git/repos/test-repo-simple"
git log --oneline --all --graph > ../../results/test-1.1-git-log.txt
git branch -a > ../../results/test-1.1-git-branches.txt
find . -type f -not -path './.git/*' | sort > ../../results/test-1.1-git-files.txt

# Test
cd "$TEST_ROOT/timber-git"
timber-git clone "$TEST_ROOT/standard-git/repos/test-repo-simple"

# Verify bare repository structure
ls -la test-repo-simple/ > results/test-1.1-tg-structure.txt
test -d test-repo-simple/.bare && echo "BARE_EXISTS" > results/test-1.1-tg-bare-check.txt
test -d test-repo-simple/main && echo "MAIN_EXISTS" >> results/test-1.1-tg-main-check.txt

# Verify git history in bare repo
cd test-repo-simple/.bare
git log --oneline --all --graph > ../../results/test-1.1-tg-bare-log.txt
git branch -a > ../../results/test-1.1-tg-bare-branches.txt

# Verify main worktree
cd ../main
find . -type f -not -path './.git' | sort > ../../results/test-1.1-tg-main-files.txt
git log --oneline > ../../results/test-1.1-tg-main-log.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Clone Structure Check:" > test-1.1-comparison.txt
cat ../timber-git/results/test-1.1-tg-bare-check.txt >> test-1.1-comparison.txt
cat ../timber-git/results/test-1.1-tg-main-check.txt >> test-1.1-comparison.txt

echo -e "\nFile Comparison:" >> test-1.1-comparison.txt
diff ../standard-git/results/test-1.1-git-files.txt \
     ../timber-git/results/test-1.1-tg-main-files.txt >> test-1.1-comparison.txt || true

echo -e "\nBranch Comparison:" >> test-1.1-comparison.txt
diff ../standard-git/results/test-1.1-git-branches.txt \
     ../timber-git/results/test-1.1-tg-bare-branches.txt >> test-1.1-comparison.txt || true

# Expected:
# - `.bare/` directory should exist containing bare repository
# - `main/` worktree should exist with all files checked out
# - Git history should be identical
# - All branches should be available in bare repository

# Verify expectations
if [ -f "../timber-git/results/test-1.1-tg-bare-check.txt" ] && \
   grep -q "BARE_EXISTS" "../timber-git/results/test-1.1-tg-bare-check.txt" && \
   grep -q "MAIN_EXISTS" "../timber-git/results/test-1.1-tg-main-check.txt" ]; then
    echo "✓ Test 1.1 PASSED"
    exit 0
else
    echo "✗ Test 1.1 FAILED"
    exit 1
fi
