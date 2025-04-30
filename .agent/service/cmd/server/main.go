package main

import (
	"agentt/internal/config"
	"agentt/internal/discovery"
	"agentt/internal/server"
	"agentt/internal/store"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// --- Configuration ---
	configPath := flag.String("config", ".agent/service/config.yaml", "Path to the configuration file.")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration from %s: %v", *configPath, err)
	}

	// --- Setup Dependencies ---
	guidanceStore := store.NewGuidanceStore()
	wchr, err := discovery.NewWatcher(cfg, guidanceStore, *configPath)
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

	// Close watcher (optional, depends if resources need explicit cleanup)
	if err := wchr.Close(); err != nil {
		log.Printf("Error closing watcher: %v", err)
	}

	// TODO: Add graceful HTTP server shutdown if needed

	log.Println("Server gracefully stopped.")
}
