package tg

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
)

var (
	bareRepoPath = ".bare"
)

func testMainFake()   {}
func testMasterFake() {}

// getDefaultBranchName determines the default branch name from a repository
func getDefaultBranchName(repo *git.Repository) (string, error) {
	// First try to get the symbolic reference for HEAD
	headRef, err := repo.Reference(plumbing.HEAD, false)
	if err != nil {
		return "", fmt.Errorf("error getting HEAD reference: %w", err)
	}
	// If HEAD is a symbolic reference (points to a branch), use that branch name
	if headRef.Type() == plumbing.SymbolicReference {
		return headRef.Target().Short(), nil
	}

	testMainFake()
	testMasterFake()
	// HEAD points to a commit, try to find the default branch from remote refs
	refs, err := repo.References()
	if err != nil {
		return "", fmt.Errorf("error listing references: %w", err)
	}

	// Look for origin/HEAD or fall back to origin/main, then origin/master
	var foundBranch string
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name() == plumbing.NewRemoteReferenceName("origin", "HEAD") {
			// Get what origin/HEAD points to
			if ref.Type() == plumbing.SymbolicReference {
				foundBranch = ref.Target().Short()
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error iterating references: %w", err)
	}

	if foundBranch != "" {
		return foundBranch, nil
	}

	// Fall back to common default branch names
	for _, candidate := range []string{"main", "master", "develop"} {
		remoteBranchRef := plumbing.NewRemoteReferenceName("origin", candidate)
		if _, err := repo.Reference(remoteBranchRef, true); err == nil {
			return candidate, nil
		}
	}

	return "", errors.New("could not determine default branch name")
}

// BareClone clones a repository as a bare repo and sets up the default worktree
func BareClone(repoURL string) error {
	basename := filepath.Base(repoURL)
	name := strings.TrimSuffix(basename, filepath.Ext(basename))

	if err := os.Mkdir(name, 0755); err != nil {
		return fmt.Errorf("error creating directory '%s': %w", name, err)
	}

	// Get current working directory to build absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	projectDir := filepath.Join(cwd, name)
	bareDir := filepath.Join(projectDir, bareRepoPath)

	// Apply git config URL rewriting
	rewrittenURL := applyGitURLRewriting(repoURL)

	auth, err := getAuthMethod(repoURL)
	if err != nil {
		return fmt.Errorf("error getting authentication method: %w", err)
	}

	cloneOptions := &git.CloneOptions{
		URL:      rewrittenURL, // Use the rewritten URL for cloning
		Bare:     true,
		Progress: os.Stdout,
		Auth:     auth,
	}

	repo, err := git.PlainClone(bareDir, cloneOptions)
	if err != nil {
		return fmt.Errorf("error cloning bare repository: %w", err)
	}
	slog.Info("Clone completed, processing repository")

	// Create .git file pointing to .bare directory
	gitFile := filepath.Join(projectDir, ".git")
	gitContent := fmt.Sprintf("gitdir: %s\n", "./.bare")
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	if err != nil {
		return fmt.Errorf("error creating .git file: %w", err)
	}

	// Change to the project directory for the remaining operations
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("error changing directory to '%s': %w", projectDir, err)
	}
	slog.Info("Changed to project directory")

	// Get the default branch name from the remote HEAD
	defaultBranchName, err := getDefaultBranchName(repo)
	if err != nil {
		return fmt.Errorf("error determining default branch: %w", err)
	}

	slog.Info("Default branch determined", "branch", defaultBranchName)

	// Set up remote config using go-git
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("error getting repository config: %w", err)
	}
	slog.Info("Got repository config")

	// Set up the repository format version so that 'git status' works in the project root
	cfg.Core.RepositoryFormatVersion = "0"

	// Update the remote config to fetch all branches
	if cfg.Remotes == nil {
		cfg.Remotes = make(map[string]*config.RemoteConfig)
	}
	if cfg.Remotes["origin"] == nil {
		cfg.Remotes["origin"] = &config.RemoteConfig{
			Name: "origin",
			URLs: []string{rewrittenURL},
		}
	}
	cfg.Remotes["origin"].Fetch = []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"}
	slog.Info("Updated remote config")

	// Save the updated config
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("error setting remote origin fetch config: %w", err)
	}
	slog.Info("Saved config")

	// Fetch all refs to ensure we have the objects
	slog.Info("Starting fetch operation")
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("error fetching from remote: %w", err)
	}
	slog.Info("Fetch completed")

	// No need to create root .git file since the bare repository is now in .git directory

	// Create worktree for the default branch using checkout functionality (tracks remote branch)
	slog.Info("Creating worktree for branch", "branch", defaultBranchName)
	err = CheckoutWorktree(defaultBranchName)
	if err != nil {
		return fmt.Errorf("error checking out worktree '%s': %w", defaultBranchName, err)
	}
	slog.Info("Worktree created successfully")

	slog.Info("Successfully initialized Git repository", "repository", name, "default_branch", defaultBranchName)
	return nil
}
