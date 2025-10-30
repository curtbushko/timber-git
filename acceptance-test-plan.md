# Acceptance Test Plan - Timber-Git Core Commands

## Overview

This document outlines comprehensive acceptance tests for timber-git's core commands. Tests compare timber-git behavior against standard git operations to ensure correctness and compatibility. **Note**: Submodule/symlink testing will be covered separately and is excluded from this plan.

## Test-Driven Development (TDD) Approach

### TDD Workflow

All acceptance tests should be developed following a TDD approach:

1. **Write the acceptance test first** - Define the expected behavior in the acceptance test
2. **Write failing unit tests** - Create unit tests that match the acceptance test scenario
3. **Define interfaces** - Create Go interfaces for dependencies and abstractions
4. **Implement minimal code** - Write just enough code to make tests pass
5. **Refactor** - Improve code quality while keeping tests green
6. **Run acceptance test** - Verify end-to-end functionality

### Go Interface Guidelines

When implementing timber-git functionality, use interfaces for:

1. **Git Operations** - Abstract go-git operations for testability
   ```go
   type GitRepository interface {
       Clone(url, path string, opts *CloneOptions) error
       CreateWorktree(path, branch string) error
       ListWorktrees() ([]WorktreeInfo, error)
       RemoveWorktree(path string) error
   }
   ```

2. **File System Operations** - Abstract file operations for testing
   ```go
   type FileSystem interface {
       MkdirAll(path string, perm os.FileMode) error
       RemoveAll(path string) error
       Stat(path string) (os.FileInfo, error)
       ReadDir(path string) ([]os.DirEntry, error)
   }
   ```

3. **Command Execution** - Abstract for testing without actual git commands
   ```go
   type CommandExecutor interface {
       Execute(name string, args ...string) (string, error)
   }
   ```

### Unit Test Structure

Each acceptance test scenario should have corresponding unit tests:

```go
// Example: Test Clone Command
func TestClone_CreatesBarePlusWorktree(t *testing.T) {
    // Arrange
    mockGit := &MockGitRepository{}
    mockFS := &MockFileSystem{}
    cloner := NewCloner(mockGit, mockFS)

    // Act
    err := cloner.Clone("test-repo", ".")

    // Assert
    assert.NoError(t, err)
    assert.True(t, mockFS.DirExists(".bare"))
    assert.True(t, mockFS.DirExists("main"))
}
```

## Implementation Checklist

### Phase 1: Test Infrastructure Setup
- [ ] Create acceptance test directory structure
- [ ] Write `setup/create-repos.sh` script
- [ ] Write `setup/cleanup.sh` script
- [ ] Write `setup/test-helpers.sh` with reusable functions
- [ ] Create master test runner `run-all-tests.sh`
- [ ] Test infrastructure works end-to-end

### Phase 2: Clone Command - TDD Cycle
- [ ] **Acceptance Test**: Write Test 1.1 (Basic Clone) shell script
- [ ] **Interfaces**: Define `GitRepository` interface
- [ ] **Interfaces**: Define `FileSystem` interface
- [ ] **Unit Tests**: Write unit tests for clone functionality
- [ ] **Implementation**: Implement clone to pass unit tests
- [ ] **Acceptance Test**: Run Test 1.1 - should pass
- [ ] **Acceptance Test**: Write Test 1.2 (Custom Directory)
- [ ] **Unit Tests**: Write unit tests for custom directory
- [ ] **Implementation**: Implement custom directory support
- [ ] **Acceptance Test**: Run Test 1.2 - should pass
- [ ] **Acceptance Test**: Write Test 1.3 (Error Handling)
- [ ] **Unit Tests**: Write unit tests for error cases
- [ ] **Implementation**: Implement error handling
- [ ] **Acceptance Test**: Run Test 1.3 - should pass
- [ ] **Refactor**: Review and refactor clone code
- [ ] **All Tests**: Run all clone acceptance tests

