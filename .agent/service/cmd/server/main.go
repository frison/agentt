package main

import (
	// "agentt/internal/config"
	// "agentt/internal/discovery"
	// "agentt/internal/server"
	// "agentt/internal/store"
	// "context"
	// "flag"
	"log"
	// "net/http"
)

func main() {
	log.Println("This standalone server entrypoint (cmd/server/main.go) is deprecated.")
	log.Println("Use 'agentt server start' instead.")
	// --- All original code commented out ---
	/*
		// --- Configuration ---
		configPath := flag.String("config", ".agent/service/config.yaml", "Path to the configuration file.")
		flag.Parse()

		cfg, err := config.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Error loading configuration from %s: %v", *configPath, err)
		}

		log.Printf("Using configuration file: %s", *configPath)

		// --- Setup Dependencies ---
		guidanceStore := store.NewGuidanceStore()

		// Create watcher and perform initial scan (fatal on duplicate ID)
		watcher, err := discovery.NewWatcher(cfg, guidanceStore, *configPath)
		if err != nil {
			log.Fatalf("Failed to create discovery watcher: %v", err)
		}
		err = watcher.InitialScan()
		if err != nil {
			// InitialScan logs fatal error internally on duplicate ID
			log.Fatalf("Error during initial scan of guidance files: %v", err)
		}

		// --- Start Watcher Goroutine ---
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Ensure cancellation propagates
		go watcher.Start(ctx)

		// --- Setup HTTP Server ---
		srv := server.NewServer(cfg, guidanceStore)

		// --- Start HTTP Server (Blocking) ---
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}

		log.Println("Server shutting down gracefully.")
		// Cleanup watcher (optional, depends on desired shutdown behavior)
		if err := watcher.Close(); err != nil {
			log.Printf("Error closing watcher: %v", err)
		}
	*/
}
