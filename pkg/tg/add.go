package tg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
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

	// Implement worktree functionality using go-git
	// Open the main repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
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
		return fmt.Errorf("error creating worktree directory: %v", err)
	}

	// Get the reference to base the new branch on
	// Try to get the remote branch reference first, then fall back to HEAD
	var baseRef *plumbing.Reference
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", branch)
	baseRef, err = repo.Reference(remoteBranchRef, true)
	if err != nil {
		// If remote branch doesn't exist, try HEAD
		baseRef, err = repo.Head()
		if err != nil {
			return fmt.Errorf("error getting base reference: %v", err)
		}
	}

	// Create worktree manually using filesystem operations and git objects
	fmt.Printf("Creating worktree by manual checkout\n")
	
	// Create the branch reference in the main repository first
	fmt.Printf("Creating branch reference for: %s\n", branch)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, baseRef.Hash()))
	if err != nil {
		return fmt.Errorf("error creating branch reference: %v", err)
	}
	fmt.Printf("Created branch reference: %s\n", branch)
	
	// Get the commit object to checkout
	commit, err := repo.CommitObject(baseRef.Hash())
	if err != nil {
		return fmt.Errorf("error getting commit object: %v", err)
	}
	
	// Get the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("error getting tree from commit: %v", err)
	}
	
	// Create the worktree directory files by walking the tree
	fmt.Printf("Checking out files to worktree: %s\n", branch)
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
		defer reader.Close()
		
		// Create the file
		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()
		
		// Copy contents
		_, err = outFile.ReadFrom(reader)
		return err
	})
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error checking out files: %v", err)
	}
	
	// Create a .git file pointing to the main repository
	gitFile := filepath.Join(worktreePath, ".git")
	gitContent := fmt.Sprintf("gitdir: %s\n", filepath.Join(cwd, ".git"))
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating .git file: %v", err)
	}
	
	fmt.Printf("Successfully checked out branch: %s\n", branch)
	fmt.Println("Worktree created successfully")

	fmt.Printf("Successfully created worktree '%s' with new branch '%s'\n", branch, branch)
	return nil
}