package content

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetItemID extracts the canonical identifier for an item.
// Priority:
// 1. 'id' field in frontmatter (if present and non-empty string).
// 2. 'title' field in frontmatter (if entityType is "behavior" and present and non-empty string).
// 3. Filename stem (filename without extension or path) as a final fallback.
// Returns the ID string or an error if the fallback filename cannot be processed.
func GetItemID(item *Item) (string, error) {
	if item == nil {
		return "", fmt.Errorf("cannot get ID from nil item")
	}

	// 1. Prefer 'id' field if present in FrontMatter
	if item.FrontMatter != nil {
		if idVal, ok := item.FrontMatter["id"].(string); ok && idVal != "" {
			return idVal, nil
		}

		// 2. Fallback to 'title' for behaviors if present in FrontMatter
		if item.EntityType == "behavior" {
			if titleVal, ok := item.FrontMatter["title"].(string); ok && titleVal != "" {
				return titleVal, nil
			}
		}
	}

	// 3. Fallback to filename stem
	if item.SourcePath == "" {
		return "", fmt.Errorf("cannot derive ID from filename: SourcePath is empty")
	}
	filename := filepath.Base(item.SourcePath)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	if baseName == "" {
		// This could happen for hidden files like '.rcp' - might need refinement
		return "", fmt.Errorf("cannot derive ID from filename: resulted in empty base name for path %s", item.SourcePath)
	}
	return baseName, nil

	// Original Error if neither id nor title found - replaced by filename fallback
	// return "", fmt.Errorf("item %s (%s) missing required 'id' or fallback 'title' field, and cannot derive from filename", item.SourcePath, item.EntityType)
}

/* // REMOVED: Prefixing logic is no longer needed.
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
*/
