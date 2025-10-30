#!/bin/bash
# Test 8.3: Move to Main - Deleted Files
# Objective: Verify `timber-git move-to-main` handles file deletions correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Create a feature branch and delete files
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b move-test-deletes 2>/dev/null || true

cd move-test-deletes

# Delete a file and stage the deletion
rm src/utils.go
git add src/utils.go

# Capture the state before move-to-main
git status --porcelain > ../../results/test-8.3-before-status.txt
find . -type f -not -path './.git*' | sort > ../../results/test-8.3-before-files.txt

# Test - Run move-to-main
echo "y" | timber-git move-to-main 2>&1 | tee ../../results/test-8.3-tg-output.txt

# Verify changes are no longer in feature branch (file should be back)
git status --porcelain > ../../results/test-8.3-after-feature-status.txt
test -f src/utils.go && echo "FEATURE_FILE_RESTORED" > ../../results/test-8.3-feature-verify.txt

# Check that main branch has the deletion
cd ../main
git status --porcelain > ../../results/test-8.3-after-main-status.txt

# Verify the file is deleted in main
if [ ! -f src/utils.go ]; then
    echo "MAIN_FILE_DELETED" > ../../results/test-8.3-main-verify.txt
else
    echo "MAIN_FILE_STILL_EXISTS" > ../../results/test-8.3-main-verify.txt
fi

# Comparison
cd "$TEST_ROOT/comparison"
echo "Move to Main (Deleted Files) Test:" > test-8.3-comparison.txt
echo -e "\nBefore move-to-main (feature branch status):" >> test-8.3-comparison.txt
cat ../timber-git/results/test-8.3-before-status.txt >> test-8.3-comparison.txt
echo -e "\nAfter move-to-main (main branch status - should show deletion):" >> test-8.3-comparison.txt
cat ../timber-git/results/test-8.3-after-main-status.txt >> test-8.3-comparison.txt

# Expected:
# - Feature branch should have file restored after move-to-main
# - Main branch should have the file deleted
# - Deletion should be staged in main

# Verify expectations
if [ -f "../timber-git/results/test-8.3-main-verify.txt" ] && \
   grep -q "MAIN_FILE_DELETED" "../timber-git/results/test-8.3-main-verify.txt" && \
   [ -f "../timber-git/results/test-8.3-feature-verify.txt" ] && \
   grep -q "FEATURE_FILE_RESTORED" "../timber-git/results/test-8.3-feature-verify.txt"; then
    echo "✓ Test 8.3 PASSED"
    exit 0
else
    echo "✗ Test 8.3 FAILED"
    cat ../timber-git/results/test-8.3-main-verify.txt 2>/dev/null || echo "Main verification file not found"
    cat ../timber-git/results/test-8.3-feature-verify.txt 2>/dev/null || echo "Feature verification file not found"
    exit 1
fi
