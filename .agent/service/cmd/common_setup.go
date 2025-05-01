package cmd

import (
	"agentt/internal/config"
	"agentt/internal/discovery"
	"agentt/internal/store"
	"fmt"
	"log"
)

// setupResult holds the results of the common setup process.
// Using a struct allows returning multiple values cleanly.
type setupResult struct {
	Cfg        *config.ServiceConfig
	Store      *store.GuidanceStore
	ConfigPath string
}

// setupDiscovery performs the common steps needed by CLI commands:
// 1. Load configuration.
// 2. Initialize the guidance store.
// 3. Create the discovery watcher.
// 4. Perform the initial scan to populate the store.
// It returns the loaded config, the populated store, the path of the loaded config, or an error.
func setupDiscovery(configPath string) (*setupResult, error) {
	// --- Configuration ---
	cfg, loadedPath, err := config.FindAndLoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}
	log.Printf("Using configuration file: %s", loadedPath)

	// --- Setup Dependencies & Load ---
	guidanceStore := store.NewGuidanceStore()
	watcher, err := discovery.NewWatcher(cfg, guidanceStore, loadedPath)
	if err != nil {
		// No need to close watcher here as it likely wasn't fully created
		return nil, fmt.Errorf("failed to create discovery watcher: %w", err)
	}
	// We are not running the watcher loop in CLI mode, so no need to defer Close()

	// --- Load Entities via Initial Scan ---
	err = watcher.InitialScan() // Populates the guidanceStore
	if err != nil {
		log.Printf("Warning during initial scan: %v", err)
		return nil, fmt.Errorf("error during initial scan of guidance files: %w", err)
	}

	return &setupResult{
		Cfg:        cfg,
		Store:      guidanceStore,
		ConfigPath: loadedPath,
	}, nil
}
