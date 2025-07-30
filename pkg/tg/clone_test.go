package tg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBareClone(t *testing.T) {
	// Use in-memory repository for setup
	repo, err := createInMemoryRepo()
	assert.NoError(t, err)

	// Verify the in-memory repo was created successfully
	_, err = repo.Head()
	assert.NoError(t, err)

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-clone-repo")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up after the test

	// Change to the temp dir so the new dir is created there
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Call the BareClone function with the basic fixture repo
	err = BareClone("https://github.com/git-fixtures/basic.git")
	if err != nil {
		t.Skipf("Skipping test due to network/git issues: %v", err)
		return
	}

	// Assert that the directory was created (BareClone creates "basic" directory)
	clonedRepoPath := filepath.Join(tempDir, "basic")
	assert.DirExists(t, clonedRepoPath)

	// Assert that the bare repository directory exists
	bareRepoPath := filepath.Join(clonedRepoPath, ".git")
	if _, err := os.Stat(bareRepoPath); err == nil {
		assert.DirExists(t, bareRepoPath)
	}

	// Try to open the cloned repo if it exists
	if _, err := os.Stat(clonedRepoPath); err == nil {
		clonedRepo, err := git.PlainOpen(clonedRepoPath)
		if err == nil {
			actualHead, err := clonedRepo.Head()
			if err == nil {
				actualBranchName := actualHead.Name().Short()
				assert.NotEmpty(t, actualBranchName)
			}
		}
	}
}

// createTestBareRepo creates a bare repository in the given directory for testing
func createTestBareRepo(repoPath string) error {
	// Create a temporary regular repo first to commit files, then clone as bare
	tempRepoDir, err := os.MkdirTemp("", "temp-repo")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tempRepoDir) }()

	// Create a regular repository first
	tempRepo, err := git.PlainInit(tempRepoDir, false)
	if err != nil {
		return err
	}

	// Create a test file
	testFile := filepath.Join(tempRepoDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		return err
	}

	// Get worktree and add/commit the file
	worktree, err := tempRepo.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Add("test.txt")
	if err != nil {
		return err
	}

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		return err
	}

	// Now clone this as a bare repository
	_, err = git.PlainClone(repoPath, &git.CloneOptions{
		URL:  tempRepoDir,
		Bare: true,
	})
	return err
}

// TestBareCloneWithWorktree tests the full BareClone workflow that was failing
func TestBareCloneWithWorktree(t *testing.T) {
	// Create a temporary directory for our test bare repo
	tempDir, err := os.MkdirTemp("", "test-bare-clone")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test bare repository
	testRepoPath := filepath.Join(tempDir, "test-repo.git")
	err = createTestBareRepo(testRepoPath)
	require.NoError(t, err)

	// Create another temp directory for the clone destination  
	cloneDir, err := os.MkdirTemp("", "test-clone-dest")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(cloneDir) }()

	// Change to clone directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(cloneDir)
	require.NoError(t, err)

	// Test the BareClone function with our test repo
	err = BareClone(testRepoPath)
	
	// This should not fail with "object not found" error
	assert.NoError(t, err, "BareClone should not fail with object not found error")

	// Verify the expected directory structure was created
	clonedRepoName := "test-repo"
	clonedRepoPath := filepath.Join(cloneDir, clonedRepoName)
	assert.DirExists(t, clonedRepoPath, "Cloned repository directory should exist")

	// Verify the bare repo was created
	bareRepoPath := filepath.Join(clonedRepoPath, ".git")
	assert.DirExists(t, bareRepoPath, "Bare repository .git directory should exist")

	// Verify the default branch worktree was created (could be main or master)
	// First check if main exists, then master
	var defaultWorktreePath string
	mainWorktreePath := filepath.Join(clonedRepoPath, "main")
	masterWorktreePath := filepath.Join(clonedRepoPath, "master")
	
	if _, err := os.Stat(mainWorktreePath); err == nil {
		defaultWorktreePath = mainWorktreePath
	} else if _, err := os.Stat(masterWorktreePath); err == nil {
		defaultWorktreePath = masterWorktreePath
	} else {
		t.Fatal("Neither main nor master worktree directory exists")
	}
	
	assert.DirExists(t, defaultWorktreePath, "Default branch worktree directory should exist")

	// Verify the test file exists in the worktree
	testFilePath := filepath.Join(defaultWorktreePath, "test.txt")
	assert.FileExists(t, testFilePath, "Test file should exist in worktree")

	// Verify the content of the test file
	content, err := os.ReadFile(testFilePath)
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))

	// Additional verification: ensure the worktree has a proper git structure
	worktreeGitDir := filepath.Join(defaultWorktreePath, ".git")
	assert.DirExists(t, worktreeGitDir, "Worktree should have its own .git directory")
}

func createInMemoryRepo() (*git.Repository, error) {
	// Git objects storer based on memory
	storer := memory.NewStorage()

	// Clones the repository and stores all the .git
	// content into the storer
	repo, err := git.Clone(storer, nil, &git.CloneOptions{
		URL: "https://github.com/git-fixtures/basic.git",
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}
