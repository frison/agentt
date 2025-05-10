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
	ID                           string                 `json:"id"`
	Type                         string                 `json:"type"`           // "behavior" or "recipe"
	Tier                         string                 `json:"tier,omitempty"` // "must" or "should" (only for behaviors)
	Body                         string                 `json:"body"`
	ResourceLocator              string                 `json:"resourceLocator"` // Generalized locator (e.g., file path, URL)
	Metadata                     map[string]interface{} `json:"metadata"`        // Frontmatter or other metadata
	LastUpdated                  time.Time              `json:"lastUpdated"`
	OriginatingBackendIdentifier string                 `json:"-"` // Identifier of the backend this entity came from, not serialized
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

// WritableBackend defines an interface for backends that support creating and updating entities.
// Implementations of this interface should ensure that operations are only performed
// if the backend instance is configured as writable.
type WritableBackend interface {
	// CreateEntity creates a new guidance entity in the backend.
	// entityData contains the frontmatter/metadata. body contains the main content.
	// It should return an error if the backend is not writable or if the entity already exists (unless force is true).
	CreateEntity(entityData map[string]interface{}, body string, force bool) error

	// UpdateEntity updates an existing guidance entity in the backend.
	// entityID is the ID of the entity to update.
	// updatedData contains the frontmatter/metadata fields to be updated. If nil, metadata is not changed.
	// updatedBody contains the new body. If nil, the body is not changed.
	// It should return an error if the backend is not writable, if the entity does not exist,
	// or if either updatedData or updatedBody is not provided (at least one must be).
	UpdateEntity(entityID string, updatedData map[string]interface{}, updatedBody *string) error
}
