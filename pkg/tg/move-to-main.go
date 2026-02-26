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
	"github.com/go-git/go-git/v6/plumbing/format/config"
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
	diffText := generateWorktreeDiff(repo, currentBranch, defaultBranch)

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
		slog.Info("Operation canceled by user")
		return nil
	}

	// 5. Apply changes to default branch
	err = applyChangesToDefaultBranch(repo, defaultBranch)
	if err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	// 6. Reset the current worktree to clean state (discard changes)
	err = resetCurrentWorktree()
	if err != nil {
		// If reset fails, just log a warning - the important part (moving changes) is done
		slog.Warn("Could not reset worktree, but changes were successfully moved", "error", err)
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

	// Get the origin URL from repository config
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("error getting repository config: %w", err)
	}

	var originURL string
	if cfg.Remotes != nil && cfg.Remotes["origin"] != nil && len(cfg.Remotes["origin"].URLs) > 0 {
		originURL = cfg.Remotes["origin"].URLs[0]
	}

	// Get authentication method
	auth, err := getAuthMethod(originURL)
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

// generateWorktreeDiff generates a diff showing only working directory changes
// This manually compares HEAD tree with filesystem to work around go-git Status() bugs
func generateWorktreeDiff(_ *git.Repository, _, _ string) string {
	var diffOutput strings.Builder
	diffOutput.WriteString("Working directory changes:\n")

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		diffOutput.WriteString("Error getting current directory\n")
		return diffOutput.String()
	}

	// Get changed files by manually comparing HEAD with filesystem
	changes, err := getWorkingDirectoryChanges(cwd)
	if err != nil {
		diffOutput.WriteString(fmt.Sprintf("Error detecting changes: %v\n", err))
		return diffOutput.String()
	}

	// Show all changed files
	for _, change := range changes {
		diffOutput.WriteString(fmt.Sprintf(" %s %s\n", change.Status, change.Path))
	}

	return diffOutput.String()
}

// FileChange represents a file change in the working directory
type FileChange struct {
	Path   string
	Status string // "M" (modified), "A" (added), "D" (deleted)
}

// getWorkingDirectoryChanges manually detects file changes by comparing HEAD with filesystem
// This works around go-git Status() bugs with worktrees and deleted files
func getWorkingDirectoryChanges(worktreePath string) ([]FileChange, error) {
	// Find the bare repository
	bareRepoPath, err := findBareRepoPath(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("error finding bare repository: %w", err)
	}

	// Open the bare repository (not the worktree)
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return nil, fmt.Errorf("error opening repository: %w", err)
	}

	// Get current branch name from worktree metadata
	branchName, err := getCurrentBranchName(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("error getting branch: %w", err)
	}

	// Get the branch reference from the bare repo
	branchRef := plumbing.NewBranchReferenceName(branchName)
	head, err := repo.Reference(branchRef, true)
	if err != nil {
		return nil, fmt.Errorf("error getting branch reference: %w", err)
	}

	// Get HEAD commit object
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("error getting commit: %w", err)
	}

	// Get HEAD tree
	headTree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("error getting tree: %w", err)
	}

	// Build map of files in HEAD
	headFiles := make(map[string]plumbing.Hash)
	err = headTree.Files().ForEach(func(file *object.File) error {
		headFiles[file.Name] = file.Hash
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking HEAD tree: %w", err)
	}

	// Build map of files in working directory
	workingFiles := make(map[string]string) // path -> hash
	err = filepath.Walk(worktreePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Only process files, not directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(worktreePath, path)
		if err != nil {
			return err
		}

		// Compute file hash
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		workingFiles[relPath] = computeHash(content)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking working directory: %w", err)
	}

	// Compare and find changes
	var changes []FileChange

	// Find modified and deleted files
	for headPath, headHash := range headFiles {
		if workingHash, exists := workingFiles[headPath]; exists {
			// File exists in both - check if modified
			if workingHash != headHash.String() {
				changes = append(changes, FileChange{Path: headPath, Status: "M"})
			}
		} else {
			// File in HEAD but not in working directory - deleted
			changes = append(changes, FileChange{Path: headPath, Status: "D"})
		}
	}

	// Find added files
	for workingPath := range workingFiles {
		if _, exists := headFiles[workingPath]; !exists {
			// Skip .git file (worktree metadata)
			if workingPath == ".git" {
				continue
			}
			// File in working directory but not in HEAD - added
			changes = append(changes, FileChange{Path: workingPath, Status: "A"})
		}
	}

	return changes, nil
}

