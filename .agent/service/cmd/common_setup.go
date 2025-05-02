package cmd

import (
	"agentt/internal/config"
	"agentt/internal/discovery"
	"agentt/internal/store"
	"fmt"
	"log/slog"
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
	// Use slog.Info (level check is handled by the default logger config)
	slog.Info("Using configuration file", "path", loadedPath)

	// --- Setup Dependencies & Load ---
	guidanceStore := store.NewGuidanceStore()
	watcher, err := discovery.NewWatcher(cfg, guidanceStore, loadedPath)
	if err != nil {
		// No need to close watcher here as it likely wasn't fully created
		return nil, fmt.Errorf("failed to create discovery watcher: %w", err)
	}
	// We are not running the watcher loop in CLI mode, so no need to defer Close()

	// --- Load Entities via Initial Scan ---
	// Use slog.Info
	slog.Info("Performing initial scan of guidance files...")

	err = watcher.InitialScan() // Populates the guidanceStore
	if err != nil {
		slog.Warn("During initial scan", "error", err) // Use slog.Warn
		return nil, fmt.Errorf("error during initial scan of guidance files: %w", err)
	}
	// Use slog.Info
	slog.Info("Initial scan complete.")

	return &setupResult{
		Cfg:        cfg,
		Store:      guidanceStore,
		ConfigPath: loadedPath,
	}, nil
}
