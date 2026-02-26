// Package cmd provides CLI commands for timber-git
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/curtbushko/timber-git/pkg/tg"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone <repository_url> [directory]",
	Short: "Clone a repository using a bare clone and git worktrees",
	Long: `Clones a give <repository_url> into a .bare directory and
creates the default branch in a worktree (usually main).
Optionally specify a target directory name.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		targetDir := ""
		if len(args) > 1 {
			targetDir = args[1]
		}
		if err := tg.BareCloneWithTarget(args[0], targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
