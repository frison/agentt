package cmd

import (
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Used for flags.
// var cfgFile string // Replaced by rootConfigPath

// Package variable to hold the value from the persistent flag
var rootConfigPath string
var quiet bool    // Flag to suppress non-warning/error logs
var verbosity int // Verbosity level controlled by -v flags

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
			// Configure logging level based on -q and -v flags

			// --- Configure slog ---
			var logLevel slog.Level
			if quiet {
				logLevel = slog.LevelError // Only errors and above (effectively)
			} else {
				switch verbosity {
				case 0:
					logLevel = slog.LevelWarn // Default: Warn+
				case 1:
					logLevel = slog.LevelInfo // -v: Info+
				default: // >= 2
					logLevel = slog.LevelDebug // -vv and above: Debug+
				}
			}

			opts := &slog.HandlerOptions{
				Level: logLevel,
				// Optional: Customize time format or remove time
				// ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// 	if a.Key == slog.TimeKey {
				// 		return slog.Attr{}
				// 	}
				// 	return a
				// },
			}

			// Use TextHandler for more human-readable CLI output
			handler := slog.NewTextHandler(os.Stderr, opts)
			logger := slog.New(handler)
			slog.SetDefault(logger)

			// Discard output from the standard log package, just in case any library uses it
			log.SetOutput(io.Discard)

			return nil
		},
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

	// Define persistent --quiet flag
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error log messages (equivalent to log level ERROR)")

	// Define persistent -v flag for verbosity (can be repeated)
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase logging verbosity (default: WARN, -v: INFO, -vv: DEBUG)")

	// Subcommands add themselves via their own init() functions.

	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)
}
