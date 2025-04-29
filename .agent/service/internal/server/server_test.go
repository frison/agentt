package server_test

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/content"
	"agent-guidance-service/internal/server"
	"agent-guidance-service/internal/store"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a test server with mock data
func setupTestServer() *server.Server {
	// Mock Config
	mockConfig := &config.ServiceConfig{
		ListenAddress: ":0", // Not used directly in handler tests
		EntityTypes: []config.EntityTypeDefinition{
			{Name: "behavior", Description: "Test Behavior", PathGlob: "*.bhv", RequiredFrontMatter: []string{"title", "tags"}},
			{Name: "recipe", Description: "Test Recipe", PathGlob: "*.rcp", RequiredFrontMatter: []string{"id", "tags"}},
		},
		LLMGuidanceFile: createMockLLMFile(), // Create a temporary LLM file
	}

	// Mock Store
	mockStore := store.NewGuidanceStore()
	itemB1 := &content.Item{SourcePath: "/test/b1.bhv", EntityType: "behavior", IsValid: true, Tier: "must", FrontMatter: map[string]interface{}{"title": "B1", "tags": []interface{}{"core"}}}
	itemB2 := &content.Item{SourcePath: "/test/b2.bhv", EntityType: "behavior", IsValid: true, Tier: "should", FrontMatter: map[string]interface{}{"title": "B2", "tags": []interface{}{"git"}}}
	itemR1 := &content.Item{SourcePath: "/test/r1.rcp", EntityType: "recipe", IsValid: true, FrontMatter: map[string]interface{}{"id": "r1", "tags": []interface{}{"core", "git"}}}
	itemInvalid := &content.Item{SourcePath: "/test/invalid.bhv", EntityType: "behavior", IsValid: false, FrontMatter: map[string]interface{}{"title": "Invalid"}}
	mockStore.AddOrUpdate(itemB1)
	mockStore.AddOrUpdate(itemB2)
	mockStore.AddOrUpdate(itemR1)
	mockStore.AddOrUpdate(itemInvalid)

	return server.NewServer(mockConfig, mockStore)
}

// Helper to create a temporary LLM guidance file for testing
var mockLLMFilePath string

func createMockLLMFile() string {
	if mockLLMFilePath != "" {
		// Clean up previous run if needed, though t.TempDir might handle it
		os.Remove(mockLLMFilePath)
	}
	// Use a fixed path or t.TempDir()
	// Using fixed path for simplicity here, but TempDir is better practice
	mockLLMFilePath = filepath.Join(os.TempDir(), "mock_llm_guidance.txt")
	content := []byte("LLM Guidance.\nEntity Types:\n{{ENTITY_TYPES_DOCUMENTATION}}End.")
	os.WriteFile(mockLLMFilePath, content, 0644)
	return mockLLMFilePath
}

func TestHandlers(t *testing.T) {
	srv := setupTestServer()
	defer os.Remove(mockLLMFilePath) // Clean up mock file

	// Setup mux manually for testing specific handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HandleHealth)
	mux.HandleFunc("/entityTypes", srv.HandleEntityTypes)
	mux.HandleFunc("/discover/", srv.HandleDiscover) // Use exported handler
	mux.HandleFunc("/llm.txt", srv.HandleLLMGuidance)

	testServer := httptest.NewServer(mux) // Pass mux to test server
	defer testServer.Close()

	// --- Test /health ---
	t.Run("HealthCheck", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/health")
		if err != nil {
			t.Fatalf("GET /health failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /health, got %d", res.StatusCode)
		}
		bodyBytes := make([]byte, 2)
		_, _ = res.Body.Read(bodyBytes)
		if string(bodyBytes) != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", string(bodyBytes))
		}
	})

	// --- Test /entityTypes ---
	t.Run("EntityTypes", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/entityTypes")
		if err != nil {
			t.Fatalf("GET /entityTypes failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /entityTypes, got %d", res.StatusCode)
		}

		var entityTypes []config.EntityTypeDefinition
		if err := json.NewDecoder(res.Body).Decode(&entityTypes); err != nil {
			t.Fatalf("Failed to decode /entityTypes response: %v", err)
		}
		if len(entityTypes) != 2 {
			t.Errorf("Expected 2 entity types, got %d", len(entityTypes))
		}
		if entityTypes[0].Name != "behavior" || entityTypes[1].Name != "recipe" {
			t.Errorf("Unexpected entity types returned: %v", entityTypes)
		}
	})

	// --- Test /llm.txt ---
	t.Run("LLMGuidance", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/llm.txt")
		if err != nil {
			t.Fatalf("GET /llm.txt failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /llm.txt, got %d", res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "*   **behavior**: Test Behavior") {
			t.Errorf("LLM guidance missing expected behavior description. Got:\n%s", bodyStr)
		}
		if !strings.Contains(bodyStr, "*   **recipe**: Test Recipe") {
			t.Errorf("LLM guidance missing expected recipe description. Got:\n%s", bodyStr)
		}
	})

	// --- Test /discover ---
	t.Run("DiscoverBehaviorsNoFilter", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/behavior")
		if err != nil {
			t.Fatalf("GET /discover/behavior failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}
		var items []*content.Item
		json.NewDecoder(res.Body).Decode(&items)
		if len(items) != 2 {
			t.Errorf("Expected 2 valid behaviors, got %d", len(items))
		}
	})

	t.Run("DiscoverRecipesTagCore", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/recipe?tag=core")
		if err != nil {
			t.Fatalf("GET /discover/recipe?tag=core failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}
		var items []*content.Item
		json.NewDecoder(res.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("Expected 1 recipe with tag 'core', got %d", len(items))
		}
		if items[0].SourcePath != "/test/r1.rcp" {
			t.Errorf("Expected recipe r1, got %s", items[0].SourcePath)
		}
	})

	t.Run("DiscoverBehaviorsTierMust", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/behavior?tier=must")
		if err != nil {
			t.Fatalf("GET /discover/behavior?tier=must failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}
		var items []*content.Item
		json.NewDecoder(res.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("Expected 1 behavior with tier 'must', got %d", len(items))
		}
		if items[0].SourcePath != "/test/b1.bhv" {
			t.Errorf("Expected behavior b1, got %s", items[0].SourcePath)
		}
	})

	t.Run("DiscoverInvalidEntityType", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/unknown")
		if err != nil {
			t.Fatalf("GET /discover/unknown failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found for invalid entity type, got %d", res.StatusCode)
		}
	})

	t.Run("DiscoverMissingEntityType", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/")
		if err != nil {
			t.Fatalf("GET /discover/ failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for missing entity type, got %d", res.StatusCode)
		}
	})

	t.Run("DiscoverNoMatch", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/discover/behavior?tag=nonexistent")
		if err != nil {
			t.Fatalf("GET /discover/behavior?tag=nonexistent failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}
		var items []*content.Item
		json.NewDecoder(res.Body).Decode(&items)
		if len(items) != 0 {
			t.Errorf("Expected 0 items for non-matching tag, got %d", len(items))
		}
	})
}
