package tg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveWorktree(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-remove-worktree")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to the temp dir
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Test removing a worktree in a directory that's not a git repo should fail
	branchName := "feature-test"
	err = RemoveWorktree(branchName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestRemoveWorktree_NotInGitRepo(t *testing.T) {
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

	// Test removing a worktree should fail
	err = RemoveWorktree("test-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestRemoveWorktree_NonexistentWorktree(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-remove-nonexistent")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to the temp dir
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Test removing a nonexistent worktree should fail with "not a git repository"
	err = RemoveWorktree("nonexistent-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}
