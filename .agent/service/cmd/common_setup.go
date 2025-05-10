package cmd

import (
	"agentt/internal/config"
	guidanceBackend "agentt/internal/guidance/backend"
	"agentt/internal/guidance/backend/localfs"
	"agentt/internal/server"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Global variables or shared state for commands
var (
	initErr                 error
	globalMultiBackend      guidanceBackend.GuidanceBackend
	allInstantiatedBackends map[string]guidanceBackend.GuidanceBackend
	globalConfig            *config.Config
	globalBackendInstance   *guidanceBackend.MultiBackend
	initState               = &sync.Once{}
)

// doInitialSetupAndParse loads the configuration and instantiates all defined backends.
// It ensures this happens only once and stores the results globally.
// Verbosity count (from flags) is passed in to configure logging correctly first.
func doInitialSetupAndParse(verbosity int) error {
	slog.Debug("Attempting to perform initial setup and parse configuration.")

	// Setup logging (idempotent or careful with multiple calls if any)
	setupLogging(verbosity, quiet)

	var loadedPath string // To store the path from which config was loaded
	// Load configuration
	// The global rootConfigPath (from cmd/root.go) should be used here.
	globalConfig, loadedPath, initErr = config.FindAndLoadConfig(rootConfigPath) // Use rootConfigPath
	if initErr != nil {
		return fmt.Errorf("failed to find/load configuration: %w", initErr)
	}
	if globalConfig == nil {
		return errors.New("configuration loaded successfully but is nil")
	}
	globalConfig.LoadedFromPath = loadedPath // Store the loaded path

	slog.Info("Configuration loaded", "path", globalConfig.LoadedFromPath)

	// Initialize backends based on the loaded configuration
	currentInstantiatedBackends := make(map[string]guidanceBackend.GuidanceBackend)
	backendInstancesForMulti := make([]guidanceBackend.GuidanceBackend, 0, len(globalConfig.Backends))
	var backendErrs []string

	seenNames := make(map[string]bool)
	for _, backendSpec := range globalConfig.Backends {
		if backendSpec.Name == "" {
			slog.Error("Configuration error: Backend found without a name. Skipping.", "spec", backendSpec)
			backendErrs = append(backendErrs, "backend at index (unnamed) has no name defined")
			continue
		}
		if seenNames[backendSpec.Name] {
			return fmt.Errorf("duplicate backend name configured: '%s'. Backend names must be unique", backendSpec.Name)
		}
		seenNames[backendSpec.Name] = true
	}
	if len(backendErrs) > 0 {
		return fmt.Errorf("errors during backend name validation: %s", strings.Join(backendErrs, "; "))
	}

	for i, backendSpec := range globalConfig.Backends {
		if backendSpec.Name == "" {
			continue
		}
		slog.Debug("Processing backend spec", "index", i, "type", backendSpec.Type, "name", backendSpec.Name, "writable", backendSpec.Writable)
		var instance guidanceBackend.GuidanceBackend
		var currentErr error

		switch backendSpec.Type {
		case "localfs":
			localFSSettings, err := backendSpec.GetLocalFSSettings()
			if err != nil {
				currentErr = fmt.Errorf("failed to parse settings for localfs backend '%s': %w", backendSpec.Name, err)
			} else {
				instance, currentErr = localfs.NewLocalFSBackend(localFSSettings, globalConfig.LoadedFromPath, globalConfig.EntityTypes, backendSpec.Writable)
			}
		default:
			currentErr = fmt.Errorf("unsupported backend type '%s' specified for backend '%s'", backendSpec.Type, backendSpec.Name)
		}

		if currentErr != nil {
			slog.Error("Failed to initialize a backend instance", "name", backendSpec.Name, "type", backendSpec.Type, "error", currentErr)
			backendErrs = append(backendErrs, fmt.Sprintf("backend '%s': %v", backendSpec.Name, currentErr))
		} else if instance != nil {
			currentInstantiatedBackends[backendSpec.Name] = instance
			backendInstancesForMulti = append(backendInstancesForMulti, instance)
			slog.Info("Successfully initialized backend instance", "name", backendSpec.Name, "type", backendSpec.Type)
		}
	}

	allInstantiatedBackends = currentInstantiatedBackends

	if len(allInstantiatedBackends) == 0 {
		if len(backendErrs) > 0 {
			return fmt.Errorf("failed to initialize any guidance backends. Errors: %s", strings.Join(backendErrs, "; "))
		}
		return fmt.Errorf("no guidance backends were successfully initialized (and no specific errors reported, check config)")
	}

	if len(backendInstancesForMulti) > 0 {
		multiBE, err := guidanceBackend.NewMultiBackend(backendInstancesForMulti)
		if err != nil {
			return fmt.Errorf("failed to create aggregate MultiBackend: %w. Individual backends might have initialized: %v", err, allInstantiatedBackends)
		}
		globalMultiBackend = multiBE
		globalBackendInstance = multiBE
	} else {
		slog.Warn("No backend instances were available to form a MultiBackend, though some individual backends might exist in allInstantiatedBackends map.")
	}

	slog.Debug("Initial setup and backend parsing complete.")
	return nil
}

func initializeSharedState(verbosity int) error {
	initState.Do(func() {
		initErr = doInitialSetupAndParse(verbosity)
		if initErr != nil {
			slog.Error("Initialization of shared state failed", "error", initErr)
		}
	})
	return initErr
}

func GetMultiBackendAndConfig(verbosity int) (*guidanceBackend.MultiBackend, *config.Config, error) {
	initState.Do(func() {
		initErr = doInitialSetupAndParse(verbosity)
	})
	if initErr != nil {
		return nil, nil, initErr
	}
	if globalMultiBackend == nil && globalConfig != nil && len(globalConfig.Backends) > 0 {
		return nil, globalConfig, fmt.Errorf("MultiBackend is nil after initialization, check backend configurations or initialization errors")
	}
	return globalBackendInstance, globalConfig, nil
}

// GetNamedBackend retrieves a specifically named backend instance from the initialized map.
// It ensures shared state is initialized but does not set up logging itself (assumes GetMultiBackendAndConfig or another entry point did).
func GetNamedBackend(name string) (guidanceBackend.GuidanceBackend, error) {
	_ = initializeSharedState(0) // Ensure init; verbosity for logging setup would have been from an earlier call.
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize shared state: %w", initErr)
	}
	if allInstantiatedBackends == nil {
		return nil, fmt.Errorf("backend map not initialized, shared state error likely occurred")
	}

	instance, exists := allInstantiatedBackends[name]
	if !exists {
		return nil, fmt.Errorf("no backend configured with name '%s'", name)
	}
	return instance, nil
}