### Phase 3: Checkout Command - TDD Cycle
- [ ] **Acceptance Test**: Write Test 2.1 (Checkout Existing Branch)
- [ ] **Interfaces**: Define `WorktreeManager` interface
- [ ] **Unit Tests**: Write unit tests for checkout existing branch
- [ ] **Implementation**: Implement checkout for existing branches
- [ ] **Acceptance Test**: Run Test 2.1 - should pass
- [ ] **Acceptance Test**: Write Test 2.2 (Create New Branch)
- [ ] **Unit Tests**: Write unit tests for -b flag
- [ ] **Implementation**: Implement -b flag support
- [ ] **Acceptance Test**: Run Test 2.2 - should pass
- [ ] **Acceptance Test**: Write Test 2.3 (FZF Integration)
- [ ] **Interfaces**: Define `BranchSelector` interface for FZF abstraction
- [ ] **Unit Tests**: Write unit tests for interactive selection
- [ ] **Implementation**: Implement FZF integration
- [ ] **Acceptance Test**: Run Test 2.3 - should pass
- [ ] **Acceptance Test**: Write Test 2.4 (Error Handling)
- [ ] **Unit Tests**: Write unit tests for invalid branch
- [ ] **Implementation**: Implement checkout error handling
- [ ] **Acceptance Test**: Run Test 2.4 - should pass
- [ ] **Refactor**: Review and refactor checkout code
- [ ] **All Tests**: Run all checkout acceptance tests

### Phase 4: List Command - TDD Cycle
- [ ] **Acceptance Test**: Write Test 3.1 (List All Worktrees)
- [ ] **Interfaces**: Define `WorktreeFormatter` interface for output
- [ ] **Unit Tests**: Write unit tests for list functionality
- [ ] **Implementation**: Implement worktree listing
- [ ] **Acceptance Test**: Run Test 3.1 - should pass
- [ ] **Refactor**: Review and refactor list code
- [ ] **All Tests**: Run all list acceptance tests

### Phase 5: Remove Command - TDD Cycle
- [ ] **Acceptance Test**: Write Test 4.1 (Remove Worktree)
- [ ] **Unit Tests**: Write unit tests for remove functionality
- [ ] **Implementation**: Implement worktree removal
- [ ] **Acceptance Test**: Run Test 4.1 - should pass
- [ ] **Acceptance Test**: Write Test 4.2 (Force Remove)
- [ ] **Unit Tests**: Write unit tests for force flag
- [ ] **Implementation**: Implement force removal
- [ ] **Acceptance Test**: Run Test 4.2 - should pass
- [ ] **Acceptance Test**: Write Test 4.3 (Error Handling)
- [ ] **Unit Tests**: Write unit tests for remove errors
- [ ] **Implementation**: Implement remove error handling
- [ ] **Acceptance Test**: Run Test 4.3 - should pass
- [ ] **Refactor**: Review and refactor remove code
- [ ] **All Tests**: Run all remove acceptance tests

### Phase 6: Status Command - TDD Cycle
- [ ] **Acceptance Test**: Write Test 5.1 (Show Status)
- [ ] **Interfaces**: Define `StatusReporter` interface
- [ ] **Unit Tests**: Write unit tests for status display
- [ ] **Implementation**: Implement status command
- [ ] **Acceptance Test**: Run Test 5.1 - should pass
- [ ] **Refactor**: Review and refactor status code
- [ ] **All Tests**: Run all status acceptance tests

### Phase 7: Git Operations Compatibility - TDD Cycle
- [ ] **Acceptance Test**: Write Test 6.1 (Commit in Worktree)
- [ ] **Unit Tests**: Write unit tests for commit operations
- [ ] **Implementation**: Verify commit functionality works
- [ ] **Acceptance Test**: Run Test 6.1 - should pass
- [ ] **Acceptance Test**: Write Test 6.2 (Branch Operations)
- [ ] **Unit Tests**: Write unit tests for branch operations
- [ ] **Implementation**: Verify branch operations work
- [ ] **Acceptance Test**: Run Test 6.2 - should pass
- [ ] **Acceptance Test**: Write Test 6.3 (Merge Operations)
- [ ] **Unit Tests**: Write unit tests for merge support
- [ ] **Implementation**: Verify merge operations work
- [ ] **Acceptance Test**: Run Test 6.3 - should pass
- [ ] **Acceptance Test**: Write Test 6.4 (Rebase Operations)
- [ ] **Unit Tests**: Write unit tests for rebase support
- [ ] **Implementation**: Verify rebase operations work
- [ ] **Acceptance Test**: Run Test 6.4 - should pass
- [ ] **All Tests**: Run all git operations acceptance tests

