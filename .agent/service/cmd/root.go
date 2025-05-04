package cmd

import (
	"fmt"
	// "io"
	// "log"
	"log/slog"
	// "os"

	"github.com/spf13/cobra"
)

// Used for flags.
// var cfgFile string // Replaced by rootConfigPath

// Package variable to hold the value from the persistent flag
var rootConfigPath string
var quiet bool    // Flag to suppress non-warning/error logs
var verbosity int // Verbosity level controlled by -v flags
var verbose bool

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			setupLogging(verbose, quiet)
			slog.Debug("Executing rootCmd PersistentPreRunE")

			// Initialize backend(s) conditionally
			// Remove IsCompletionCommand() check
			if cmd.Name() != "server" && cmd.Name() != "help" && cmd.Name() != "llm" { // Add other commands to skip? Maybe completion commands?
				slog.Debug("Initializing backend service for command", "command", cmd.Name())
				err := initializeBackend()
				if err != nil {
					return fmt.Errorf("backend initialization failed: %w", err)
				}
				slog.Debug("Backend service initialized successfully")
			} else {
				slog.Debug("Skipping backend initialization for command", "command", cmd.Name())
			}
			return nil
		},
	}
)

// Execute executes the root command.
func Execute() error {
	// Call Execute which might exit, so os is needed
	return rootCmd.Execute()
}

func init() {
	// Define persistent --config flag on the root command
	// The value will be stored in the rootConfigPath package variable.
	rootCmd.PersistentFlags().StringVarP(&rootConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")

	// Define persistent --quiet flag
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error log messages (equivalent to log level ERROR)")

	// Define persistent -v flag for verbosity (can be repeated)
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase logging verbosity (default: WARN, -v: INFO, -vv: DEBUG)")

	// Subcommands add themselves via their own init() functions.
	// Ensure all commands are added here if not done in their own init.
	rootCmd.AddCommand(summaryCmd) // Ensure these are still added
	rootCmd.AddCommand(detailsCmd)
	rootCmd.AddCommand(llmCmd)
	rootCmd.AddCommand(serverCmd)
	// serverCmd adds its own subcommands (like start) in its init
}
