package server_test

import (
	"agentt/internal/config"
	// "agentt/internal/content" // REMOVED
	"agentt/internal/server"
	// "agentt/internal/store" // REMOVED
	"agentt/internal/guidance/backend" // ADDED
	"bytes"                            // ADDED for details POST request
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	// "strings" // REMOVED - Unused
	"testing"
	"time" // ADDED for backend.Entity
)

// --- Mock Guidance Backend ---

type mockGuidanceBackend struct {
	MockSummaries []backend.Summary
	MockEntities  map[string]backend.Entity // Store by ID for easy lookup
}

func (m *mockGuidanceBackend) Initialize(config map[string]interface{}) error {
	// No-op for mock
	return nil
}

func (m *mockGuidanceBackend) GetSummary() ([]backend.Summary, error) {
	// Return a copy to avoid test modifications affecting the mock
	summariesCopy := make([]backend.Summary, len(m.MockSummaries))
	copy(summariesCopy, m.MockSummaries)
	return summariesCopy, nil
}

func (m *mockGuidanceBackend) GetDetails(ids []string) ([]backend.Entity, error) {
	found := make([]backend.Entity, 0)
	for _, id := range ids {
		if entity, ok := m.MockEntities[id]; ok {
			found = append(found, entity)
		}
	}
	return found, nil
}

// newMockGuidanceBackend creates a mock backend with sample data.
func newMockGuidanceBackend() *mockGuidanceBackend {
	summaryB1 := backend.Summary{ID: "bhv-B1", Type: "behavior", Tier: "must", Tags: []string{"core"}, Description: "Behavior B1"}
	summaryB2 := backend.Summary{ID: "bhv-B2", Type: "behavior", Tier: "should", Tags: []string{"git"}, Description: "Behavior B2"}
	summaryR1 := backend.Summary{ID: "rcp-r1", Type: "recipe", Tags: []string{"core", "git"}, Description: "Recipe R1"}

	entityB1 := backend.Entity{ID: "bhv-B1", Type: "behavior", Tier: "must", Body: "Body B1", ResourceLocator: "/test/b1.bhv", Metadata: map[string]interface{}{"title": "B1", "tags": []interface{}{"core"}}, LastUpdated: time.Now()}
	entityB2 := backend.Entity{ID: "bhv-B2", Type: "behavior", Tier: "should", Body: "Body B2", ResourceLocator: "/test/b2.bhv", Metadata: map[string]interface{}{"title": "B2", "tags": []interface{}{"git"}}, LastUpdated: time.Now()}
	entityR1 := backend.Entity{ID: "rcp-r1", Type: "recipe", Body: "Body R1", ResourceLocator: "/test/r1.rcp", Metadata: map[string]interface{}{"id": "r1", "tags": []interface{}{"core", "git"}}, LastUpdated: time.Now()}

	return &mockGuidanceBackend{
		MockSummaries: []backend.Summary{summaryB1, summaryB2, summaryR1},
		MockEntities: map[string]backend.Entity{
			"bhv-B1": entityB1,
			"bhv-B2": entityB2,
			"rcp-r1": entityR1,
		},
	}
}

// --- Test Setup ---

// Helper to create a test server with mock backend
func setupTestServer() *server.Server {
	// Mock Config (remains mostly the same)
	mockConfig := &config.ServiceConfig{
		ListenAddress: ":0",
		EntityTypes: []config.EntityTypeDefinition{
			{Name: "behavior", Description: "Test Behavior"},
			{Name: "recipe", Description: "Test Recipe"},
		},
		// Backend config not directly used by server logic itself, but mock needs data
	}

	// Create Mock Backend
	mockBackend := newMockGuidanceBackend()

	// Pass backend to server
	return server.NewServer(mockConfig, mockBackend)
}

// --- Test Cases ---

func TestHandlers(t *testing.T) {
	srv := setupTestServer()

	// Setup mux manually for testing specific handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HandleHealth)
	mux.HandleFunc("/entityTypes", srv.HandleEntityTypes)
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

		// Check if it matches the embedded content directly (placeholder logic removed from handler)
		if bodyStr != server.LLMServerHelpContent {
			t.Errorf("LLM guidance mismatch.\nExpected:\n%s\nGot:\n%s", server.LLMServerHelpContent, bodyStr)
		}
	})

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

		// Expect backend.Summary type now
		var summaries []backend.Summary
		if err := json.NewDecoder(res.Body).Decode(&summaries); err != nil {
			t.Fatalf("Failed to decode /summary response: %v", err)
		}

		// Expect 3 summaries from mock backend
		if len(summaries) != 3 {
			t.Errorf("Expected 3 summaries, got %d", len(summaries))
		}

		// Basic check for one item (R1)
		foundR1 := false
		for _, s := range summaries {
			if s.ID == "rcp-r1" && s.Type == "recipe" { // Check ID from mock
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
		// Request details for bhv-B1 and rcp-r1
		requestBody := `{"ids": ["bhv-B1", "rcp-r1"]}`
		// Use bytes.NewBuffer for request body
		res, err := http.Post(testServer.URL+"/details", "application/json", bytes.NewBufferString(requestBody))
		if err != nil {
			t.Fatalf("POST /details failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /details, got %d", res.StatusCode)
		}

		// Expect backend.Entity type now
		var details []backend.Entity
		if err := json.NewDecoder(res.Body).Decode(&details); err != nil {
			t.Fatalf("Failed to decode /details response: %v", err)
		}

		// Expect 2 entities from mock backend
		if len(details) != 2 {
			t.Errorf("Expected 2 detail items, got %d", len(details))
		}

		// Check types and ResourceLocator
		foundB1 := false
		foundR1 := false
		for _, item := range details {
			if item.ID == "bhv-B1" && item.Type == "behavior" && item.ResourceLocator == "/test/b1.bhv" {
				foundB1 = true
			}
			if item.ID == "rcp-r1" && item.Type == "recipe" && item.ResourceLocator == "/test/r1.rcp" {
				foundR1 = true
			}
		}
		if !foundB1 || !foundR1 {
			t.Errorf("Did not find expected items in details response. Found B1: %t, Found R1: %t", foundB1, foundR1)
		}
	})

	// Add test case for details not found
	t.Run("DetailsEndpoint_NotFound", func(t *testing.T) {
		requestBody := `{"ids": ["non-existent-id"]}`
		res, err := http.Post(testServer.URL+"/details", "application/json", bytes.NewBufferString(requestBody))
		if err != nil {
			t.Fatalf("POST /details (not found) failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /details (not found), got %d", res.StatusCode)
		}

		var details []backend.Entity
		if err := json.NewDecoder(res.Body).Decode(&details); err != nil {
			t.Fatalf("Failed to decode /details (not found) response: %v", err)
		}

		if len(details) != 0 {
			t.Errorf("Expected 0 detail items for non-existent ID, got %d", len(details))
		}
	})
}
