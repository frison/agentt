package cmd

import (
	"agentt/internal/config"
	guidanceBackend "agentt/internal/guidance/backend"
	"agentt/internal/guidance/backend/localfs"
	"agentt/internal/server"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Global variables or shared state for commands
var (
	initErr  error
	initOnce sync.Once
	// Store the instantiated backend globally (use with caution)
	globalBackend guidanceBackend.GuidanceBackend
	// Store the loaded config globally after first load
	globalConfig *config.Config
)

// initializeBackend loads the configuration and instantiates the backend.
// It ensures this happens only once and stores the results globally.
// Verbosity count (from flags) is passed in to configure logging correctly.
func initializeBackend(verbosity int) (guidanceBackend.GuidanceBackend, *config.Config, error) {
	initOnce.Do(func() {
		cfg, err := config.FindAndLoadConfig()
		if err != nil {
			initErr = fmt.Errorf("failed to load configuration: %w", err)
			return
		}
		if cfg == nil {
			initErr = fmt.Errorf("configuration loaded as nil without error")
			return
		}
		globalConfig = cfg
		setupLogging(verbosity)

		initializedBackends := make([]guidanceBackend.GuidanceBackend, 0, len(cfg.Backends))
		var backendErrs []error

		for i, backendSpec := range cfg.Backends {
			slog.Debug("Processing backend spec", "index", i, "type", backendSpec.Type, "name", backendSpec.Name)
			var instance guidanceBackend.GuidanceBackend
			var currentErr error

			if backendSpec.Type == "localfs" {
				localFSSettings, err := backendSpec.GetLocalFSSettings()
				if err != nil {
					currentErr = fmt.Errorf("failed to parse settings for localfs backend %d (%s): %w", i, backendSpec.Name, err)
				} else {
					instance, currentErr = localfs.NewLocalFSBackend(localFSSettings, cfg.LoadedFromPath, cfg.EntityTypes)
				}
			} else {
				currentErr = fmt.Errorf("unsupported backend type '%s' specified for backend %d (%s)", backendSpec.Type, i, backendSpec.Name)
			}

			if currentErr != nil {
				slog.Error("Failed to initialize a backend instance", "index", i, "type", backendSpec.Type, "name", backendSpec.Name, "error", currentErr)
				backendErrs = append(backendErrs, currentErr)
			} else if instance != nil {
				initializedBackends = append(initializedBackends, instance)
				slog.Info("Successfully initialized backend instance", "index", i, "type", backendSpec.Type, "name", backendSpec.Name)
			}
		}

		if len(initializedBackends) == 0 {
			if len(backendErrs) > 0 {
				initErr = fmt.Errorf("failed to initialize any guidance backends; last error: %w", backendErrs[len(backendErrs)-1])
			} else {
				initErr = fmt.Errorf("no guidance backends were successfully initialized (and no specific errors reported)")
			}
			return
		}

		multiBE, err := guidanceBackend.NewMultiBackend(initializedBackends)
		if err != nil {
			initErr = fmt.Errorf("failed to create aggregate backend: %w", err)
			return
		}
		globalBackend = multiBE
	})

	return globalBackend, globalConfig, initErr
}

// GetBackend returns the initialized guidance backend and config, triggering initialization if needed.
// Verbosity count must be passed from the command flags.
// Returns nil interfaces and error if initialization fails.
func GetBackendAndConfig(verbosity int) (guidanceBackend.GuidanceBackend, *config.Config, error) {
	return initializeBackend(verbosity)
}

// setupLogging configures the global slog logger based on verbosity count.
func setupLogging(verbosity int) {
	var level slog.Level
	switch verbosity {
	case 0:
		level = slog.LevelWarn
	case 1:
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	h := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(h))
	slog.Debug("Logging configured", "level", level.String(), "verbosity_count", verbosity)
}

// setupAndRunServer starts the server using the provided config and backend instance.
func setupAndRunServer(cfg *config.Config, backendInstance guidanceBackend.GuidanceBackend) error {
	if cfg == nil {
		return fmt.Errorf("cannot start server: config is nil")
	}
	if backendInstance == nil {
		return fmt.Errorf("cannot start server: backend instance is nil")
	}
	s := server.NewServer(cfg, backendInstance)
	slog.Info("Attempting to start server...")
	return s.ListenAndServe()
}
