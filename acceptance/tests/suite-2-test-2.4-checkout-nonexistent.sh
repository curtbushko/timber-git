#!/bin/bash
# Test 2.4: Checkout Non-Existent Branch
# Objective: Verify error handling for invalid branch name

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git/test-repo-simple"
set +e  # Temporarily disable exit on error since we expect this to fail
timber-git checkout nonexistent-branch 2>&1 | tee ../results/test-2.4-tg-error.txt
EXIT_CODE=${PIPESTATUS[0]}
set -e  # Re-enable exit on error
echo "Exit code: $EXIT_CODE" >> ../results/test-2.4-tg-error.txt

# Verify no worktree created
test -d nonexistent-branch && echo "ERROR: DIR_EXISTS" > ../results/test-2.4-tg-no-dir.txt || echo "CORRECT: NO_DIR" > ../results/test-2.4-tg-no-dir.txt

# Expected:
# - Clear error message
# - Non-zero exit code
# - No worktree directory created

# Verify expectations
if [ $EXIT_CODE -ne 0 ] && \
   [ -f "../results/test-2.4-tg-no-dir.txt" ] && \
   grep -q "NO_DIR" "../results/test-2.4-tg-no-dir.txt"; then
    echo "✓ Test 2.4 PASSED"
    exit 0
else
    echo "✗ Test 2.4 FAILED"
    exit 1
fi
