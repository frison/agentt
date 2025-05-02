package backend

import "time"

// Summary provides a concise overview of a guidance entity.
type Summary struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`           // "behavior" or "recipe"
	Tier        string   `json:"tier,omitempty"` // "must" or "should" (only for behaviors)
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description"`
}

// Entity represents the full details of a guidance entity.
type Entity struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`           // "behavior" or "recipe"
	Tier            string                 `json:"tier,omitempty"` // "must" or "should" (only for behaviors)
	Body            string                 `json:"body"`
	ResourceLocator string                 `json:"resourceLocator"` // Generalized locator (e.g., file path, URL)
	Metadata        map[string]interface{} `json:"metadata"`        // Frontmatter or other metadata
	LastUpdated     time.Time              `json:"lastUpdated"`
	// Tags and Description omitted as they are available in Summary and redundant here based on Plan 0.1
}

// GuidanceBackend defines the interface for fetching guidance entities
// from different storage mechanisms (e.g., local filesystem, remote API).
type GuidanceBackend interface {
	// Initialize prepares the backend with necessary configuration.
	Initialize(config map[string]interface{}) error

	// GetSummary retrieves summaries for all available guidance entities.
	GetSummary() ([]Summary, error)

	// GetDetails retrieves the full details for the specified entity IDs.
	GetDetails(ids []string) ([]Entity, error)
}
