package cmd

import (
	"agentt/internal/config"
	"agentt/internal/guidance/backend"
	"agentt/internal/guidance/backend/localfs"
	"fmt"
	"log/slog"
	"path/filepath"
)

// setupResult holds the results of the common setup process.
// Using a struct allows returning multiple values cleanly.
type setupResult struct {
	Cfg        *config.ServiceConfig
	Backend    backend.GuidanceBackend
	ConfigPath string
}

// setupDiscovery performs the common steps needed by CLI commands:
// 1. Load configuration.
// 2. Initialize the appropriate guidance backend based on config.
// It returns the loaded config, the initialized backend, the config path, or an error.
func setupDiscovery(configPath string) (*setupResult, error) {
	// --- Configuration ---
	cfg, loadedPath, err := config.FindAndLoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}
	// Use slog.Info (level check is handled by the default logger config)
	slog.Info("Using configuration file", "path", loadedPath)

	// --- Backend Initialization ---
	var selectedBackend backend.GuidanceBackend
	var backendInitErr error

	backendType := cfg.Backend.Type
	slog.Info("Attempting to initialize guidance backend", "type", backendType)

	switch backendType {
	case "localfs":
		localBackend := localfs.NewLocalFilesystemBackend()

		// Prepare backend-specific config map for Initialize
		// Pass necessary info: RootDir (relative to config), EntityTypeDefs, etc.
		backendConfigMap := make(map[string]interface{})
		if cfg.Backend.Settings != nil {
			backendConfigMap = cfg.Backend.Settings // Start with settings from config.yaml
		}
		// Add required info derived from main config
		var absoluteRootDir string
		if filepath.IsAbs(cfg.Backend.RootDir) {
			absoluteRootDir = cfg.Backend.RootDir
			slog.Debug("Using absolute rootDir from config", "path", absoluteRootDir)
		} else {
			absoluteRootDir = filepath.Join(cfg.LoadedFromDir, cfg.Backend.RootDir)
			slog.Debug("Resolved relative rootDir", "configDir", cfg.LoadedFromDir, "relativeRootDir", cfg.Backend.RootDir, "absolutePath", absoluteRootDir)
		}
		backendConfigMap["rootDir"] = absoluteRootDir     // Use the calculated absolute path
		backendConfigMap["entityTypes"] = cfg.EntityTypes // Pass the definitions
		// Add other potential global settings if needed
		// backendConfigMap["requireExplicitID"] = true // Example

		backendInitErr = localBackend.Initialize(backendConfigMap)
		selectedBackend = localBackend

	default:
		backendInitErr = fmt.Errorf("unsupported backend type specified in configuration: '%s'", backendType)
	}

	if backendInitErr != nil {
		slog.Error("Failed to initialize guidance backend", "type", backendType, "error", backendInitErr)
		return nil, fmt.Errorf("failed to initialize guidance backend (type: %s): %w", backendType, backendInitErr)
	}
	slog.Info("Guidance backend initialized successfully", "type", backendType)

	return &setupResult{
		Cfg:        cfg,
		Backend:    selectedBackend,
		ConfigPath: loadedPath,
	}, nil
}
