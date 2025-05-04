package server_test

import (
	"agentt/internal/config"
	// "agentt/internal/content" // REMOVED
	"agentt/internal/server"
	// "agentt/internal/store" // REMOVED
	"agentt/internal/guidance/backend" // Use package path for mocks generated within
	// "agentt/internal/guidance/backend/mocks" // REMOVED
	// "bytes" // REMOVED - Unused
	"encoding/json"
	// "io" // Unused
	"net/http"
	"net/http/httptest"
	"reflect" // Added for deep comparison
	"sort"    // Added for ElementsMatch replacement
	"strings" // Added back for details body
	"testing"
	// "time" // REMOVED - Unused
	"os"
	"path/filepath"

	"agentt/internal/guidance/backend/localfs"
	"github.com/golang/mock/gomock"
	"time"
)

// --- Test Setup ---

// Helper to create a server instance with a gomock backend
func setupTestServer(t *testing.T) (*server.Server, *backend.MockGuidanceBackend, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockBackend := backend.NewMockGuidanceBackend(ctrl) // Corrected: Use type defined in package

	// Use the new config.Config struct
	mockConfig := &config.Config{
		ListenAddress: ":0",
		EntityTypes: []config.EntityType{
			{Name: "behavior", Description: "Behavior Type", RequiredFields: []string{"id", "tier"}}, // Corrected description
			{Name: "recipe", Description: "Recipe Type", RequiredFields: []string{"id"}},             // Corrected description
		},
		Backends: []config.BackendSpec{
			{ // Define BackendSpec correctly
				Type: "mock", // Or a relevant type for the test setup
				Settings: map[string]interface{}{
					"rootDir":         ".",                                    // Example setting
					"entityLocations": map[string]string{"behavior": "*.bhv"}, // Example setting
				},
			},
		},
	}

	srv := server.NewServer(mockConfig, mockBackend) // Pass mock backend
	return srv, mockBackend, ctrl                    // Return srv, mock, ctrl
}

// --- Test Cases ---

