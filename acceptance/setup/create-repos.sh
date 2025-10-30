#!/bin/bash
# Creates all test repositories with proper structure

set -e

SETUP_DIR="$(cd "$(dirname "$0")" && pwd)"
STANDARD_GIT_DIR="$SETUP_DIR/../standard-git/repos"
TIMBER_GIT_DIR="$SETUP_DIR/../timber-git"

# Create workspace directories
mkdir -p "$STANDARD_GIT_DIR"
mkdir -p "$STANDARD_GIT_DIR/results"
mkdir -p "$TIMBER_GIT_DIR/results"
mkdir -p "$SETUP_DIR/../comparison"

# Function to create a simple repository
create_simple_repo() {
    local repo_name=$1
    local branches=$2

    echo "Creating $repo_name..."

    # Remove existing repository if it exists
    if [ -d "$STANDARD_GIT_DIR/$repo_name" ]; then
        echo "  Removing existing $repo_name..."
        rm -rf "$STANDARD_GIT_DIR/$repo_name"
    fi

    # Create in standard git workspace
    mkdir -p "$STANDARD_GIT_DIR/$repo_name"
    cd "$STANDARD_GIT_DIR/$repo_name"
    git init

    # Create initial content
    echo "# $repo_name" > README.md
    mkdir -p src
    cat > src/main.go <<'EOF'
// Package main is the entry point for the application.
package main
EOF
    cat > src/utils.go <<'EOF'
// Package utils provides utility functions.
package utils
EOF
    echo "module github.com/test/$repo_name" > go.mod

    git add .
    git commit -m "Initial commit"
    git branch -M main

    # Create additional branches with changes
    for branch in $(echo $branches | tr ',' ' '); do
        if [ "$branch" != "main" ]; then
            git checkout -b $branch
            echo "// $branch changes" >> src/main.go
            echo "Updated in $branch" >> README.md
            git add .
            git commit -m "Changes for $branch"
            git checkout main
        fi
    done
}

# Function to create complex repository with merge history
create_complex_repo() {
    local repo_name=$1

    echo "Creating $repo_name with complex history..."

    # Remove existing repository if it exists
    if [ -d "$STANDARD_GIT_DIR/$repo_name" ]; then
        echo "  Removing existing $repo_name..."
        rm -rf "$STANDARD_GIT_DIR/$repo_name"
    fi

    mkdir -p "$STANDARD_GIT_DIR/$repo_name"
    cd "$STANDARD_GIT_DIR/$repo_name"
    git init

    # Initial commit
    echo "# $repo_name" > README.md
    mkdir -p cmd pkg docs
    cat > cmd/main.go <<'EOF'
// Package main is the entry point for the application.
package main
EOF
    cat > pkg/lib.go <<'EOF'
// Package pkg provides library functions.
package pkg
EOF
    echo "# Documentation" > docs/README.md
    git add .
    git commit -m "Initial commit"
    git branch -M main

    # Create develop branch
    git checkout -b develop
    echo "// develop work" >> cmd/main.go
    git add cmd/main.go
    git commit -m "Develop branch work"

    # Create feature branch from develop
    git checkout -b feature-x
    echo "// feature x" >> pkg/lib.go
    git add pkg/lib.go
    git commit -m "Add feature x"

    # Merge feature into develop
    git checkout develop
    git merge feature-x -m "Merge feature-x into develop"

    # Create hotfix from main
    git checkout main
    git checkout -b hotfix-123
    echo "// hotfix" >> cmd/main.go
    git add cmd/main.go
    git commit -m "Hotfix #123"

    # Merge hotfix into main
    git checkout main
    git merge hotfix-123 -m "Merge hotfix-123"

    git checkout main
}

# Create test repositories
create_simple_repo "test-repo-simple" "main,feature-a,feature-b,dev"
create_complex_repo "test-repo-complex"

echo "Test repositories created successfully!"
