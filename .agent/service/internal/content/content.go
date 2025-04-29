package content

import "time"

// Item represents a single piece of discovered guidance content (behavior, recipe, etc.).
type Item struct {
	// EntityType is the configured type name (e.g., "behavior", "recipe").
	EntityType string `json:"entityType"`
	// SourcePath is the absolute file path from which this item was loaded.
	SourcePath string `json:"sourcePath"`
	// FrontMatter holds all key-value pairs parsed from the YAML frontmatter.
	FrontMatter map[string]interface{} `json:"frontMatter"`
	// Body holds the raw content of the file after the frontmatter.
	Body string `json:"body,omitempty"` // Often not needed in discovery results, maybe make optional?
	// IsValid indicates if the item passed validation (e.g., required frontmatter present).
	IsValid bool `json:"isValid"`
	// ValidationErrors contains reasons why IsValid is false. Empty if IsValid is true.
	ValidationErrors []string `json:"validationErrors,omitempty"`
	// LastUpdated is the timestamp when the item was last loaded or updated.
	LastUpdated time.Time `json:"lastUpdated"`

	// --- Metadata inferred during parsing ---

	// Tier is specific to 'behavior' types, inferred from the path ("must" or "should").
	Tier string `json:"tier,omitempty"` // Only populated for behaviors
}
