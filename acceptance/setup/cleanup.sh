#!/bin/bash
# Cleanup test environment

set -e

SETUP_DIR="$(cd "$(dirname "$0")" && pwd)"
TEST_ROOT="$SETUP_DIR/.."

echo "Cleaning up test environment..."

# Remove workspaces but keep setup
rm -rf "$TEST_ROOT/standard-git/repos"
rm -rf "$TEST_ROOT/standard-git/results"
rm -rf "$TEST_ROOT/timber-git/results"
rm -rf "$TEST_ROOT/timber-git"/*.git
rm -rf "$TEST_ROOT/timber-git"/test-repo-*
rm -rf "$TEST_ROOT/timber-git"/custom-name
rm -rf "$TEST_ROOT/timber-git"/does-not-exist
rm -rf "$TEST_ROOT/timber-git"/existing-dir
rm -rf "$TEST_ROOT/timber-git"/nonexistent-repo
rm -rf "$TEST_ROOT/comparison"

# Recreate empty directories
mkdir -p "$TEST_ROOT/standard-git/repos"
mkdir -p "$TEST_ROOT/standard-git/results"
mkdir -p "$TEST_ROOT/timber-git/results"
mkdir -p "$TEST_ROOT/comparison"

echo "Cleanup complete. Run create-repos.sh to recreate test environment."
