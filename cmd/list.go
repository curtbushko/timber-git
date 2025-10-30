package cmd

import (
	"fmt"
	"os"

	"github.com/curtbushko/timber-git/pkg/tg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all worktrees in the repository",
	Long: `List all worktrees in the repository with their paths and branches.

This is equivalent to running: git worktree list

Example:
  timber-git list`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := tg.ListWorktrees(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
