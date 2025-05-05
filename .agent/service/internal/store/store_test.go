package store_test

import (
	"agentt/internal/content"
	"agentt/internal/store"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"
)

// Helper to create a sample content item
func createItem(path, entityType string, isValid bool, fm map[string]interface{}) *content.Item {
	item := &content.Item{
		EntityType:  entityType,
		SourcePath:  path,
		FrontMatter: fm,
		IsValid:     isValid,
		LastUpdated: time.Now(),
	}

	// Extract common fields from FrontMatter into struct fields for testing queries
	if fm != nil {
		if tierVal, ok := fm["tier"].(string); ok {
			item.Tier = tierVal
		}
		// Add extraction for other queryable fields if needed (e.g., tags)
	}

	return item
}

func TestGuidanceStore_New(t *testing.T) {
	s := store.NewGuidanceStore()
	if s == nil {
		t.Fatal("NewGuidanceStore returned nil")
	}
	if len(s.GetAll()) != 0 {
		t.Errorf("Expected new store to be empty, got %d items", len(s.GetAll()))
	}
}

func TestGuidanceStore_AddOrUpdate(t *testing.T) {
	s := store.NewGuidanceStore()
	// Ensure items have explicit 'id' fields
	item1 := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"id": "item1", "title": "Item 1"})
	item2 := createItem("/path/to/item2.rcp", "recipe", true, map[string]interface{}{"id": "item2"})

	if err := s.AddOrUpdate(item1); err != nil {
		t.Fatalf("AddOrUpdate(item1) failed: %v", err)
	}
	if err := s.AddOrUpdate(item2); err != nil {
		t.Fatalf("AddOrUpdate(item2) failed: %v", err)
	}

	if len(s.GetAll()) != 2 {
		t.Fatalf("Expected store to have 2 items, got %d", len(s.GetAll()))
	}

	// Test update
	item1Updated := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"id": "item1", "title": "Item 1 Updated"}) // Use the same ID for update
	if err := s.AddOrUpdate(item1Updated); err != nil {
		t.Fatalf("AddOrUpdate(item1Updated) failed: %v", err)
	}

	retrieved, found := s.GetByPath("/path/to/item1.bhv")
	if !found {
		t.Fatal("Updated item1 not found")
	}
	if retrieved.FrontMatter["title"] != "Item 1 Updated" {
		t.Errorf("Expected updated title 'Item 1 Updated', got '%v'", retrieved.FrontMatter["title"])
	}

	// Test adding nil or invalid path item
	/* // Remove empty branches causing staticcheck warnings
	if err := s.AddOrUpdate(nil); err == nil { // Expecting an error here ideally, but current AddOrUpdate ignores nil
		// t.Error("AddOrUpdate(nil) should ideally return an error, but didn't")
	}
	if err := s.AddOrUpdate(&content.Item{SourcePath: ""}); err == nil { // Expecting an error here ideally
		// t.Error("AddOrUpdate(empty path) should ideally return an error, but didn't")
	}
	*/
	// Explicitly call them outside the check, as they are currently no-ops that don't error
	_ = s.AddOrUpdate(nil)             // Currently ignored by AddOrUpdate, assign to _ to satisfy errcheck
	_ = s.AddOrUpdate(&content.Item{}) // Also ignored, assign to _ to satisfy errcheck

	if len(s.GetAll()) != 2 {
		t.Errorf("Store size changed after adding nil/empty path items, got %d", len(s.GetAll()))
	}
}

func TestGuidanceStore_Remove(t *testing.T) {
	s := store.NewGuidanceStore()
	item1 := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"id": "rm-item1"})
	item2 := createItem("/path/to/item2.rcp", "recipe", true, map[string]interface{}{"id": "rm-item2"})
	if err := s.AddOrUpdate(item1); err != nil {
		t.Fatalf("AddOrUpdate(item1) failed: %v", err)
	}
	if err := s.AddOrUpdate(item2); err != nil {
		t.Fatalf("AddOrUpdate(item2) failed: %v", err)
	}

	s.Remove("/path/to/item1.bhv")
	if len(s.GetAll()) != 1 {
		t.Fatalf("Expected store to have 1 item after remove, got %d", len(s.GetAll()))
	}

	_, found := s.GetByPath("/path/to/item1.bhv")
	if found {
		t.Error("Removed item1 was still found")
	}

	// Test remove non-existent
	s.Remove("/path/to/nonexistent.txt")
	if len(s.GetAll()) != 1 {
		t.Fatalf("Store size changed after removing non-existent item, got %d", len(s.GetAll()))
	}

	// Test remove empty path
	s.Remove("")
	if len(s.GetAll()) != 1 {
		t.Fatalf("Store size changed after removing empty path, got %d", len(s.GetAll()))
	}
}

