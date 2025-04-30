package cmd

import (
	"github.com/spf13/cobra"
)

// CliCmd is the root command for CLI-specific operations (excluding the server).
var CliCmd = &cobra.Command{
	Use:   "cli", // Or perhaps just add commands directly to root? Let's stick with cli for now.
	Short: "Interact with the Agent Guidance definitions via the command line.",
	Long: `Provides commands to view summaries, details, and other information
about the agent guidance behaviors and recipes without running the server.`,
	// No Run function needed for a parent command
}

func init() {
	// Add the cli command tree to the root command
	rootCmd.AddCommand(CliCmd)
}
