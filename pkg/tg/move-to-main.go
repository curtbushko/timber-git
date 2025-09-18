package tg

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// MoveToDefaultBranch moves changes from the current worktree to the default branch
func MoveToDefaultBranch() error {
	// 1. Verify we're in a worktree directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Check if we're in a timber-git worktree by looking for .git file and .bare directory
	gitFile := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitFile); os.IsNotExist(err) {
		return errors.New("not in a git worktree directory")
	}

	// Find the bare repository path by reading .git file
	bareRepoPath, err := findBareRepoPath(cwd)
	if err != nil {
		return fmt.Errorf("error finding bare repository: %w", err)
	}

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	// Get the default branch name
	defaultBranch, err := getDefaultBranchName(repo)
	if err != nil {
		return fmt.Errorf("error determining default branch: %w", err)
	}

	// Get current branch name from the worktree
	currentBranch, err := getCurrentBranchName(cwd)
	if err != nil {
		return fmt.Errorf("error determining current branch: %w", err)
	}

	if currentBranch == defaultBranch {
		return fmt.Errorf("already on default branch '%s'", defaultBranch)
	}

	slog.Info("Moving changes between branches", "from", currentBranch, "to", defaultBranch)

	// 2. Pull the latest changes to the default branch
	err = pullDefaultBranch(repo, defaultBranch)
	if err != nil {
		return fmt.Errorf("error pulling default branch: %w", err)
	}

	// 3. Generate diff between current worktree and default branch
	diffText, err := generateWorktreeDiff(repo, currentBranch, defaultBranch)
	if err != nil {
		return fmt.Errorf("error generating diff: %w", err)
	}

	if strings.TrimSpace(diffText) == "" {
		slog.Info("No changes to move")
		return nil
	}

	// 4. Show diff and ask for confirmation
	fmt.Println("Changes to be moved:")
	fmt.Println("===================")
	fmt.Println(diffText)
	fmt.Println("===================")

	if !confirmChanges() {
		slog.Info("Operation cancelled by user")
		return nil
	}

	// 5. Apply changes to default branch
	err = applyChangesToDefaultBranch(repo, defaultBranch)
	if err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	slog.Info("Successfully moved changes", "to_branch", defaultBranch)
	return nil
}

// findBareRepoPath finds the path to the bare repository from a worktree
func findBareRepoPath(worktreePath string) (string, error) {
	gitFile := filepath.Join(worktreePath, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		return "", fmt.Errorf("error reading .git file: %w", err)
	}

	gitdir := strings.TrimSpace(string(content))
	if !strings.HasPrefix(gitdir, "gitdir: ") {
		return "", errors.New("invalid .git file format")
	}

	gitdirPath := strings.TrimPrefix(gitdir, "gitdir: ")
	if !filepath.IsAbs(gitdirPath) {
		gitdirPath = filepath.Join(worktreePath, gitdirPath)
	}

	// For worktrees, gitdir points to .bare/worktrees/<branch>
	// We need to go up two levels to get to .bare
	bareRepoPath := filepath.Join(gitdirPath, "..", "..")
	bareRepoPath, err = filepath.Abs(bareRepoPath)
	if err != nil {
		return "", fmt.Errorf("error resolving bare repo path: %w", err)
	}

	return bareRepoPath, nil
}

// getCurrentBranchName gets the current branch name from the worktree
func getCurrentBranchName(worktreePath string) (string, error) {
	gitFile := filepath.Join(worktreePath, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		return "", fmt.Errorf("error reading .git file: %w", err)
	}

	gitdir := strings.TrimSpace(string(content))
	if !strings.HasPrefix(gitdir, "gitdir: ") {
		return "", errors.New("invalid .git file format")
	}

	gitdirPath := strings.TrimPrefix(gitdir, "gitdir: ")
	if !filepath.IsAbs(gitdirPath) {
		gitdirPath = filepath.Join(worktreePath, gitdirPath)
	}

	// Read HEAD file from worktree metadata
	headFile := filepath.Join(gitdirPath, "HEAD")
	headContent, err := os.ReadFile(headFile)
	if err != nil {
		return "", fmt.Errorf("error reading HEAD file: %w", err)
	}

	headLine := strings.TrimSpace(string(headContent))
	if !strings.HasPrefix(headLine, "ref: refs/heads/") {
		return "", errors.New("HEAD does not point to a branch")
	}

	branchName := strings.TrimPrefix(headLine, "ref: refs/heads/")
	return branchName, nil
}

