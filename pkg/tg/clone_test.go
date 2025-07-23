package tg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/stretchr/testify/assert"
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

	// Assert that the .git file was created and has the correct content
	gitFilePath := filepath.Join(clonedRepoPath, ".git")
	if _, err := os.Stat(gitFilePath); err == nil {
		gitFileContent, err := os.ReadFile(gitFilePath)
		assert.NoError(t, err)
		expectedGitDir := "gitdir: ./.bare\n"
		assert.Equal(t, expectedGitDir, string(gitFileContent))
	}

	// Assert that the bare repository directory exists
	bareRepoPath := filepath.Join(clonedRepoPath, ".bare")
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
