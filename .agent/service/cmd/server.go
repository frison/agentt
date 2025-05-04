package cmd

import (
	"github.com/spf13/cobra"
	// "fmt" // Removed: Unused
	// "log/slog" // Removed: Unused
	// "agentt/internal/server" // Removed: Unused after removing serverStartCmd logic
	// "agentt/internal/config" // Removed: Unused after removing serverStartCmd logic
	// "agentt/internal/guidance/backend" // Removed: Unused after removing serverStartCmd logic
)

// serverCmd represents the base command when called without any subcommands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the Agent Guidance HTTP server",
	Long:  `Starts and manages the HTTP server component that serves guidance definitions over an API.`,
	// PersistentPreRunE: func(cmd *cobra.Command, args []string) error { ... } // Optional setup for server commands
}

/* REMOVED: Unused variable
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP server",
	Long:  `Launches the HTTP server to listen for requests according to the configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialization should happen in Root's PersistentPreRunE or here
		// Ensure backend is initialized before starting server
		if err := initializeBackend(); err != nil {
			return fmt.Errorf("backend initialization failed before starting server: %w", err)
		}

		// Configuration should be loaded by initializeBackend
		if loadedConfig == nil {
			return fmt.Errorf("configuration not loaded, cannot start server")
		}

		// Assuming globalBackendService is populated by initializeBackend
		if len(globalBackendService) == 0 {
			return fmt.Errorf("no backends initialized, cannot start server")
		}

		// Pass the first backend for now (or adapt server.NewServer if needed)
		// TODO: Adapt server to handle multiple backends if necessary
		guidanceBackend := globalBackendService[0]

		srv := server.NewServer(loadedConfig, guidanceBackend)
		slog.Info("Starting server", "address", loadedConfig.ListenAddress)
		if err := srv.Start(); err != nil {
			return fmt.Errorf("server failed to start: %w", err)
		}
		return nil
	},
}
*/

func init() {
	// serverStartCmd specific flags can be added here if needed
	// serverCmd.AddCommand(serverStartCmd) is done in cmd/root.go
}
