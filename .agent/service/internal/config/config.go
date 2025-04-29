package config

import (
	"fmt"
	"os"

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
	// LLMGuidanceFile is the path to the text file served at /llm.txt.
	LLMGuidanceFile string `yaml:"llmGuidanceFile"`
}

const defaultLLMGuidancePath = ".agent/service/llm_guidance.txt"

// LoadConfig loads the service configuration from a YAML file.
func LoadConfig(configPath string) (*ServiceConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("configuration path cannot be empty")
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
	if cfg.LLMGuidanceFile == "" {
		cfg.LLMGuidanceFile = defaultLLMGuidancePath
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
		// Ensure absolute path for glob pattern relative to config file dir? Or assume relative to CWD? Assuming CWD for now.
		// If we assume relative to config dir:
		// configDir := filepath.Dir(configPath)
		// cfg.EntityTypes[i].PathGlob = filepath.Join(configDir, cfg.EntityTypes[i].PathGlob)

		if entityNames[cfg.EntityTypes[i].Name] {
			return nil, fmt.Errorf("duplicate entity type name detected: '%s'", cfg.EntityTypes[i].Name)
		}
		entityNames[cfg.EntityTypes[i].Name] = true
	}

	return &cfg, nil
}
