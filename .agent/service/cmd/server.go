package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the base command when called without any subcommands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the Agent Guidance HTTP server",
	Long:  `Starts and manages the HTTP server component that serves guidance definitions over an API.`,
	RunE:  runServer, // Assign the run function
	// PersistentPreRunE: func(cmd *cobra.Command, args []string) error { ... } // Optional setup for server commands
}

func init() {
	// No server specific flags needed on the parent command itself
	// serverStartCmd specific flags can be added here if needed
	// serverCmd.AddCommand(serverStartCmd) is done in cmd/root.go
}

// runServer executes the server command logic.
func runServer(cmd *cobra.Command, args []string) error {
	// Get verbosity level from root command's persistent flag
	verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")

	// Get the initialized backend and config
	backendInstance, cfg, err := GetBackendAndConfig(verbosity)
	if err != nil {
		// Handle initialization error (already logged by GetBackendAndConfig/initializeBackend)
		return fmt.Errorf("initialization failed: %w", err)
	}
	if backendInstance == nil || cfg == nil {
		return fmt.Errorf("initialization returned nil backend or config without specific error")
	}

	// Pass the obtained config and backend to the setup function
	return setupAndRunServer(cfg, backendInstance)
}
