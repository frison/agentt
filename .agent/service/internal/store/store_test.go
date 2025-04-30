package store_test

import (
	"agentt/internal/content"
	"agentt/internal/store"
	"reflect"
	"sort"
	"testing"
	"time"
)

// Helper to create a sample content item
func createItem(path, entityType string, isValid bool, fm map[string]interface{}) *content.Item {
	return &content.Item{
		EntityType:  entityType,
		SourcePath:  path,
		FrontMatter: fm,
		IsValid:     isValid,
		LastUpdated: time.Now(),
	}
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
	item1 := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"title": "Item 1"})
	item2 := createItem("/path/to/item2.rcp", "recipe", true, map[string]interface{}{"id": "item2"})

	s.AddOrUpdate(item1)
	s.AddOrUpdate(item2)

	if len(s.GetAll()) != 2 {
		t.Fatalf("Expected store to have 2 items, got %d", len(s.GetAll()))
	}

	// Test update
	item1Updated := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"title": "Item 1 Updated"})
	s.AddOrUpdate(item1Updated)

	if len(s.GetAll()) != 2 {
		t.Fatalf("Expected store to still have 2 items after update, got %d", len(s.GetAll()))
	}

	retrieved, found := s.GetByPath("/path/to/item1.bhv")
	if !found {
		t.Fatal("Updated item1 not found")
	}
	if retrieved.FrontMatter["title"] != "Item 1 Updated" {
		t.Errorf("Expected updated title 'Item 1 Updated', got '%v'", retrieved.FrontMatter["title"])
	}

	// Test adding nil or invalid path item
	s.AddOrUpdate(nil)
	s.AddOrUpdate(&content.Item{SourcePath: ""})
	if len(s.GetAll()) != 2 {
		t.Errorf("Store size changed after adding nil/empty path items, got %d", len(s.GetAll()))
	}
}

func TestGuidanceStore_Remove(t *testing.T) {
	s := store.NewGuidanceStore()
	item1 := createItem("/path/to/item1.bhv", "behavior", true, nil)
	item2 := createItem("/path/to/item2.rcp", "recipe", true, nil)
	s.AddOrUpdate(item1)
	s.AddOrUpdate(item2)

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
	item1 := createItem("/path/to/item1.bhv", "behavior", true, map[string]interface{}{"title": "Get Me"})
	s.AddOrUpdate(item1)

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

	itemB1 := createItem("/path/b1.bhv", "behavior", true, nil)
	itemB2 := createItem("/path/b2.bhv", "behavior", true, nil)
	itemR1 := createItem("/path/r1.rcp", "recipe", true, nil)
	itemBInvalid := createItem("/path/b3.bhv", "behavior", false, nil) // Should be ignored

	s.AddOrUpdate(itemB1)
	s.AddOrUpdate(itemB2)
	s.AddOrUpdate(itemR1)
	s.AddOrUpdate(itemBInvalid)

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

	itemB1 := createItem("/path/b1.bhv", "behavior", true, nil)
	itemB1.Tier = "must"
	itemB2 := createItem("/path/b2.bhv", "behavior", true, nil)
	itemB2.Tier = "should"
	itemB3 := createItem("/path/b3.bhv", "behavior", true, nil)
	itemB3.Tier = "must"
	itemR1 := createItem("/path/r1.rcp", "recipe", true, nil) // No tier

	s.AddOrUpdate(itemB1)
	s.AddOrUpdate(itemB2)
	s.AddOrUpdate(itemB3)
	s.AddOrUpdate(itemR1)

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter for tier must",
			filters:       map[string]interface{}{"tier": "must"},
			expectedPaths: []string{"/path/b1.bhv", "/path/b3.bhv"},
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

	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{"id": "r1", "priority": 10})
	itemR2 := createItem("/path/r2.rcp", "recipe", true, map[string]interface{}{"id": "r2", "priority": 20})
	itemR3 := createItem("/path/r3.rcp", "recipe", true, map[string]interface{}{"id": "r3", "priority": 10})
	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{"id": "b1", "priority": 10}) // Add ID for testing

	s.AddOrUpdate(itemR1)
	s.AddOrUpdate(itemR2)
	s.AddOrUpdate(itemR3)
	s.AddOrUpdate(itemB1)

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter by recipe id r2",
			filters:       map[string]interface{}{"id": "r2"},
			expectedPaths: []string{"/path/r2.rcp"},
		},
		{
			name:          "Filter by priority 10",
			filters:       map[string]interface{}{"priority": 10},
			expectedPaths: []string{"/path/r1.rcp", "/path/r3.rcp", "/path/b1.bhv"},
		},
		{
			name:          "Filter by id r1 and priority 10",
			filters:       map[string]interface{}{"id": "r1", "priority": 10},
			expectedPaths: []string{"/path/r1.rcp"},
		},
		{
			name:          "Filter by non-existent priority",
			filters:       map[string]interface{}{"priority": 99},
			expectedPaths: []string{},
		},
		{
			name:          "Filter by non-existent frontmatter key",
			filters:       map[string]interface{}{"custom": "value"},
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

func TestQuery_FilterByTag(t *testing.T) {
	s := store.NewGuidanceStore()

	itemB1 := createItem("/path/b1.bhv", "behavior", true, map[string]interface{}{"tags": []interface{}{"core", "setup"}, "priority": 1})
	itemB2 := createItem("/path/b2.bhv", "behavior", true, map[string]interface{}{"tags": []interface{}{"git"}, "priority": 100})
	itemR1 := createItem("/path/r1.rcp", "recipe", true, map[string]interface{}{"tags": []interface{}{"core", "git"}, "id": "r1"})
	itemR2 := createItem("/path/r2.rcp", "recipe", true, map[string]interface{}{"tags": []interface{}{"core"}, "id": "r2"})
	itemR3NoTags := createItem("/path/r3.rcp", "recipe", true, map[string]interface{}{"id": "r3"})

	s.AddOrUpdate(itemB1)
	s.AddOrUpdate(itemB2)
	s.AddOrUpdate(itemR1)
	s.AddOrUpdate(itemR2)
	s.AddOrUpdate(itemR3NoTags)

	tests := []struct {
		name          string
		filters       map[string]interface{}
		expectedPaths []string
	}{
		{
			name:          "Filter by tag core",
			filters:       map[string]interface{}{"tag": "core"},
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp", "/path/r2.rcp"},
		},
		{
			name:          "Filter by tag git",
			filters:       map[string]interface{}{"tag": "git"},
			expectedPaths: []string{"/path/b2.bhv", "/path/r1.rcp"},
		},
		{
			name:          "Filter by tag setup",
			filters:       map[string]interface{}{"tag": "setup"},
			expectedPaths: []string{"/path/b1.bhv"},
		},
		{
			name:          "Filter by non-existent tag",
			filters:       map[string]interface{}{"tag": "unknown"},
			expectedPaths: []string{},
		},
		{
			name:    "Filter by tag on item with no tags field",
			filters: map[string]interface{}{"tag": "core"},
			// This should not include /path/r3.rcp
			expectedPaths: []string{"/path/b1.bhv", "/path/r1.rcp", "/path/r2.rcp"},
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
