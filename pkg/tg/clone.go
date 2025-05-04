package tg

import "errors"

var (
	bareRepoPath := ".bare"
)

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <repository_url> [directory_name]")
		os.Exit(1)
	}

	url := os.Args[1]
	basename := filepath.Base(url)
	name := ""
	if len(os.Args) > 2 {
		name = os.Args[2]
	} else {
		name = strings.TrimSuffix(basename, filepath.Ext(basename))
	}

	if err := os.Mkdir(name, 0755); err != nil {
		fmt.Printf("Error creating directory '%s': %v\n", name, err)
		os.Exit(1)
	}

	if err := os.Chdir(name); err != nil {
		fmt.Printf("Error changing directory to '%s': %v\n", name, err)
		os.Exit(1)
	}

	repo, err := git.PlainClone(bareRepoPath, false, &git.CloneOptions{
		URL:      url,
		Bare:     true,
		Progress: os.Stdout, // Optional: show progress
	})
	if err != nil {
		fmt.Printf("Error cloning bare repository: %v\n", err)
		os.Exit(1)
	}

	// Get the default branch (HEAD) of the bare repository
	headRef, err := repo.Head()
	if err != nil {
		fmt.Printf("Error getting HEAD of the bare repository: %v\n", err)
		os.Exit(1)
	}
	defaultBranchName := headRef.Short() // Get the short name (e.g., "main", "master")

	gitFilePath := ".git"
	content := fmt.Sprintf("gitdir: ./%s\n", bareRepoPath)
	if err := os.WriteFile(gitFilePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing to '%s': %v\n", gitFilePath, err)
		os.Exit(1)
	}

	cfg, err := repo.Config()
	if err != nil {
		fmt.Printf("Error getting repository config: %v\n", err)
		os.Exit(1)
	}

	cfg.Raw.Section("remote \"origin\"").Set("fetch", "+refs/heads/*:refs/remotes/origin/*")

	if err := repo.Storer.SetConfig(cfg); err != nil {
		fmt.Printf("Error setting remote origin fetch config: %v\n", err)
		os.Exit(1)
	}

	_, err = repo.WorktreeAdd(&git.WorktreeAddOptions{
		Name:   defaultBranchName,
		Path:   ".",
		Branch: headRef.Name(), // Use the actual reference name
	})
	if err != nil {
		// If the worktree already exists, try a simple 'git worktree add <branch_name>'
		cmd := exec.Command("git", "worktree", "add", defaultBranchName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error adding worktree '%s': %v\n", defaultBranchName, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Successfully initialized Git repository in '%s' with default branch '%s'\n", name, defaultBranchName)
}