### Phase 8: Edge Cases - TDD Cycle
- [ ] **Acceptance Test**: Write Test 7.1 (Clone Exists)
- [ ] **Unit Tests**: Write unit tests for existing directory
- [ ] **Implementation**: Implement directory existence check
- [ ] **Acceptance Test**: Run Test 7.1 - should pass
- [ ] **Acceptance Test**: Write Test 7.2 (Multiple Worktrees)
- [ ] **Unit Tests**: Write unit tests for concurrent worktrees
- [ ] **Implementation**: Verify multiple worktrees work
- [ ] **Acceptance Test**: Run Test 7.2 - should pass
- [ ] **All Tests**: Run all edge case acceptance tests

### Phase 9: Integration and Documentation
- [ ] Run full acceptance test suite
- [ ] Fix any failing tests
- [ ] Document test results
- [ ] Update README with testing instructions
- [ ] Create CI/CD integration for automated testing
- [ ] Review test coverage and add missing scenarios

## Key Interfaces for Implementation

### 1. GitRepository Interface

```go
// GitRepository abstracts go-git operations for testability
type GitRepository interface {
    // Clone clones a repository to the specified path
    Clone(url, targetPath string, bare bool) error

    // CreateWorktree creates a new worktree for a branch
    CreateWorktree(branch, path string) error

    // ListWorktrees returns all worktrees for this repository
    ListWorktrees() ([]WorktreeInfo, error)

    // RemoveWorktree removes a worktree
    RemoveWorktree(path string, force bool) error

    // GetBranches returns all available branches
    GetBranches() ([]string, error)

    // BranchExists checks if a branch exists
    BranchExists(name string) (bool, error)

    // CreateBranch creates a new branch
    CreateBranch(name string) error
}

type WorktreeInfo struct {
    Path   string
    Branch string
    Clean  bool
}
```

### 2. FileSystem Interface

```go
// FileSystem abstracts file operations for testability
type FileSystem interface {
    // MkdirAll creates a directory and all parent directories
    MkdirAll(path string, perm os.FileMode) error

    // RemoveAll removes a path and all children
    RemoveAll(path string) error

    // Stat returns file info
    Stat(path string) (os.FileInfo, error)

    // Exists checks if a path exists
    Exists(path string) bool

    // ReadDir reads directory contents
    ReadDir(path string) ([]os.DirEntry, error)

    // IsDir checks if path is a directory
    IsDir(path string) bool
}
```

### 3. BranchSelector Interface

```go
// BranchSelector abstracts branch selection (FZF, CLI, etc.)
type BranchSelector interface {
    // SelectBranch prompts user to select a branch
    SelectBranch(branches []string) (string, error)

    // IsAvailable checks if selector is available (e.g., FZF installed)
    IsAvailable() bool
}

// Implementations:
// - FZFSelector - uses fzf for interactive selection
// - PromptSelector - uses simple CLI prompt
// - MockSelector - for testing
```

### 4. OutputFormatter Interface

```go
// OutputFormatter formats command output
type OutputFormatter interface {
    // FormatWorktreeList formats worktree list output
    FormatWorktreeList(worktrees []WorktreeInfo) string

    // FormatError formats error messages
    FormatError(err error) string

    // FormatSuccess formats success messages
    FormatSuccess(message string) string
}
```

## TDD Best Practices for Timber-Git

### 1. Write Tests First
Always write acceptance test scenarios before implementing features. This ensures:
- Clear understanding of expected behavior
- No over-engineering
- Good test coverage from the start

### 2. Use Interfaces for External Dependencies
Mock all external dependencies:
- Git operations (go-git)
- File system operations
- Network operations
- User input (FZF, prompts)

### 3. Test Structure Pattern

```go
// Follow Arrange-Act-Assert pattern
func TestFeature_Scenario(t *testing.T) {
    // Arrange - set up test dependencies
    mockGit := &MockGitRepository{}
    mockFS := &MockFileSystem{}
    sut := NewFeature(mockGit, mockFS)

    // Act - execute the functionality
    result, err := sut.DoSomething()

    // Assert - verify expectations
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 4. Mock Implementation Example

```go
type MockGitRepository struct {
    CloneFunc         func(url, path string, bare bool) error
    CreateWorktreeFunc func(branch, path string) error
    ListWorktreesFunc  func() ([]WorktreeInfo, error)
}

func (m *MockGitRepository) Clone(url, path string, bare bool) error {
    if m.CloneFunc != nil {
        return m.CloneFunc(url, path, bare)
    }
    return nil
}