// ListWritableBackendNames returns a list of names for all configured and successfully instantiated writable backends.
func ListWritableBackendNames() ([]string, error) {
	_ = initializeSharedState(0) // Ensures init.
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize shared state: %w", initErr)
	}
	if globalConfig == nil || allInstantiatedBackends == nil {
		return nil, fmt.Errorf("config or backend map not initialized, shared state error likely occurred")
	}

	names := make([]string, 0)
	for _, spec := range globalConfig.Backends {
		if spec.Writable {
			// Check if this writable backend was actually instantiated successfully
			if _, exists := allInstantiatedBackends[spec.Name]; exists {
				names = append(names, spec.Name)
			}
		}
	}
	return names, nil
}

// GetDefaultWritableBackend returns a single WritableBackend if exactly one is configured and writable.
// Returns an error if zero or multiple writable backends are found.
func GetDefaultWritableBackend() (guidanceBackend.WritableBackend, error) {
	_ = initializeSharedState(0) // Ensures init.
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize shared state: %w", initErr)
	}
	if globalConfig == nil || allInstantiatedBackends == nil {
		return nil, fmt.Errorf("config or backend map not initialized, shared state error likely occurred")
	}

	writableBackends := make([]guidanceBackend.WritableBackend, 0)
	writableBackendNames := make([]string, 0)

	for backendName, instance := range allInstantiatedBackends {
		// Find the original spec to check its Writable flag
		var originalSpec *config.BackendSpec
		for i := range globalConfig.Backends {
			if globalConfig.Backends[i].Name == backendName {
				originalSpec = &globalConfig.Backends[i]
				break
			}
		}

		if originalSpec != nil && originalSpec.Writable {
			if wb, ok := instance.(guidanceBackend.WritableBackend); ok {
				writableBackends = append(writableBackends, wb)
				writableBackendNames = append(writableBackendNames, backendName)
			} else {
				slog.Warn("Instantiated backend is marked writable in config but does not implement WritableBackend interface", "name", backendName)
			}
		}
	}

	if len(writableBackends) == 0 {
		return nil, fmt.Errorf("no writable backends configured or initialized successfully")
	}
	if len(writableBackends) > 1 {
		return nil, fmt.Errorf("multiple writable backends configured (%s). Please specify one using --backend-target", strings.Join(writableBackendNames, ", "))
	}
	return writableBackends[0], nil
}

// GetNamedWritableBackend retrieves a specifically named backend and ensures it's writable.
func GetNamedWritableBackend(name string) (guidanceBackend.WritableBackend, error) {
	_ = initializeSharedState(0) // Ensure init.
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize shared state: %w", initErr)
	}
	if globalConfig == nil {
		return nil, fmt.Errorf("global config not initialized, shared state error likely occurred")
	}

	instance, err := GetNamedBackend(name) // This already checks if allInstantiatedBackends is nil
	if err != nil {
		return nil, err // Error from GetNamedBackend (e.g., not found)
	}

	// Find the original spec to check its Writable flag from config
	var originalSpec *config.BackendSpec
	for i := range globalConfig.Backends {
		if globalConfig.Backends[i].Name == name {
			originalSpec = &globalConfig.Backends[i]
			break
		}
	}

	if originalSpec == nil {
		// This case should theoretically be caught by GetNamedBackend if the backend wasn't even in allInstantiatedBackends
		// because it wouldn't be in globalConfig.Backends either, or its name wouldn't match.
		// However, as a safeguard if GetNamedBackend somehow found an instance not tied to a spec name (which is unlikely with current logic).
		return nil, fmt.Errorf("internal error: backend spec not found for existing instance '%s'", name)
	}

	if !originalSpec.Writable {
		return nil, fmt.Errorf("backend '%s' is configured but not marked as writable", name)
	}

	wb, ok := instance.(guidanceBackend.WritableBackend)
	if !ok {
		return nil, fmt.Errorf("backend '%s' is writable in config but does not implement WritableBackend interface", name)
	}

	return wb, nil
}

// setupLogging configures the global slog logger based on verbosity count.
func setupLogging(verbosity int, quietFlag bool) {
	var level slog.Level

	if quietFlag {
		level = slog.LevelError
	} else {
		switch verbosity {
		case 0:
			level = slog.LevelWarn
		case 1:
			level = slog.LevelInfo
		default:
			level = slog.LevelDebug
		}
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
