package server

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/store"
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
}

// NewServer creates a new HTTP server instance.
func NewServer(cfg *config.ServiceConfig, store *store.GuidanceStore) *Server {
	return &Server{
		cfg:   cfg,
		store: store,
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/entityTypes", s.handleEntityTypes)
	mux.HandleFunc("/discover/", s.handleDiscover) // Note the trailing slash for path parameters
	mux.HandleFunc("/llm.txt", s.handleLLMGuidance)

	log.Printf("Starting HTTP server on %s...", s.cfg.ListenAddress)
	return http.ListenAndServe(s.cfg.ListenAddress, mux)
}

// handleHealth returns a simple 200 OK.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleEntityTypes returns the configured entity types.
func (s *Server) handleEntityTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, s.cfg.EntityTypes)
}

// handleDiscover returns discovered items, filtered by entity type and query params.
func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract entity type from path: /discover/{entityType}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/discover/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Missing entity type in path (e.g., /discover/behavior)", http.StatusBadRequest)
		return
	}
	entityType := pathParts[0]

	// Validate entity type against config
	validEntityType := false
	for _, et := range s.cfg.EntityTypes {
		if et.Name == entityType {
			validEntityType = true
			break
		}
	}
	if !validEntityType {
		http.Error(w, fmt.Sprintf("Unknown entity type: %s", entityType), http.StatusNotFound)
		return
	}

	// Build filters from query parameters
	filters := make(map[string]interface{})
	filters["entityType"] = entityType // Always filter by path entity type

	queryParams := r.URL.Query()
	for key, values := range queryParams {
		if len(values) > 0 {
			// Use the first value for simplicity. Handle multi-value params if needed.
			filters[key] = values[0]
		}
	}

	// Query the store
	results := s.store.Query(filters)

	s.writeJSONResponse(w, http.StatusOK, results)
}

// handleLLMGuidance serves the LLM guidance text file with replaced placeholders.
func (s *Server) handleLLMGuidance(w http.ResponseWriter, r *http.Request) {
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
		entityDocs.WriteString(fmt.Sprintf("*   **%s**: %s (Discover via `/discover/%s`)
", et.Name, et.Description, et.Name))
	}

	// Replace placeholder
	outputBytes := bytes.Replace(contentBytes, []byte("{{ENTITY_TYPES_DOCUMENTATION}}"), []byte(entityDocs.String()), 1)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(outputBytes)
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