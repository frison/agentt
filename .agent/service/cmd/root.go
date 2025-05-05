package cmd

import (
	"log" // Add standard log package

	"github.com/spf13/cobra"
)

// Used for flags.
// var cfgFile string // Replaced by rootConfigPath

// Package variable to hold the value from the persistent flag
var rootConfigPath string
var quiet bool    // Flag to suppress non-warning/error logs
var verbosity int // Verbosity level controlled by -v flags
// var verbose bool // REMOVED: Unused, verbosity controlled by count flag

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
			// Initialize logging and backend ONCE before any command runs.
			// GetBackendAndConfig handles the sync.Once logic.
			// Pass the verbosity flag value.
			// We ignore the returned backend/config here, just ensuring initialization happens.
			_, _, err := GetBackendAndConfig(verbosity)
			if err != nil {
				// Log the error using standard log before slog might be configured
				log.Printf("Initialization failed: %v", err)
				// We might want to return the error here to stop execution
				// return err
			}
			// Logging is now configured by initializeBackend based on verbosity flags.
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
	// cobra.OnInitialize(initConfig) // REMOVED: Config loaded via GetBackend() called in PersistentPreRunE
	// Define flags and configuration settings here.
	// Persistent flags are global for the application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agentt.yaml or ./.agentt.yaml)")

	// Define persistent flags used by setupLogging (now read from config)
	// rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output (INFO level)")
	// rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational output (WARN level)")

	// Define persistent --config flag on the root command
	// The value will be stored in the rootConfigPath package variable.
	rootCmd.PersistentFlags().StringVarP(&rootConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")

	// Define persistent --quiet flag
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error log messages (equivalent to log level ERROR)")

	// Define persistent -v flag for verbosity (can be repeated)
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase logging verbosity (default: WARN, -v: INFO, -vv: DEBUG)")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(llmCmd)
}
