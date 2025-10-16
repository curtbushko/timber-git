package tg

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v6"
)

// SubmoduleHandler defines the interface for handling git submodules
type SubmoduleHandler interface {
	// InitSubmodules initializes all submodules in a repository recursively
	InitSubmodules(repo *git.Repository) error
	// UpdateSubmodules updates all submodules in a repository recursively
	UpdateSubmodules(repo *git.Repository) error
}

// DefaultSubmoduleHandler implements SubmoduleHandler using go-git
type DefaultSubmoduleHandler struct{}

// NewSubmoduleHandler creates a new default submodule handler
func NewSubmoduleHandler() SubmoduleHandler {
	return &DefaultSubmoduleHandler{}
}

// InitSubmodules initializes all submodules in a repository recursively
func (h *DefaultSubmoduleHandler) InitSubmodules(repo *git.Repository) error {
	if repo == nil {
		return errors.New("repository cannot be nil")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return fmt.Errorf("error getting submodules: %w", err)
	}

	// Initialize each submodule
	for _, submodule := range submodules {
		if err := submodule.Init(); err != nil {
			return fmt.Errorf("error initializing submodule %s: %w", submodule.Config().Name, err)
		}
	}

	return nil
}

// UpdateSubmodules updates all submodules in a repository recursively
func (h *DefaultSubmoduleHandler) UpdateSubmodules(repo *git.Repository) error {
	if repo == nil {
		return errors.New("repository cannot be nil")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return fmt.Errorf("error getting submodules: %w", err)
	}

	// Update each submodule recursively
	for _, submodule := range submodules {
		if err := submodule.Update(&git.SubmoduleUpdateOptions{
			Init:              true,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); err != nil {
			return fmt.Errorf("error updating submodule %s: %w", submodule.Config().Name, err)
		}
	}

	return nil
}

// initializeSubmodules is a helper function that initializes and updates submodules in a repository
func initializeSubmodules(repo *git.Repository) error {
	handler := NewSubmoduleHandler()

	// First initialize the submodules
	if err := handler.InitSubmodules(repo); err != nil {
		return fmt.Errorf("error initializing submodules: %w", err)
	}

	// Then update them to fetch content
	if err := handler.UpdateSubmodules(repo); err != nil {
		return fmt.Errorf("error updating submodules: %w", err)
	}

	return nil
}
