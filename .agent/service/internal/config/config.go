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

// FindAndLoadConfig searches for the configuration file in standard locations
// (current dir, .agent/service/, user config dir) and loads it.
// It returns the loaded configuration or an error if not found or invalid.
func FindAndLoadConfig() (*Config, error) {
	searchPaths := []string{"."} // Start with current directory
	wd, err := os.Getwd()
	if err == nil {
		for i := 0; i < 3; i++ { // Check up to 3 parent levels
			parent := filepath.Dir(wd)
			if parent == wd { // Reached root or error
				break
			}
			searchPaths = append(searchPaths, parent)
			wd = parent
		}
	}

	for _, dir := range searchPaths {
		configPath := filepath.Join(dir, ConfigDirName, DefaultConfigFileName)
		slog.Debug("Checking for config file", "path", configPath)
		if _, err := os.Stat(configPath); err == nil {
			slog.Info("Found configuration file", "path", configPath)
			// Found the file, now load it using the specific path loader
			return LoadConfig(configPath)
		} else if !errors.Is(err, os.ErrNotExist) {
			// Log error if it's something other than file not found
			slog.Warn("Error checking for config file", "path", configPath, "error", err)
		}
	}

	slog.Error("Configuration file not found in standard locations")
	return nil, errors.New("configuration file (.agent/service/config.yaml) not found in current directory or parent directories")
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
