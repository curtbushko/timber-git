package tg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

// ListWorktrees lists all worktrees in the repository
// This is equivalent to: git worktree list
func ListWorktrees() error {
	// Check if we're in a timber-git repository by looking for .bare directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	gitPath := filepath.Join(cwd, bareRepoPath)
	if _, statErr := os.Stat(gitPath); os.IsNotExist(statErr) {
		return errors.New("not a timber-git repository (or any of the parent directories)")
	}

	// Open the bare repository
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	// Get HEAD to determine which branch is the default
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD: %w", err)
	}

	// Read worktrees directory to find all worktrees
	worktreesDir := filepath.Join(gitPath, "worktrees")
	var worktrees []worktreeInfo

	// Check if worktrees directory exists
	if _, err := os.Stat(worktreesDir); err == nil { //nolint:nestif // nested logic needed for worktree enumeration
		entries, err := os.ReadDir(worktreesDir)
		if err != nil {
			return fmt.Errorf("error reading worktrees directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Read the gitdir file to get the worktree path
			gitdirPath := filepath.Join(worktreesDir, entry.Name(), "gitdir")
			content, err := os.ReadFile(gitdirPath)
			if err != nil {
				continue
			}

			// The gitdir file contains path to the worktree's .git file
			gitFile := string(content)
			// Remove trailing newline and the .git part
			worktreePath := filepath.Dir(gitFile[:len(gitFile)-1])

			// Read HEAD file to get the branch
			headPath := filepath.Join(worktreesDir, entry.Name(), "HEAD")
			headContent, err := os.ReadFile(headPath)
			if err != nil {
				continue
			}

			branch := parseHeadRef(string(headContent))

			worktrees = append(worktrees, worktreeInfo{
				Path:   worktreePath,
				Branch: branch,
				Head:   string(headContent),
			})
		}
	}

	// Print worktrees in format similar to git worktree list
	for _, wt := range worktrees {
		hash := getShortHash(repo, wt.Branch)
		fmt.Printf("%s %s [%s]\n", wt.Path, hash, wt.Branch)
	}

	// If no worktrees found, the bare repo itself might be the only one
	if len(worktrees) == 0 {
		hash := head.Hash().String()[:7]
		branch := head.Name().Short()
		fmt.Printf("%s %s [%s]\n", gitPath, hash, branch)
	}

	return nil
}

type worktreeInfo struct {
	Path   string
	Branch string
	Head   string
}

// parseHeadRef extracts branch name from HEAD ref
func parseHeadRef(headContent string) string {
	// HEAD content is usually "ref: refs/heads/branch-name\n"
	if len(headContent) > 16 && headContent[:16] == "ref: refs/heads/" {
		branch := headContent[16:]
		// Remove trailing newline
		if len(branch) > 0 && branch[len(branch)-1] == '\n' {
			branch = branch[:len(branch)-1]
		}
		return branch
	}
	return "detached"
}

// getShortHash gets the short commit hash for a branch
func getShortHash(repo *git.Repository, branch string) string {
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return "0000000"
	}
	return ref.Hash().String()[:7]
}
