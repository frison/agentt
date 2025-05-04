package cmd

import (
	// "agentt/internal/config" // Unused after refactor
	// "agentt/internal/content" // REMOVED
	// "agentt/internal/discovery" // Unused after refactor
	// "agentt/internal/store" // Unused after refactor
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"agentt/internal/guidance/backend" // Add this import back
	// "agentt/internal/guidance/backend" // REMOVED - Unused
	"github.com/spf13/cobra"
)

var (
// summaryConfigPath string // REMOVED - Use rootConfigPath from root.go
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Displays a summary of all discovered guidance entities (behaviors, recipes)",
	Long: `Scans the configured backend(s) for guidance entities and outputs a JSON
array summarizing each valid entity found.

The summary includes the entity ID, type, tier (if applicable), description, and tags.
Duplicate entity IDs found across different backends will be noted with a warning.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Backend initialization is now handled by rootCmd.PersistentPreRunE
		if len(globalBackendService) == 0 {
			return fmt.Errorf("internal error: no backend service available")
		}
		slog.Info("Fetching summaries from initialized backends", "backend_count", len(globalBackendService))

		allSummaries := make([]backend.Summary, 0)
		seenIDs := make(map[string]string) // Map ID -> source backend info (e.g., type/name)

		for i, service := range globalBackendService {
			slog.Debug("Fetching summary from backend", "index", i)
			summaries, err := service.GetSummary()
			if err != nil {
				slog.Error("Failed to get summary from a backend", "index", i, "error", err)
				// Decide whether to fail or just continue with results from other backends
				// Continuing for now, but logging the error.
				continue
			}
			slog.Debug("Received summaries from backend", "index", i, "count", len(summaries))

			// Merge summaries and check for duplicates
			for _, summary := range summaries {
				if existingSource, exists := seenIDs[summary.ID]; exists {
					// Found duplicate ID
					slog.Warn("Duplicate entity ID found across backends",
						"id", summary.ID,
						"source1", existingSource,
						"source2", fmt.Sprintf("backend %d", i), // Improve source info if possible
					)
					// Decide strategy: skip duplicate, merge, error? Skipping for now.
				} else {
					allSummaries = append(allSummaries, summary)
					seenIDs[summary.ID] = fmt.Sprintf("backend %d", i) // Store source info
				}
			}
		}

		slog.Info("Total summaries collected", "count", len(allSummaries))

		// Output the combined summary as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ") // Pretty print
		if err := encoder.Encode(allSummaries); err != nil {
			slog.Error("Failed to encode summary to JSON", "error", err)
			return fmt.Errorf("failed to encode summary to JSON: %w", err)
		}

		return nil
	},
}

func init() {
	// No command-specific flags for summary currently
	// rootCmd.AddCommand(summaryCmd) // AddCommand is now done in root.go's init
}

// Placeholder: Define SummaryItem structure based on Phase 1
// type SummaryItem struct {
//	 ID          string   `json:"id"`
//	 Type        string   `json:"type"` // "behavior" or "recipe"
//	 Tier        string   `json:"tier,omitempty"` // "must" or "should" for behaviors
//	 Tags        []string `json:"tags"`
//	 Description string   `json:"description"`
//}

// Placeholder: Mapping function if needed
// func PrepareSummary(entities []loader.Entity) []SummaryItem {
// 	// ... logic ...
// }