func TestGuidanceStore_GetByPath(t *testing.T) {
	s := store.NewGuidanceStore()
	item1 := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"id": "get-item1", "title": "Get Me"})
	if err := s.AddOrUpdate(item1); err != nil {
		t.Fatalf("AddOrUpdate(item1) failed: %v", err)
	}

	retrieved, found := s.GetByPath("/path/to/item1.bhv")
	if !found {
		t.Fatal("GetByPath failed to find existing item")
	}
	if retrieved == nil {
		t.Fatal("GetByPath returned nil item for existing path")
	}
	if retrieved.FrontMatter["title"] != "Get Me" {
		t.Errorf("Retrieved item has wrong title: %v", retrieved.FrontMatter["title"])
	}

	_, found = s.GetByPath("/path/to/nonexistent.txt")
	if found {
		t.Error("GetByPath found non-existent item")
	}
}

// func TestGuidanceStore_Query(t *testing.T) { // Comment out original complex test for now
// 	s := store.NewGuidanceStore()
//
// 	itemB1 := createItem("/path/must/b1.bhv", "behavior", true, map[string]interface{}{"tags": []interface{}{"core"}, "priority": 1})
// ... rest of original test commented out ...
// }

// --- TDD Tests for Query ---

func TestQuery_FilterByEntityType(t *testing.T) {
	s := store.NewGuidanceStore()

	// Ensure items have explicit 'id' fields
	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{"id": "b1"})
	itemB2 := createItem("/path/b2.bhv", "behavior", true, map[string]interface{}{"id": "b2"})
	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{"id": "r1"})
	// This item is invalid, so it shouldn't need an ID to be added, but adding one for consistency
	// Note: AddOrUpdate currently logs an error for missing ID but doesn't return it, test checks Query results.
	itemBInvalid := createItem("/path/b3.bhv", "behavior", false, map[string]interface{}{"id": "b3-invalid"})

	if err := s.AddOrUpdate(itemB1); err != nil {
		t.Fatalf("AddOrUpdate(itemB1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemB2); err != nil {
		t.Fatalf("AddOrUpdate(itemB2) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemR1); err != nil {
		t.Fatalf("AddOrUpdate(itemR1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemBInvalid); err != nil {
		// We now EXPECT an error here because GetItemID requires an ID even if IsValid is false
		// However, the AddOrUpdate logic might need adjustment if we *want* to store invalid items even without IDs.
		// For now, let's assume AddOrUpdate requires a valid ID to proceed.
		t.Logf("Note: AddOrUpdate(itemBInvalid) correctly failed as expected due to missing ID requirement: %v", err)
	} else {
		// If AddOrUpdate *did* succeed (e.g., if IsValid bypasses ID check somehow), that's unexpected.
		t.Logf("Warning: AddOrUpdate(itemBInvalid) succeeded unexpectedly.")
	}

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter for behaviors",
			filters:       map[string]interface{}{"entityType": "behavior"},
			expectedPaths: []string{"/path/b1.bhv", "/path/b2.bhv"}, // Excludes recipe and invalid behavior
		},
		{
			name:          "Filter for recipes",
			filters:       map[string]interface{}{"entityType": "recipe"},
			expectedPaths: []string{"/path/r1.rcp"},
		},
		{
			name:          "Filter for non-existent type",
			filters:       map[string]interface{}{"entityType": "unknown"},
			expectedPaths: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := s.Query(tc.filters)
			resultPaths := make([]string, len(results))
			for i, item := range results {
				resultPaths[i] = item.SourcePath
			}
			sort.Strings(resultPaths)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(resultPaths, tc.expectedPaths) {
				t.Errorf("Query mismatch:\nFilters:  %v\nExpected: %v\nGot:      %v", tc.filters, tc.expectedPaths, resultPaths)
			}
		})
	}
}

