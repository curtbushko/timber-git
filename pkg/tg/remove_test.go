package tg

import (
	"os"
	"os/exec"
	"testing"

	"github.com/go-git/go-git/v6"
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

	// Initialize a git repository
	_, err = git.PlainInit(".", false)
	assert.NoError(t, err)

	// Create an initial commit using git commands (more reliable for testing)
	// Create a file
	testFile := "test.txt"
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	assert.NoError(t, err)

	// Add and commit using git commands
	cmd := exec.Command("git", "add", testFile)
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Env = append(os.Environ(), 
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User", 
		"GIT_COMMITTER_EMAIL=test@example.com")
	err = cmd.Run()
	assert.NoError(t, err)

	// First, add a worktree to remove
	branchName := "feature-test"
	err = AddWorktree(branchName)
	assert.NoError(t, err)

	// Verify the worktree directory was created
	_, err = os.Stat(branchName)
	assert.NoError(t, err)

	// Test removing the worktree
	err = RemoveWorktree(branchName)
	assert.NoError(t, err)

	// Verify the worktree directory was removed
	_, err = os.Stat(branchName)
	assert.True(t, os.IsNotExist(err))

	// Verify the branch was deleted by checking git branches
	cmd = exec.Command("git", "branch", "-a")
	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.NotContains(t, string(output), branchName)
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

	// Initialize a git repository
	_, err = git.PlainInit(".", false)
	assert.NoError(t, err)

	// Create an initial commit
	testFile := "test.txt"
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	assert.NoError(t, err)

	cmd := exec.Command("git", "add", testFile)
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Env = append(os.Environ(), 
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User", 
		"GIT_COMMITTER_EMAIL=test@example.com")
	err = cmd.Run()
	assert.NoError(t, err)

	// Test removing a nonexistent worktree should fail
	err = RemoveWorktree("nonexistent-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}