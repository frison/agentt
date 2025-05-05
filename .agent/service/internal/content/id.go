package content

import (
	"fmt"
)

// GetItemID extracts the canonical identifier for an item EXCLUSIVELY from the
// 'id' field in the frontmatter. It returns an error if the 'id' field is
// missing, not a string, or empty.
func GetItemID(item *Item) (string, error) {
	if item == nil {
		return "", fmt.Errorf("cannot get ID from nil item")
	}

	// Require 'id' field in FrontMatter
	if item.FrontMatter == nil {
		path := item.SourcePath
		if path == "" {
			path = "(unknown source)"
		}
		return "", fmt.Errorf("item '%s' missing frontmatter, cannot determine ID", path)
	}

	idVal, ok := item.FrontMatter["id"].(string)
	if !ok {
		path := item.SourcePath
		if path == "" {
			path = "(unknown source)"
		}
		// Check if 'id' key exists but is wrong type
		if _, exists := item.FrontMatter["id"]; exists {
			return "", fmt.Errorf("item '%s' has non-string 'id' field in frontmatter (type: %T)", path, item.FrontMatter["id"])
		}
		return "", fmt.Errorf("item '%s' missing required 'id' field in frontmatter", path)
	}

	if idVal == "" {
		path := item.SourcePath
		if path == "" {
			path = "(unknown source)"
		}
		return "", fmt.Errorf("item '%s' has empty 'id' field in frontmatter", path)
	}

	return idVal, nil
}
