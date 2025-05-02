package server

import (
	"agentt/internal/config"
	// "agentt/internal/content" // REMOVE if unused after refactor
	// "agentt/internal/store" // REMOVED
	"agentt/internal/guidance/backend" // ADDED
	// "bytes" // REMOVED - Unused
	_ "embed" // Import embed
	"encoding/json"
	"fmt"
	// "log" // REMOVED (use slog)
	"log/slog" // ADDED
	"net/http"
	// "strings" // REMOVED - Unused
	"time" // Add time package
)

//go:embed llm_server_help.txt
var LLMServerHelpContent string // Embedded server protocol/help text (Exported for testing)

// Server wraps the HTTP server dependencies and handlers.
type Server struct {
	cfg *config.ServiceConfig
	// store *store.GuidanceStore // REPLACED
	backend backend.GuidanceBackend // ADDED
	// Adding a testing.T field ONLY for debugging the Query issue
	// t *testing.T // REMOVE THIS AFTER DEBUGGING
}

// NewServer creates a new HTTP server instance.
// Adjust signature if adding testing.T
func NewServer(cfg *config.ServiceConfig, backend backend.GuidanceBackend) *Server {
	return &Server{
		cfg:     cfg,
		backend: backend, // Use backend
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

	// Use slog
	slog.Info("Starting HTTP server", "address", s.cfg.ListenAddress)
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

// LoggingMiddleware logs request details using slog.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		// Use slog
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", duration,
			"source", r.RemoteAddr,
		)
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

// HandleLLMGuidance serves the embedded LLM guidance text file.
func (s *Server) HandleLLMGuidance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simplified: Remove placeholder replacement as it seems outdated/unused
	contentBytes := []byte(LLMServerHelpContent)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(contentBytes)
}

// HandleSummary returns a JSON array of backend.Summary for all entities.
func (s *Server) HandleSummary(w http.ResponseWriter, r *http.Request) {
	// Use slog
	slog.Debug("Received request", "path", r.URL.Path, "method", r.Method)
	if r.Method != http.MethodGet {
		slog.Warn("Method not allowed", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Call backend GetSummary
	summaries, err := s.backend.GetSummary()
	if err != nil {
		slog.Error("Failed to get summaries from backend", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("Prepared summaries for response", "count", len(summaries), "path", r.URL.Path)
	s.writeJSONResponse(w, http.StatusOK, summaries) // Return summaries directly
}

// --- Request/Response Structs ---

type DetailsRequest struct {
	IDs []string `json:"ids"`
}

// HandleDetails returns full details for requested item IDs using the backend.
func (s *Server) HandleDetails(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Received request", "path", r.URL.Path, "method", r.Method)
	if r.Method != http.MethodPost {
		slog.Warn("Method not allowed", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request body for details", "error", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.IDs) == 0 {
		slog.Warn("Empty IDs array in details request")
		http.Error(w, "Request body must contain a non-empty 'ids' array", http.StatusBadRequest)
		return
	}
	slog.Info("Processing details request", "requested_ids_count", len(req.IDs))

	// Call backend GetDetails
	entities, err := s.backend.GetDetails(req.IDs)
	if err != nil {
		slog.Error("Failed to get details from backend", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log how many were found vs requested
	slog.Info("Returning details response", "found_count", len(entities), "requested_count", len(req.IDs))

	s.writeJSONResponse(w, http.StatusOK, entities) // Return entities directly
}

// writeJSONResponse is a helper to marshal data to JSON and write the response.
func (s *Server) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal JSON response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}