// computeHash computes git blob hash for file content
func computeHash(content []byte) string {
	hasher := plumbing.NewHasher(config.SHA1, plumbing.BlobObject, int64(len(content)))
	_, _ = hasher.Write(content)
	return hasher.Sum().String()
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

// applyChangesToDefaultBranch applies the working directory changes to the default branch worktree
func applyChangesToDefaultBranch(repo *git.Repository, defaultBranch string) error {
	slog.Info("Applying changes to branch working directory", "branch", defaultBranch)

	// Verify the repository has the default branch
	_, err := repo.Reference(plumbing.NewBranchReferenceName(defaultBranch), true)
	if err != nil {
		return fmt.Errorf("default branch %s not found in repository: %w", defaultBranch, err)
	}

	// Get current working directory to understand the source changes
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Get the default branch worktree path (assuming timber-git structure)
	cwdParent := filepath.Dir(cwd)
	defaultBranchWorktreePath := filepath.Join(cwdParent, defaultBranch)

	// Check if default branch worktree exists
	if _, err := os.Stat(defaultBranchWorktreePath); os.IsNotExist(err) {
		return fmt.Errorf("default branch worktree not found at %s", defaultBranchWorktreePath)
	}

	// Get changed files manually (works around go-git Status() bugs)
	changes, err := getWorkingDirectoryChanges(cwd)
	if err != nil {
		return fmt.Errorf("error detecting changes: %w", err)
	}

	// Apply each change to the default branch worktree
	for _, change := range changes {
		srcPath := filepath.Join(cwd, change.Path)
		dstPath := filepath.Join(defaultBranchWorktreePath, change.Path)

		switch change.Status {
		case "D": // Deleted file
			// Remove from destination
			if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("error removing file %s: %w", dstPath, err)
			}
			slog.Info("Removed file", "path", dstPath)

		case "A", "M": // Added or Modified file
			// Ensure destination directory exists
			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", dstDir, err)
			}

			// Copy file content
			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("error reading source file %s: %w", srcPath, err)
			}

			if err := os.WriteFile(dstPath, srcContent, 0644); err != nil {
				return fmt.Errorf("error writing to destination file %s: %w", dstPath, err)
			}

			slog.Info("Copied file", "from", srcPath, "to", dstPath)
		}
	}

	slog.Info("Successfully applied changes to working directory", "branch", defaultBranch, "path", defaultBranchWorktreePath)
	return nil
}

// generateWorkingDirDiff creates a diff representation for working directory changes
func generateWorkingDirDiff(status git.Status) string {
	var diffOutput strings.Builder
	diffOutput.WriteString("Working directory changes:\n")
	
	hasChanges := false
	for fileName, fileStatus := range status {
		// Only include files that have actual working directory or staging changes:
		// - Modified files in working directory (worktree != Unmodified)
		// - Staged changes (staging != Unmodified and staging != Untracked)
		// Exclude:
		// - Untracked files (??) - these are not part of git yet
		// - Files that are clean (both staging and worktree are Unmodified)
		// - Files that only exist (Added) but haven't been modified in working dir
		
		if fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked {
			// Skip untracked files
			continue
		}
		
		if fileStatus.Staging == git.Unmodified && fileStatus.Worktree == git.Unmodified {
			// Skip unmodified files
			continue
		}
		
		
		hasChanges = true
		
		// Format the status codes
		stagingStatus := string(fileStatus.Staging)
		worktreeStatus := string(fileStatus.Worktree)
		if stagingStatus == " " {
			stagingStatus = " "
		}
		if worktreeStatus == " " {
			worktreeStatus = " "
		}
		diffOutput.WriteString(fmt.Sprintf("%s%s %s\n", stagingStatus, worktreeStatus, fileName))
	}
	
	if !hasChanges {
		diffOutput.WriteString("No changes to move.\n")
	}

	return diffOutput.String()
}

// resetCurrentWorktree resets the current worktree to clean state
func resetCurrentWorktree() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Get current branch name
	branchName, err := getCurrentBranchName(cwd)
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
	}

	// Find and open the bare repository
	bareRepoPath, err := findBareRepoPath(cwd)
	if err != nil {
		return fmt.Errorf("error finding bare repository: %w", err)
	}

	bareRepo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("error opening bare repository: %w", err)
	}

	// Get the branch reference from the bare repo
	branchRef := plumbing.NewBranchReferenceName(branchName)
	head, err := bareRepo.Reference(branchRef, true)
	if err != nil {
		slog.Warn("Could not get HEAD reference, attempting worktree metadata", "error", err)
		return fmt.Errorf("error getting branch reference: %w", err)
	}

	// Manually reset the worktree by checking out files from the commit
	err = checkoutFilesToWorktreeFromRepoWithBranch(bareRepo, head.Hash(), cwd, branchName)
	if err != nil {
		return fmt.Errorf("error resetting worktree files: %w", err)
	}

	slog.Info("Reset worktree to clean state")
	return nil
}

// generateCommittedDiff generates a diff between committed changes on branches
func generateCommittedDiff(repo *git.Repository, currentBranch, defaultBranch string) (string, error) {
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