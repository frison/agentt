package config_test

import (
	"agentt/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a temporary config file
func createTempConfig(t *testing.T, content string, dir string, filename string) string {
	t.Helper()
	tempDir := dir
	if tempDir == "" {
		tempDir = t.TempDir()
	}
	filePath := filepath.Join(tempDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file %s: %v", filePath, err)
	}
	return filePath
}

func TestFindAndLoadConfig_ExplicitPath(t *testing.T) {
	content := `
entityTypes:
  - name: behavior
    pathGlob: "./behaviors/*.bhv"
listenAddress: ":8081"
backend:
  type: localfs
  rootDir: "."
`
	configPath := createTempConfig(t, content, "", "test-config.yaml")

	cfg, loadedPath, err := config.FindAndLoadConfig(configPath)
	if err != nil {
		t.Fatalf("FindAndLoadConfig failed with explicit path: %v", err)
	}
	if loadedPath != configPath {
		t.Errorf("Expected loaded path '%s', got '%s'", configPath, loadedPath)
	}
	if len(cfg.EntityTypes) != 1 || cfg.EntityTypes[0].PathGlob != "./behaviors/*.bhv" {
		t.Errorf("Expected EntityTypes[0].PathGlob './behaviors/*.bhv', got '%+v'", cfg.EntityTypes)
	}
	if cfg.ListenAddress != ":8081" {
		t.Errorf("Expected ListenAddress ':8081', got '%s'", cfg.ListenAddress)
	}
}

func TestFindAndLoadConfig_EnvVar(t *testing.T) {
	content := `
entityTypes: [{name: recipe, pathGlob: "env-recipes/*.rcp"}]
backend:
  type: localfs
  rootDir: "."
`
	configPath := createTempConfig(t, content, "", "env-config.yaml")

	t.Setenv("AGENTT_CONFIG", configPath)

	cfg, loadedPath, err := config.FindAndLoadConfig("") // No explicit path
	if err != nil {
		t.Fatalf("FindAndLoadConfig failed with env var: %v", err)
	}
	if loadedPath != configPath {
		t.Errorf("Expected loaded path '%s', got '%s'", configPath, loadedPath)
	}
	if len(cfg.EntityTypes) != 1 || cfg.EntityTypes[0].PathGlob != "env-recipes/*.rcp" {
		t.Errorf("Expected EntityTypes[0].PathGlob 'env-recipes/*.rcp', got '%+v'", cfg.EntityTypes)
	}
}

func TestFindAndLoadConfig_NotFound(t *testing.T) {
	// Change to a temporary directory where no config files exist
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	baseDir := t.TempDir()
	err = os.Chdir(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chdir(originalWD)
		t.Setenv("AGENTT_CONFIG", "") // Clean up env var
	})

	// Ensure no flag, no env var, and no default files exist IN the temp dir
	// by pointing the env var to a specific non-existent path
	nonExistentPath := "/path/to/surely/non/existent/config/file.yaml"
	t.Setenv("AGENTT_CONFIG", nonExistentPath)

	_, _, err = config.FindAndLoadConfig("") // No explicit path, will use env var
	if err == nil {
		t.Fatalf("Expected an error when no config file found, got nil")
	}
	// Check if the error message contains the expected substring for env-var-not-found
	expectedSubString := "config file specified by environment variable"
	if !strings.Contains(err.Error(), expectedSubString) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedSubString, err)
	}
}

