package tg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/config"
	"github.com/go-git/go-git/plumbing"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestCloneRepo(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-clone-repo")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Use an in-memory file system for the bare repo
	fs := memfs.New()

	// Create a bare repository in memory
	repo, err := git.Init(memory.NewStorage(), fs)
	assert.NoError(t, err)

	// Create a file and commit it to the initial branch (e.g., "main")
	wt, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = wt.Create("test.txt")
	assert.NoError(t, err)
	_, err = wt.Add("test.txt")
	assert.NoError(t, err)
	_, err = wt.Commit("Initial commit", &git.CommitOptions{Author: &git.Signature{Name: "Test User", Email: "test@example.com"}})
	assert.NoError(t, err)

	// Get the HEAD ref
	headRef, err := repo.Head()
	assert.NoError(t, err)
	defaultBranchName := headRef.Short()

	// Create a new branch "feature/test" and switch to it.
	err = repo.CreateBranch(&config.Branch{Name: "feature/test", Remote: "origin", Merge: "refs/heads/feature/test"})
	assert.NoError(t, err)

	// Switch to the new branch
	err = wt.Checkout(&git.CheckoutOptions{Branch: plumbing.ReferenceName("refs/heads/feature/test")})
	assert.NoError(t, err)

	// Add a file and commit to the new branch
	_, err = wt.Create("feature.txt")
	assert.NoError(t, err)
	_, err = wt.Add("feature.txt")
	assert.NoError(t, err)
	_, err = wt.Commit("Add feature file", &git.CommitOptions{Author: &git.Signature{Name: "Test User", Email: "test@example.com"}})
	assert.NoError(t, err)

	// Get the URL for the in-memory repo (we'll use a special format for memory)
	repoURL := "memory://" + headRef.String() // Use the ref string as a unique identifier

	// Call the function to test (assuming it's in the same package)
	dirName := "cloned_repo"
	err = os.Chdir(tempDir) // Change to the temp dir so the new dir is created there.
	assert.NoError(t, err)
	// Create the directory *before* calling the logic, as the logic creates and chdirs.
	err = os.Mkdir(dirName, 0755)
	assert.NoError(t, err)

	// Capture standard output, so we can assert against it.
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cloneRepo(repoURL, dirName) // Call the function directly
	w.Close()
	os.Stdout = originalStdout

	assert.NoError(t, err)

	// Assert that the directory was created
	clonedRepoPath := filepath.Join(tempDir, dirName)
	assert.DirExists(t, clonedRepoPath)

	// Assert that the .git file was created and has the correct content
	gitFilePath := filepath.Join(clonedRepoPath, ".git")
	gitFileContent, err := os.ReadFile(gitFilePath)
	assert.NoError(t, err)
	expectedGitDir := fmt.Sprintf("gitdir: ./.bare\n")
	assert.Equal(t, expectedGitDir, string(gitFileContent))

	// Assert that the worktree was created with the correct branch
	clonedRepo, err := git.PlainOpen(clonedRepoPath, nil)
	assert.NoError(t, err)
	wt, err = clonedRepo.Worktree()
	assert.NoError(t, err)
	actualBranch, err := wt.Branch()
	assert.NoError(t, err)
	assert.Equal(t, defaultBranchName, actualBranch.Short()) // Check the branch name

	// Check for the files from both branches.
	_, err = os.Stat(filepath.Join(clonedRepoPath, "test.txt"))
	assert.NoError(t, err, "test.txt should exist")
	_, err = os.Stat(filepath.Join(clonedRepoPath, "feature.txt"))
	assert.NoError(t, err, "feature.txt should exist")
}

func TestConvert(t *testing.T) {
	cases := []struct {
		name   string
		actual string
		want   string
	}{
		{
			name:   "Standard git repo",
			actual: "",
			want:   "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			repo, err := createInMemoryRepo()
			assert.NoError(t, err)

			repo, err := loadRepo(tmpDir)
			assert.NoError(t, err)

			status, err := isWorktree(repo)
			assert.NoError(t, err)

			got, err := convertRepo(repo)
			assert.NoError(t, err)

		})
	}
}

func createInMemoryRepo() *git.Repository {
	// Filesystem abstraction based on memory
	fs := memfs.New()
	// Git objects storer based on memory
	storer := memory.NewStorage()

	// Clones the repository into the worktree (fs) and stores all the .git
	// content into the storer
	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: "https://github.com/git-fixtures/basic.git",
	})
	if err != nil {
		return nil, ErrCreateNewRepo
	}

	return repo, nil
}