func TestServer_HandleHealth(t *testing.T) {
	srv, _, ctrl := setupTestServer(t) // Adjusted call site (ignore mockBackend)
	defer ctrl.Finish()

	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	srv.HandleHealth(rr, req)
	// Standard library checks
	if rr.Code != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	expectedBody := "OK"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

func TestServer_HandleEntityTypes(t *testing.T) {
	srv, _, ctrl := setupTestServer(t)
	defer ctrl.Finish()

	req, _ := http.NewRequest("GET", "/entity-types", nil)
	rr := httptest.NewRecorder()
	srv.HandleEntityTypes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var entityTypes []config.EntityType // Expect a slice (JSON array)
	err := json.NewDecoder(rr.Body).Decode(&entityTypes)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Define expected slice based on setupTestServer config
	expectedTypes := []config.EntityType{
		{Name: "behavior", Description: "Behavior Type", RequiredFields: []string{"id", "tier"}},
		{Name: "recipe", Description: "Recipe Type", RequiredFields: []string{"id"}},
	}

	// Sort both slices by Name for consistent comparison
	sort.Slice(expectedTypes, func(i, j int) bool {
		return expectedTypes[i].Name < expectedTypes[j].Name
	})
	sort.Slice(entityTypes, func(i, j int) bool {
		return entityTypes[i].Name < entityTypes[j].Name
	})

	// Use reflect.DeepEqual for slice comparison after sorting
	if !reflect.DeepEqual(expectedTypes, entityTypes) {
		t.Errorf("Returned entity types do not match expected (after sorting).\nExpected: %+v\nActual:   %+v", expectedTypes, entityTypes)
	}
}

func TestServer_HandleSummary_Success(t *testing.T) {
	server, mockBackend, ctrl := setupTestServer(t) // Correct call site
	defer ctrl.Finish()

	// Define Expected Data (remains the same)
	expectedSummaries := []backend.Summary{
		{ID: "bhv1", Type: "behavior", Tier: "must", Description: "Behavior 1"},
		{ID: "rcp1", Type: "recipe", Description: "Recipe 1"},
	}

	// Set Mock Expectations (remains the same, now uses gomock object)
	mockBackend.EXPECT().GetSummary().Return(expectedSummaries, nil).Times(1)

	// Perform Request (remains the same)
	req := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()
	server.HandleSummary(w, req)

	// Assert Results (remains the same)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	var summaries []backend.Summary
	err := json.NewDecoder(resp.Body).Decode(&summaries)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Replace ElementsMatch with custom logic (sort and compare)
	sortSummaries(expectedSummaries)
	sortSummaries(summaries)
	if !reflect.DeepEqual(expectedSummaries, summaries) {
		t.Errorf("Returned summaries do not match expected (after sorting).\nExpected: %+v\nActual:   %+v", expectedSummaries, summaries)
	}
}

func TestServer_HandleDetails_Success(t *testing.T) {
	server, mockBackend, ctrl := setupTestServer(t) // Correct call site
	defer ctrl.Finish()

	// Define Expected Data (remains the same, corrected Metadata slightly)
	expectedIDs := []string{"bhv1", "rcp1"}
	expectedEntities := []backend.Entity{
		{ID: "bhv1", Type: "behavior", Tier: "must", Body: "Body 1", Metadata: map[string]interface{}{"id": "bhv1", "tier": "must"}},
		{ID: "rcp1", Type: "recipe", Body: "Body 2", Metadata: map[string]interface{}{"id": "rcp1"}},
	}

	// Set Mock Expectations (remains the same, now uses gomock object)
	mockBackend.EXPECT().GetDetails(expectedIDs).Return(expectedEntities, nil).Times(1)

	// Perform Request (remains the same)
	body := strings.NewReader(`{"ids":["bhv1", "rcp1"]}`)
	req := httptest.NewRequest(http.MethodPost, "/details", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleDetails(w, req)

	// Assert Results (remains the same)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	var entities []backend.Entity
	err := json.NewDecoder(resp.Body).Decode(&entities)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Replace ElementsMatch with custom logic (sort and compare)
	sortEntities(expectedEntities)
	sortEntities(entities)
	if !reflect.DeepEqual(expectedEntities, entities) {
		t.Errorf("Returned entities do not match expected (after sorting).\nExpected: %+v\nActual:   %+v", expectedEntities, entities)
	}
}

// Removed duplicate TestServer_HandleEntityTypes_Success

// --- Helper functions for sorting --- //

func sortSummaries(summaries []backend.Summary) {
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ID < summaries[j].ID
	})
}

func sortEntities(entities []backend.Entity) {
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})
}

