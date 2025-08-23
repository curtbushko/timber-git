// Package cmd provides the CLI commands for timber-git
package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// themeCmd represents the theme command
var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Configure the theme that the command will use",
	Long: `This tool uses flair (github.com/curtbushko/flair) to configure themes. Flair
sets themes in your $HOME/.config/flair directory and allows for universal theme settings
for tools that support it. Flair currently only supports Charmbracelet based tools.`,
	Run: func(_ *cobra.Command, _ []string) {
		slog.Info("theme called")
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// themeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// themeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
