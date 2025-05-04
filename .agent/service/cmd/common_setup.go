package cmd

import (
	"agentt/internal/config"
	"agentt/internal/guidance/backend"
	"agentt/internal/guidance/backend/localfs"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// globalBackendService holds the initialized backend(s) for use by commands.
// Changed to a slice to support multiple backends.
var globalBackendService []backend.GuidanceBackend

// var loadedConfig *config.Config // REMOVED: Store loaded config for access if needed

// setupLogging configures the slog logger based on verbosity flags.
func setupLogging(verbose, quiet bool) {
	logLevel := slog.LevelWarn
	if quiet {
		logLevel = slog.LevelError
	} else if verbose {
		logLevel = slog.LevelInfo
		// Check for higher verbosity (e.g., -vv)
		if strings.Count(os.Args[0], "-v") > 1 || containsFlag(os.Args, "--debug") { // Simple check, might need refinement
			logLevel = slog.LevelDebug
		}
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))
	slog.Debug("Logging initialized", "level", logLevel.String())
}

// initializeBackend loads configuration and sets up the guidance backend(s).
// Now populates globalBackendService with a list of backends.
func initializeBackend() error {
	cfg, err := config.FindAndLoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	// loadedConfig = cfg // Store globally

	globalBackendService = make([]backend.GuidanceBackend, 0, len(cfg.Backends))

	for i, backendSpec := range cfg.Backends {
		slog.Info("Initializing backend", "index", i, "type", backendSpec.Type, "name", backendSpec.Name)
		var instance backend.GuidanceBackend
		var initErr error

		switch backendSpec.Type {
		case "localfs":
			localFSSettings, err := backendSpec.GetLocalFSSettings()
			if err != nil {
				initErr = fmt.Errorf("failed to parse settings for localfs backend %d (%s): %w", i, backendSpec.Name, err)
			} else {
				instance, initErr = localfs.NewLocalFSBackend(localFSSettings, cfg.LoadedFromPath, cfg.EntityTypes)
			}
		// case "database":
		// 	 // dbSettings, err := backendSpec.GetDatabaseSettings() ...
		// 	 // instance, initErr = database.NewDatabaseBackend(...) ...
		default:
			initErr = fmt.Errorf("unsupported backend type '%s' specified for backend %d (%s)", backendSpec.Type, i, backendSpec.Name)
		}

		if initErr != nil {
			// Decide whether to fail hard or just log and skip this backend
			slog.Error("Failed to initialize backend", "index", i, "type", backendSpec.Type, "name", backendSpec.Name, "error", initErr)
			// Option 1: Fail hard
			return fmt.Errorf("failed to initialize backend %d (%s): %w", i, backendSpec.Name, initErr)
			// Option 2: Log and continue (might lead to incomplete results)
			// continue
		}

		globalBackendService = append(globalBackendService, instance)
		slog.Info("Backend initialized successfully", "index", i, "type", backendSpec.Type, "name", backendSpec.Name)
	}

	if len(globalBackendService) == 0 {
		return errors.New("no guidance backends were successfully initialized")
	}

	return nil
}

// containsFlag is a simple helper to check for debug flags (could be improved)
func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
