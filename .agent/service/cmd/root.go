package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Used for flags.
// var cfgFile string // Replaced by rootConfigPath

// Package variable to hold the value from the persistent flag
var rootConfigPath string

var (
	rootCmd = &cobra.Command{
		Use:   "agentt",
		Short: "Agent Guidance Service and CLI (config: flag > env > defaults)",
		Long: `Agentt provides an HTTP server for agent guidance discovery
and a command-line interface for interacting with the same definitions.

Configuration is loaded using the --config flag, the AGENTT_CONFIG environment
variable, or by searching default paths (./config.yaml, ./.agent/service/config.yaml, etc.)
relative to the current directory, in that order of precedence.`,
		// Silence errors and usage so main.go can handle them and exit non-zero.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Define persistent --config flag on the root command
	// The value will be stored in the rootConfigPath package variable.
	rootCmd.PersistentFlags().StringVarP(&rootConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")

	// Subcommands add themselves via their own init() functions.

	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)
}

// Helper function for handling errors
func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
