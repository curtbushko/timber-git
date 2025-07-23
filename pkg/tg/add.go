package tg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v6"
)

// AddWorktree creates a new worktree with a new branch
// This is equivalent to: git worktree add <branch> -b <branch>
func AddWorktree(branch string) error {
	// Check if we're in a git repository by looking for .git file or directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	gitPath := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// Verify this is a valid git repository using go-git
	_, err = git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	// Use git command for worktree operations since go-git v6 doesn't have WorktreeAdd
	cmd := exec.Command("git", "worktree", "add", branch, "-b", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error adding worktree '%s': %v", branch, err)
	}

	fmt.Printf("Successfully created worktree '%s' with new branch '%s'\n", branch, branch)
	return nil
}