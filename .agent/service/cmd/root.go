package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	rootConfigPath string
	quiet          bool
	verbosity      int
)

var rootCmd = &cobra.Command{
	Use:   "agentt",
	Short: "Agent Guidance Service and CLI (config: flag > env > defaults)",
	Long: `Agentt provides an HTTP server for agent guidance discovery
and a command-line interface for interacting with the same definitions.

Configuration is loaded using the --config flag, the AGENTT_CONFIG environment
variable, or by searching default paths (./config.yaml, ./.agent/service/config.yaml, etc.)
relative to the current directory, in that order of precedence.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := GetBackendAndConfig(verbosity)
		if err != nil {
			log.Printf("Initialization failed: %v", err)
			return err
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error log messages (equivalent to log level ERROR)")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase logging verbosity (default: WARN, -v: INFO, -vv: DEBUG)")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(llmCmd)
}
