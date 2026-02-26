package tg

import (
	"os"
	"path/filepath"
	"strings"
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

func TestGenerateCommittedDiff(t *testing.T) {
	// Create an in-memory repository for testing
	repo, err := git.Init(memory.NewStorage(), nil)
	require.NoError(t, err)

	// Create a blob for a simple file
	blob1Content := "Hello World\n"
	_ = &object.Blob{} // blob1 struct not directly used, we create encoded object below
	
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
	diff, err := generateCommittedDiff(repo, "feature", "main")
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

func TestGenerateWorkingDirDiff_FilteringLogic(t *testing.T) {
	tests := []struct {
		name     string
		status   git.Status
		expected bool // whether the file should be included
	}{
		{
			name: "untracked file should be excluded",
			status: git.Status{
				"untracked.txt": &git.FileStatus{
					Staging:  git.Untracked,
					Worktree: git.Untracked,
				},
			},
			expected: false,
		},
		{
			name: "modified file should be included",
			status: git.Status{
				"modified.txt": &git.FileStatus{
					Staging:  git.Unmodified,
					Worktree: git.Modified,
				},
			},
			expected: true,
		},
		{
			name: "staged file should be included",
			status: git.Status{
				"staged.txt": &git.FileStatus{
					Staging:  git.Added,
					Worktree: git.Unmodified,
				},
			},
			expected: true,
		},
		{
			name: "unmodified file should be excluded",
			status: git.Status{
				"clean.txt": &git.FileStatus{
					Staging:  git.Unmodified,
					Worktree: git.Unmodified,
				},
			},
			expected: false,
		},
		{
			name: "deleted file should be included",
			status: git.Status{
				"deleted.txt": &git.FileStatus{
					Staging:  git.Deleted,
					Worktree: git.Unmodified,
				},
			},
			expected: true,
		},
		{
			name: "mixed status - many untracked files should not pollute output",
			status: git.Status{
				"go.mod":     &git.FileStatus{Staging: git.Untracked, Worktree: git.Untracked},
				"go.sum":     &git.FileStatus{Staging: git.Untracked, Worktree: git.Untracked},
				".gitignore": &git.FileStatus{Staging: git.Untracked, Worktree: git.Untracked},
				"LICENSE":    &git.FileStatus{Staging: git.Untracked, Worktree: git.Untracked},
				"README.md":  &git.FileStatus{Staging: git.Unmodified, Worktree: git.Modified},
				"main.go":    &git.FileStatus{Staging: git.Added, Worktree: git.Unmodified},
			},
			expected: true, // Only README.md and main.go should be included
		},
	}

	for _, tt := range tests { //nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			diff := generateWorkingDirDiff(tt.status)
			
			if tt.name == "mixed status - many untracked files should not pollute output" {
				// Should only contain the 2 meaningful changes
				assert.Contains(t, diff, "README.md")
				assert.Contains(t, diff, "main.go")
				// Should NOT contain untracked files
				assert.NotContains(t, diff, "go.mod")
				assert.NotContains(t, diff, ".gitignore")
				assert.NotContains(t, diff, "LICENSE")
				return
			}
			
			if tt.expected {
				// Should contain at least one meaningful file change
				assert.NotEqual(t, "Working directory changes:\nNo changes to move.\n", diff)
			} else {
				// Should indicate no changes to move
				assert.Contains(t, diff, "No changes to move")
			}
		})
	}
}

