package store

import (
	// "agentt/internal/content" // Original incorrect path?
	"agentt/internal/content" // Correct path relative to module 'agentt'
	"errors"                  // Re-add import
	"fmt"
	"log" // Added for logging
	"sync"
)

// ErrDuplicateID indicates an attempt to add an item with an ID that already exists but has a different source path.
var ErrDuplicateID = errors.New("duplicate guidance entity ID detected") // Re-add variable

// GuidanceStore holds the discovered and parsed guidance items in memory.
type GuidanceStore struct {
	mu          sync.RWMutex
	itemsByPath map[string]*content.Item // Keyed by absolute SourcePath
	itemsByID   map[string]*content.Item // Keyed by canonical Item ID
}

// NewGuidanceStore creates a new, empty guidance store.
func NewGuidanceStore() *GuidanceStore {
	return &GuidanceStore{
		itemsByPath: make(map[string]*content.Item),
		itemsByID:   make(map[string]*content.Item),
	}
}

// AddOrUpdate upserts a content item into the store.
// It checks for duplicate IDs and returns ErrDuplicateID if detected.
func (s *GuidanceStore) AddOrUpdate(item *content.Item) error {
	if item == nil || item.SourcePath == "" {
		log.Println("Warning: Attempted to add nil or empty-path item to store")
		return nil // Ignore invalid items, not an error condition for the store itself
	}

	newItemID, err := content.GetItemID(item)
	if err != nil {
		log.Printf("Error: Cannot determine ID for item %s: %v. Item not added to store.", item.SourcePath, err)
		return fmt.Errorf("failed to get ID for %s: %w", item.SourcePath, err)
	}

	s.mu.Lock() // Full lock for modification
	defer s.mu.Unlock()

	// Check for existing item with the same ID but different path
	if existingItem, idExists := s.itemsByID[newItemID]; idExists && existingItem.SourcePath != item.SourcePath {
		return fmt.Errorf("%w: ID '%s' used by '%s' and '%s'",
			ErrDuplicateID, newItemID, item.SourcePath, existingItem.SourcePath)
	}

	// If an item with the same SourcePath already exists, check if its ID changed
	if oldItem, pathExists := s.itemsByPath[item.SourcePath]; pathExists {
		oldItemID, oldErr := content.GetItemID(oldItem)
		// If the old item had a valid ID and it's different from the new ID, remove the old ID entry
		if oldErr == nil && oldItemID != newItemID {
			delete(s.itemsByID, oldItemID)
		}
	}

	// Add/Update in both maps
	s.itemsByPath[item.SourcePath] = item
	s.itemsByID[newItemID] = item
	return nil // Success
}

// Remove deletes an item from the store by its source path.
func (s *GuidanceStore) Remove(sourcePath string) {
	if sourcePath == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the item first to find its ID
	if itemToRemove, exists := s.itemsByPath[sourcePath]; exists {
		itemID, err := content.GetItemID(itemToRemove)
		if err == nil {
			delete(s.itemsByID, itemID) // Remove from ID index
		}
		delete(s.itemsByPath, sourcePath) // Remove from path index
	}
}

// GetAll returns a slice of all items currently in the store.
func (s *GuidanceStore) GetAll() []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allItems := make([]*content.Item, 0, len(s.itemsByPath)) // Iterate path map
	for _, item := range s.itemsByPath {
		allItems = append(allItems, item)
	}
	return allItems
}

// GetByPath retrieves a single item by its source path.
func (s *GuidanceStore) GetByPath(sourcePath string) (*content.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.itemsByPath[sourcePath]
	return item, found
}

// GetByID retrieves a single item by its canonical ID.
func (s *GuidanceStore) GetByID(id string) (*content.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.itemsByID[id]
	return item, found
}

// Query performs filtering on the items in the store.
// Filters is a map where key is the field name (e.g., "entityType", "tier", or any key in FrontMatter)
// and value is the desired value.
// NOTE: This still iterates through all items for general filtering.
// Optimization for direct ID lookup should use GetByID.
func (s *GuidanceStore) Query(filters map[string]interface{}) []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*content.Item, 0)

itemLoop:
	// Iterate over one of the maps, e.g., by path
	for _, item := range s.itemsByPath {
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
