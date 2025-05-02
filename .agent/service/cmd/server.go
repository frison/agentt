package cmd

import (
	// "agentt/internal/config" // REMOVED - Unused
	"agentt/internal/server"
	// "agentt/internal/guidance/backend" // REMOVED - Unused
	// "context" // REMOVED - Unused
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	// configPath string // REMOVED - Use rootConfigPath from root.go

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Manage the Agent Guidance HTTP server",
		Long:  `Commands to start, stop, or manage the Agent Guidance HTTP server.`,
	}

	serverStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Agent Guidance HTTP server",
		Long: `Starts the Agent Guidance HTTP server.
Backend is initialized on startup based on config.
Uses the configuration specified via --config flag, AGENTT_CONFIG env var, or default search paths.`,
		Run: startServer,
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)
	// Config flag is persistent on root
}

func startServer(cmd *cobra.Command, args []string) {
	// --- Use common setup to get config and initialized backend ---
	// Standard log will be replaced by slog via root PersistentPreRunE
	setupRes, err := setupDiscovery(rootConfigPath)
	if err != nil {
		// Use Fatalf or equivalent with slog? Cobra handles printing errors, maybe just return.
		// For now, stick to Fatal as it guarantees exit on setup failure.
		// slog.Error("Server setup failed", "error", err) // Slog might not be configured yet if PersistentPreRunE fails?
		log.Fatalf("Server setup failed: %v", err) // Keep standard log fatal here for now
	}
	cfg := setupRes.Cfg
	guidanceBackend := setupRes.Backend // Get the initialized backend

	// --- REMOVED: Old Store/Watcher Setup ---
	// guidanceStore := store.NewGuidanceStore()
	// wchr, err := discovery.NewWatcher(cfg, guidanceStore, loadedPath)
	// ...
	// err = wchr.InitialScan()
	// ...

	// --- Setup HTTP Server (passing backend) ---
	srv := server.NewServer(cfg, guidanceBackend) // Pass backend instead of store

	// --- REMOVED: Watcher Start ---
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// wchr.Start(ctx)

	// --- Start HTTP Server ---
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server", "address", cfg.ListenAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			serverErrChan <- err // Report error
		}
		close(serverErrChan) // Signal clean shutdown
	}()

	// --- Graceful Shutdown Handling ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	slog.Info("Agent Guidance Service started. Press Ctrl+C to shutdown.")

	select {
	case <-quit: // Wait for shutdown signal
		slog.Info("Received shutdown signal...")
	case err := <-serverErrChan: // Wait for server error
		if err != nil {
			// Logged already, just exit.
			// Consider returning error from startServer if Cobra handles exit codes well?
			os.Exit(1) // Exit if server failed critically
		}
		// If channel closed without error, server shut down cleanly (maybe via a /shutdown endpoint later?)
		slog.Info("Server stopped gracefully.")
		return // Exit cleanly
	}

	slog.Info("Shutting down server...")

	// --- REMOVED: Watcher Shutdown ---
	// cancel()
	// if err := wchr.Close(); err != nil {
	// 	slog.Error("Error closing watcher", "error", err)
	// }

	// TODO: Implement graceful HTTP server shutdown (using context)
	// Example:
	// shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer shutdownCancel()
	// if err := srv.Shutdown(shutdownCtx); err != nil {
	// 	slog.Error("HTTP server graceful shutdown failed", "error", err)
	// }

	slog.Info("Server shutdown complete.")
}
