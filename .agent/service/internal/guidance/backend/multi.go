package backend

import (
	"fmt"
	"log/slog"
	"sync"
)

// MultiBackend aggregates multiple GuidanceBackend implementations into one.
// It calls the corresponding methods on all underlying backends and merges results.
type MultiBackend struct {
	backends []GuidanceBackend
}

// Ensure MultiBackend implements the GuidanceBackend interface
var _ GuidanceBackend = (*MultiBackend)(nil)

// NewMultiBackend creates a new MultiBackend instance.
func NewMultiBackend(backends []GuidanceBackend) (*MultiBackend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("cannot create MultiBackend with zero underlying backends")
	}
	return &MultiBackend{backends: backends}, nil
}

// GetSummary calls GetSummary on all underlying backends and aggregates the results.
// It logs warnings for duplicate entity IDs found across different backends.
func (mb *MultiBackend) GetSummary() ([]Summary, error) {
	allSummaries := make([]Summary, 0)
	seenIDs := make(map[string]string) // id -> source backend identifier (e.g., index)
	var errors []error
	var mu sync.Mutex // Mutex to protect shared slices/maps during concurrent access if added later

	slog.Debug("MultiBackend GetSummary started", "backend_count", len(mb.backends))

	// Potential for parallelization later using goroutines and waitgroups
	for i, be := range mb.backends {
		backendIdentifier := fmt.Sprintf("backend_%d", i) // Simple identifier for logging
		slog.Debug("Calling GetSummary on underlying backend", "identifier", backendIdentifier)
		summaries, err := be.GetSummary()
		if err != nil {
			slog.Error("Error getting summary from underlying backend", "identifier", backendIdentifier, "error", err)
			mu.Lock()
			errors = append(errors, fmt.Errorf("backend %s GetSummary failed: %w", backendIdentifier, err))
			mu.Unlock()
			continue // Skip this backend
		}

		mu.Lock()
		for _, s := range summaries {
			if existingSource, found := seenIDs[s.ID]; found {
				slog.Warn("Duplicate entity ID found across backends", "id", s.ID, "source1", existingSource, "source2", backendIdentifier)
				// Skip adding the duplicate summary
				continue
			}
			allSummaries = append(allSummaries, s)
			seenIDs[s.ID] = backendIdentifier
		}
		mu.Unlock()
	}

	slog.Debug("MultiBackend GetSummary finished", "total_summaries", len(allSummaries), "errors_encountered", len(errors))

	// Combine errors if any occurred
	if len(errors) > 0 {
		// Depending on desired behavior, we might return partial results + error, or just error.
		// Returning partial results for now.
		// Consider using a multi-error type if more structured error handling is needed.
		combinedErr := fmt.Errorf("%d errors occurred during GetSummary across backends", len(errors))
		for _, e := range errors {
			combinedErr = fmt.Errorf("%w; %w", combinedErr, e) // Simple wrapping
		}
		return allSummaries, combinedErr
	}

	return allSummaries, nil
}

// GetDetails calls GetDetails on all underlying backends and aggregates the results.
// It ensures only one Entity per unique ID is returned (first one found wins).
func (mb *MultiBackend) GetDetails(ids []string) ([]Entity, error) {
	allEntities := make([]Entity, 0, len(ids))
	foundIDs := make(map[string]bool)
	var errors []error
	var mu sync.Mutex

	slog.Debug("MultiBackend GetDetails started", "backend_count", len(mb.backends), "requested_ids_count", len(ids))

	for i, be := range mb.backends {
		backendIdentifier := fmt.Sprintf("backend_%d", i)
		// Determine which IDs *might* be in this backend (optimization possible later)
		// For now, just request all originally requested IDs from each backend.
		slog.Debug("Calling GetDetails on underlying backend", "identifier", backendIdentifier)
		entities, err := be.GetDetails(ids)
		if err != nil {
			slog.Error("Error getting details from underlying backend", "identifier", backendIdentifier, "error", err)
			mu.Lock()
			errors = append(errors, fmt.Errorf("backend %s GetDetails failed: %w", backendIdentifier, err))
			mu.Unlock()
			continue // Skip this backend
		}

		mu.Lock()
		for _, e := range entities {
			if !foundIDs[e.ID] { // If we haven't already found details for this ID
				// Populate the OriginatingBackendIdentifier
				entityToStore := e // Make a copy to avoid modifying the slice element directly if it's a pointer later
				entityToStore.OriginatingBackendIdentifier = backendIdentifier
				allEntities = append(allEntities, entityToStore)
				foundIDs[e.ID] = true
			}
		}
		mu.Unlock()
	}

	slog.Debug("MultiBackend GetDetails finished", "entities_found", len(allEntities), "errors_encountered", len(errors))

	if len(errors) > 0 {
		combinedErr := fmt.Errorf("%d errors occurred during GetDetails across backends", len(errors))
		for _, e := range errors {
			combinedErr = fmt.Errorf("%w; %w", combinedErr, e)
		}
		return allEntities, combinedErr
	}

	return allEntities, nil
}
