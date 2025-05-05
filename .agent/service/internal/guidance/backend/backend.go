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

// GuidanceBackend defines the interface for loading guidance entities.
//
// //go:generate mockgen -destination=mock_backend.go -package=backend -source=backend.go GuidanceBackend // Temporarily commented out
type GuidanceBackend interface {
	// GetSummary returns summaries for all available entities.
	GetSummary() ([]Summary, error)
	// GetDetails returns full details for the specified entity IDs.
	GetDetails(ids []string) ([]Entity, error)
}