// ... implement other interface methods
```

### 5. Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestClone_ErrorCases(t *testing.T) {
    tests := []struct {
        name        string
        url         string
        targetPath  string
        setupMock   func(*MockGitRepository, *MockFileSystem)
        expectedErr string
    }{
        {
            name:        "repository does not exist",
            url:         "invalid-repo",
            targetPath:  "test",
            setupMock:   func(g *MockGitRepository, f *MockFileSystem) {
                g.CloneFunc = func(url, path string, bare bool) error {
                    return errors.New("repository not found")
                }
            },
            expectedErr: "repository not found",
        },
        {
            name:        "target directory exists",
            url:         "valid-repo",
            targetPath:  "existing",
            setupMock:   func(g *MockGitRepository, f *MockFileSystem) {
                f.ExistsFunc = func(path string) bool { return true }
            },
            expectedErr: "directory already exists",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockGit := &MockGitRepository{}
            mockFS := &MockFileSystem{}
            tt.setupMock(mockGit, mockFS)

            cloner := NewCloner(mockGit, mockFS)
            err := cloner.Clone(tt.url, tt.targetPath)

            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
        })
    }
}
```

### 6. Integration Tests vs Unit Tests

- **Unit Tests**: Test individual functions with mocked dependencies
- **Integration Tests**: Test multiple components together (still use go-git in memory)
- **Acceptance Tests**: End-to-end tests using actual git commands and file system

### 7. Continuous Testing During Development

Run tests continuously during development:

```bash
# Watch mode for unit tests
go test ./... -v

# Run specific test
go test ./pkg/tg -run TestClone_CreatesBarePlusWorktree -v

# Run with coverage
go test ./... -cover

# Run acceptance tests
cd acceptance-tests && bash run-all-tests.sh
```

## Test Environment Setup

### Prerequisites

```bash
# Required tools
- git (standard CLI)
- timber-git (tg)
- SSH keys configured for GitHub access (optional for remote testing)
```

### Shared Test Infrastructure

All tests will use a common setup to minimize duplication:

```bash
# Test directory structure
~/acceptance-tests/
├── setup/                      # Shared setup scripts and fixtures
│   ├── create-repos.sh        # Creates test repositories
│   ├── cleanup.sh             # Cleanup script
│   └── fixtures/              # Fixture data
├── standard-git/              # Standard git operations workspace
│   ├── repos/
│   └── results/               # Captured state/outputs
├── timber-git/                # Timber-git operations workspace
│   └── results/               # Captured state/outputs
└── comparison/                # Comparison results
    ├── test-001-results.txt
    ├── test-002-results.txt
    └── summary.md
```

### Test Repository Structure

Create the following test repositories for reuse across tests:

```bash
# Repository: test-repo-simple
# - Simple repository with multiple branches
# - Branches: main, feature-a, feature-b, dev
# - Files: README.md, src/main.go, src/utils.go, go.mod
# - Commits on different branches

# Repository: test-repo-complex
# - Repository with more complex history
# - Branches: main, develop, feature-x, hotfix-123
# - Files: Multiple directories with various file types
# - Merge commits, feature branches
```

### Setup Scripts

#### `setup/create-repos.sh`

```bash
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

    # Create in standard git workspace
    mkdir -p "$STANDARD_GIT_DIR/$repo_name"
    cd "$STANDARD_GIT_DIR/$repo_name"
    git init

    # Create initial content
    echo "# $repo_name" > README.md
    mkdir -p src
    echo "package main" > src/main.go
    echo "package utils" > src/utils.go
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

    mkdir -p "$STANDARD_GIT_DIR/$repo_name"
    cd "$STANDARD_GIT_DIR/$repo_name"
    git init

    # Initial commit
    echo "# $repo_name" > README.md
    mkdir -p cmd pkg docs
    echo "package main" > cmd/main.go
    echo "package pkg" > pkg/lib.go
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
```

#### `setup/cleanup.sh`

```bash
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
rm -rf "$TEST_ROOT/comparison"

echo "Cleanup complete. Run create-repos.sh to recreate test environment."
```

## Test Cases

### Test Suite 1: Clone Command

#### Test 1.1: Basic Clone - Create Bare Repository

**Objective**: Verify `tg clone` creates a bare repository with proper structure.

**Setup**:
```bash
cd ~/acceptance-tests/standard-git/repos
cd test-repo-simple
git log --oneline --all --graph > ../../results/test-1.1-git-log.txt
git branch -a > ../../results/test-1.1-git-branches.txt
find . -type f -not -path './.git/*' | sort > ../../results/test-1.1-git-files.txt
```

