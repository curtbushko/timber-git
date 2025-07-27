package tg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
)

var (
	bareRepoPath = ".git"
)

func BareClone(repoURL string) error {

	basename := filepath.Base(repoURL)
	name := strings.TrimSuffix(basename, filepath.Ext(basename))

	if err := os.Mkdir(name, 0755); err != nil {
		return fmt.Errorf("error creating directory '%s': %v", name, err)
	}

	// Get current working directory to build absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	projectDir := filepath.Join(cwd, name)
	bareDir := filepath.Join(projectDir, bareRepoPath)

	// Apply git config URL rewriting
	rewrittenURL, err := applyGitURLRewriting(repoURL)
	if err != nil {
		return fmt.Errorf("error applying git URL rewriting: %v", err)
	}

	auth, err := getAuthMethod(repoURL)
	if err != nil {
		return fmt.Errorf("error getting authentication method: %v", err)
	}

	cloneOptions := &git.CloneOptions{
		URL:      rewrittenURL, // Use the rewritten URL for cloning
		Bare:     true,
		Progress: os.Stdout,
		Auth:     auth,
	}

	repo, err := git.PlainClone(bareDir, cloneOptions)
	if err != nil {
		return fmt.Errorf("error cloning bare repository: %v", err)
	}

	// Change to the project directory for the remaining operations
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("error changing directory to '%s': %v", projectDir, err)
	}

	// Get the default branch (HEAD) of the bare repository
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD of the bare repository: %v", err)
	}
	defaultBranchName := headRef.Name().Short() // Get the short name (e.g., "main", "master")

	// Use git command to set the fetch config since the go-git v6 config API has changed
	configCmd := exec.Command("git", "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	configCmd.Dir = "."
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("error setting remote origin fetch config: %v", err)
	}

	// Use git command for worktree operations since go-git v6 doesn't have WorktreeAdd
	cmd := exec.Command("git", "worktree", "add", defaultBranchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error adding worktree '%s': %v", defaultBranchName, err)
	}

	fmt.Printf("Successfully initialized Git repository in '%s' with default branch '%s'\n", name, defaultBranchName)
	return nil
}
