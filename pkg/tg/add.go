// Package tg provides core timber-git functionality
package tg

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/format/index"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// AddMainWorktree creates the main worktree in the current directory
// This is used during clone to set up the primary working directory
func AddMainWorktree(branch string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	gitPath := filepath.Join(cwd, bareRepoPath)
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return errors.New("not a git repository (or any of the parent directories)")
	}

	// Open the bare repository
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	// Get the reference for the branch
	branchRef := plumbing.NewBranchReferenceName(branch)
	var baseRef *plumbing.Reference
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", branch)
	baseRef, err = repo.Reference(remoteBranchRef, true)
	if err != nil {
		// If remote branch doesn't exist, try HEAD
		baseRef, err = repo.Head()
		if err != nil {
			return fmt.Errorf("error getting base reference: %w", err)
		}
	}

	// Create the branch reference in the main repository
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, baseRef.Hash()))
	if err != nil {
		return fmt.Errorf("error creating branch reference: %w", err)
	}

	// Checkout files to the current directory
	err = checkoutFilesToWorktreeFromRepoWithBranch(repo, baseRef.Hash(), cwd, branch)
	if err != nil {
		return fmt.Errorf("error checking out files: %w", err)
	}

	return nil
}

// AddWorktree creates a new worktree with a new branch
// This is equivalent to: git worktree add <branch> -b <branch>
// If the worktree path already has files (like during clone), it will set up git in the current directory
func AddWorktree(branch string) error {
	// Check if we're in a git repository by looking for .git file or directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	gitPath := filepath.Join(cwd, bareRepoPath)
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return errors.New("not a git repository (or any of the parent directories)")
	}

	// Verify this is a valid git repository using go-git
	_, err = git.PlainOpen(gitPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	// Implement worktree functionality using go-git
	// Open the main repository
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	// Check if branch already exists locally (not in bare repo)
	branchRef := plumbing.NewBranchReferenceName(branch)

	// Check if worktree directory already exists
	worktreePath := filepath.Join(cwd, branch)
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", branch)
	}

	// Create the worktree directory
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("error creating worktree directory: %w", err)
	}

	// Get the reference to base the new branch on (always use HEAD for new local branches)
	baseRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD reference: %w", err)
	}

	// Create worktree manually using filesystem operations and git objects
	slog.Info("Creating worktree by manual checkout")

	// Create the branch reference in the main repository first
	slog.Info("Creating branch reference", "branch", branch)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, baseRef.Hash()))
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating branch reference: %w", err)
	}
	slog.Info("Created branch reference", "branch", branch)

	// Create proper worktree .git directory structure
	err = setupWorktreeGitDir(worktreePath, cwd, branch)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error setting up worktree git directory: %w", err)
	}

	// Use go-git to checkout files properly to the worktree from the main repo
	slog.Info("Checking out files to worktree using go-git", "branch", branch)
	err = checkoutFilesToWorktreeFromRepoWithBranch(repo, baseRef.Hash(), worktreePath, branch)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error checking out files: %w", err)
	}

	slog.Info("Successfully checked out branch", "branch", branch)
	slog.Info("Worktree created successfully")

	slog.Info("Successfully created worktree with new branch", "worktree", branch, "branch", branch)
	return nil
}

