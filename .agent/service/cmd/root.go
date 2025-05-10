package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	rootConfigPath string
	quiet          bool
	verbosity      int
)

var rootCmd = &cobra.Command{
	Use:   "agentt",
	Short: "Agentt is a CLI tool for managing agent guidance definitions.",
	Long: `Agentt allows you to interact with and manage behaviors, recipes,
and other guidance artifacts used by AI agents.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbosity, _ := cmd.Flags().GetCount("verbose")

		_, _, err := GetMultiBackendAndConfig(verbosity)
		if err != nil {
			return fmt.Errorf("failed to initialize backend: %w", err)
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