**Test**:
```bash
cd ~/acceptance-tests/timber-git
tg clone ~/acceptance-tests/standard-git/repos/test-repo-simple

# Verify bare repository structure
ls -la test-repo-simple/ > results/test-1.1-tg-structure.txt
test -d test-repo-simple/.bare && echo "BARE_EXISTS" > results/test-1.1-tg-bare-check.txt
test -d test-repo-simple/main && echo "MAIN_EXISTS" > results/test-1.1-tg-main-check.txt

# Verify git history in bare repo
cd test-repo-simple/.bare
git log --oneline --all --graph > ../../results/test-1.1-tg-bare-log.txt
git branch -a > ../../results/test-1.1-tg-bare-branches.txt

# Verify main worktree
cd ../main
find . -type f -not -path './.git' | sort > ../../results/test-1.1-tg-main-files.txt
git log --oneline > ../../results/test-1.1-tg-main-log.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Clone Structure Check:" > test-1.1-comparison.txt
cat ../timber-git/results/test-1.1-tg-bare-check.txt >> test-1.1-comparison.txt
cat ../timber-git/results/test-1.1-tg-main-check.txt >> test-1.1-comparison.txt

echo -e "\nFile Comparison:" >> test-1.1-comparison.txt
diff ../standard-git/results/test-1.1-git-files.txt \
     ../timber-git/results/test-1.1-tg-main-files.txt >> test-1.1-comparison.txt || true

echo -e "\nBranch Comparison:" >> test-1.1-comparison.txt
diff ../standard-git/results/test-1.1-git-branches.txt \
     ../timber-git/results/test-1.1-tg-bare-branches.txt >> test-1.1-comparison.txt || true
```

**Expected**:
- `.bare/` directory should exist containing bare repository
- `main/` worktree should exist with all files checked out
- Git history should be identical
- All branches should be available in bare repository

---

#### Test 1.2: Clone with Custom Target Directory

**Objective**: Verify clone respects custom target directory specification.

**Test**:
```bash
cd ~/acceptance-tests/timber-git
tg clone ~/acceptance-tests/standard-git/repos/test-repo-simple custom-name

# Verify custom directory created
test -d custom-name && echo "CUSTOM_DIR_EXISTS" > results/test-1.2-tg-custom-check.txt
test -d custom-name/.bare && echo "BARE_EXISTS" >> results/test-1.2-tg-custom-check.txt
test -d custom-name/main && echo "MAIN_EXISTS" >> results/test-1.2-tg-custom-check.txt
```

**Expected**:
- Repository should be cloned to `custom-name/` directory
- Same bare + worktree structure

---

#### Test 1.3: Clone Non-Existent Repository

**Objective**: Verify error handling for invalid repository path.

**Test**:
```bash
cd ~/acceptance-tests/timber-git
tg clone ~/acceptance-tests/standard-git/repos/does-not-exist 2>&1 | tee results/test-1.3-tg-error.txt
echo "Exit code: $?" >> results/test-1.3-tg-error.txt

# Verify no partial directory created
ls -la | grep "does-not-exist" > results/test-1.3-tg-no-dir.txt || echo "NO_DIR" > results/test-1.3-tg-no-dir.txt
```

**Expected**:
- Clear error message
- Non-zero exit code
- No partial directory structure created

---

### Test Suite 2: Checkout Command (Add Worktree)

#### Test 2.1: Checkout Existing Branch

**Objective**: Verify `tg checkout` creates worktree for existing branch.

**Setup**:
```bash
cd ~/acceptance-tests/standard-git/repos/test-repo-simple
git checkout feature-a
find . -type f -not -path './.git/*' | sort > ../../results/test-2.1-git-files.txt
git log --oneline > ../../results/test-2.1-git-log.txt
git checkout main
```

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout feature-a

# Verify worktree created
test -d feature-a && echo "WORKTREE_EXISTS" > ../results/test-2.1-tg-worktree-check.txt
ls -la > ../results/test-2.1-tg-worktrees.txt

# Verify files in worktree
cd feature-a
find . -type f -not -path './.git' | sort > ../../results/test-2.1-tg-files.txt
git log --oneline > ../../results/test-2.1-tg-log.txt
git branch --show-current > ../../results/test-2.1-tg-branch.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Checkout Comparison:" > test-2.1-comparison.txt
echo -e "\nFiles:" >> test-2.1-comparison.txt
diff ../standard-git/results/test-2.1-git-files.txt \
     ../timber-git/results/test-2.1-tg-files.txt >> test-2.1-comparison.txt || true

