#!/bin/bash
# Helper functions for test execution

# Run a single test
run_test() {
    local test_name=$1
    local test_script=$2

    echo "Running: $test_name"
    if bash "$test_script"; then
        echo "  ✓ PASSED"
        return 0
    else
        echo "  ✗ FAILED"
        return 1
    fi
}

# Run a test suite
run_test_suite() {
    local suite_name=$1
    local pattern=$2

    echo ""
    echo "========================================="
    echo "$suite_name"
    echo "========================================="

    for test_file in $pattern; do
        if [ -f "$test_file" ]; then
            TOTAL=$((TOTAL + 1))
            if run_test "$(basename $test_file)" "$test_file"; then
                PASSED=$((PASSED + 1))
            else
                FAILED=$((FAILED + 1))
            fi
        fi
    done
}

# Compare files and report
compare_files() {
    local file1=$1
    local file2=$2
    local description=$3

    if diff -q "$file1" "$file2" > /dev/null; then
        echo "  ✓ $description: MATCH"
        return 0
    else
        echo "  ✗ $description: DIFFER"
        diff "$file1" "$file2" || true
        return 1
    fi
}

# Verify directory exists
verify_dir_exists() {
    local dir_path=$1
    local description=$2

    if [ -d "$dir_path" ]; then
        echo "  ✓ $description: EXISTS"
        return 0
    else
        echo "  ✗ $description: MISSING"
        return 1
    fi
}

# Verify file contains expected content
verify_content() {
    local file_path=$1
    local expected=$2
    local description=$3

    if grep -q "$expected" "$file_path"; then
        echo "  ✓ $description: FOUND"
        return 0
    else
        echo "  ✗ $description: NOT FOUND"
        cat "$file_path"
        return 1
    fi
}
