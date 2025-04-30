package server_test

import (
	"agentt/internal/config"
	"agentt/internal/content"
	"agentt/internal/server"
	"agentt/internal/store"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		// LLMGuidanceFile: createMockLLMFile(), // REMOVED: Create a temporary LLM file
	}

	// Mock Store
	mockStore := store.NewGuidanceStore()
	// IDs in FrontMatter should NOT have prefixes; the prefixing happens during summary/details generation.
	// However, the tests currently EXPECT prefixed IDs in the output.
	// The Query function relies on FrontMatter fields for filtering.
	// Let's adjust the items to have the ID/Title fields that the prefixing logic uses.
	itemB1 := &content.Item{SourcePath: "/test/b1.bhv", EntityType: "behavior", IsValid: true, Tier: "must", FrontMatter: map[string]interface{}{"title": "B1", "tags": []interface{}{"core"}}} // Use title as base ID for behavior
	itemB2 := &content.Item{SourcePath: "/test/b2.bhv", EntityType: "behavior", IsValid: true, Tier: "should", FrontMatter: map[string]interface{}{"title": "B2", "tags": []interface{}{"git"}}} // Use title as base ID for behavior
	itemR1 := &content.Item{SourcePath: "/test/r1.rcp", EntityType: "recipe", IsValid: true, FrontMatter: map[string]interface{}{"id": "r1", "tags": []interface{}{"core", "git"}}}     // Use id as base ID for recipe
	itemInvalid := &content.Item{SourcePath: "/test/invalid.bhv", EntityType: "behavior", IsValid: false, FrontMatter: map[string]interface{}{"title": "Invalid"}}
	mockStore.AddOrUpdate(itemB1)
	mockStore.AddOrUpdate(itemB2)
	mockStore.AddOrUpdate(itemR1)
	mockStore.AddOrUpdate(itemInvalid)

	return server.NewServer(mockConfig, mockStore)
}

/* // REMOVED Mock LLM File helper
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
*/

func TestHandlers(t *testing.T) {
	srv := setupTestServer()
	// defer os.Remove(mockLLMFilePath) // REMOVED: Clean up mock file

	// Setup mux manually for testing specific handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HandleHealth)
	mux.HandleFunc("/entityTypes", srv.HandleEntityTypes)
	// mux.HandleFunc("/discover/", srv.HandleDiscover) // REMOVED handler registration
	mux.HandleFunc("/llm.txt", srv.HandleLLMGuidance)
	mux.HandleFunc("/summary", srv.HandleSummary)
	mux.HandleFunc("/details", srv.HandleDetails)

	testServer := httptest.NewServer(mux)
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

		// Construct expected output using embedded text and mock config
		// Note the extra newline added by the loop in the handler
		expectedEntityTypeDocs := "*   **behavior**: Test Behavior \n*   **recipe**: Test Recipe \n"
		expectedOutput := strings.Replace(server.LLMServerHelpContent, "{{ENTITY_TYPES_DOCUMENTATION}}", expectedEntityTypeDocs, 1)

		if bodyStr != expectedOutput {
			t.Errorf("LLM guidance mismatch.\nExpected:\n%s\nGot:\n%s", expectedOutput, bodyStr)
		}
		/* // REMOVED old check
		if !strings.Contains(bodyStr, "*   **behavior**: Test Behavior") {
			t.Errorf("LLM guidance missing expected behavior description. Got:\n%s", bodyStr)
		}
		if !strings.Contains(bodyStr, "*   **recipe**: Test Recipe") {
			t.Errorf("LLM guidance missing expected recipe description. Got:\n%s", bodyStr)
		}
		*/
	})

	/* // --- Test /discover --- // REMOVED
	t.Run("DiscoverBehaviorsNoFilter", func(t *testing.T) {
		// ... removed test ...
	})

	t.Run("DiscoverRecipesTagCore", func(t *testing.T) {
		// ... removed test ...
	})

	t.Run("DiscoverBehaviorsTierMust", func(t *testing.T) {
		// ... removed test ...
	})

	t.Run("DiscoverInvalidEntityType", func(t *testing.T) {
		// ... removed test ...
	})

	t.Run("DiscoverMissingEntityType", func(t *testing.T) {
		// ... removed test ...
	})

	t.Run("DiscoverNoMatch", func(t *testing.T) {
		// ... removed test ...
	})
	*/ // END REMOVED /discover tests

	// --- Test /summary ---
	t.Run("SummaryEndpoint", func(t *testing.T) {
		res, err := http.Get(testServer.URL + "/summary")
		if err != nil {
			t.Fatalf("GET /summary failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /summary, got %d", res.StatusCode)
		}

		var summaries []content.ItemSummary
		if err := json.NewDecoder(res.Body).Decode(&summaries); err != nil {
			t.Fatalf("Failed to decode /summary response: %v", err)
		}

		// Expect 3 valid items (b1, b2, r1)
		if len(summaries) != 3 {
			t.Errorf("Expected 3 summaries, got %d", len(summaries))
		}

		// Basic check for one item
		foundR1 := false
		for _, s := range summaries {
			if s.ID == "rcp-r1" && s.Type == "recipe" {
				foundR1 = true
				if len(s.Tags) != 2 || s.Tags[0] != "core" || s.Tags[1] != "git" {
					t.Errorf("Recipe r1 summary has incorrect tags: %v", s.Tags)
				}
				break
			}
		}
		if !foundR1 {
			t.Error("Did not find summary for recipe r1")
		}
	})

	// --- Test /details ---
	t.Run("DetailsEndpoint_Found", func(t *testing.T) {
		// Request details for b1 (using title as ID) and r1
		requestBody := `{"ids": ["bhv-B1", "rcp-r1"]}`
		res, err := http.Post(testServer.URL+"/details", "application/json", strings.NewReader(requestBody))
		if err != nil {
			t.Fatalf("POST /details failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /details, got %d", res.StatusCode)
		}

		var details []*content.Item
		if err := json.NewDecoder(res.Body).Decode(&details); err != nil {
			t.Fatalf("Failed to decode /details response: %v", err)
		}

		if len(details) != 2 {
			t.Errorf("Expected 2 detail items, got %d", len(details))
		}

		// Check types
		foundB1 := false
		foundR1 := false
		for _, item := range details {
			if item.SourcePath == "/test/b1.bhv" && item.EntityType == "behavior" {
				foundB1 = true
			}
			if item.SourcePath == "/test/r1.rcp" && item.EntityType == "recipe" {
				foundR1 = true
			}
		}
		if !foundB1 || !foundR1 {
			t.Errorf("Did not find expected items in details response. Found B1: %t, Found R1: %t", foundB1, foundR1)
		}
	})

	t.Run("DetailsEndpoint_NotFound", func(t *testing.T) {
		requestBody := `{"ids": ["nonexistent"]}`
		res, err := http.Post(testServer.URL+"/details", "application/json", strings.NewReader(requestBody))
		if err != nil {
			t.Fatalf("POST /details failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", res.StatusCode)
		}
		var details []*content.Item
		json.NewDecoder(res.Body).Decode(&details)
		if len(details) != 0 {
			t.Errorf("Expected 0 detail items for non-existent ID, got %d", len(details))
		}
	})

	t.Run("DetailsEndpoint_BadRequest", func(t *testing.T) {
		requestBody := `{"wrong_key": []}`
		res, err := http.Post(testServer.URL+"/details", "application/json", strings.NewReader(requestBody))
		if err != nil {
			t.Fatalf("POST /details failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for invalid body, got %d", res.StatusCode)
		}
	})
}
