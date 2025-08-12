package tg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
)

var (
	bareRepoPath = ".bare"
)

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
	fmt.Println("Clone completed, processing repository...")

	// Change to the project directory for the remaining operations
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("error changing directory to '%s': %w", projectDir, err)
	}
	fmt.Println("Changed to project directory")


	// Get the default branch (HEAD) of the bare repository
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD of the bare repository: %w", err)
	}
	defaultBranchName := headRef.Name().Short() // Get the short name (e.g., "main", "master")
	fmt.Printf("Default branch: %s\n", defaultBranchName)

	// Set up remote config using go-git
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("error getting repository config: %w", err)
	}
	fmt.Println("Got repository config")

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
	fmt.Println("Updated remote config")

	// Save the updated config
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("error setting remote origin fetch config: %w", err)
	}
	fmt.Println("Saved config")

	// Fetch all refs to ensure we have the objects
	fmt.Println("Starting fetch operation...")
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("error fetching from remote: %w", err)
	}
	fmt.Println("Fetch completed")


	// Create worktree for the default branch directly in the current directory
	fmt.Printf("Creating worktree for branch: %s\n", defaultBranchName)
	err = AddWorktree(defaultBranchName)
	if err != nil {
		return fmt.Errorf("error adding worktree '%s': %w", defaultBranchName, err)
	}
	fmt.Println("Worktree created successfully")

	fmt.Printf("Successfully initialized Git repository in '%s' with default branch '%s'\n", name, defaultBranchName)
	return nil
}
