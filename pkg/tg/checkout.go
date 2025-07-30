package tg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/storage/filesystem"
	"github.com/go-git/go-billy/v6/osfs"
)

// SelectBranchWithFzf presents an interactive branch selector
// Since the fzf library API is complex, this implementation provides a fallback
// that shows branches in a simple numbered list for selection
func SelectBranchWithFzf() (string, error) {
	// Open the git repository to get remote branches
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("error opening repository: %v", err)
	}

	// Get all remote references
	remoteRefs, err := repo.References()
	if err != nil {
		return "", fmt.Errorf("error getting references: %v", err)
	}

	var branches []string
	err = remoteRefs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/") {
			branchName := strings.TrimPrefix(ref.Name().String(), "refs/remotes/origin/")
			// Skip HEAD reference
			if branchName != "HEAD" {
				branches = append(branches, branchName)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error iterating references: %v", err)
	}

	if len(branches) == 0 {
		return "", fmt.Errorf("no remote branches found")
	}

	// Display branches with numbers
	fmt.Println("Available branches:")
	for i, branch := range branches {
		fmt.Printf("%d) %s\n", i+1, branch)
	}

	// Get user selection
	fmt.Print("Select branch number (or press Enter to cancel): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %v", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		// User cancelled
		return "", nil
	}

	// Parse selection
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	if selection < 1 || selection > len(branches) {
		return "", fmt.Errorf("selection out of range: %d", selection)
	}

	return branches[selection-1], nil
}

// CheckoutWorktree creates a new worktree from an existing remote branch
// This is equivalent to: git worktree add <branch> -B <branch> "origin/<branch>"
func CheckoutWorktree(branch string) error {
	// Check if we're in a git repository by looking for .git file or directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	gitPath := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// Open the main repository using go-git
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	// Check if the remote branch exists
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", branch)
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err != nil {
		return fmt.Errorf("remote branch 'origin/%s' not found: %v", branch, err)
	}

	// Check if local directory already exists
	worktreePath := filepath.Join(cwd, branch)
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", branch)
	}

	// Create the worktree directory
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("error creating worktree directory: %v", err)
	}

	// Create filesystem for the new worktree
	worktreeFS := osfs.New(worktreePath)
	
	// Create git storage for the worktree
	gitDir := filepath.Join(worktreePath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return fmt.Errorf("error creating .git directory: %v", err)
	}
	
	gitFS := osfs.New(gitDir)
	storage := filesystem.NewStorage(gitFS, nil)

	// Clone the repository to create the worktree
	// Use the remote reference hash to create a local branch reference for cloning
	localBranchRef := plumbing.NewBranchReferenceName(branch)
	worktreeRepo, err := git.Clone(storage, worktreeFS, &git.CloneOptions{
		URL:           ".", // Clone from current repo
		ReferenceName: localBranchRef,
		SingleBranch:  true,
	})
	if err != nil {
		// Clean up on failure
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating worktree: %v", err)
	}

	// Create local branch tracking the remote branch
	localBranchRef = plumbing.NewBranchReferenceName(branch)
	err = worktreeRepo.Storer.SetReference(plumbing.NewHashReference(localBranchRef, remoteRef.Hash()))
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error creating local branch: %v", err)
	}

	// Set up branch config to track the remote branch
	cfg, err := worktreeRepo.Config()
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error getting repository config: %v", err)
	}

	// Initialize Branches map if nil
	if cfg.Branches == nil {
		cfg.Branches = make(map[string]*config.Branch)
	}
	
	// Add branch configuration
	cfg.Branches[branch] = &config.Branch{
		Name:   branch,
		Remote: "origin",
		Merge:  plumbing.NewBranchReferenceName(branch),
	}

	// Save the updated config
	err = worktreeRepo.Storer.SetConfig(cfg)
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error saving branch configuration: %v", err)
	}

	// Get the worktree and checkout the branch
	worktree, err := worktreeRepo.Worktree()
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: localBranchRef,
	})
	if err != nil {
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("error checking out branch: %v", err)
	}

	fmt.Printf("Successfully created worktree '%s' tracking 'origin/%s'\n", branch, branch)
	return nil
}