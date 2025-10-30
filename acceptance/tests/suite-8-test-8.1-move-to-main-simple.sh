#!/bin/bash
# Test 8.1: Move to Main - Simple File Changes
# Objective: Verify `timber-git move-to-main` moves uncommitted changes from feature branch to main

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Create a feature branch and make changes
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b move-test-feature 2>/dev/null || true

cd move-test-feature

# Make some changes to files
echo "// Added feature code" >> src/main.go
echo "// Added utility function" >> src/utils.go
echo "Feature documentation update" >> README.md

# Capture the state before move-to-main
git status --porcelain > ../../results/test-8.1-before-status.txt
find . -type f -not -path './.git' | sort > ../../results/test-8.1-before-files.txt

# Test - Run move-to-main (this should move changes without committing in feature branch)
echo "y" | timber-git move-to-main 2>&1 | tee ../../results/test-8.1-tg-output.txt

# Verify changes are no longer in feature branch
git status --porcelain > ../../results/test-8.1-after-feature-status.txt

# Check that main branch now has the changes
cd ../main
git status --porcelain > ../../results/test-8.1-after-main-status.txt

# Verify the files have the changes
grep -q "Added feature code" src/main.go && echo "MAIN_HAS_MAIN_GO" > ../../results/test-8.1-main-verify.txt
grep -q "Added utility function" src/utils.go && echo "MAIN_HAS_UTILS_GO" >> ../../results/test-8.1-main-verify.txt
grep -q "Feature documentation update" README.md && echo "MAIN_HAS_README" >> ../../results/test-8.1-main-verify.txt

# Comparison with git behavior
cd "$TEST_ROOT/comparison"
echo "Move to Main Test:" > test-8.1-comparison.txt
echo -e "\nBefore move-to-main (feature branch status):" >> test-8.1-comparison.txt
cat ../timber-git/results/test-8.1-before-status.txt >> test-8.1-comparison.txt
echo -e "\nAfter move-to-main (feature branch status - should be clean):" >> test-8.1-comparison.txt
cat ../timber-git/results/test-8.1-after-feature-status.txt >> test-8.1-comparison.txt
echo -e "\nAfter move-to-main (main branch status - should have changes):" >> test-8.1-comparison.txt
cat ../timber-git/results/test-8.1-after-main-status.txt >> test-8.1-comparison.txt

# Expected:
# - Feature branch should be clean after move-to-main
# - Main branch should have the uncommitted changes
# - Files should be identical between branches

# Verify expectations
if [ -f "../timber-git/results/test-8.1-main-verify.txt" ] && \
   grep -q "MAIN_HAS_MAIN_GO" "../timber-git/results/test-8.1-main-verify.txt" && \
   grep -q "MAIN_HAS_UTILS_GO" "../timber-git/results/test-8.1-main-verify.txt" && \
   grep -q "MAIN_HAS_README" "../timber-git/results/test-8.1-main-verify.txt"; then
    echo "✓ Test 8.1 PASSED"
    exit 0
else
    echo "✗ Test 8.1 FAILED"
    cat ../timber-git/results/test-8.1-main-verify.txt 2>/dev/null || echo "Verification file not found"
    exit 1
fi
