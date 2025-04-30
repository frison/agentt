package content

import (
	"fmt"
)

// getBaseID extracts the base identifier from an item's front matter.
// It uses the 'id' field primarily, falling back to 'title' for behaviors.
// Returns the base ID string or an error if neither is found or valid.
func getBaseID(item *Item) (string, error) {
	if item == nil || item.FrontMatter == nil {
		return "", fmt.Errorf("cannot get base ID from nil item or frontmatter")
	}

	// Prefer 'id' field if present
	if idVal, ok := item.FrontMatter["id"].(string); ok && idVal != "" {
		return idVal, nil
	}

	// Fallback to 'title' for behaviors
	if item.EntityType == "behavior" {
		if titleVal, ok := item.FrontMatter["title"].(string); ok && titleVal != "" {
			return titleVal, nil
		}
	}

	// If neither found
	return "", fmt.Errorf("item %s (%s) missing required 'id' or fallback 'title' field", item.SourcePath, item.EntityType)
}

// GetPrefixedID returns the standard prefixed ID (e.g., bhv-ID, rcp-ID) for the item.
// It extracts the base ID using getBaseID and applies the correct prefix.
// Returns the prefixed ID string or an error if the base ID cannot be determined.
func GetPrefixedID(item *Item) (string, error) {
	baseID, err := getBaseID(item)
	if err != nil {
		return "", err // Propagate error from getBaseID
	}

	switch item.EntityType {
	case "behavior":
		return "bhv-" + baseID, nil
	case "recipe":
		return "rcp-" + baseID, nil
	default:
		// Maybe return baseID for unknown types, or error?
		// Let's return baseID for now, but log a warning or reconsider later.
		// log.Printf("Warning: Generating ID without prefix for unknown entity type '%s'", item.EntityType)
		return baseID, nil
	}
}