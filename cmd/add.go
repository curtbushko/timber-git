package cmd

import (
	"fmt"
	"os"

	"github.com/curtbushko/timber-git/pkg/tg"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new worktree with a new branch",
	Long: `Create a new worktree at the specified branch name and create a new branch with the same name.

This is equivalent to running: git worktree add <branch> -b <branch>

Example:
  timber-git add feature-xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]

		if err := tg.AddWorktree(branch); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}