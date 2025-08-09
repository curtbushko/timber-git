package tg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// AddMainWorktree creates the main worktree in the current directory
// This is used during clone to set up the primary working directory
func AddMainWorktree(branch string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	gitPath := filepath.Join(cwd, ".bare")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
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
			return fmt.Errorf("error getting base reference: %v", err)
		}
	}

	// Create the branch reference in the main repository
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, baseRef.Hash()))
	if err != nil {
		return fmt.Errorf("error creating branch reference: %v", err)
	}

	// Create proper worktree .git directory structure for main worktree
	err = setupMainWorktreeGitDir(cwd, branch)
	if err != nil {
		return fmt.Errorf("error setting up main worktree git directory: %v", err)
	}

	// Checkout files to the current directory
	err = checkoutFilesToWorktreeFromRepo(repo, baseRef.Hash(), cwd)
	if err != nil {
		return fmt.Errorf("error checking out files: %v", err)
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
		return fmt.Errorf("error getting current directory: %v", err)
	}

	gitPath := filepath.Join(cwd, ".bare")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// Verify this is a valid git repository using go-git
	_, err = git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	// Implement worktree functionality using go-git
	// Open the main repository
	repo, err := git.PlainOpen(bareRepoPath)
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
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating branch reference: %v", err)
	}
	fmt.Printf("Created branch reference: %s\n", branch)

	// Create proper worktree .git directory structure
	err = setupWorktreeGitDir(worktreePath, cwd, branch)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error setting up worktree git directory: %v", err)
	}

	// Use go-git to checkout files properly to the worktree from the main repo
	fmt.Printf("Checking out files to worktree using go-git: %s\n", branch)
	err = checkoutFilesToWorktreeFromRepo(repo, baseRef.Hash(), worktreePath)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error checking out files: %v", err)
	}

	fmt.Printf("Successfully checked out branch: %s\n", branch)
	fmt.Println("Worktree created successfully")

	fmt.Printf("Successfully created worktree '%s' with new branch '%s'\n", branch, branch)
	return nil
}

// setupWorktreeGitDir creates a proper .git file structure for a worktree
func setupWorktreeGitDir(worktreePath, mainRepoPath, branch string) error {
	// Create worktree entry in main repo
	worktreesDir := filepath.Join(mainRepoPath, ".bare", "worktrees", branch)
	err := os.MkdirAll(worktreesDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create worktrees directory: %v", err)
	}

	// Create gitdir file pointing to worktree
	gitdirFile := filepath.Join(worktreesDir, "gitdir")
	err = os.WriteFile(gitdirFile, []byte(filepath.Join(worktreePath, ".git")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to create gitdir file: %v", err)
	}

	// Create HEAD file in worktree metadata
	headFile := filepath.Join(worktreesDir, "HEAD")
	headContent := fmt.Sprintf("ref: refs/heads/%s\n", branch)
	err = os.WriteFile(headFile, []byte(headContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create HEAD file: %v", err)
	}

	// Create .git file (not directory) pointing to the worktree metadata
	gitFile := filepath.Join(worktreePath, ".git")
	// Use absolute path to .bare/worktrees/<branch>
	absoluteWorktreesDir := filepath.Join(mainRepoPath, ".bare", "worktrees", branch)
	gitContent := fmt.Sprintf("gitdir: %s\n", absoluteWorktreesDir)
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .git file: %v", err)
	}

	return nil
}


// setupMainWorktreeGitDir creates a proper .git file structure for the main worktree
func setupMainWorktreeGitDir(worktreePath, branch string) error {
	// Create worktree entry in main repo
	worktreesDir := filepath.Join(worktreePath, ".bare", "worktrees", "main")
	err := os.MkdirAll(worktreesDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create main worktrees directory: %v", err)
	}

	// Create gitdir file pointing to main worktree
	gitdirFile := filepath.Join(worktreesDir, "gitdir")
	err = os.WriteFile(gitdirFile, []byte(filepath.Join(worktreePath, ".git")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to create gitdir file: %v", err)
	}

	// Create HEAD file in worktree metadata
	headFile := filepath.Join(worktreesDir, "HEAD")
	headContent := fmt.Sprintf("ref: refs/heads/%s\n", branch)
	err = os.WriteFile(headFile, []byte(headContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create HEAD file: %v", err)
	}

	// Create .git file (not directory) pointing to the worktree metadata
	gitFile := filepath.Join(worktreePath, ".git")
	gitContent := "gitdir: .bare/worktrees/main\n"
	err = os.WriteFile(gitFile, []byte(gitContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .git file: %v", err)
	}

	return nil
}


// checkoutFilesToWorktreeFromRepo checks out files using go-git directly from main repo
func checkoutFilesToWorktreeFromRepo(repo *git.Repository, commitHash plumbing.Hash, worktreePath string) error {
	// Get the commit object from the main repository
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("error getting commit object: %v", err)
	}

	// Get the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("error getting tree from commit: %v", err)
	}

	// Walk the tree and create files using go-git's file iteration
	return tree.Files().ForEach(func(file *object.File) error {
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
				fmt.Printf("Warning: failed to close reader: %v\n", closeErr)
			}
		}()

		// Create the file
		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := outFile.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close file: %v\n", closeErr)
			}
		}()

		// Copy contents
		_, err = outFile.ReadFrom(reader)
		return err
	})
}

