package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the base command when called without any subcommands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Agentt HTTP server",
	Long:  `Starts the HTTP server to provide an API for agent guidance definitions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")
		// For the server, we typically want the MultiBackend to serve all configured backends.
		backendInstance, cfg, err := GetMultiBackendAndConfig(verbosity)
		if err != nil {
			return fmt.Errorf("failed to initialize backend for server: %w", err)
		}
		if cfg == nil || backendInstance == nil {
			return fmt.Errorf("critical error: config or backend instance is nil after initialization")
		}

		// The liveness and readiness probes might depend on backend health in the future.
		// For now, they are simple HTTP 200s.
		return setupAndRunServer(cfg, backendInstance)
	},
	// PersistentPreRunE: func(cmd *cobra.Command, args []string) error { ... } // Optional setup for server commands
}

func init() {
	// No server specific flags needed on the parent command itself
	// serverStartCmd specific flags can be added here if needed
	// serverCmd.AddCommand(serverStartCmd) is done in cmd/root.go
	rootCmd.AddCommand(serverCmd)
}