// Helper to setup test server with a real LocalFS backend
func setupTestServerWithLocalFS(t *testing.T) (*server.Server, backend.GuidanceBackend) { // Return real backend interface
	testConfig := &config.Config{
		ListenAddress: ":0",
		EntityTypes: []config.EntityType{
			{Name: "behavior", RequiredFields: []string{"id", "tier"}},
			{Name: "recipe", RequiredFields: []string{"id"}},
		},
		Backends: []config.BackendSpec{
			{
				Type: "localfs", // Use localfs type
				Settings: map[string]interface{}{
					"rootDir": "testdata/localfs_content", // Point to test data
					"entityLocations": map[string]string{
						"behavior": "*.bhv",
						"recipe":   "*.rcp",
					},
				},
			},
		},
		LoadedFromPath: ".", // Mock the loaded path for relative resolution
	}

	// Setup testdata directory
	rootDir := "testdata/localfs_content"
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", rootDir, err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "bhv1.bhv"), []byte("---\nid: bhv1\ntier: must\n---\nBehavior Body 1"), 0644); err != nil {
		t.Fatalf("Failed to write test file bhv1.bhv: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "rcp1.rcp"), []byte("---\nid: rcp1\n---\nRecipe Body 1"), 0644); err != nil {
		t.Fatalf("Failed to write test file rcp1.rcp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll("testdata") })

	// Extract settings and create the real backend
	backendSettings, err := testConfig.Backends[0].GetLocalFSSettings()
	if err != nil {
		t.Fatalf("Failed to get localfs settings: %v", err)
	}
	fsBackend, err := localfs.NewLocalFSBackend(backendSettings, testConfig.LoadedFromPath, testConfig.EntityTypes)
	if err != nil {
		t.Fatalf("Failed to create localfs backend: %v", err)
	}

	// Pass the single backend instance, not a slice
	srv := server.NewServer(testConfig, fsBackend)
	return srv, fsBackend // Return the real backend
}

// --- LocalFS Integration Test Cases ---

// Test Summary endpoint with a real LocalFS backend
func TestServer_HandleSummary_LocalFS(t *testing.T) {
	// Use the helper that sets up a real backend
	srv, _ := setupTestServerWithLocalFS(t) // Ignore backend return value here

	// Perform Request
	req := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()
	srv.HandleSummary(w, req)

	// Assert Results
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	var summaries []backend.Summary
	err := json.NewDecoder(resp.Body).Decode(&summaries)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Define expected based on files created in setupTestServerWithLocalFS
	expectedSummaries := []backend.Summary{
		{ID: "bhv1", Type: "behavior", Tier: "must"},
		{ID: "rcp1", Type: "recipe"},
	}

	// Sort and compare
	sortSummaries(expectedSummaries)
	sortSummaries(summaries)
	if !reflect.DeepEqual(expectedSummaries, summaries) {
		t.Errorf("Returned summaries do not match expected (after sorting).\nExpected: %+v\nActual:   %+v", expectedSummaries, summaries)
	}
}

// Test Details endpoint with a real LocalFS backend
func TestServer_HandleDetails_LocalFS(t *testing.T) {
	// Use the helper that sets up a real backend
	srv, _ := setupTestServerWithLocalFS(t)

	// Perform Request
	body := strings.NewReader(`{"ids":["bhv1", "rcp1"]}`)
	req := httptest.NewRequest(http.MethodPost, "/details", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.HandleDetails(w, req)

	// Assert Results
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	var entities []backend.Entity
	err := json.NewDecoder(resp.Body).Decode(&entities)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Define expected based on files created in setupTestServerWithLocalFS
	expectedEntities := []backend.Entity{
		{ID: "bhv1", Type: "behavior", Tier: "must", Body: "Behavior Body 1", Metadata: map[string]interface{}{"id": "bhv1", "tier": "must"}, ResourceLocator: filepath.Join("testdata/localfs_content", "bhv1.bhv")},
		{ID: "rcp1", Type: "recipe", Body: "Recipe Body 1", Metadata: map[string]interface{}{"id": "rcp1"}, ResourceLocator: filepath.Join("testdata/localfs_content", "rcp1.rcp")},
	}

	// Normalize ResourceLocator paths before comparing
	for i := range expectedEntities {
		absPath, _ := filepath.Abs(expectedEntities[i].ResourceLocator)
		expectedEntities[i].ResourceLocator = absPath
	}
	for i := range entities {
		absPath, _ := filepath.Abs(entities[i].ResourceLocator)
		entities[i].ResourceLocator = absPath
	}

	// Sort and compare
	sortEntities(expectedEntities)
	sortEntities(entities)

	// Zero out LastUpdated fields before comparison as they are dynamic
	zeroTime := time.Time{}
	for i := range expectedEntities {
		expectedEntities[i].LastUpdated = zeroTime
	}
	for i := range entities {
		entities[i].LastUpdated = zeroTime
	}

	if !reflect.DeepEqual(expectedEntities, entities) {
		t.Errorf("Returned entities do not match expected (after sorting and zeroing timestamps).\nExpected: %+v\nActual:   %+v", expectedEntities, entities)
	}
}