echo -e "\nLog:" >> test-2.1-comparison.txt
diff ../standard-git/results/test-2.1-git-log.txt \
     ../timber-git/results/test-2.1-tg-log.txt >> test-2.1-comparison.txt || true
```

**Expected**:
- `feature-a/` worktree should exist
- Files should match git checkout of feature-a
- Current branch should be feature-a

---

#### Test 2.2: Checkout with -b Flag (Create New Branch)

**Objective**: Verify `tg checkout -b` creates new branch and worktree.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout -b new-feature

# Verify worktree created
test -d new-feature && echo "WORKTREE_EXISTS" > ../results/test-2.2-tg-worktree-check.txt

cd new-feature
git branch --show-current > ../../results/test-2.2-tg-branch.txt
git log --oneline -1 > ../../results/test-2.2-tg-log.txt

# Verify branch exists in bare repo
cd ../.bare
git branch -a | grep new-feature > ../results/test-2.2-tg-branch-exists.txt
```

**Expected**:
- `new-feature/` worktree should exist
- Current branch should be new-feature
- Branch should exist in bare repository
- Should be based on current main branch

---

#### Test 2.3: Checkout Interactive (FZF Mode)

**Objective**: Verify `tg checkout` without args triggers FZF selection (if available).

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple

# This test requires FZF and is interactive
# For automated testing, verify it lists branches
tg checkout --help 2>&1 | tee ../results/test-2.3-tg-help.txt

# Verify behavior when FZF not available
which fzf > ../results/test-2.3-fzf-available.txt || echo "FZF_NOT_FOUND" > ../results/test-2.3-fzf-available.txt
```

**Expected**:
- If FZF available, should show interactive selector
- If FZF not available, should show error or list branches
- Should document FZF requirement

---

#### Test 2.4: Checkout Non-Existent Branch

**Objective**: Verify error handling for invalid branch name.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout nonexistent-branch 2>&1 | tee ../results/test-2.4-tg-error.txt
echo "Exit code: $?" >> ../results/test-2.4-tg-error.txt

# Verify no worktree created
test -d nonexistent-branch && echo "ERROR: DIR_EXISTS" > ../results/test-2.4-tg-no-dir.txt || echo "CORRECT: NO_DIR" > ../results/test-2.4-tg-no-dir.txt
```

**Expected**:
- Clear error message
- Non-zero exit code
- No worktree directory created

---

### Test Suite 3: List Command

#### Test 3.1: List All Worktrees

**Objective**: Verify `tg list` shows all worktrees.

**Setup**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
# Ensure multiple worktrees exist
tg checkout feature-a 2>/dev/null || true
tg checkout feature-b 2>/dev/null || true
```

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg list > ../results/test-3.1-tg-list.txt

# Also get actual directory listing
ls -la > ../results/test-3.1-tg-dirs.txt

# Compare with git worktree list
cd .bare
git worktree list > ../../results/test-3.1-git-worktree-list.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Worktree Listing:" > test-3.1-comparison.txt
echo -e "\ntg list output:" >> test-3.1-comparison.txt
cat ../timber-git/results/test-3.1-tg-list.txt >> test-3.1-comparison.txt
echo -e "\ngit worktree list output:" >> test-3.1-comparison.txt
cat ../timber-git/results/test-3.1-git-worktree-list.txt >> test-3.1-comparison.txt
```

**Expected**:
- Should list all worktrees (main, feature-a, feature-b)
- Should show branch names and paths
- Output should align with `git worktree list`

---

### Test Suite 4: Remove Command

#### Test 4.1: Remove Worktree

**Objective**: Verify `tg remove` deletes worktree correctly.

**Setup**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout -b temp-branch 2>/dev/null || true
test -d temp-branch && echo "SETUP: WORKTREE_EXISTS" > ../results/test-4.1-tg-setup.txt
```

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg remove temp-branch

# Verify worktree removed
test -d temp-branch && echo "ERROR: STILL_EXISTS" > ../results/test-4.1-tg-removed.txt || echo "REMOVED" > ../results/test-4.1-tg-removed.txt

# Verify branch still exists (or doesn't, depending on flags)
cd .bare
git branch -a > ../../results/test-4.1-tg-branches.txt
```

