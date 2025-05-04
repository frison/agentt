package cmd

import (
	"agentt/internal/filter"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	// "agentt/internal/guidance/backend"
	"github.com/spf13/cobra"
)

var (
// Re-use filterQuery variable declared in details.go (or root.go eventually)
// idsFilterQuery string // No need for a separate variable
)

// idsCmd represents the ids command
var idsCmd = &cobra.Command{
	Use:   "ids",
	Short: "Lists the IDs of guidance entities matching a filter query",
	Long: `Scans the configured backend(s) for guidance entities and outputs a JSON
array containing only the IDs of entities that match the provided filter query.

This is useful for obtaining a list of relevant entity IDs based on criteria
without retrieving the full summaries.

The --filter flag is required.
Example:
  agentt ids --filter "tier:must tag:scope:core -type:recipe"
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Validate Flags ---
		if filterQuery == "" {
			return errors.New("the --filter flag is required for the ids command")
		}

		// --- Backend Initialization Check ---
		if len(globalBackendService) == 0 {
			return fmt.Errorf("internal error: no backend service available")
		}
		slog.Info("Fetching summaries to extract IDs based on filter", "backend_count", len(globalBackendService), "query", filterQuery)

		// --- Parse Filter ---
		parsedFilter, err := filter.ParseFilter(filterQuery)
		if err != nil {
			slog.Error("Failed to parse filter query", "query", filterQuery, "error", err)
			return fmt.Errorf("failed to parse filter query: %w", err)
		}
		if parsedFilter == nil {
			// An empty/trivial filter would match everything.
			slog.Warn("Empty or trivial filter parsed, potentially listing all entity IDs.")
			// Let it proceed, it will list all IDs.
		} else {
			slog.Info("Applying filter to find matching IDs", "parsed", parsedFilter.String())
		}

		// --- Fetch Summaries and Filter IDs ---
		matchingIDs := make([]string, 0)
		seenIDs := make(map[string]bool) // Use map to prevent duplicate IDs in output

		for i, service := range globalBackendService {
			slog.Debug("Fetching summary from backend for ID extraction", "index", i)
			summaries, err := service.GetSummary()
			if err != nil {
				slog.Error("Failed to get summary from a backend", "index", i, "error", err)
				continue // Skip this backend on error
			}

			for _, summary := range summaries {
				if _, seen := seenIDs[summary.ID]; seen {
					continue // Already added this ID from another backend
				}

				passedFilter := true
				if parsedFilter != nil {
					passedFilter = parsedFilter.Evaluate(summary)
				}

				if passedFilter {
					matchingIDs = append(matchingIDs, summary.ID)
					seenIDs[summary.ID] = true // Mark as seen
				}
			}
		}

		slog.Info("Total matching IDs found", "count", len(matchingIDs))

		// --- Output IDs as JSON ---
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ") // Pretty print JSON array
		if err := encoder.Encode(matchingIDs); err != nil {
			slog.Error("Failed to encode IDs to JSON", "error", err)
			return fmt.Errorf("failed to encode IDs to JSON: %w", err)
		}

		return nil
	},
}

func init() {
	// Add the required filter flag
	// It reuses the filterQuery variable defined in details.go/root.go
	idsCmd.Flags().StringVar(&filterQuery, "filter", "", "Filter query to select entities (required)")
	_ = idsCmd.MarkFlagRequired("filter") // Mark as required for this command

	// Add command to root later in root.go
}