func TestMoveToDefaultBranch_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, func())
		expectedErr string
	}{
		{
			name: "not in git worktree directory",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				// No .git file
				return tempDir, func() {}
			},
			expectedErr: "not in a git worktree directory",
		},
		{
			name: "invalid .git file format",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitFile := filepath.Join(tempDir, ".git")
				err := os.WriteFile(gitFile, []byte("invalid content"), 0644)
				require.NoError(t, err)
				return tempDir, func() {}
			},
			expectedErr: "invalid .git file format",
		},
		{
			name: "bare repository not found",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				
				// Create .git file pointing to non-existent directory
				gitFile := filepath.Join(tempDir, ".git")
				err := os.WriteFile(gitFile, []byte("gitdir: /nonexistent/path"), 0644)
				require.NoError(t, err)
				
				return tempDir, func() {}
			},
			expectedErr: "error opening repository",
		},
	}

	for _, tt := range tests { //nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			testDir, cleanup := tt.setupFunc(t)
			defer cleanup()
			
			// Save original directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalDir)
				require.NoError(t, err)
			}()
			
			// Change to test directory
			err = os.Chdir(testDir)
			require.NoError(t, err)
			
			// Test the function
			err = MoveToDefaultBranch()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestMoveToDefaultBranch_NoChangesToMove(t *testing.T) {
	// Create an in-memory repository for simpler testing
	repo, err := git.Init(memory.NewStorage(), nil)
	require.NoError(t, err)
	
	// Create a commit to checkout files to the worktree
	// Create a simple blob and tree
	blobContent := "test content\n"
	blobObj := repo.Storer.NewEncodedObject()
	blobObj.SetType(plumbing.BlobObject)
	blobObj.SetSize(int64(len(blobContent)))
	blobWriter, err := blobObj.Writer()
	require.NoError(t, err)
	_, err = blobWriter.Write([]byte(blobContent))
	require.NoError(t, err)
	err = blobWriter.Close()
	require.NoError(t, err)
	blobHash, err := repo.Storer.SetEncodedObject(blobObj)
	require.NoError(t, err)
	
	// Create tree with the blob
	tree := &object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: "test.txt",
				Mode: filemode.Regular,
				Hash: blobHash,
			},
		},
	}
	treeObj := repo.Storer.NewEncodedObject()
	treeObj.SetType(plumbing.TreeObject)
	err = tree.Encode(treeObj)
	require.NoError(t, err)
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	require.NoError(t, err)
	
	// Create commit
	commit := &object.Commit{
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
		TreeHash: treeHash,
	}
	commitObj := repo.Storer.NewEncodedObject()
	commitObj.SetType(plumbing.CommitObject)
	err = commit.Encode(commitObj)
	require.NoError(t, err)
	commitHash, err := repo.Storer.SetEncodedObject(commitObj)
	require.NoError(t, err)
	
	// Create both branches pointing to same commit (no diff)
	mainRef := plumbing.NewBranchReferenceName("main")
	featureRef := plumbing.NewBranchReferenceName("feature")
	err = repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commitHash))
	require.NoError(t, err)
	err = repo.Storer.SetReference(plumbing.NewHashReference(featureRef, commitHash))
	require.NoError(t, err)
	
	// Set HEAD to main (default branch)
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, mainRef)
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)
	
	// Test generateCommittedDiff directly since both branches have same content
	diff, err := generateCommittedDiff(repo, "feature", "main")
	require.NoError(t, err)
	
	// Should be empty diff since both branches point to same commit
	assert.Empty(t, strings.TrimSpace(diff))
}

// Test that simulates the bug where many files are incorrectly shown
func TestMoveToDefaultBranch_ManyUntrackedFiles(t *testing.T) {
	// This test creates a scenario similar to what the user experienced
	// where many untracked files are incorrectly included in the move operation
	tempDir := t.TempDir()
	
	// Create .bare directory and initialize repository
	bareRepoDir := filepath.Join(tempDir, ".bare")
	repo, err := git.PlainInit(bareRepoDir, true)
	require.NoError(t, err)
	
	// Create worktree structure
	worktreeDir := filepath.Join(tempDir, ".bare", "worktrees", "feature")
	err = os.MkdirAll(worktreeDir, 0755)
	require.NoError(t, err)
	
	// Create HEAD file
	headFile := filepath.Join(worktreeDir, "HEAD")
	err = os.WriteFile(headFile, []byte("ref: refs/heads/feature\n"), 0644)
	require.NoError(t, err)
	
	// Create .git file
	gitFile := filepath.Join(tempDir, ".git")
	err = os.WriteFile(gitFile, []byte("gitdir: "+worktreeDir), 0644)
	require.NoError(t, err)
	
	// Create many files that would typically be untracked in a new project
	untrackedFiles := []string{
		"go.mod", "go.sum", ".gitignore", "LICENSE", "README.md",
		"main.go", "cmd/root.go", "cmd/add.go", "pkg/tg/clone.go",
		"Makefile", "flake.nix", "flake.lock", ".envrc",
	}
	
	for _, fileName := range untrackedFiles {
		filePath := filepath.Join(tempDir, fileName)
		dir := filepath.Dir(filePath)
		if dir != tempDir {
			err = os.MkdirAll(dir, 0755)
			require.NoError(t, err)
		}
		err = os.WriteFile(filePath, []byte("content of "+fileName), 0644)
		require.NoError(t, err)
	}
	
	// Create branch references
	mainRef := plumbing.NewBranchReferenceName("main")
	featureRef := plumbing.NewBranchReferenceName("feature")
	dummyHash := plumbing.NewHash("1234567890123456789012345678901234567890")
	err = repo.Storer.SetReference(plumbing.NewHashReference(mainRef, dummyHash))
	require.NoError(t, err)
	err = repo.Storer.SetReference(plumbing.NewHashReference(featureRef, dummyHash))
	require.NoError(t, err)
	
	// Set HEAD to main (default branch)
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, mainRef)
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)
	
	// Test the problematic scenario by simulating what generateWorktreeDiff would see
	status := make(git.Status)
	for _, fileName := range untrackedFiles {
		status[fileName] = &git.FileStatus{
			Staging:  git.Untracked,
			Worktree: git.Untracked,
		}
	}
	
	// This should NOT include untracked files in the diff
	diff := generateWorkingDirDiff(status)
	assert.Contains(t, diff, "No changes to move")
	
	// None of the untracked files should appear in the diff
	for _, fileName := range untrackedFiles {
		assert.NotContains(t, diff, fileName, "Untracked file %s should not appear in diff", fileName)
	}
}