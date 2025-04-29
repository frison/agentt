package store

import (
	"agent-guidance-service/internal/content"
	"fmt"
	"sync"
)

// GuidanceStore holds the discovered and parsed guidance items in memory.
type GuidanceStore struct {
	mu    sync.RWMutex
	items map[string]*content.Item // Keyed by absolute SourcePath
}

// NewGuidanceStore creates a new, empty guidance store.
func NewGuidanceStore() *GuidanceStore {
	return &GuidanceStore{
		items: make(map[string]*content.Item),
	}
}

// AddOrUpdate upserts a content item into the store.
func (s *GuidanceStore) AddOrUpdate(item *content.Item) {
	if item == nil || item.SourcePath == "" {
		return // Ignore invalid items
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item.SourcePath] = item
}

// Remove deletes an item from the store by its source path.
func (s *GuidanceStore) Remove(sourcePath string) {
	if sourcePath == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, sourcePath)
}

// GetAll returns a slice of all items currently in the store.
func (s *GuidanceStore) GetAll() []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allItems := make([]*content.Item, 0, len(s.items))
	for _, item := range s.items {
		allItems = append(allItems, item)
	}
	return allItems
}

// GetByPath retrieves a single item by its source path.
func (s *GuidanceStore) GetByPath(sourcePath string) (*content.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.items[sourcePath]
	return item, found
}

// Query performs filtering on the items in the store.
// Filters is a map where key is the field name (e.g., "entityType", "tier", or any key in FrontMatter) and value is the desired value.
func (s *GuidanceStore) Query(filters map[string]interface{}) []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*content.Item, 0)

itemLoop:
	for _, item := range s.items {
		// Only return valid items by default through Query?
		// Or allow querying invalid items too? Assuming only valid for now.
		if !item.IsValid {
			continue
		}

		for key, expectedValue := range filters {
			var actualValue interface{}
			found := false

			// Check top-level fields first
			switch key {
			case "entityType":
				actualValue = item.EntityType
				found = true
			case "sourcePath":
				actualValue = item.SourcePath
				found = true
			case "tier": // Specific to behaviors
				actualValue = item.Tier
				found = true
			default:
				// Check FrontMatter
				actualValue, found = item.FrontMatter[key]
			}

			if !found {
				continue itemLoop // Field not found in this item, cannot match filter
			}

			// Basic type-insensitive comparison (improve if needed, e.g., for numeric ranges)
			if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
				continue itemLoop // Value does not match
			}
		}
		// If we reach here, all filters matched
		results = append(results, item)
	}

	return results
}