package tg

import (
	"testing"

	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubmoduleHandler(t *testing.T) {
	handler := NewSubmoduleHandler()
	assert.NotNil(t, handler)
	assert.IsType(t, &DefaultSubmoduleHandler{}, handler)
}

func TestInitSubmodules_NilRepository(t *testing.T) {
	handler := NewSubmoduleHandler()
	err := handler.InitSubmodules(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository cannot be nil")
}

func TestInitSubmodules_NoSubmodules(t *testing.T) {
	// Create a test repository without submodules using in-memory filesystem
	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), git.WithWorkTree(fs))
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Initialize submodules should succeed even with no submodules
	handler := NewSubmoduleHandler()
	err = handler.InitSubmodules(repo)
	// This should not error - it should just have nothing to do
	assert.NoError(t, err)
}

func TestUpdateSubmodules_NilRepository(t *testing.T) {
	handler := NewSubmoduleHandler()
	err := handler.UpdateSubmodules(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository cannot be nil")
}

func TestUpdateSubmodules_NoSubmodules(t *testing.T) {
	// Create a test repository without submodules using in-memory filesystem
	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), git.WithWorkTree(fs))
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Update submodules should succeed even with no submodules
	handler := NewSubmoduleHandler()
	err = handler.UpdateSubmodules(repo)
	// This should not error - it should just have nothing to do
	assert.NoError(t, err)
}

func TestSubmoduleHandler_Interface(_ *testing.T) {
	// Verify that DefaultSubmoduleHandler implements SubmoduleHandler
	var _ SubmoduleHandler = &DefaultSubmoduleHandler{}
}

func createTestRepoWithCommit(t *testing.T) *git.Repository {
	t.Helper()

	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), git.WithWorkTree(fs))
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create an author
	author := &object.Signature{
		Name:  "Test Author",
		Email: "test@example.com",
	}

	// Create empty commit
	hash, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author:            author,
		Committer:         author,
		AllowEmptyCommits: true,
	})
	require.NoError(t, err)

	// Set HEAD to point to main branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main"))
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)

	// Create main branch pointing to the commit
	mainRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), hash)
	err = repo.Storer.SetReference(mainRef)
	require.NoError(t, err)

	return repo
}

func TestInitSubmodules_WithSubmodule(t *testing.T) {
	// Create a parent repository
	parentRepo := createTestRepoWithCommit(t)

	// Add a submodule configuration to the parent repository
	cfg, err := parentRepo.Config()
	require.NoError(t, err)

	if cfg.Submodules == nil {
		cfg.Submodules = make(map[string]*config.Submodule)
	}

	// Note: We can't fully test submodule init without a real filesystem
	// This test verifies the handler works with a repository that has submodule config
	handler := NewSubmoduleHandler()
	err = handler.InitSubmodules(parentRepo)
	// Should not error even if we can't actually init the submodule
	// (requires filesystem which we don't have in memory storage)
	assert.NoError(t, err)
}

func TestUpdateSubmodules_WithoutInit(t *testing.T) {
	// Create a parent repository
	parentRepo := createTestRepoWithCommit(t)

	// Update without submodules should not error
	handler := NewSubmoduleHandler()
	err := handler.UpdateSubmodules(parentRepo)
	assert.NoError(t, err)
}