**Expected**:
- `temp-branch/` directory should be removed
- Git worktree should be removed
- Branch handling depends on implementation (may or may not delete branch)

---

#### Test 4.2: Remove with Force Flag

**Objective**: Verify `tg remove -f` removes worktree even with uncommitted changes.

**Setup**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout -b dirty-branch 2>/dev/null || true
cd dirty-branch
echo "uncommitted" >> README.md
git status --porcelain > ../../results/test-4.2-tg-status.txt
```

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple

# Try without force (should fail)
tg remove dirty-branch 2>&1 | tee ../results/test-4.2-tg-error-no-force.txt
test -d dirty-branch && echo "STILL_EXISTS" > ../results/test-4.2-tg-no-force-check.txt

# Try with force
tg remove -f dirty-branch 2>&1 | tee ../results/test-4.2-tg-force.txt
test -d dirty-branch && echo "ERROR: STILL_EXISTS" > ../results/test-4.2-tg-removed.txt || echo "REMOVED" > ../results/test-4.2-tg-removed.txt
```

**Expected**:
- Without `-f`: Should error and preserve worktree
- With `-f`: Should remove worktree despite uncommitted changes

---

#### Test 4.3: Remove Non-Existent Worktree

**Objective**: Verify error handling when removing non-existent worktree.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg remove nonexistent 2>&1 | tee ../results/test-4.3-tg-error.txt
echo "Exit code: $?" >> ../results/test-4.3-tg-error.txt
```

**Expected**:
- Clear error message
- Non-zero exit code

---

### Test Suite 5: Status and Info Commands

#### Test 5.1: Show Repository Status

**Objective**: Verify `tg status` shows repository and worktree information.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg status > ../results/test-5.1-tg-status.txt

# Also capture git status for comparison
cd main
git status > ../../results/test-5.1-git-status-main.txt
cd ../feature-a
git status > ../../results/test-5.1-git-status-feature-a.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Status Output:" > test-5.1-comparison.txt
cat ../timber-git/results/test-5.1-tg-status.txt >> test-5.1-comparison.txt
```

**Expected**:
- Should show all worktrees and their status
- Should indicate clean/dirty state
- Should show current branches

---

### Test Suite 6: Git Operations in Worktrees

#### Test 6.1: Commit in Worktree

**Objective**: Verify git commits work correctly in timber-git worktrees.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple/main
echo "// new code" >> src/main.go
git add src/main.go
git commit -m "Test commit in worktree"
git log -1 --oneline > ../../results/test-6.1-tg-commit.txt
git show HEAD --stat > ../../results/test-6.1-tg-show.txt

# Verify commit appears in bare repo
cd ../.bare
git log main -1 --oneline > ../results/test-6.1-tg-bare-commit.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Commit Comparison:" > test-6.1-comparison.txt
diff ../timber-git/results/test-6.1-tg-commit.txt \
     ../timber-git/results/test-6.1-tg-bare-commit.txt >> test-6.1-comparison.txt || true
```

**Expected**:
- Commit should succeed in worktree
- Commit should appear in bare repository
- Both should show same commit

---

#### Test 6.2: Branch Operations

**Objective**: Verify git branch operations work in worktrees.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple/main
git branch test-branch-from-worktree
git branch > ../../results/test-6.2-tg-branches.txt

# Verify branch visible in bare repo
cd ../.bare
git branch > ../results/test-6.2-tg-bare-branches.txt
```

**Comparison**:
```bash
cd ~/acceptance-tests/comparison
echo "Branch Operations:" > test-6.2-comparison.txt
diff ../timber-git/results/test-6.2-tg-branches.txt \
     ../timber-git/results/test-6.2-tg-bare-branches.txt >> test-6.2-comparison.txt || true
```

**Expected**:
- Branch creation should work from worktree
- Branches should be visible in bare repository

---

#### Test 6.3: Merge Operations

**Objective**: Verify git merge works in worktrees.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout -b merge-test
cd merge-test
echo "// merge source" >> src/main.go
git add src/main.go
git commit -m "Commit to merge"

