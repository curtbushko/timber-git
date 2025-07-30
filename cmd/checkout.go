package cmd

import (
	"fmt"
	"os"

	"github.com/curtbushko/timber-git/pkg/tg"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout [branch]",
	Short: "Create a new worktree from an existing remote branch",
	Long: `Create a new worktree at the specified branch name and create a local branch 
tracking the remote branch.

This is equivalent to running: git worktree add <branch> -B <branch> "origin/<branch>"

If no branch is specified, an interactive branch selector will be shown.

Example:
  timber-git checkout feature-xyz
  timber-git checkout  # interactive mode`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var branch string
		var err error

		if len(args) == 0 {
			// Interactive mode - use fzf to select branch
			branch, err = tg.SelectBranchWithFzf()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting branch: %v\n", err)
				os.Exit(1)
			}
			if branch == "" {
				// User cancelled selection
				os.Exit(0)
			}
		} else {
			branch = args[0]
		}

		if err := tg.CheckoutWorktree(branch); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}