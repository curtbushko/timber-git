package tg

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	fzf "github.com/junegunn/fzf/src"
)

// SelectBranchWithFzf presents an interactive branch selector using FZF
func SelectBranchWithFzf() (string, error) {
	// Check if we're in a timber-git repository by looking for .git directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	gitPath := filepath.Join(cwd, bareRepoPath)
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", errors.New("not a timber-git repository (or any of the parent directories)")
	}

	// Open the bare repository to get remote branches
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return "", fmt.Errorf("error opening repository: %w", err)
	}

	// Get branches sorted by commit date (most recent first)
	branches, err := getBranchesSortedByDate(repo)
	if err != nil {
		return "", fmt.Errorf("error getting branches: %w", err)
	}

	if len(branches) == 0 {
		return "", errors.New("no remote branches found")
	}

	// Use FZF for selection
	return selectBranchWithFzf(branches, repo)
}

// getBranchesSortedByDate gets remote branches sorted by commit date
func getBranchesSortedByDate(repo *git.Repository) ([]string, error) {
	type branchInfo struct {
		name string
		date time.Time
	}

	var branchInfos []branchInfo
	remoteRefs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("error getting references: %w", err)
	}

	err = remoteRefs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/") {
			branchName := strings.TrimPrefix(ref.Name().String(), "refs/remotes/origin/")
			// Skip HEAD reference
			if branchName != "HEAD" {
				// Get commit date for sorting
				commit, err := repo.CommitObject(ref.Hash())
				if err == nil {
					branchInfos = append(branchInfos, branchInfo{
						name: branchName,
						date: commit.Committer.When,
					})
				} else {
					// If we can't get commit info, add with zero time
					branchInfos = append(branchInfos, branchInfo{
						name: branchName,
						date: time.Time{},
					})
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating references: %w", err)
	}

	// Sort by commit date (most recent first)
	sort.Slice(branchInfos, func(i, j int) bool {
		return branchInfos[i].date.After(branchInfos[j].date)
	})

	// Extract branch names
	branches := make([]string, len(branchInfos))
	for i, info := range branchInfos {
		branches[i] = info.name
	}

	return branches, nil
}

// selectBranchWithFzf uses the FZF library for interactive selection
func selectBranchWithFzf(branches []string, repo *git.Repository) (string, error) {
	// Create input channel with branch data
	inputChan := make(chan string)
	go func() {
		defer close(inputChan)
		for _, branch := range branches {
			inputChan <- branch
		}
	}()

	// Create output channel to receive selected items
	outputChan := make(chan string)
	var selected string
	go func() {
		for s := range outputChan {
			selected = strings.TrimSpace(s)
		}
	}()

	// Parse fzf options to match fzf.sh configuration
	options, err := fzf.ParseOptions(
		true, // load default options
		[]string{
			"--ansi",
			"--border-label=| Branches |",
			"--height=90%",
			"--border=rounded",
			"--margin=2,2,2,2",
			"--prompt=checkout worktree: ",
			"--preview-window=top:40%",
			"--preview=" + buildPreviewCommand(repo),
			"--bind=j:down,k:up,ctrl-j:preview-down,ctrl-k:preview-up,ctrl-f:preview-page-down,ctrl-b:preview-page-up,esc:abort",
		},
	)
	if err != nil {
		return "", fmt.Errorf("error parsing FZF options: %w", err)
	}

	// Configure input and output channels
	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	code, err := fzf.Run(options)
	if err != nil {
		return "", fmt.Errorf("FZF error: %w", err)
	}

	if code != fzf.ExitOk {
		return "", nil // User cancelled or other exit
	}

	if selected == "" {
		return "", nil // User cancelled
	}

	return selected, nil
}

// buildPreviewCommand creates a preview command that shows git log for the branch
func buildPreviewCommand(_ *git.Repository) string {
	// Since we can't use external git command, we'll create a simple preview
	// For now, return empty string - we'll need to implement custom preview
	return ""
}


// CheckoutWorktree creates a new worktree from an existing remote branch
// This is equivalent to: git worktree add <branch> -B <branch> "origin/<branch>"
func CheckoutWorktree(branch string) error {
	// Check if we're in a git repository by looking for .git directory
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

	// Check if the remote branch exists
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", branch)
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err != nil {
		return fmt.Errorf("remote branch 'origin/%s' not found: %w", branch, err)
	}

	// Check if local directory already exists
	worktreePath := filepath.Join(cwd, branch)
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", branch)
	}

	// Create the worktree directory
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("error creating worktree directory: %w", err)
	}

	// Create the local branch reference in the bare repository
	localBranchRef := plumbing.NewBranchReferenceName(branch)
	err = repo.Storer.SetReference(plumbing.NewHashReference(localBranchRef, remoteRef.Hash()))
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating local branch: %w", err)
	}

	// Create proper worktree .git directory structure
	err = setupWorktreeGitDir(worktreePath, cwd, branch)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error setting up worktree git directory: %w", err)
	}

	// Use go-git to checkout files properly to the worktree from the main repo
	err = checkoutFilesToWorktreeFromRepoWithBranch(repo, remoteRef.Hash(), worktreePath, branch)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error checking out files: %w", err)
	}

	// Initialize and update submodules if they exist
	err = initializeSubmodules(repo)
	if err != nil {
		slog.Warn("Failed to initialize submodules", "error", err)
		// Don't fail the checkout if submodule initialization fails
	}

	slog.Info("Successfully created worktree tracking origin branch", "worktree", branch, "origin_branch", branch)
	return nil
}