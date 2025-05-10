package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Constants for configuration file discovery
const (
	DefaultConfigFileName = "config.yaml"
	ConfigDirName         = ".agent/service"
)

// LoadConfig reads and parses the configuration file from the specified path.
func LoadConfig(configPath string) (*Config, error) {
	slog.Debug("Attempting to load configuration", "path", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read configuration file", "path", configPath, "error", err)
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		slog.Error("Failed to parse configuration YAML", "path", configPath, "error", err)
		return nil, fmt.Errorf("failed to parse config YAML %s: %w", configPath, err)
	}

	// Store the path from which the config was loaded (absolute path recommended)
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		slog.Warn("Failed to get absolute path for config file, storing as is", "path", configPath, "error", err)
		absPath = configPath // Fallback to original path
	}
	cfg.LoadedFromPath = absPath

	// Perform general validation
	if err := validateConfig(&cfg); err != nil {
		slog.Error("Configuration validation failed", "path", configPath, "error", err)
		return nil, fmt.Errorf("configuration validation failed for %s: %w", configPath, err)
	}

	slog.Info("Configuration loaded successfully", "path", configPath)
	return &cfg, nil
}

// getUserConfigDir returns the path to the user-specific configuration directory.
// It checks for XDG_CONFIG_HOME first, then defaults to ~/.config/agentt or ~/.agentt.
func getUserConfigDir() (string, error) {
	var configHome string
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configHome = filepath.Join(xdgConfigHome, "agentt")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		// Prefer .config/agentt, but also check .agentt for backward compatibility or user preference
		configHome = filepath.Join(home, ".config", "agentt")
		// As a fallback, consider ~/.agentt directly if ~/.config/agentt doesn't exist or isn't standard.
		// For simplicity in this function, we'll just return the primary path.
		// The caller (FindAndLoadConfig) can attempt multiple user-specific paths if needed.
	}
	// Ensure the directory exists, or create it.
	// For FindAndLoad, we only need to know the path, not create it.
	// if err := os.MkdirAll(configHome, 0750); err != nil {
	// 	return "", fmt.Errorf("failed to create user config directory %s: %w", configHome, err)
	// }
	return configHome, nil
}

