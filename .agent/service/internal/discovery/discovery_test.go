package discovery_test

import (
	"agentt/internal/config"
	"agentt/internal/discovery"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper function to create a mock EntityTypeDefinition
func mockBehaviorDef() config.EntityTypeDefinition {
	return config.EntityTypeDefinition{
		Name:                "behavior",
		PathGlob:            "*.bhv",
		RequiredFrontMatter: []string{"title", "priority", "description", "tags"},
		FileExtensionHint:   ".bhv",
	}
}

func mockRecipeDef() config.EntityTypeDefinition {
	return config.EntityTypeDefinition{
		Name:                "recipe",
		PathGlob:            "*.rcp",
		RequiredFrontMatter: []string{"id", "title", "priority", "description", "tags"},
		FileExtensionHint:   ".rcp",
	}
}

func TestParseFile_ValidBehavior(t *testing.T) {
	filePath := filepath.Join("testdata", "valid_behavior.bhv")
	entityDef := mockBehaviorDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		t.Fatalf("ParseFile failed for valid behavior: %v", err)
	}
	if item == nil {
		t.Fatal("ParseFile returned nil item for valid behavior")
	}

	if !item.IsValid {
		t.Errorf("Expected IsValid to be true, got false. Errors: %v", item.ValidationErrors)
	}
	if item.EntityType != "behavior" {
		t.Errorf("Expected EntityType 'behavior', got '%s'", item.EntityType)
	}
	if !strings.HasSuffix(item.SourcePath, "testdata/valid_behavior.bhv") {
		t.Errorf("Unexpected SourcePath: %s", item.SourcePath)
	}
	if item.Body != "This is the body of the behavior." {
		t.Errorf("Expected Body 'This is the body of the behavior.', got '%s'", item.Body)
	}

	expectedFM := map[string]interface{}{
		"title":        "Valid Behavior",
		"priority":     1,
		"description":  "A test behavior file.",
		"tags":         []interface{}{"test", "core"},
		"custom_field": "some value",
	}
	if !reflect.DeepEqual(item.FrontMatter, expectedFM) {
		t.Errorf("FrontMatter mismatch:\nExpected: %v\nGot:      %v", expectedFM, item.FrontMatter)
	}

	// Test Tier inference (adjust path if needed for test setup)
	// Assuming the test runs where the parent dir isn't must/should
	if item.Tier != "" {
		t.Errorf("Expected Tier to be empty for testdata path, got '%s'", item.Tier)
	}
}

func TestParseFile_ValidRecipe(t *testing.T) {
	filePath := filepath.Join("testdata", "valid_recipe.rcp")
	entityDef := mockRecipeDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		t.Fatalf("ParseFile failed for valid recipe: %v", err)
	}
	if item == nil {
		t.Fatal("ParseFile returned nil item for valid recipe")
	}

	if !item.IsValid {
		t.Errorf("Expected IsValid to be true, got false. Errors: %v", item.ValidationErrors)
	}
	if item.EntityType != "recipe" {
		t.Errorf("Expected EntityType 'recipe', got '%s'", item.EntityType)
	}
	if item.Body != "Recipe body here." {
		t.Errorf("Expected Body 'Recipe body here.', got '%s'", item.Body)
	}

	expectedFM := map[string]interface{}{
		"id":          "valid-recipe",
		"title":       "Valid Recipe",
		"priority":    10,
		"description": "A test recipe file.",
		"tags":        []interface{}{"test", "recipe"},
	}
	if !reflect.DeepEqual(item.FrontMatter, expectedFM) {
		t.Errorf("FrontMatter mismatch:\nExpected: %v\nGot:      %v", expectedFM, item.FrontMatter)
	}
}

func TestParseFile_Invalid_MissingRequired(t *testing.T) {
	filePath := filepath.Join("testdata", "invalid_behavior_missing_priority.bhv")
	entityDef := mockBehaviorDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		t.Fatalf("ParseFile failed unexpectedly: %v", err)
	}
	if item == nil {
		t.Fatal("ParseFile returned nil item")
	}

	if item.IsValid {
		t.Error("Expected IsValid to be false due to missing priority, got true")
	}
	if len(item.ValidationErrors) == 0 {
		t.Error("Expected validation errors for missing priority, got none")
	} else if !strings.Contains(item.ValidationErrors[0], "Missing required frontmatter key: 'priority'") {
		t.Errorf("Expected error about missing 'priority', got: %v", item.ValidationErrors)
	}
}

func TestParseFile_Invalid_YamlSyntax(t *testing.T) {
	filePath := filepath.Join("testdata", "invalid_yaml.bhv")
	entityDef := mockBehaviorDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		t.Fatalf("ParseFile failed unexpectedly: %v", err)
	}
	if item == nil {
		t.Fatal("ParseFile returned nil item")
	}

	if item.IsValid {
		t.Error("Expected IsValid to be false due to invalid YAML, got true")
	}
	if len(item.ValidationErrors) == 0 {
		t.Error("Expected validation errors for invalid YAML, got none")
	} else if !strings.Contains(item.ValidationErrors[0], "YAML parsing error") {
		t.Errorf("Expected error about YAML parsing, got: %v", item.ValidationErrors)
	}
}

func TestParseFile_Invalid_NoFrontmatter(t *testing.T) {
	filePath := filepath.Join("testdata", "no_frontmatter.bhv")
	entityDef := mockBehaviorDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		t.Fatalf("ParseFile failed unexpectedly: %v", err)
	}
	if item == nil {
		t.Fatal("ParseFile returned nil item")
	}

	if item.IsValid {
		t.Error("Expected IsValid to be false due to missing frontmatter, got true")
	}
	if len(item.ValidationErrors) == 0 {
		t.Error("Expected validation errors for missing frontmatter, got none")
	} else if !strings.Contains(item.ValidationErrors[0], "No valid YAML frontmatter detected") {
		t.Errorf("Expected error about missing frontmatter, got: %v", item.ValidationErrors)
	}
	if item.Body != "Just body content, no frontmatter." { // Ensure body is still captured
		t.Errorf("Expected body to be captured even with no frontmatter, got '%s'", item.Body)
	}
}

func TestParseFile_NotFound(t *testing.T) {
	filePath := filepath.Join("testdata", "does_not_exist.bhv")
	entityDef := mockBehaviorDef()

	item, err := discovery.ParseFile(filePath, entityDef)
	if err != nil {
		// Depending on how ParseFile handles file not found, adjust check
		// If it returns nil, nil -> check item == nil
		// If it returns an error -> check err != nil
		// Assuming it returns nil, nil based on parseFile logic
		t.Fatalf("ParseFile failed unexpectedly for non-existent file: %v", err)
	}
	if item != nil {
		t.Fatal("Expected nil item for non-existent file, got an item")
	}
	// If ParseFile is designed to return the error, check:
	// if err == nil {
	// 	 t.Fatal("Expected file not found error, got nil")
	// }
}

// TODO: Add tests for Watcher InitialScan and event handling
func TestWatcher(t *testing.T) {
	t.Skip("Test not implemented")
}