# Switch to main and merge
cd ../main
git merge merge-test -m "Test merge in worktree" > ../../results/test-6.3-tg-merge.txt
git log --oneline --graph -5 > ../../results/test-6.3-tg-log.txt
```

**Expected**:
- Merge should succeed
- Merge commit should appear in history
- Changes should be integrated

---

#### Test 6.4: Rebase Operations

**Objective**: Verify git rebase works in worktrees.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple

# Create divergent branches
tg checkout -b rebase-test
cd rebase-test
echo "// rebase test" >> src/main.go
git add src/main.go
git commit -m "Commit for rebase"

# Add commit to main
cd ../main
echo "// main change" >> README.md
git add README.md
git commit -m "Main change"

# Rebase feature on main
cd ../rebase-test
git rebase main > ../../results/test-6.4-tg-rebase.txt 2>&1
git log --oneline --graph -5 > ../../results/test-6.4-tg-log.txt
```

**Expected**:
- Rebase should succeed
- Commits should be replayed on top of main
- Linear history should result

---

### Test Suite 7: Edge Cases

#### Test 7.1: Clone Already Exists

**Objective**: Verify behavior when cloning to existing directory.

**Test**:
```bash
cd ~/acceptance-tests/timber-git
mkdir -p existing-dir
tg clone ~/acceptance-tests/standard-git/repos/test-repo-simple existing-dir 2>&1 | tee results/test-7.1-tg-error.txt
echo "Exit code: $?" >> results/test-7.1-tg-error.txt
```

**Expected**:
- Should error cleanly
- Should not overwrite existing directory

---

#### Test 7.2: Multiple Concurrent Worktrees

**Objective**: Verify multiple worktrees can exist simultaneously without conflicts.

**Test**:
```bash
cd ~/acceptance-tests/timber-git/test-repo-simple
tg checkout feature-a 2>/dev/null || true
tg checkout feature-b 2>/dev/null || true
tg checkout dev 2>/dev/null || true

# Make changes in each
cd feature-a
echo "// feature a work" >> src/main.go
git add src/main.go && git commit -m "Work in feature-a"

cd ../feature-b
echo "// feature b work" >> src/utils.go
git add src/utils.go && git commit -m "Work in feature-b"

cd ../dev
echo "// dev work" >> README.md
git add README.md && git commit -m "Work in dev"

# Verify all commits independent
cd ..
tg list > ../results/test-7.2-tg-list.txt
cd .bare
git log --all --oneline --graph > ../results/test-7.2-tg-all-logs.txt
```

**Expected**:
- All worktrees should work independently
- Commits should not interfere with each other
- Each branch should have correct commits

---

## Test Execution Framework

### Master Test Runner

Create `run-all-tests.sh`:

```bash
#!/bin/bash
# Master test runner for acceptance tests

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Source helper functions
source setup/test-helpers.sh

# Initialize
echo "========================================="
echo "Timber-Git Acceptance Test Suite"
echo "========================================="
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
```

### Helper Functions

Create `setup/test-helpers.sh`:

```bash
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
```

## Success Criteria

### Overall Acceptance Criteria

1. **Clone Functionality**: Successfully clones repositories creating bare + main worktree structure
2. **Worktree Management**: Can create, list, and remove worktrees correctly
3. **Git Compatibility**: All standard git operations work in timber-git worktrees
4. **Error Handling**: All error cases handled gracefully with clear messages
5. **Data Integrity**: No data loss or corruption across operations

### Test Pass Criteria

- **Suite 1**: All clone tests pass - repositories cloned correctly
- **Suite 2**: All checkout tests pass - worktrees created properly
- **Suite 3**: List command shows accurate worktree information
- **Suite 4**: Remove command cleans up worktrees correctly
- **Suite 5**: Status commands provide useful information
- **Suite 6**: Git operations work identically to standard git
- **Suite 7**: Edge cases handled gracefully

### Documentation Requirements

Each test failure should generate:
- Clear description of what failed
- Expected vs actual results
- Logs and error messages
- Steps to reproduce

## Maintenance and Extension

### Adding New Tests

1. Create test script in appropriate `tests/suite-N-` directory
2. Follow naming convention: `suite-N-test-description.sh`
3. Use helper functions from `test-helpers.sh`
4. Update this document with test description

### CI/CD Integration

This test suite can be integrated into CI/CD:

```yaml
# Example GitHub Actions workflow
acceptance-tests:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - name: Install timber-git
      run: make install
    - name: Run acceptance tests
      run: cd acceptance-tests && bash run-all-tests.sh
    - name: Upload results
      uses: actions/upload-artifact@v2
      with:
        name: acceptance-test-results
        path: acceptance-tests/comparison/
```
