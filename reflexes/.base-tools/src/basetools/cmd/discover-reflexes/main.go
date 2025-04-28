package main

import (
	"encoding/json"
	"fmt"
	"nhi/basetools/pkg/discoverytypes" // Import our new type
	"nhi/basetools/pkg/manifesttypes" // Import existing manifest types
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	reflexRoot      = "reflexes" // Base directory to search within
	manifestFileName = "manifest.yml"
	skipDir         = "bin"      // Directory to skip within reflexes/
)

func main() {
	reflexes, err := findAndParseReflexes(reflexRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering reflexes: %v\n", err)
		os.Exit(1)
	}

	outputJSON, err := json.MarshalIndent(reflexes, "", "  ") // Use indent for readability
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(outputJSON))
}

// findAndParseReflexes walks the root directory, finds manifest.yml files,
// parses them, and returns a slice of DiscoveredReflex structs.
func findAndParseReflexes(rootDir string) ([]discoverytypes.DiscoveredReflex, error) {
	var discovered []discoverytypes.DiscoveredReflex

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %w", path, err)
		}

		// Skip the root directory itself
		if path == rootDir {
			return nil
		}

		// Skip the designated skip directory (e.g., "bin") directly under rootDir
		if info.IsDir() && filepath.Base(path) == skipDir && filepath.Dir(path) == rootDir {
			fmt.Fprintf(os.Stderr, "DEBUG: Skipping directory: %s\n", path)
			return filepath.SkipDir
		}

		// Process only manifest files
		if !info.IsDir() && info.Name() == manifestFileName {
			fmt.Fprintf(os.Stderr, "DEBUG: Found manifest: %s\n", path)
			reflex, err := parseManifest(path, rootDir)
			if err != nil {
				// Log error but continue walking to find other manifests
				fmt.Fprintf(os.Stderr, "Error parsing manifest %s: %v\n", path, err)
				return nil // Continue walking
			}
			discovered = append(discovered, reflex)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %w", rootDir, err)
	}

	return discovered, nil
}

// parseManifest reads and parses a single manifest.yml file.
func parseManifest(manifestPath string, rootDir string) (discoverytypes.DiscoveredReflex, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return discoverytypes.DiscoveredReflex{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest manifesttypes.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return discoverytypes.DiscoveredReflex{}, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Calculate relative path
	relPath, err := filepath.Rel(rootDir, filepath.Dir(manifestPath))
	if err != nil {
		// Should generally not happen if manifestPath is within rootDir
		relPath = filepath.Dir(manifestPath)
		fmt.Fprintf(os.Stderr, "Warning: Could not calculate relative path for %s: %v\n", manifestPath, err)
	}

	// Convert input/output paths to map[string]interface{} for flexibility in JSON
	inputsMap := make(map[string]interface{})
	for k, v := range manifest.InputPaths {
		inputsMap[k] = v
	}
	outputsMap := make(map[string]interface{})
	for k, v := range manifest.OutputPaths {
		outputsMap[k] = v
	}

	discovered := discoverytypes.DiscoveredReflex{
		Path:        relPath,
		Name:        manifest.Name,
		Description: manifest.Description,
		Inputs:      inputsMap,
		Outputs:     outputsMap,
	}

	// Handle cases where maps might be empty after conversion
	if len(discovered.Inputs) == 0 {
		discovered.Inputs = nil // Ensure it marshals as omitted/null, not {}
	}
	if len(discovered.Outputs) == 0 {
		discovered.Outputs = nil // Ensure it marshals as omitted/null, not {}
	}

	return discovered, nil
}