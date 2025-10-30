#!/bin/bash
# Test 8.4: Move to Main - Mixed Changes (Add, Modify, Delete)
# Objective: Verify `timber-git move-to-main` handles all types of changes together

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Setup - Create a feature branch with mixed changes
cd "$TEST_ROOT/timber-git/test-repo-simple"
timber-git checkout -b move-test-mixed 2>/dev/null || true

cd move-test-mixed

# 1. Modify existing files
echo "// Modified main code" >> src/main.go

# 2. Add new file
mkdir -p tests
echo "package tests" > tests/test.go
echo "// Test code" >> tests/test.go
git add tests/test.go

# 3. Delete a file
rm go.mod
git add go.mod

# 4. Modify another file
echo "Updated README content" >> README.md

# Capture the state before move-to-main
git status --porcelain > ../../results/test-8.4-before-status.txt
git diff --stat HEAD > ../../results/test-8.4-before-diff.txt || true

# Test - Run move-to-main
echo "y" | timber-git move-to-main 2>&1 | tee ../../results/test-8.4-tg-output.txt

# Verify changes are no longer in feature branch
git status --porcelain > ../../results/test-8.4-after-feature-status.txt

# Check that main branch now has all the changes
cd ../main
git status --porcelain > ../../results/test-8.4-after-main-status.txt

# Verify each type of change:
# 1. Modified file
grep -q "Modified main code" src/main.go && echo "MODIFIED_MAIN_GO" > ../../results/test-8.4-main-verify.txt

# 2. New file
test -f tests/test.go && echo "NEW_TEST_GO" >> ../../results/test-8.4-main-verify.txt
grep -q "Test code" tests/test.go && echo "NEW_TEST_CONTENT" >> ../../results/test-8.4-main-verify.txt

# 3. Deleted file
if [ ! -f go.mod ]; then
    echo "DELETED_GO_MOD" >> ../../results/test-8.4-main-verify.txt
fi

# 4. Modified another file
grep -q "Updated README content" README.md && echo "MODIFIED_README" >> ../../results/test-8.4-main-verify.txt

# Comparison
cd "$TEST_ROOT/comparison"
echo "Move to Main (Mixed Changes) Test:" > test-8.4-comparison.txt
echo -e "\nBefore move-to-main (feature branch status):" >> test-8.4-comparison.txt
cat ../timber-git/results/test-8.4-before-status.txt >> test-8.4-comparison.txt
echo -e "\nAfter move-to-main (main branch status):" >> test-8.4-comparison.txt
cat ../timber-git/results/test-8.4-after-main-status.txt >> test-8.4-comparison.txt

# Expected:
# - All modifications should be in main
# - New files should be in main
# - Deleted files should be deleted in main
# - Feature branch should be clean

# Verify expectations
VERIFY_FILE="../timber-git/results/test-8.4-main-verify.txt"
if [ -f "$VERIFY_FILE" ] && \
   grep -q "MODIFIED_MAIN_GO" "$VERIFY_FILE" && \
   grep -q "NEW_TEST_GO" "$VERIFY_FILE" && \
   grep -q "NEW_TEST_CONTENT" "$VERIFY_FILE" && \
   grep -q "DELETED_GO_MOD" "$VERIFY_FILE" && \
   grep -q "MODIFIED_README" "$VERIFY_FILE"; then
    echo "✓ Test 8.4 PASSED"
    exit 0
else
    echo "✗ Test 8.4 FAILED"
    cat "$VERIFY_FILE" 2>/dev/null || echo "Verification file not found"
    exit 1
fi
