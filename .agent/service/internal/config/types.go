package config

import (
	"fmt"
)

// Config represents the top-level configuration structure.
type Config struct {
	ListenAddress string        `yaml:"listenAddress,omitempty"`
	EntityTypes   []EntityType  `yaml:"entityTypes"`
	Backends      []BackendSpec `yaml:"backends"`
	// Store the path from which this config was loaded
	LoadedFromPath string `yaml:"-"` // Ignored by YAML parser
}

// EntityType defines the structure and metadata expected for a type of entity.
type EntityType struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description,omitempty"`
	RequiredFields []string `yaml:"requiredFields"`
}

// BackendSpec defines the configuration for a single guidance backend.
// Using a map for settings allows flexibility for different backend types.
type BackendSpec struct {
	Name     string                 `yaml:"name,omitempty"`
	Type     string                 `yaml:"type"` // e.g., "localfs", "database"
	Settings map[string]interface{} `yaml:"settings"`
}

// LocalFSBackendSettings represents the specific settings for a 'localfs' backend.
// We'll extract these from BackendSpec.Settings when needed.
type LocalFSBackendSettings struct {
	RootDir         string            `yaml:"rootDir"`
	EntityLocations map[string]string `yaml:"entityLocations"`
}

// --- Helper functions for extracting typed settings ---

// GetLocalFSSettings attempts to parse the BackendSpec's Settings into LocalFSBackendSettings.
// Refactored to manually extract and type-check fields.
func (bs *BackendSpec) GetLocalFSSettings() (LocalFSBackendSettings, error) {
	var settings LocalFSBackendSettings

	// Extract rootDir
	if rootDirVal, ok := bs.Settings["rootDir"]; ok {
		if rootDirStr, ok := rootDirVal.(string); ok {
			settings.RootDir = rootDirStr
		} else {
			return settings, fmt.Errorf("failed to parse localfs settings: 'rootDir' must be a string, got %T", rootDirVal)
		}
	} // Missing rootDir is allowed here, will default later

	// Extract entityLocations
	if locVal, ok := bs.Settings["entityLocations"]; ok {
		if locMap, ok := locVal.(map[string]interface{}); ok { // YAML parser often uses map[string]interface{} for nested maps
			settings.EntityLocations = make(map[string]string)
			for k, v := range locMap {
				if vStr, okStr := v.(string); okStr {
					settings.EntityLocations[k] = vStr
				} else {
					return settings, fmt.Errorf("failed to parse localfs settings: value for entityLocations key '%s' must be a string, got %T", k, v)
				}
			}
		} else if locMapStr, ok := locVal.(map[string]string); ok { // Handle case where it's already map[string]string
			settings.EntityLocations = locMapStr
		} else {
			return settings, fmt.Errorf("failed to parse localfs settings: 'entityLocations' must be a map, got %T", locVal)
		}
	} // Missing entityLocations is allowed here, validated later

	return settings, nil
}

// Helper to check slice contains string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// Note: Need to import "fmt" for the helper function.
