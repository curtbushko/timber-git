#!/bin/bash
# Test 8.2: Move to Main - New Files
# Objective: Verify `timber-git move-to-main` moves new files from feature branch to main

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Create a feature branch and add new files
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b move-test-newfiles 2>/dev/null || true

cd move-test-newfiles

# Create new files
mkdir -p src/features
echo "package features" > src/features/new_feature.go
echo "// New feature implementation" >> src/features/new_feature.go

mkdir -p docs
echo "# New Documentation" > docs/feature.md
echo "This is new documentation for the feature" >> docs/feature.md

# Stage the new files (git add)
git add src/features/new_feature.go docs/feature.md

# Capture the state before move-to-main
git status --porcelain > ../../results/test-8.2-before-status.txt
find . -type f -not -path './.git*' | sort > ../../results/test-8.2-before-files.txt

# Test - Run move-to-main
echo "y" | timber-git move-to-main 2>&1 | tee ../../results/test-8.2-tg-output.txt

# Verify changes are no longer in feature branch
git status --porcelain > ../../results/test-8.2-after-feature-status.txt

# Check that main branch now has the new files
cd ../main
git status --porcelain > ../../results/test-8.2-after-main-status.txt

# Verify the new files exist in main
test -f src/features/new_feature.go && echo "MAIN_HAS_NEW_FEATURE_GO" > ../../results/test-8.2-main-verify.txt
test -f docs/feature.md && echo "MAIN_HAS_FEATURE_MD" >> ../../results/test-8.2-main-verify.txt

# Verify content
grep -q "New feature implementation" src/features/new_feature.go && echo "CONTENT_VERIFIED_GO" >> ../../results/test-8.2-main-verify.txt
grep -q "new documentation" docs/feature.md && echo "CONTENT_VERIFIED_MD" >> ../../results/test-8.2-main-verify.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Move to Main (New Files) Test:" > test-8.2-comparison.txt
echo -e "\nBefore move-to-main (feature branch status):" >> test-8.2-comparison.txt
cat ../timber-git/results/test-8.2-before-status.txt >> test-8.2-comparison.txt
echo -e "\nAfter move-to-main (main branch status - should have new files):" >> test-8.2-comparison.txt
cat ../timber-git/results/test-8.2-after-main-status.txt >> test-8.2-comparison.txt

# Expected:
# - Feature branch should be clean after move-to-main
# - Main branch should have the new files
# - Files should exist with correct content

# Verify expectations
if [ -f "../timber-git/results/test-8.2-main-verify.txt" ] && \
   grep -q "MAIN_HAS_NEW_FEATURE_GO" "../timber-git/results/test-8.2-main-verify.txt" && \
   grep -q "MAIN_HAS_FEATURE_MD" "../timber-git/results/test-8.2-main-verify.txt" && \
   grep -q "CONTENT_VERIFIED_GO" "../timber-git/results/test-8.2-main-verify.txt" && \
   grep -q "CONTENT_VERIFIED_MD" "../timber-git/results/test-8.2-main-verify.txt"; then
    echo "✓ Test 8.2 PASSED"
    exit 0
else
    echo "✗ Test 8.2 FAILED"
    cat ../timber-git/results/test-8.2-main-verify.txt 2>/dev/null || echo "Verification file not found"
    exit 1
fi
