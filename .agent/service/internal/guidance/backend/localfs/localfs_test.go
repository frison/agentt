package localfs

import (
	"agentt/internal/config"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Helper to create temp test files/dirs
func createTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	tempDir := t.TempDir()
	for relPath, content := range files {
		absPath := filepath.Join(tempDir, relPath)
		absDir := filepath.Dir(absPath)
		if err := os.MkdirAll(absDir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", absDir, err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", absPath, err)
		}
	}
	return tempDir
}

// Helper to create default entity type definitions for tests
func getDefaultEntityDefs() []config.EntityTypeDefinition {
	return []config.EntityTypeDefinition{
		{Name: "behavior", PathGlob: "behaviors/**/*.bhv", RequiredFrontMatter: []string{"id", "title"}},
		{Name: "recipe", PathGlob: "recipes/*.rcp", RequiredFrontMatter: []string{"id", "title"}},
	}
}

// --- Initialize Tests ---

func TestInitialize_Valid(t *testing.T) {
	files := map[string]string{
		"behaviors/must/b1.bhv": "---\nid: b1\ntitle: Behavior 1\n---\nBody B1",
		"recipes/r1.rcp":        "---\nid: r1\ntitle: Recipe 1\n---\nBody R1",
	}
	tempDir := createTestFiles(t, files)

	backend := NewLocalFilesystemBackend()
	configMap := map[string]interface{}{
		"rootDir":           tempDir,
		"entityTypes":       getDefaultEntityDefs(),
		"requireExplicitID": true,
	}

	err := backend.Initialize(configMap)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	summaries, _ := backend.GetSummary()
	if len(summaries) != 2 {
		t.Errorf("Expected 2 summaries, got %d", len(summaries))
	}

	details, _ := backend.GetDetails([]string{"b1", "r1"})
	if len(details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(details))
	}
}

func TestInitialize_DuplicateID(t *testing.T) {
	files := map[string]string{
		"behaviors/must/common.bhv": "---\nid: common-id\ntitle: Behavior 1\n---\nBody B1",
		"recipes/common.rcp":        "---\nid: common-id\ntitle: Recipe 1\n---\nBody R1",
	}
	tempDir := createTestFiles(t, files)

	backend := NewLocalFilesystemBackend()
	configMap := map[string]interface{}{
		"rootDir":           tempDir,
		"entityTypes":       getDefaultEntityDefs(),
		"requireExplicitID": true,
	}

	err := backend.Initialize(configMap)
	if err == nil {
		t.Fatal("Expected Initialize to fail with duplicate ID, but got nil error")
	}
	if !errors.Is(err, ErrDuplicateID) {
		t.Errorf("Expected error type ErrDuplicateID, got %v", err)
	}
	t.Logf("Got expected error: %v", err) // Log error for context
}

// --- parseAndValidateFile Tests ---

func TestParseAndValidateFile_ValidBehavior(t *testing.T) {
	content := "---\nid: b1\ntitle: Valid Behavior\ntags: [tag1, tag2]\n---\nBehavior Body"
	tempDir := createTestFiles(t, map[string]string{"behaviors/must/b1.bhv": content})
	filePath := filepath.Join(tempDir, "behaviors/must/b1.bhv")

	backend := NewLocalFilesystemBackend()
	backend.requireExplicitID = true // Need to set this for the test
	entityDef := config.EntityTypeDefinition{Name: "behavior", PathGlob: "behaviors/**/*.bhv", RequiredFrontMatter: []string{"id", "title"}}

	entity, err := backend.parseAndValidateFile(filePath, "behavior", entityDef)
	if err != nil {
		t.Fatalf("parseAndValidateFile failed: %v", err)
	}
	if entity == nil {
		t.Fatal("Expected valid entity, got nil")
	}

	if entity.ID != "b1" {
		t.Errorf("Expected ID b1, got %s", entity.ID)
	}
	if entity.Type != "behavior" {
		t.Errorf("Expected Type behavior, got %s", entity.Type)
	}
	if entity.Tier != "must" { // Test tier inference
		t.Errorf("Expected Tier must (inferred from path), got %s", entity.Tier)
	}
	if entity.Body != "Behavior Body" {
		t.Errorf("Expected Body 'Behavior Body', got '%s'", entity.Body)
	}
	if len(entity.Metadata) == 0 {
		t.Fatal("Expected metadata, got empty map")
	}
	if title, _ := entity.Metadata["title"].(string); title != "Valid Behavior" {
		t.Errorf("Expected title 'Valid Behavior', got '%s'", title)
	}
	if tags, ok := entity.Metadata["tags"].([]interface{}); !ok || len(tags) != 2 {
		t.Errorf("Expected tags [tag1, tag2], got %v", entity.Metadata["tags"])
	}
}

