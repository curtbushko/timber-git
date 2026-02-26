package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/curtbushko/timber-git/pkg/tg"
)

var removeCmd = &cobra.Command{
	Use:   "remove <worktree>",
	Short: "Remove a worktree and delete its branch",
	Long: `Remove a worktree from the repository and delete the associated branch.

This is equivalent to running: git worktree remove <worktree> && git branch -D <worktree>

Example:
  timber-git remove feature-xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		worktree := args[0]

		if err := tg.RemoveWorktree(worktree); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
