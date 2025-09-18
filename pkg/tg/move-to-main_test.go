package tg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindBareRepoPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a .git file with gitdir content
	gitFile := filepath.Join(tempDir, ".git")
	gitContent := "gitdir: /some/path/.bare/worktrees/feature-branch"
	err := os.WriteFile(gitFile, []byte(gitContent), 0644)
	require.NoError(t, err)
	
	// Test finding bare repo path
	bareRepoPath, err := findBareRepoPath(tempDir)
	require.NoError(t, err)
	
	// Should resolve to the parent directory of worktrees
	expected := filepath.Join("/some/path/.bare")
	assert.Equal(t, expected, bareRepoPath)
}

func TestGetCurrentBranchName(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create .bare/worktrees/feature-branch directory
	worktreeDir := filepath.Join(tempDir, ".bare", "worktrees", "feature-branch")
	err := os.MkdirAll(worktreeDir, 0755)
	require.NoError(t, err)
	
	// Create HEAD file in worktree metadata
	headFile := filepath.Join(worktreeDir, "HEAD")
	headContent := "ref: refs/heads/feature-branch\n"
	err = os.WriteFile(headFile, []byte(headContent), 0644)
	require.NoError(t, err)
	
	// Create .git file pointing to worktree metadata
	gitFile := filepath.Join(tempDir, ".git")
	gitContent := "gitdir: " + worktreeDir
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	require.NoError(t, err)
	
	// Test getting current branch name
	branchName, err := getCurrentBranchName(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "feature-branch", branchName)
}

func TestGenerateWorktreeDiff(t *testing.T) {
	// Create an in-memory repository for testing
	repo, err := git.Init(memory.NewStorage(), nil)
	require.NoError(t, err)

	// Create a blob for a simple file
	blob1Content := "Hello World\n"
	blob1 := &object.Blob{}
	blob1.Size = int64(len(blob1Content))
	
	// Create encoded object for blob1
	blob1Obj := repo.Storer.NewEncodedObject()
	blob1Obj.SetType(plumbing.BlobObject)
	blob1Obj.SetSize(int64(len(blob1Content)))
	writer1, err := blob1Obj.Writer()
	require.NoError(t, err)
	_, err = writer1.Write([]byte(blob1Content))
	require.NoError(t, err)
	err = writer1.Close()
	require.NoError(t, err)
	
	blob1Hash, err := repo.Storer.SetEncodedObject(blob1Obj)
	require.NoError(t, err)

	// Create a second blob with different content
	blob2Content := "Hello World Modified\n"
	blob2Obj := repo.Storer.NewEncodedObject()
	blob2Obj.SetType(plumbing.BlobObject)
	blob2Obj.SetSize(int64(len(blob2Content)))
	writer2, err := blob2Obj.Writer()
	require.NoError(t, err)
	_, err = writer2.Write([]byte(blob2Content))
	require.NoError(t, err)
	err = writer2.Close()
	require.NoError(t, err)
	
	blob2Hash, err := repo.Storer.SetEncodedObject(blob2Obj)
	require.NoError(t, err)

	// Create tree1 with first blob
	tree1 := &object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: "file.txt",
				Mode: filemode.Regular,
				Hash: blob1Hash,
			},
		},
	}
	
	tree1Obj := repo.Storer.NewEncodedObject()
	tree1Obj.SetType(plumbing.TreeObject)
	err = tree1.Encode(tree1Obj)
	require.NoError(t, err)
	tree1Hash, err := repo.Storer.SetEncodedObject(tree1Obj)
	require.NoError(t, err)

	// Create tree2 with second blob 
	tree2 := &object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: "file.txt",
				Mode: filemode.Regular,
				Hash: blob2Hash,
			},
		},
	}
	
	tree2Obj := repo.Storer.NewEncodedObject()
	tree2Obj.SetType(plumbing.TreeObject)
	err = tree2.Encode(tree2Obj)
	require.NoError(t, err)
	tree2Hash, err := repo.Storer.SetEncodedObject(tree2Obj)
	require.NoError(t, err)

	// Create commits
	commit1 := &object.Commit{
		Author: object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Message:  "Initial commit",
		TreeHash: tree1Hash,
	}
	
	commit1Obj := repo.Storer.NewEncodedObject()
	commit1Obj.SetType(plumbing.CommitObject)
	err = commit1.Encode(commit1Obj)
	require.NoError(t, err)
	commit1Hash, err := repo.Storer.SetEncodedObject(commit1Obj)
	require.NoError(t, err)

	commit2 := &object.Commit{
		Author: object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Message:     "Modified file",
		TreeHash:    tree2Hash,
		ParentHashes: []plumbing.Hash{commit1Hash},
	}
	
	commit2Obj := repo.Storer.NewEncodedObject()
	commit2Obj.SetType(plumbing.CommitObject)
	err = commit2.Encode(commit2Obj)
	require.NoError(t, err)
	commit2Hash, err := repo.Storer.SetEncodedObject(commit2Obj)
	require.NoError(t, err)

	// Create branch references
	mainRef := plumbing.NewBranchReferenceName("main")
	featureRef := plumbing.NewBranchReferenceName("feature")
	
	err = repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commit1Hash))
	require.NoError(t, err)
	err = repo.Storer.SetReference(plumbing.NewHashReference(featureRef, commit2Hash))
	require.NoError(t, err)

	// Test diff generation
	diff, err := generateWorktreeDiff(repo, "feature", "main")
	require.NoError(t, err)
	
	// Should contain diff content showing the change
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "file.txt")
}

func TestMoveToDefaultBranch_AlreadyOnDefaultBranch(t *testing.T) {
	// This test verifies the logic exists in the code
	// Since setting up a full git repository is complex, we'll test the validation logic
	
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create .bare/worktrees/main directory
	worktreeDir := filepath.Join(tempDir, ".bare", "worktrees", "main")
	err := os.MkdirAll(worktreeDir, 0755)
	require.NoError(t, err)
	
	// Create HEAD file pointing to main branch
	headFile := filepath.Join(worktreeDir, "HEAD")
	headContent := "ref: refs/heads/main\n"
	err = os.WriteFile(headFile, []byte(headContent), 0644)
	require.NoError(t, err)
	
	// Create .git file pointing to worktree metadata
	gitFile := filepath.Join(tempDir, ".git")
	gitContent := "gitdir: " + worktreeDir
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	require.NoError(t, err)
	
	// Initialize a bare git repository in .bare
	bareRepoDir := filepath.Join(tempDir, ".bare")
	repo, err := git.PlainInit(bareRepoDir, true)
	require.NoError(t, err)
	
	// Create a main branch reference
	mainRef := plumbing.NewBranchReferenceName("main")
	// Create a dummy hash for the reference
	dummyHash := plumbing.NewHash("1234567890123456789012345678901234567890")
	err = repo.Storer.SetReference(plumbing.NewHashReference(mainRef, dummyHash))
	require.NoError(t, err)
	
	// Create HEAD reference pointing to main
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, mainRef)
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)
	
	// Change to the temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Test that the command fails when already on default branch
	err = MoveToDefaultBranch()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already on default branch")
}

func TestPullDefaultBranch(t *testing.T) {
	// Create an in-memory repository for testing
	repo, err := git.Init(memory.NewStorage(), nil)
	require.NoError(t, err)
	
	// Create a default branch
	defaultBranch := "main"
	
	// Since we can't actually fetch in an in-memory repo, 
	// this test just ensures the function doesn't crash
	// In a real scenario, this would test network operations
	err = pullDefaultBranch(repo, defaultBranch)
	// We expect this to fail since there's no remote, but it shouldn't panic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remote not found")
}