#!/bin/bash
# Master test runner for acceptance tests

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Source helper functions
source setup/test-helpers.sh

# Initialize
echo "========================================="
echo "Timber-Git Acceptance Test Suite"
echo "========================================="
echo ""

# Cleanup any previous test runs first
echo "Cleaning up previous test runs..."
bash setup/cleanup.sh
echo ""

# Setup
echo "Setting up test environment..."
bash setup/create-repos.sh
echo ""

# Track results
PASSED=0
FAILED=0
TOTAL=0

# Run test suites
run_test_suite "Suite 1: Clone Command" "tests/suite-1-*.sh"
run_test_suite "Suite 2: Checkout Command" "tests/suite-2-*.sh"
run_test_suite "Suite 3: List Command" "tests/suite-3-*.sh"
run_test_suite "Suite 4: Remove Command" "tests/suite-4-*.sh"
run_test_suite "Suite 5: Status Commands" "tests/suite-5-*.sh"
run_test_suite "Suite 6: Git Operations" "tests/suite-6-*.sh"
run_test_suite "Suite 7: Edge Cases" "tests/suite-7-*.sh"
run_test_suite "Suite 8: Move to Main Command" "tests/suite-8-*.sh"

# Summary
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Total Tests: $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

# Cleanup
echo "Cleaning up..."
bash setup/cleanup.sh

# Exit with appropriate code
if [ $FAILED -eq 0 ]; then
    echo "All tests passed!"
    exit 0
else
    echo "Some tests failed!"
    exit 1
fi
