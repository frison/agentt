package server

import (
	"agentt/internal/config"
	"agentt/internal/content"
	"agentt/internal/store"
	"bytes"
	_ "embed" // Import embed
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time" // Add time package
)

//go:embed llm_server_help.txt
var LLMServerHelpContent string // Embedded server protocol/help text (Exported for testing)

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
	// Wrap the mux with the logging middleware before starting the server
	loggedMux := LoggingMiddleware(mux)
	return http.ListenAndServe(s.cfg.ListenAddress, loggedMux)
}

// --- Middleware ---

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK} // Default to 200 OK
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs request details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use custom response writer to capture status code
		rw := newResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		// Log request details after processing
		duration := time.Since(start)
		log.Printf("Request: %s %s | Status: %d | Duration: %s | Source: %s",
			r.Method, r.URL.Path, rw.statusCode, duration, r.RemoteAddr)
	})
}

// --- Handlers ---

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

// HandleLLMGuidance serves the embedded LLM guidance text file with replaced placeholders.
func (s *Server) HandleLLMGuidance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use embedded content directly
	contentBytes := []byte(LLMServerHelpContent)

	// Build entity type documentation
	var entityDocs strings.Builder
	for _, et := range s.cfg.EntityTypes {
		// Note: This placeholder replacement logic might become obsolete or need rethinking
		// if the server help text no longer needs dynamic parts.
		entityDocs.WriteString(fmt.Sprintf("*   **%s**: %s \n", et.Name, et.Description))
	}

	// Replace placeholder
	outputBytes := bytes.Replace(contentBytes, []byte("{{ENTITY_TYPES_DOCUMENTATION}}"), []byte(entityDocs.String()), 1)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(outputBytes)
}

// HandleSummary returns a JSON array of ItemSummary for all valid items.
func (s *Server) HandleSummary(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /summary")
	if r.Method != http.MethodGet {
		log.Printf("Method Not Allowed for /summary: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	allValidItems := s.store.Query(map[string]interface{}{})
	log.Printf("Retrieved %d valid items from store for /summary", len(allValidItems))

	summaries := make([]content.ItemSummary, 0, len(allValidItems))
	for _, item := range allValidItems {
		// Use the new utility function to get the prefixed ID
		itemID, err := content.GetItemID(item)
		if err != nil {
			log.Printf("Warning: Skipping item for summary: could not get ID for %s: %v", item.SourcePath, err)
			continue
		}

		// Description and Tags logic remains the same
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
			ID:          itemID,
			Type:        item.EntityType,
			Tier:        item.Tier, // Will be empty if not a behavior or not inferred
			Tags:        tags,
			Description: description,
		})
	}

	log.Printf("Prepared %d summaries for /summary response", len(summaries))
	s.writeJSONResponse(w, http.StatusOK, summaries)
}

// --- Request/Response Structs ---

type DetailsRequest struct {
	IDs []string `json:"ids"`
}

// HandleDetails returns full details for requested item IDs.
// NOTE: This uses POST instead of GET for pragmatic reasons. While GET is semantically
// correct for data retrieval, sending a potentially large list of IDs is cleaner
// and avoids potential URL length limits when passed in the request body.
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

	// Fetch items directly by ID using the optimized store method
	foundItems := make([]*content.Item, 0, len(req.IDs))
	for _, id := range req.IDs {
		if item, found := s.store.GetByID(id); found {
			// Optionally double-check IsValid here, although GetByID fetches directly
			// if item.IsValid { // Query already filters by IsValid, GetByID does not implicitly
			// Let's check IsValid status before adding to results, as GetByID bypasses Query filters
			if item.IsValid {
				foundItems = append(foundItems, item)
			} else {
				log.Printf("Warning: Requested ID '%s' found but corresponds to an invalid item ('%s'). Skipping.", id, item.SourcePath)
			}
		} else {
			// Log IDs that were requested but not found (optional)
			log.Printf("Warning: Requested ID '%s' not found in store.", id)
		}
	}

	// --- Old Iteration Logic (Removed) ---
	// allValidItemsForDetails := s.store.Query(map[string]interface{}{}) // Fetch fresh for details
	// foundItems := make([]*content.Item, 0)
	// requestedIDsSet := make(map[string]bool)
	// for _, id := range req.IDs {
	// 	requestedIDsSet[id] = true
	// }
	//
	// for _, item := range allValidItemsForDetails { // Use the fresh list
	// 	itemDetailID, err := content.GetItemID(item) // Use the canonical ID function
	// 	if err != nil {
	// 		// Cannot generate an ID for this item, so cannot match it.
	// 		log.Printf("Warning: Skipping item for details: could not get ID for %s: %v", item.SourcePath, err)
	// 		continue
	// 	}
	//
	// 	// Check if this generated canonical ID was requested
	// 	if _, found := requestedIDsSet[itemDetailID]; found { // Check map lookup directly
	// 		foundItems = append(foundItems, item)
	// 		delete(requestedIDsSet, itemDetailID) // Mark the canonical ID as found
	// 	}
	// }
	// --- End Old Logic ---

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