func TestFindAndLoadConfig_InvalidYAML(t *testing.T) {
	content := `entityTypes: [{name: valid}]
listenAddress: { invalid ` // Invalid YAML
	configPath := createTempConfig(t, content, "", "invalid.yaml")
	// Explicitly use the invalid path via flag/env var to bypass default search
	t.Setenv("AGENTT_CONFIG", configPath)
	t.Cleanup(func() { t.Setenv("AGENTT_CONFIG", "") })

	_, _, err := config.FindAndLoadConfig("") // Use env var
	if err == nil {
		t.Fatalf("Expected an error for invalid YAML, got nil")
	}
	// Check if the error message contains the underlying yaml parsing error signature
	expectedSubString := "yaml:"
	if !strings.Contains(err.Error(), expectedSubString) {
		t.Errorf("Expected error to contain yaml parsing error ('%s'), got: %v", expectedSubString, err)
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	testFilePath := "test_valid_config.yaml"
	cfg, err := config.LoadConfig(testFilePath)
	if err != nil {
		t.Fatalf("LoadConfig failed for valid config %s: %v", testFilePath, err)
	}

	if cfg.ListenAddress != ":9090" {
		t.Errorf("Expected ListenAddress :9090, got %s", cfg.ListenAddress)
	}
	if len(cfg.EntityTypes) != 1 {
		t.Fatalf("Expected 1 entity type, got %d", len(cfg.EntityTypes))
	}

	et := cfg.EntityTypes[0]
	if et.Name != "test_entity" {
		t.Errorf("Expected entity name test_entity, got %s", et.Name)
	}
	if et.PathGlob != "/tmp/test_*.entity" {
		t.Errorf("Expected path glob /tmp/test_*.entity, got %s", et.PathGlob)
	}
	if len(et.RequiredFrontMatter) != 2 || et.RequiredFrontMatter[0] != "id" || et.RequiredFrontMatter[1] != "tags" {
		t.Errorf("Expected required frontmatter [id, tags], got %v", et.RequiredFrontMatter)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Create a temp file with minimal config to test defaults
	content := `
entityTypes:
  - name: "minimal"
    pathGlob: "*.minimal"
    requiredFrontMatter: ["title"]
backend:
  type: localfs
  rootDir: "."
`
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "minimal_config.yaml")
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write minimal config file: %v", err)
	}

	cfg, err := config.LoadConfig(testFilePath)
	if err != nil {
		t.Fatalf("LoadConfig failed for minimal config: %v", err)
	}

	// Check defaults
	if cfg.ListenAddress != ":8080" {
		t.Errorf("Expected default ListenAddress :8080, got %s", cfg.ListenAddress)
	}
}

func TestLoadConfig_Invalid_NotFound(t *testing.T) {
	_, err := config.LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent config file, got nil")
	}
}

func TestLoadConfig_Invalid_EmptyPath(t *testing.T) {
	_, err := config.LoadConfig("")
	if err == nil {
		t.Fatal("Expected error for empty config path, got nil")
	}
}

func TestLoadConfig_Invalid_Syntax(t *testing.T) {
	// Create a temp file with invalid YAML
	content := `listenAddress: :8080
entityTypes: [
  name: invalid
`
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "invalid_syntax.yaml")
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid syntax config file: %v", err)
	}

	_, err = config.LoadConfig(testFilePath)
	if err == nil {
		t.Fatalf("Expected YAML parsing error for %s, got nil", testFilePath)
	}
}

func TestLoadConfig_Invalid_MissingEntityName(t *testing.T) {
	testFilePath := "test_invalid_config_missing_name.yaml"
	_, err := config.LoadConfig(testFilePath)
	if err == nil {
		t.Fatalf("Expected error for missing entity name in %s, got nil", testFilePath)
	}
	t.Logf("Got expected error for missing name: %v", err)
}

func TestLoadConfig_Invalid_DuplicateEntityName(t *testing.T) {
	testFilePath := "test_invalid_config_duplicate_name.yaml"
	_, err := config.LoadConfig(testFilePath)
	if err == nil {
		t.Fatalf("Expected error for duplicate entity name in %s, got nil", testFilePath)
	}
	t.Logf("Got expected error for duplicate name: %v", err)
}

func TestLoadConfig_Invalid_MissingPathGlob(t *testing.T) {
	content := `
entityTypes:
  - name: "no_glob"
    requiredFrontMatter: ["title"]
`
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "missing_glob.yaml")
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write missing glob config file: %v", err)
	}
	_, err = config.LoadConfig(testFilePath)
	if err == nil {
		t.Fatalf("Expected error for missing pathGlob in %s, got nil", testFilePath)
	}
	t.Logf("Got expected error for missing pathGlob: %v", err)
}

func TestLoadConfig_Invalid_NoEntityTypes(t *testing.T) {
	content := `listenAddress: ":8080"`
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "no_entities.yaml")
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write no entities config file: %v", err)
	}
	_, err = config.LoadConfig(testFilePath)
	if err == nil {
		t.Fatalf("Expected error for no entityTypes defined in %s, got nil", testFilePath)
	}
	t.Logf("Got expected error for no entityTypes: %v", err)
}
