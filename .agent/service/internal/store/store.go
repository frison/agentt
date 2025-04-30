package store

import (
	"agentt/internal/content"
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
// Filters is a map where key is the field name (e.g., "entityType", "tier", or any key in FrontMatter)
// and value is the desired value.
func (s *GuidanceStore) Query(filters map[string]interface{}) []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*content.Item, 0)

itemLoop:
	for _, item := range s.items {
		if !item.IsValid {
			continue
		}

		// Check if the item matches ALL provided filters
		for filterKey, expectedFilterValue := range filters {
			match := false

			switch filterKey {
			case "entityType":
				if fmt.Sprintf("%v", item.EntityType) == fmt.Sprintf("%v", expectedFilterValue) {
					match = true
				}
			case "tier":
				if fmt.Sprintf("%v", item.Tier) == fmt.Sprintf("%v", expectedFilterValue) {
					match = true
				}
			case "tag": // Special handling for tag - must check FrontMatter["tags"]
				if actualValue, found := item.FrontMatter["tags"]; found { // Look for "tags" plural
					if tagsSlice, sliceOk := actualValue.([]interface{}); sliceOk {
						if expectedTagStr, filterOk := expectedFilterValue.(string); filterOk {
							for _, tagInItem := range tagsSlice {
								if tagStr, itemOk := tagInItem.(string); itemOk && tagStr == expectedTagStr {
									match = true
									break
								}
							}
						}
					}
				}
			default: // Handle other frontmatter keys
				if actualValue, found := item.FrontMatter[filterKey]; found {
					if fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedFilterValue) {
						match = true
					}
				}
			}

			// If this specific filter key did not match, skip the whole item
			if !match {
				continue itemLoop
			}

		} // End of loop over filters for one item

		results = append(results, item)

	} // End of loop over all items

	return results
}