func TestParseAndValidateFile_ValidRecipe(t *testing.T) {
	content := "---\nid: r1\ntitle: Valid Recipe\n---\nRecipe Body"
	tempDir := createTestFiles(t, map[string]string{"recipes/r1.rcp": content})
	filePath := filepath.Join(tempDir, "recipes/r1.rcp")

	backend := NewLocalFilesystemBackend()
	backend.requireExplicitID = true
	entityDef := config.EntityTypeDefinition{Name: "recipe", PathGlob: "recipes/*.rcp", RequiredFrontMatter: []string{"id", "title"}}

	entity, err := backend.parseAndValidateFile(filePath, "recipe", entityDef)
	if err != nil {
		t.Fatalf("parseAndValidateFile failed: %v", err)
	}
	if entity == nil {
		t.Fatal("Expected valid entity, got nil")
	}

	if entity.ID != "r1" || entity.Type != "recipe" || entity.Body != "Recipe Body" {
		t.Errorf("Unexpected fields: %+v", entity)
	}
}

func TestParseAndValidateFile_MissingRequiredID(t *testing.T) {
	content := "---\ntitle: Missing ID\n---\nBody"
	tempDir := createTestFiles(t, map[string]string{"recipes/bad_id.rcp": content})
	filePath := filepath.Join(tempDir, "recipes/bad_id.rcp")

	backend := NewLocalFilesystemBackend()
	backend.requireExplicitID = true // Explicitly require ID
	entityDef := config.EntityTypeDefinition{Name: "recipe", PathGlob: "recipes/*.rcp", RequiredFrontMatter: []string{"id", "title"}}

	entity, err := backend.parseAndValidateFile(filePath, "recipe", entityDef)
	if err == nil {
		t.Fatal("Expected error for missing ID, got nil")
	}
	if entity != nil {
		t.Errorf("Expected nil entity for missing ID, got %+v", entity)
	}
	t.Logf("Got expected error: %v", err)
}

func TestParseAndValidateFile_InvalidYAML(t *testing.T) {
	content := "---\nid: bad-yaml\ntitle: Bad YAML\ntags: [tag1 tag2] : oops\n---\nBody"
	tempDir := createTestFiles(t, map[string]string{"behaviors/should/by.bhv": content})
	filePath := filepath.Join(tempDir, "behaviors/should/by.bhv")

	backend := NewLocalFilesystemBackend()
	backend.requireExplicitID = true
	entityDef := config.EntityTypeDefinition{Name: "behavior", PathGlob: "behaviors/**/*.bhv", RequiredFrontMatter: []string{"id", "title"}}

	entity, err := backend.parseAndValidateFile(filePath, "behavior", entityDef)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
	if entity != nil {
		t.Errorf("Expected nil entity for invalid YAML, got %+v", entity)
	}
	t.Logf("Got expected error: %v", err)
}

func TestParseAndValidateFile_TierInferenceOverride(t *testing.T) {
	// Tier inferred as 'must' from path, but explicitly 'should' in frontmatter
	content := "---\nid: b-override\ntitle: Tier Override\ntier: should\n---\nBody"
	tempDir := createTestFiles(t, map[string]string{"behaviors/must/b-override.bhv": content})
	filePath := filepath.Join(tempDir, "behaviors/must/b-override.bhv")

	backend := NewLocalFilesystemBackend()
	backend.requireExplicitID = true
	entityDef := config.EntityTypeDefinition{Name: "behavior", PathGlob: "behaviors/**/*.bhv", RequiredFrontMatter: []string{"id", "title"}}

	entity, err := backend.parseAndValidateFile(filePath, "behavior", entityDef)
	if err != nil {
		t.Fatalf("parseAndValidateFile failed: %v", err)
	}
	if entity == nil {
		t.Fatal("Expected valid entity, got nil")
	}

	if entity.Tier != "should" { // Check override
		t.Errorf("Expected Tier 'should' (from frontmatter), got '%s'", entity.Tier)
	}
}
