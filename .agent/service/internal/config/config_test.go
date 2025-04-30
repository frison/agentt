package config_test

import (
	"agentt/internal/config"
	"os"
	"path/filepath"
	"testing"
)

// TODO: Add tests for LoadConfig
func TestLoadConfig(t *testing.T) {
	t.Skip("Test not implemented")
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