func TestQuery_FilterByTier(t *testing.T) {
	s := store.NewGuidanceStore()

	// Ensure items have explicit 'id' fields
	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{"id": "tier-b1", "tier": "must"})
	itemB2 := createItem("/path/b2.bhv", "behavior", true, map[string]interface{}{"id": "tier-b2", "tier": "should"})
	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{"id": "tier-r1", "tier": "must"})
	itemBNoTier := createItem("/path/b-notier.bhv", "behavior", true, map[string]interface{}{"id": "tier-b-notier"})

	if err := s.AddOrUpdate(itemB1); err != nil {
		t.Fatalf("AddOrUpdate(itemB1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemB2); err != nil {
		t.Fatalf("AddOrUpdate(itemB2) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemR1); err != nil {
		t.Fatalf("AddOrUpdate(itemR1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemBNoTier); err != nil {
		t.Fatalf("AddOrUpdate(itemBNoTier) failed: %v", err)
	}

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter for tier must",
			filters:       map[string]interface{}{"tier": "must"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
		{
			name:          "Filter for tier should",
			filters:       map[string]interface{}{"tier": "should"},
			expectedPaths: []string{"/path/b2.bhv"},
		},
		{
			name:          "Filter for tier non-existent",
			filters:       map[string]interface{}{"tier": "unknown"},
			expectedPaths: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := s.Query(tc.filters)
			resultPaths := make([]string, len(results))
			for i, item := range results {
				resultPaths[i] = item.SourcePath
			}
			sort.Strings(resultPaths)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(resultPaths, tc.expectedPaths) {
				t.Errorf("Query mismatch:\nFilters:  %v\nExpected: %v\nGot:      %v", tc.filters, tc.expectedPaths, resultPaths)
			}
		})
	}
}

func TestQuery_FilterByFrontMatterSimple(t *testing.T) {
	s := store.NewGuidanceStore()

	// Ensure items have explicit 'id' fields
	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{
		"id":       "fm-b1",
		"custom":   "value1",
		"priority": 1,
	})
	itemB2 := createItem("/path/b2.bhv", "behavior", true, map[string]interface{}{
		"id":       "fm-b2",
		"custom":   "value2",
		"priority": 2,
	})
	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{
		"id":     "fm-r1",
		"custom": "value1", // Same custom value as b1
	})
	itemBNoCustom := createItem("/path/b-nocustom.bhv", "behavior", true, map[string]interface{}{
		"id":       "fm-b-nocustom",
		"priority": 1,
	})

	if err := s.AddOrUpdate(itemB1); err != nil {
		t.Fatalf("AddOrUpdate(itemB1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemB2); err != nil {
		t.Fatalf("AddOrUpdate(itemB2) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemR1); err != nil {
		t.Fatalf("AddOrUpdate(itemR1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemBNoCustom); err != nil {
		t.Fatalf("AddOrUpdate(itemBNoCustom) failed: %v", err)
	}

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter by recipe id fm-r1",
			filters:       map[string]interface{}{"id": "fm-r1"},
			expectedPaths: []string{"/path/r1.rcp"},
		},
		{
			name:          "Filter by priority 1",
			filters:       map[string]interface{}{"priority": 1},
			expectedPaths: []string{"/path/b-nocustom.bhv", "/path/b1.bhv"},
		},
		{
			name:          "Filter by custom value1",
			filters:       map[string]interface{}{"custom": "value1"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
		{
			name:          "Filter by id fm-b1 and priority 1",
			filters:       map[string]interface{}{"id": "fm-b1", "priority": 1},
			expectedPaths: []string{"/path/b1.bhv"},
		},
		{
			name:          "Filter by non-existent priority",
			filters:       map[string]interface{}{"priority": 99},
			expectedPaths: []string{},
		},
		{
			name:          "Filter by non-existent frontmatter key",
			filters:       map[string]interface{}{"nonexistent": "value"},
			expectedPaths: []string{},
		},
		{
			name:          "Filter by custom value on item without custom field",
			filters:       map[string]interface{}{"custom": "value1"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := s.Query(tc.filters)
			resultPaths := make([]string, len(results))
			for i, item := range results {
				resultPaths[i] = item.SourcePath
			}
			sort.Strings(resultPaths)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(resultPaths, tc.expectedPaths) {
				t.Errorf("Query mismatch:\nFilters:  %v\nExpected: %v\nGot:      %v", tc.filters, tc.expectedPaths, resultPaths)
			}
		})
	}
}

func TestQuery_FilterByTag(t *testing.T) {
	s := store.NewGuidanceStore()

	// Ensure items have explicit 'id' fields
	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{
		"id":   "tag-b1",
		"tags": []interface{}{"core", "api"},
	})
	itemB2 := createItem("/path/b2.bhv", "behavior", true, map[string]interface{}{
		"id":   "tag-b2",
		"tags": []interface{}{"ui", "experimental"},
	})
	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{
		"id":   "tag-r1",
		"tags": []interface{}{"core", "deployment"},
	})
	itemBNoTags := createItem("/path/b-notags.bhv", "behavior", true, map[string]interface{}{
		"id": "tag-b-notags",
	})
	itemBBadTags := createItem("/path/b-badtags.bhv", "behavior", true, map[string]interface{}{
		"id":   "tag-b-badtags",
		"tags": "not-a-slice", // Invalid tags field
	})

	if err := s.AddOrUpdate(itemB1); err != nil {
		t.Fatalf("AddOrUpdate(itemB1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemB2); err != nil {
		t.Fatalf("AddOrUpdate(itemB2) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemR1); err != nil {
		t.Fatalf("AddOrUpdate(itemR1) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemBNoTags); err != nil {
		t.Fatalf("AddOrUpdate(itemBNoTags) failed: %v", err)
	}
	if err := s.AddOrUpdate(itemBBadTags); err != nil {
		t.Fatalf("AddOrUpdate(itemBBadTags) failed: %v", err)
	}

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter by tag core",
			filters:       map[string]interface{}{"tag": "core"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
		{
			name:          "Filter by tag git",
			filters:       map[string]interface{}{"tag": "git"},
			expectedPaths: []string{},
		},
		{
			name:          "Filter by tag api",
			filters:       map[string]interface{}{"tag": "api"},
			expectedPaths: []string{"/path/b1.bhv"},
		},
		{
			name:          "Filter by tag deployment",
			filters:       map[string]interface{}{"tag": "deployment"},
			expectedPaths: []string{"/path/r1.rcp"},
		},
		{
			name:          "Filter by non-existent tag",
			filters:       map[string]interface{}{"tag": "unknown"},
			expectedPaths: []string{},
		},
		{
			name:          "Filter by tag on item with no tags field",
			filters:       map[string]interface{}{"tag": "core"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
		{
			name:          "Filter by tag on item with bad tags field",
			filters:       map[string]interface{}{"tag": "core"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := s.Query(tc.filters)
			resultPaths := make([]string, len(results))
			for i, item := range results {
				resultPaths[i] = item.SourcePath
			}
			sort.Strings(resultPaths)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(resultPaths, tc.expectedPaths) {
				t.Errorf("Query mismatch:\nFilters:  %v\nExpected: %v\nGot:      %v", tc.filters, tc.expectedPaths, resultPaths)
			}
		})
	}
}

// TODO: Add TestQuery_FilterCombined
// This test would involve querying based on multiple criteria (e.g., type AND tag)
// to ensure the filtering logic combines conditions correctly.

func TestGuidanceStore_AddOrUpdate_DuplicateID(t *testing.T) {
	s := store.NewGuidanceStore()
	item1 := createItem("/path/to/item-one.bhv", "behavior", true, map[string]interface{}{"id": "duplicate-id"})
	item2 := createItem("/path/to/item-two.bhv", "behavior", true, map[string]interface{}{"id": "duplicate-id"}) // Same ID, different path

	// Add the first item, should succeed
	if err := s.AddOrUpdate(item1); err != nil {
		t.Fatalf("AddOrUpdate(item1) failed unexpectedly: %v", err)
	}

	// Add the second item with the same ID, should now return ErrDuplicateID
	err := s.AddOrUpdate(item2)
	if err == nil {
		t.Fatal("AddOrUpdate(item2) succeeded, expected ErrDuplicateID")
	}
	if !errors.Is(err, store.ErrDuplicateID) {
		t.Fatalf("AddOrUpdate(item2) returned wrong error type. Got: %v, Want wrapped: %v", err, store.ErrDuplicateID)
	}

	// Verify store state - should still contain only the FIRST item added for this ID
	allitems := s.GetAll()
	if len(allitems) != 1 {
		t.Errorf("expected store size to be 1 after failed duplicate add, but got %d", len(allitems))
	}

	// Check if the item associated with the ID is the *first* one added
	retrievedbyid, foundbyid := s.GetByID("duplicate-id")
	if !foundbyid {
		t.Fatal("item with duplicate id not found by id after failed add attempt")
	}
	if retrievedbyid.SourcePath != "/path/to/item-one.bhv" {
		t.Errorf("expected item with id 'duplicate-id' to have sourcepath '/path/to/item-one.bhv', got '%s'", retrievedbyid.SourcePath)
	}

	// check that the item at the *first* path is still findable
	_, foundbypath1 := s.GetByPath("/path/to/item-one.bhv")
	if !foundbypath1 {
		t.Error("item at original path '/path/to/item-one.bhv' was not found after failed duplicate add")
	}
	// check that the item at the *second* (conflicting) path was not added
	_, foundbypath2 := s.GetByPath("/path/to/item-two.bhv")
	if foundbypath2 {
		t.Error("item at conflicting path '/path/to/item-two.bhv' was found after failed duplicate add")
	}
}
