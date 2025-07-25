package tg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/storage/filesystem"
)

// RemoveWorktree removes an existing worktree and its associated branch
// This is equivalent to: git worktree remove <worktree> && git branch -D <worktree>
func RemoveWorktree(worktree string) error {
	// Check if we're in a git repository by looking for .git file or directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	gitPath := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// Open the repository using go-git
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	// Get the storage to access worktrees
	storage := repo.Storer.(*filesystem.Storage)
	
	// Get the worktree path - check if it's an absolute path or relative
	var worktreePath string
	if filepath.IsAbs(worktree) {
		worktreePath = worktree
	} else {
		// Assume it's relative to the current directory
		worktreePath = filepath.Join(cwd, "..", worktree)
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			// Try as absolute path within current directory
			worktreePath = filepath.Join(cwd, worktree)
		}
	}

	// Check if the worktree directory exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree '%s' does not exist", worktree)
	}

	// Remove the worktree directory
	if err := os.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("error removing worktree directory '%s': %v", worktreePath, err)
	}

	// Delete the branch associated with the worktree
	branchRef := plumbing.NewBranchReferenceName(worktree)
	if err := repo.Storer.RemoveReference(branchRef); err != nil {
		// Don't fail if branch doesn't exist, just warn
		fmt.Printf("Warning: could not delete branch '%s': %v\n", worktree, err)
	}

	// Clean up worktree references from git internals
	// This involves removing entries from .git/worktrees/<name>/
	gitWorktreesPath := filepath.Join(storage.Filesystem().Root(), "worktrees")
	if _, err := os.Stat(gitWorktreesPath); err == nil {
		// Find and remove the worktree reference directory
		entries, err := os.ReadDir(gitWorktreesPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					// Check if this worktree entry points to our removed directory
					gitlinkPath := filepath.Join(gitWorktreesPath, entry.Name(), "gitdir")
					if content, err := os.ReadFile(gitlinkPath); err == nil {
						if string(content) == filepath.Join(worktreePath, ".git")+"\n" {
							// Remove this worktree reference
							os.RemoveAll(filepath.Join(gitWorktreesPath, entry.Name()))
							break
						}
					}
				}
			}
		}
	}

	fmt.Printf("Successfully removed worktree '%s' and deleted branch '%s'\n", worktree, worktree)
	return nil
}