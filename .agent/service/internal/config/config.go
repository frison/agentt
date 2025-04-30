package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EntityTypeDefinition describes a type of guidance entity the service should manage.
type EntityTypeDefinition struct {
	// Name is the unique identifier for this entity type (e.g., "behavior", "recipe"). Used in API paths.
	Name string `yaml:"name"`
	// PathGlob is the file path pattern used to discover files for this entity type.
	PathGlob string `yaml:"pathGlob"`
	// RequiredFrontMatter is a list of front matter keys that MUST be present for a file to be considered valid.
	RequiredFrontMatter []string `yaml:"requiredFrontMatter"`
	// Description provides a human-readable explanation of this entity type for API documentation.
	Description string `yaml:"description"`
	// FileExtensionHint is the expected file extension (e.g. ".bhv"), used for logging and potentially faster filtering. Optional.
	FileExtensionHint string `yaml:"fileExtensionHint"`
}

// ServiceConfig holds the overall configuration for the agent guidance service.
type ServiceConfig struct {
	// ListenAddress is the address and port the HTTP server listens on (e.g., ":8080").
	ListenAddress string `yaml:"listenAddress"`
	// EntityTypes defines the different kinds of guidance content the service will discover and serve.
	EntityTypes []EntityTypeDefinition `yaml:"entityTypes"`
}

const configEnvVar = "AGENTT_CONFIG"

// Default search paths relative to CWD
var defaultConfigSearchPaths = []string{
	"config.yaml",
	"agentt.yaml",
	".agent/service/config.yaml", // Keep this for compatibility during transition?
	".agentt/config.yaml",
}

// FindAndLoadConfig determines the config path based on flag, env var, or search paths,
// then loads and returns the configuration.
func FindAndLoadConfig(configFlagValue string) (*ServiceConfig, string, error) {
	configPath := ""

	// 1. Check Flag
	if configFlagValue != "" {
		configPath = configFlagValue
		// Verify flag path exists
		if _, err := os.Stat(configPath); err != nil {
			return nil, "", fmt.Errorf("config file specified by flag (--config) not found at '%s': %w", configPath, err)
		}
		// Path found via flag
	} else {
		// 2. Check Environment Variable
		envPath := os.Getenv(configEnvVar)
		if envPath != "" {
			configPath = envPath
			// Verify env var path exists
			if _, err := os.Stat(configPath); err != nil {
				return nil, "", fmt.Errorf("config file specified by environment variable (%s) not found at '%s': %w", configEnvVar, configPath, err)
			}
			// Path found via env var
		} else {
			// 3. Check Default Search Paths (relative to CWD)
			for _, searchPath := range defaultConfigSearchPaths {
				if _, err := os.Stat(searchPath); err == nil {
					configPath = searchPath
					break // Found first valid path
				}
			}
			// Path found via search (or is empty if none found)
		}
	}

	if configPath == "" {
		return nil, "", fmt.Errorf("configuration file not found (checked flag, env var '%s', and default paths relative to CWD: %v)", configEnvVar, defaultConfigSearchPaths)
	}

	// Ensure path is absolute before loading (LoadConfig assumes this)
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get absolute path for config file '%s': %w", configPath, err)
	}

	cfg, err := LoadConfig(absPath)
	if err != nil {
		return nil, absPath, err // Return error from LoadConfig
	}

	return cfg, absPath, nil // Return loaded config and the path that was used
}

// LoadConfig loads the service configuration from a YAML file.
// Assumes configPath is an absolute path.
func LoadConfig(configPath string) (*ServiceConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("internal error: LoadConfig called with empty path")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	var cfg ServiceConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", configPath, err)
	}

	// Basic validation and defaults
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = ":8080" // Default port
	}

	if len(cfg.EntityTypes) == 0 {
		return nil, fmt.Errorf("config file '%s' must define at least one entity type", configPath)
	}

	entityNames := make(map[string]bool)
	for i := range cfg.EntityTypes {
		if cfg.EntityTypes[i].Name == "" {
			return nil, fmt.Errorf("entity type definition at index %d is missing required 'name'", i)
		}
		if cfg.EntityTypes[i].PathGlob == "" {
			return nil, fmt.Errorf("entity type '%s' is missing required 'pathGlob'", cfg.EntityTypes[i].Name)
		}

		if entityNames[cfg.EntityTypes[i].Name] {
			return nil, fmt.Errorf("duplicate entity type name detected: '%s'", cfg.EntityTypes[i].Name)
		}
		entityNames[cfg.EntityTypes[i].Name] = true
	}

	return &cfg, nil
}