// setupWorktreeGitDir creates a proper .git file structure for a worktree
func setupWorktreeGitDir(worktreePath, mainRepoPath, branch string) error {
	// Sanitize branch name for worktree reference (replace slashes with dashes)
	sanitizedBranch := sanitizeBranchNameForWorktree(branch)

	// Create worktree entry in main repo
	worktreesDir := filepath.Join(mainRepoPath, bareRepoPath, "worktrees", sanitizedBranch)
	err := os.MkdirAll(worktreesDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Create gitdir file pointing to worktree
	gitdirFile := filepath.Join(worktreesDir, "gitdir")
	err = os.WriteFile(gitdirFile, []byte(filepath.Join(worktreePath, ".git")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to create gitdir file: %w", err)
	}

	// Create HEAD file in worktree metadata
	headFile := filepath.Join(worktreesDir, "HEAD")
	headContent := fmt.Sprintf("ref: refs/heads/%s\n", branch)
	err = os.WriteFile(headFile, []byte(headContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create HEAD file: %w", err)
	}

	// Create commondir file pointing to the main bare repository
	commondirFile := filepath.Join(worktreesDir, "commondir")
	commondirContent := "../..\n"
	err = os.WriteFile(commondirFile, []byte(commondirContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create commondir file: %w", err)
	}

	// Create .git file (not directory) pointing to the worktree metadata
	gitFile := filepath.Join(worktreePath, ".git")
	// Use absolute path to .bare/worktrees/<sanitized-branch>
	absoluteWorktreesDir := filepath.Join(mainRepoPath, bareRepoPath, "worktrees", sanitizedBranch)
	gitContent := fmt.Sprintf("gitdir: %s\n", absoluteWorktreesDir)
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .git file: %w", err)
	}

	return nil
}

// checkoutFilesToWorktreeFromRepoWithBranch checks out files using go-git directly from main repo
func checkoutFilesToWorktreeFromRepoWithBranch(repo *git.Repository, commitHash plumbing.Hash, worktreePath, branch string) error {
	// Get the commit object from the main repository
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("error getting commit object: %w", err)
	}

	// Get the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("error getting tree from commit: %w", err)
	}

	// Create index entries for proper git status
	indexEntries := make([]*index.Entry, 0)

	// Walk the tree and create files using go-git's file iteration
	err = tree.Files().ForEach(func(file *object.File) error {
		// Get file path relative to worktree
		filePath := filepath.Join(worktreePath, file.Name)

		// Create directory structure
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Get file contents
		reader, err := file.Reader()
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := reader.Close(); closeErr != nil {
				slog.Warn("Failed to close reader", "error", closeErr)
			}
		}()

		// Create the file
		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := outFile.Close(); closeErr != nil {
				slog.Warn("Failed to close file", "error", closeErr)
			}
		}()

		// Copy contents
		_, err = outFile.ReadFrom(reader)
		if err != nil {
			return err
		}

		// Set correct file permissions from git object
		// Convert git filemode to os.FileMode
		perm := os.FileMode(file.Mode) & 0777
		err = os.Chmod(filePath, perm)
		if err != nil {
			return err
		}

		// Get file info for index entry
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		// Create index entry
		entry := &index.Entry{
			Hash:       file.Hash,
			Name:       file.Name,
			Mode:       file.Mode,
			ModifiedAt: fileInfo.ModTime(),
			CreatedAt:  fileInfo.ModTime(),
			Dev:        0,
			Inode:      0,
			UID:        0,
			GID:        0,
			Size:       uint32(fileInfo.Size()),
		}
		indexEntries = append(indexEntries, entry)

		return nil
	})
	if err != nil {
		return err
	}

	// Create and write the index file
	idx := &index.Index{
		Version: 2,
		Entries: indexEntries,
	}

	// Determine the correct index path based on worktree structure
	var indexPath string
	if worktreePath == "." {
		// For the main worktree in current directory
		indexPath = filepath.Join(bareRepoPath, "index")
	} else {
		// For branch worktrees, index goes in the worktree metadata directory
		// Use the sanitized branch name for the worktree reference
		sanitizedBranchName := sanitizeBranchNameForWorktree(branch)
		indexPath = filepath.Join(bareRepoPath, "worktrees", sanitizedBranchName, "index")
	}

	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("error creating index file: %w", err)
	}
	defer func() {
		if closeErr := indexFile.Close(); closeErr != nil {
			slog.Warn("Failed to close index file", "error", closeErr)
		}
	}()

	encoder := index.NewEncoder(indexFile)
	err = encoder.Encode(idx)
	if err != nil {
		return fmt.Errorf("error encoding index: %w", err)
	}

	return nil
}

// sanitizeBranchNameForWorktree converts branch names with slashes to use dashes
// for worktree reference directories to avoid filesystem path conflicts
func sanitizeBranchNameForWorktree(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}