// FindAndLoadConfig searches for and loads the configuration file based on priority:
//  1. Path specified by configPathFromFlag.
//  2. Path specified by AGENTT_CONFIG environment variable.
//  3. Default locations:
//     a. ./agentt.yaml (or DefaultConfigFileName in current dir)
//     b. <UserConfigDir>/config.yaml (e.g., ~/.config/agentt/config.yaml)
//     c. /etc/agentt/config.yaml
//     d. Original search: ./.agent/service/config.yaml and its parents.
//
// It returns the loaded Config, the path it was loaded from, and an error if any occurs.
// If a specific path is provided (flag or env) and it doesn't exist or is invalid, it's a hard error.
func FindAndLoadConfig(configPathFromFlag string) (*Config, string, error) {
	slog.Debug("Starting configuration search", "flagPath", configPathFromFlag)

	// 1. Try configPathFromFlag
	if configPathFromFlag != "" {
		slog.Info("Attempting to load configuration from flag-specified path", "path", configPathFromFlag)
		cfg, err := LoadConfig(configPathFromFlag)
		if err != nil {
			// If the file explicitly specified by flag is not found, or any other error.
			return nil, "", fmt.Errorf("failed to load configuration from %s (specified by flag): %w", configPathFromFlag, err)
		}
		return cfg, configPathFromFlag, nil
	}

	// 2. Try AGENTT_CONFIG environment variable
	envPath := os.Getenv("AGENTT_CONFIG")
	if envPath != "" {
		slog.Info("Attempting to load configuration from AGENTT_CONFIG environment variable", "path", envPath)
		cfg, err := LoadConfig(envPath)
		if err != nil {
			// If the file explicitly specified by env var is not found, or any other error.
			return nil, "", fmt.Errorf("failed to load configuration from %s (specified by AGENTT_CONFIG): %w", envPath, err)
		}
		return cfg, envPath, nil
	}

	slog.Debug("No specific config path from flag or env var. Searching default locations.")
	// 3. Try default locations
	defaultPaths := []string{}

	// 3a. ./agentt.yaml (or DefaultConfigFileName)
	defaultPaths = append(defaultPaths, DefaultConfigFileName) // e.g., "config.yaml" in current dir
	// Also try "agentt.yaml" explicitly in current dir, as it's a common convention
	defaultPaths = append(defaultPaths, "agentt.yaml")

	// 3b. User config directory
	userConfigDirPath, err := getUserConfigDir()
	if err == nil { // If we successfully got a user config dir path
		defaultPaths = append(defaultPaths, filepath.Join(userConfigDirPath, DefaultConfigFileName))
		// Also try ~/.agentt/config.yaml as a common alternative
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			defaultPaths = append(defaultPaths, filepath.Join(homeDir, ".agentt", DefaultConfigFileName))
		}
	} else {
		slog.Warn("Could not determine user config directory", "error", err)
	}

	// 3c. /etc/agentt/config.yaml
	defaultPaths = append(defaultPaths, filepath.Join("/etc", "agentt", DefaultConfigFileName))

	for _, path := range defaultPaths {
		slog.Debug("Checking default location for config file", "path", path)
		if _, err := os.Stat(path); err == nil {
			slog.Info("Found configuration file in a default location", "path", path)
			cfg, loadErr := LoadConfig(path)
			if loadErr != nil {
				slog.Warn("Found config file but failed to load/parse", "path", path, "error", loadErr)
				return nil, path, fmt.Errorf("found configuration at %s but failed to load: %w", path, loadErr)
			}
			return cfg, path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("Error checking for config file at default location (not ErrNotExist)", "path", path, "error", err)
		}
	}

	// 3d. Original search logic: ./.agent/service/config.yaml and parents
	// This was the original behavior if nothing else was found.
	// It might be too specific or overlap, consider if this is still desired after the above.
	// For now, retaining it as the last resort of the "default" searches.
	slog.Debug("Checking legacy default location: .agent/service/config.yaml in current/parent directories")
	searchLegacyPaths := []string{"."}
	wd, err := os.Getwd()
	if err == nil {
		currentWd := wd
		for i := 0; i < 3; i++ { // Check up to 3 parent levels
			parent := filepath.Dir(currentWd)
			if parent == currentWd { // Reached root or error
				break
			}
			searchLegacyPaths = append(searchLegacyPaths, parent)
			currentWd = parent
		}
	}

	for _, dir := range searchLegacyPaths {
		legacyPath := filepath.Join(dir, ConfigDirName, DefaultConfigFileName) // e.g. parent/.agent/service/config.yaml
		slog.Debug("Checking legacy default location for config file", "path", legacyPath)
		if _, err := os.Stat(legacyPath); err == nil {
			slog.Info("Found configuration file in a legacy default location", "path", legacyPath)
			cfg, loadErr := LoadConfig(legacyPath)
			if loadErr != nil {
				slog.Warn("Found legacy config file but failed to load/parse", "path", legacyPath, "error", loadErr)
				return nil, legacyPath, fmt.Errorf("found legacy configuration at %s but failed to load: %w", legacyPath, loadErr)
			}
			return cfg, legacyPath, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("Error checking for config file at legacy default location (not ErrNotExist)", "path", legacyPath, "error", err)
		}
	}

	slog.Error("Configuration file not found after checking all specified and default locations.")
	return nil, "", errors.New("configuration file not found")
}

// validateConfig performs basic structural validation on the loaded configuration.
// Backend-specific validation happens later during backend initialization.
func validateConfig(cfg *Config) error {
	if len(cfg.EntityTypes) == 0 {
		return errors.New("config validation error: 'entityTypes' field is required and cannot be empty")
	}
	if len(cfg.Backends) == 0 {
		return errors.New("config validation error: 'backends' field is required and cannot be empty")
	}

	entityTypeNames := make(map[string]bool)
	for i, et := range cfg.EntityTypes {
		if et.Name == "" {
			return fmt.Errorf("config validation error: entityTypes[%d]: 'name' field is required", i)
		}
		if entityTypeNames[et.Name] {
			return fmt.Errorf("config validation error: duplicate entity type name '%s' found", et.Name)
		}
		entityTypeNames[et.Name] = true

		if len(et.RequiredFields) == 0 {
			// Allow empty requiredFields for now, maybe enforce later if needed.
			// return fmt.Errorf("config validation error: entityTypes[%d] (name='%s'): 'requiredFields' cannot be empty", i, et.Name)
		} else {
			// Ensure 'id' is always a required field if requiredFields is not empty
			if !contains(et.RequiredFields, "id") {
				return fmt.Errorf("config validation error: entityTypes[%d] (name='%s'): 'requiredFields' must include 'id'", i, et.Name)
			}
		}
	}

	for i, be := range cfg.Backends {
		if be.Type == "" {
			return fmt.Errorf("config validation error: backends[%d]: 'type' field is required", i)
		}
		// Note: Backend-specific validation (e.g., checking 'rootDir' for 'localfs')
		// is deferred to the backend instantiation phase (Phase 2 of the plan).
	}

	return nil
}

// Helper to check slice contains string (assuming it's defined in types.go or here)
// func contains(slice []string, str string) bool { ... }
