package tg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddWorktree(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-add-worktree")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to the temp dir
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Test adding a worktree in a directory that's not a git repo should fail
	branchName := "feature-test"
	err = AddWorktree(branchName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestAddWorktree_NotInGitRepo(t *testing.T) {
	// Create a temporary directory that's not a git repo
	tempDir, err := os.MkdirTemp("", "test-not-git-repo")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to the temp dir
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Test adding a worktree should fail
	err = AddWorktree("test-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

// TestAddWorktreeErrorHandling tests that AddWorktree provides proper error messages
// instead of the previous "object not found" error when there are issues
func TestAddWorktreeErrorHandling(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "test-add-worktree-errors")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Create an empty .git directory to simulate a partial git setup
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	assert.NoError(t, err)

	// Test adding a worktree - this should fail gracefully
	// instead of with an "object not found" error
	branchName := "test-branch"
	err = AddWorktree(branchName)
	
	// The important thing is that we get a meaningful error, not "object not found"
	assert.Error(t, err, "AddWorktree should fail with empty git directory")
	assert.NotContains(t, err.Error(), "object not found", "Should not get 'object not found' error")
}


