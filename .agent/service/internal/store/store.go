package store

import (
	// "agentt/internal/content" // Original incorrect path?
	"agentt/internal/content" // Correct path relative to module 'agentt'
	"errors"                  // Re-add import
	"fmt"
	"log" // Added for logging
	"reflect"
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
	// Make a copy to avoid returning internal slice directly
	all := make([]*content.Item, 0, len(s.itemsByPath))
	for _, item := range s.itemsByPath {
		all = append(all, item)
	}
	return all
}

// GetByPath retrieves a single item by its source path.
func (s *GuidanceStore) GetByPath(sourcePath string) (*content.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.itemsByPath[sourcePath]
	return item, found
}

// GetByID retrieves an item by its canonical ID.
func (s *GuidanceStore) GetByID(id string) (*content.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.itemsByID[id]
	return item, found
}

// Query searches the store based on provided filter criteria.
// Filters is a map where keys are field names (e.g., "entityType", "tier", "tag", or any frontmatter key)
// and values are the desired values to match.
// For "tag", the value should be a single tag string; the item matches if it contains that tag.
func (s *GuidanceStore) Query(filters map[string]interface{}) []*content.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*content.Item
	for _, item := range s.itemsByPath { // Iterate through path map to get unique items
		if !item.IsValid { // Ignore invalid items during query
			continue
		}
		if matchesFilters(item, filters) {
			results = append(results, item)
		}
	}
	return results
}

// matchesFilters checks if a single item matches all provided filter criteria.
func matchesFilters(item *content.Item, filters map[string]interface{}) bool {
	for key, filterValue := range filters {
		matches := false
		switch key {
		case "entityType":
			if strVal, ok := filterValue.(string); ok && item.EntityType == strVal {
				matches = true
			}
		case "tier":
			if strVal, ok := filterValue.(string); ok && item.Tier == strVal {
				matches = true
			}
		case "tag":
			filterTag, okFilter := filterValue.(string)
			itemTags, okItem := item.FrontMatter["tags"].([]interface{})
			if okFilter && okItem {
				for _, itemTagRaw := range itemTags {
					if itemTagStr, okTagStr := itemTagRaw.(string); okTagStr && itemTagStr == filterTag {
						matches = true
						break // Found the tag, no need to check further item tags for *this* filter tag
					}
				}
			}
		default: // Assume it's a frontmatter key
			if item.FrontMatter != nil {
				if itemValue, exists := item.FrontMatter[key]; exists {
					// Simple comparison for now, might need DeepEqual for complex types
					if reflect.DeepEqual(itemValue, filterValue) {
						matches = true
					}
				}
			}
		}

		if !matches {
			return false // If any filter doesn't match, the item doesn't match overall
		}
	}
	return true // All filters matched
}
