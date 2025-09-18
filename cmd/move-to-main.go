package cmd

import (
	"fmt"
	"os"

	"github.com/curtbushko/timber-git/pkg/tg"
	"github.com/spf13/cobra"
)

var moveToMainCmd = &cobra.Command{
	Use:   "move-to-main",
	Short: "Move changes from current worktree to the default branch",
	Long: `Move changes from the current worktree to the default branch by:
1. Pulling the latest changes to the default branch
2. Generating a diff of the changes between worktrees
3. Applying the changes to the default branch after user confirmation

This command must be run from within a worktree directory.

Example:
  timber-git move-to-main`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if err := tg.MoveToDefaultBranch(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(moveToMainCmd)
}