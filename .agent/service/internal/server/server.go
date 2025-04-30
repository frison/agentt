package server

import (
	"agentt/internal/config"
	"agentt/internal/content"
	"agentt/internal/store"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Server wraps the HTTP server dependencies and handlers.
type Server struct {
	cfg   *config.ServiceConfig
	store *store.GuidanceStore
	// Adding a testing.T field ONLY for debugging the Query issue
	// t *testing.T // REMOVE THIS AFTER DEBUGGING
}

// NewServer creates a new HTTP server instance.
// Adjust signature if adding testing.T
func NewServer(cfg *config.ServiceConfig, store *store.GuidanceStore) *Server {
	return &Server{
		cfg:   cfg,
		store: store,
		// t: t, // REMOVE THIS
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/entityTypes", s.HandleEntityTypes)
	// mux.HandleFunc("/discover/", s.HandleDiscover) // DEPRECATED/REMOVED
	mux.HandleFunc("/llm.txt", s.HandleLLMGuidance)
	// Add new summary endpoint
	mux.HandleFunc("/summary", s.HandleSummary)
	// Add new details endpoint
	mux.HandleFunc("/details", s.HandleDetails)

	log.Printf("Starting HTTP server on %s...", s.cfg.ListenAddress)
	return http.ListenAndServe(s.cfg.ListenAddress, mux)
}

// HandleHealth returns a simple 200 OK.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleEntityTypes returns the configured entity types.
func (s *Server) HandleEntityTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, s.cfg.EntityTypes)
}

/* // HandleDiscover DEPRECATED/REMOVED
func (s *Server) HandleDiscover(w http.ResponseWriter, r *http.Request) {
	// ... old handler code removed ...
}
*/

// HandleLLMGuidance serves the LLM guidance text file with replaced placeholders.
func (s *Server) HandleLLMGuidance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	contentBytes, err := os.ReadFile(s.cfg.LLMGuidanceFile)
	if err != nil {
		log.Printf("Error reading LLM guidance file '%s': %v", s.cfg.LLMGuidanceFile, err)
		http.Error(w, "Internal Server Error: Could not load LLM guidance", http.StatusInternalServerError)
		return
	}

	// Build entity type documentation
	var entityDocs strings.Builder
	for _, et := range s.cfg.EntityTypes {
		entityDocs.WriteString(fmt.Sprintf("*   **%s**: %s (Discover via `/discover/%s`)\n", et.Name, et.Description, et.Name))
	}

	// Replace placeholder
	outputBytes := bytes.Replace(contentBytes, []byte("{{ENTITY_TYPES_DOCUMENTATION}}"), []byte(entityDocs.String()), 1)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(outputBytes)
}

// HandleSummary returns a JSON array of ItemSummary for all valid items.
func (s *Server) HandleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	allValidItems := s.store.Query(map[string]interface{}{}) // Fix Query call

	summaries := make([]content.ItemSummary, 0, len(allValidItems))
	for _, item := range allValidItems {
		// Extract common fields
		id := ""
		if idVal, ok := item.FrontMatter["id"].(string); ok {
			id = idVal
		} else if titleVal, ok := item.FrontMatter["title"].(string); ok && item.EntityType == "behavior" {
			// Fallback to title for behaviors if no ID
			id = titleVal // Or generate a more unique one?
		}

		description := ""
		if descVal, ok := item.FrontMatter["description"].(string); ok {
			description = descVal
		}

		tags := []string{}
		if tagsVal, ok := item.FrontMatter["tags"].([]interface{}); ok {
			for _, tagInterface := range tagsVal {
				if tagStr, strOk := tagInterface.(string); strOk {
					tags = append(tags, tagStr)
				}
			}
		}

		summaries = append(summaries, content.ItemSummary{
			ID:          id,
			Type:        item.EntityType,
			Tier:        item.Tier, // Will be empty if not a behavior or not inferred
			Tags:        tags,
			Description: description,
		})
	}

	s.writeJSONResponse(w, http.StatusOK, summaries)
}

// --- Request/Response Structs ---

type DetailsRequest struct {
	IDs []string `json:"ids"`
}

// HandleDetails returns full details for requested item IDs.
func (s *Server) HandleDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.IDs) == 0 {
		http.Error(w, "Request body must contain a non-empty 'ids' array", http.StatusBadRequest)
		return
	}

	// Fetch all valid items to search through (could optimize if store supports direct ID lookup)
	// allValidItems := s.store.Query(map[string]interface{}{}) // Get all valid items - Already fetched above, remove redundant call?
	// Let's reuse the existing allValidItems fetched for HandleSummary for efficiency, assuming HandleDetails might be called closely. Or fetch fresh? Fetch fresh for isolation.
	allValidItemsForDetails := s.store.Query(map[string]interface{}{}) // Fetch fresh for details
	foundItems := make([]*content.Item, 0)
	requestedIDsSet := make(map[string]bool)
	for _, id := range req.IDs {
		requestedIDsSet[id] = true
	}

	for _, item := range allValidItemsForDetails { // Use the fresh list
		itemID := ""
		if idVal, ok := item.FrontMatter["id"].(string); ok {
			itemID = idVal
		} else if titleVal, ok := item.FrontMatter["title"].(string); ok && item.EntityType == "behavior" {
			// Use title as fallback ID for behaviors, consistent with /summary
			itemID = titleVal
		}

		if itemID != "" && requestedIDsSet[itemID] {
			foundItems = append(foundItems, item)
			delete(requestedIDsSet, itemID) // Mark as found
		}
	}

	// Log IDs that were requested but not found (optional)
	// for id := range requestedIDsSet {
	// 	log.Printf("Warning: Requested ID '%s' not found in store.", id)
	// }

	s.writeJSONResponse(w, http.StatusOK, foundItems)
}

// writeJSONResponse is a helper to marshal data to JSON and write the response.
func (s *Server) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ") // Use indent for readability
	if err != nil {
		log.Printf("Error marshaling JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}