// pullDefaultBranch pulls the latest changes for the default branch
func pullDefaultBranch(repo *git.Repository, defaultBranch string) error {
	slog.Info("Pulling latest changes", "branch", defaultBranch)

	// Get authentication method
	auth, err := getAuthMethod("")
	if err != nil {
		slog.Warn("Could not get auth method, continuing without auth", "error", err)
		auth = nil
	}

	// Fetch from origin
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("error fetching from origin: %w", err)
	}

	// Update local default branch to match origin
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", defaultBranch)
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err != nil {
		return fmt.Errorf("error getting remote branch reference: %w", err)
	}

	localBranchRef := plumbing.NewBranchReferenceName(defaultBranch)
	err = repo.Storer.SetReference(plumbing.NewHashReference(localBranchRef, remoteRef.Hash()))
	if err != nil {
		return fmt.Errorf("error updating local branch: %w", err)
	}

	slog.Info("Successfully updated branch to latest", "branch", defaultBranch)
	return nil
}

// generateWorktreeDiff generates a diff between the current worktree and default branch
func generateWorktreeDiff(repo *git.Repository, currentBranch, defaultBranch string) (string, error) {
	// Get commits for both branches
	currentBranchRef := plumbing.NewBranchReferenceName(currentBranch)
	currentRef, err := repo.Reference(currentBranchRef, true)
	if err != nil {
		return "", fmt.Errorf("error getting current branch reference: %w", err)
	}

	defaultBranchRef := plumbing.NewBranchReferenceName(defaultBranch)
	defaultRef, err := repo.Reference(defaultBranchRef, true)
	if err != nil {
		return "", fmt.Errorf("error getting default branch reference: %w", err)
	}

	// Get commit objects
	currentCommit, err := repo.CommitObject(currentRef.Hash())
	if err != nil {
		return "", fmt.Errorf("error getting current commit: %w", err)
	}

	defaultCommit, err := repo.CommitObject(defaultRef.Hash())
	if err != nil {
		return "", fmt.Errorf("error getting default commit: %w", err)
	}

	// Get trees
	currentTree, err := currentCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("error getting current tree: %w", err)
	}

	defaultTree, err := defaultCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("error getting default tree: %w", err)
	}

	// Generate patch between trees
	patch, err := defaultTree.Patch(currentTree)
	if err != nil {
		return "", fmt.Errorf("error generating patch: %w", err)
	}

	return patch.String(), nil
}

// confirmChanges prompts the user to confirm the changes
func confirmChanges() bool {
	fmt.Print("Do you want to apply these changes to the default branch? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// applyChangesToDefaultBranch applies the diff to the default branch
func applyChangesToDefaultBranch(repo *git.Repository, defaultBranch string) error {
	slog.Info("Applying changes to branch", "branch", defaultBranch)

	// Get the default branch reference and commit
	defaultBranchRef := plumbing.NewBranchReferenceName(defaultBranch)
	defaultRef, err := repo.Reference(defaultBranchRef, true)
	if err != nil {
		return fmt.Errorf("error getting default branch reference: %w", err)
	}

	defaultCommit, err := repo.CommitObject(defaultRef.Hash())
	if err != nil {
		return fmt.Errorf("error getting default commit: %w", err)
	}

	// Get current branch commit (source of changes)
	currentBranch, err := getCurrentBranchName(".")
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
	}

	currentBranchRef := plumbing.NewBranchReferenceName(currentBranch)
	currentRef, err := repo.Reference(currentBranchRef, true)
	if err != nil {
		return fmt.Errorf("error getting current branch reference: %w", err)
	}

	currentCommit, err := repo.CommitObject(currentRef.Hash())
	if err != nil {
		return fmt.Errorf("error getting current commit: %w", err)
	}

	// Create a new commit on the default branch with the changes from current branch
	// This effectively "applies" the diff by using the current branch's tree
	currentTree, err := currentCommit.Tree()
	if err != nil {
		return fmt.Errorf("error getting current tree: %w", err)
	}

	// Create commit object
	newCommit := &object.Commit{
		Author: object.Signature{
			Name:  "timber-git",
			Email: "timber-git@local",
			When:  currentCommit.Author.When,
		},
		Committer: object.Signature{
			Name:  "timber-git",
			Email: "timber-git@local", 
			When:  currentCommit.Committer.When,
		},
		Message:  fmt.Sprintf("Move changes from %s to %s\n\n%s", currentBranch, defaultBranch, currentCommit.Message),
		TreeHash: currentTree.Hash,
		ParentHashes: []plumbing.Hash{defaultCommit.Hash},
	}

	// Encode and store the commit
	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.CommitObject)
	
	writer, err := obj.Writer()
	if err != nil {
		return fmt.Errorf("error getting object writer: %w", err)
	}
	
	err = newCommit.Encode(obj)
	if err != nil {
		return fmt.Errorf("error encoding commit: %w", err)
	}
	
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("error storing commit: %w", err)
	}

	// Update the default branch reference
	err = repo.Storer.SetReference(plumbing.NewHashReference(defaultBranchRef, commitHash))
	if err != nil {
		return fmt.Errorf("error updating default branch reference: %w", err)
	}

	slog.Info("Successfully applied changes to branch", "branch", defaultBranch)
	return nil
}