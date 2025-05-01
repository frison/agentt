package cmd

import (
	"agentt/internal/config"
	"agentt/internal/discovery"
	"agentt/internal/server"
	"agentt/internal/store"
	"context"
	"log"
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
		Long: `Starts the Agent Guidance HTTP server and the file watcher.
Uses the configuration specified via --config flag, AGENTT_CONFIG env var, or default search paths.`,
		Run: startServer,
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)

	// REMOVED flag definition - Now persistent on root
	// serverStartCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")
}

func startServer(cmd *cobra.Command, args []string) {
	// --- Configuration ---
	// Use rootConfigPath directly from root.go
	cfg, loadedPath, err := config.FindAndLoadConfig(rootConfigPath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	log.Printf("Using configuration file: %s", loadedPath)

	// --- Setup Dependencies ---
	guidanceStore := store.NewGuidanceStore()
	wchr, err := discovery.NewWatcher(cfg, guidanceStore, loadedPath)
	if err != nil {
		log.Fatalf("Error creating file watcher: %v", err)
	}

	srv := server.NewServer(cfg, guidanceStore)

	// --- Initial Scan & Start Watcher ---
	err = wchr.InitialScan()
	if err != nil {
		// Log the error but attempt to start anyway, watcher might still work
		log.Printf("Warning during initial scan: %v", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wchr.Start(ctx) // Start the watcher in the background

	// --- Start HTTP Server ---
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// --- Graceful Shutdown Handling ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Agent Guidance Service started. Press Ctrl+C to shutdown.")

	<-quit // Wait for shutdown signal

	log.Println("Shutting down server...")

	// Cancel context for watcher
	cancel()

	// Close watcher
	if err := wchr.Close(); err != nil {
		log.Printf("Error closing watcher: %v", err)
	}

	// TODO: Add graceful HTTP server shutdown if needed

	log.Println("Server gracefully stopped.")
}
