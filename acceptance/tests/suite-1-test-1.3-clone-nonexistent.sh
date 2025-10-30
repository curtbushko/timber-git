#!/bin/bash
# Test 1.3: Clone Non-Existent Repository
# Objective: Verify error handling for invalid repository path

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SCRIPT_DIR/.."

# Test
cd "$TEST_ROOT/timber-git"
set +e  # Temporarily disable exit on error since we expect this to fail
timber-git clone "$TEST_ROOT/standard-git/repos/does-not-exist" 2>&1 | tee results/test-1.3-tg-error.txt
EXIT_CODE=${PIPESTATUS[0]}
set -e  # Re-enable exit on error
echo "Exit code: $EXIT_CODE" >> results/test-1.3-tg-error.txt

# Verify no partial directory created
ls -la | grep "does-not-exist" > results/test-1.3-tg-no-dir.txt || echo "NO_DIR" > results/test-1.3-tg-no-dir.txt

# Expected:
# - Clear error message
# - Non-zero exit code
# - No partial directory structure created

# Verify expectations
if [ $EXIT_CODE -ne 0 ] && \
   [ -f "results/test-1.3-tg-no-dir.txt" ] && \
   grep -q "NO_DIR" "results/test-1.3-tg-no-dir.txt" ]; then
    echo "✓ Test 1.3 PASSED"
    exit 0
else
    echo "✗ Test 1.3 FAILED"
    exit 1
fi
