package cmd

// "agentt/internal/config" // Unused after refactor
// "agentt/internal/discovery" // Unused after refactor
// "agentt/internal/store" // Unused after refactor
import (
	// "agentt/internal/guidance/backend" // REMOVED - Unused
	"agentt/internal/filter" // Import filter package
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"agentt/internal/guidance/backend"
	"github.com/spf13/cobra"
)

var (
	entityIDs   []string
	filterQuery string // Add variable for filter flag
)

// detailsCmd represents the details command
var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Displays full details for specified entity IDs or entities matching a filter",
	Long: `Retrieves and outputs the full details (including body content) for guidance entities.

You can specify entities either by providing one or more --id flags OR by providing
a --filter query.

If using --filter, the command first finds summaries matching the filter across all
backends, then retrieves details for those specific IDs.
Examples:
  agentt details --id bhv-code-style --id rcp-git-commit
  agentt details --filter "tier:must tag:scope:core"

Details are fetched from all configured backend(s).
Output is a JSON array containing the full entity details for each ID found.
If an ID is found in multiple backends, a warning is logged, but only the first
instance found is included in the output.`, // Updated long description
	Args: cobra.NoArgs, // Ensure no positional args, IDs/filter must come from flags
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Validate Flags --- START ---
		hasIDs := len(entityIDs) > 0
		hasFilter := filterQuery != ""

		if !hasIDs && !hasFilter {
			return errors.New("either one or more --id flags or a --filter flag must be provided")
		}
		if hasIDs && hasFilter {
			return errors.New("cannot use both --id flags and --filter flag simultaneously")
		}
		// --- Validate Flags --- END ---

		// Backend initialization is now handled by rootCmd.PersistentPreRunE
		if len(globalBackendService) == 0 {
			return fmt.Errorf("internal error: no backend service available")
		}

		idsToFetch := entityIDs // Default to using --id flags if provided

		// --- Handle Filter Query --- START ---
		if hasFilter {
			slog.Info("Filter query provided, finding matching IDs first", "query", filterQuery)
			parsedFilter, err := filter.ParseFilter(filterQuery)
			if err != nil {
				slog.Error("Failed to parse filter query", "query", filterQuery, "error", err)
				return fmt.Errorf("failed to parse filter query: %w", err)
			}
			if parsedFilter == nil {
				// Empty filter might mean match all? Or should be an error?
				// Let's assume it means match all summaries for now to get all details.
				slog.Warn("Empty or trivial filter parsed, potentially fetching all details. Consider a more specific filter.")
				// Proceed without specific IDs means GetDetails needs to handle empty list? No, fetch all summaries.
			}

			// Fetch all summaries first
			matchingIDs := make(map[string]bool) // Use map to handle duplicates implicitly
			seenSummaryIDs := make(map[string]string)
			slog.Debug("Fetching all summaries to apply filter")

			for i, service := range globalBackendService {
				summaries, err := service.GetSummary()
				if err != nil {
					slog.Error("Failed to get summary from a backend (for filter)", "index", i, "error", err)
					continue // Skip this backend
				}
				for _, summary := range summaries {
					if _, exists := seenSummaryIDs[summary.ID]; exists {
						continue // Already processed this ID from another backend
					}
					seenSummaryIDs[summary.ID] = fmt.Sprintf("backend %d", i)

					if parsedFilter == nil || parsedFilter.Evaluate(summary) {
						matchingIDs[summary.ID] = true
					}
				}
			}

			// Convert map keys to slice for GetDetails
			idsToFetch = make([]string, 0, len(matchingIDs))
			for id := range matchingIDs {
				idsToFetch = append(idsToFetch, id)
			}

			if len(idsToFetch) == 0 {
				slog.Info("No entities found matching the filter query", "query", filterQuery)
				fmt.Println("[]") // Output empty JSON array
				return nil        // Successful exit, just no results
			}
			slog.Info("Found matching IDs via filter", "count", len(idsToFetch), "ids", idsToFetch)
		}
		// --- Handle Filter Query --- END ---

		slog.Info("Fetching details from initialized backends", "backend_count", len(globalBackendService), "requested_ids", idsToFetch)

		allEntities := make([]backend.Entity, 0, len(idsToFetch))
		foundIDs := make(map[string]string) // Map ID -> source backend info

		for i, service := range globalBackendService {
			slog.Debug("Fetching details from backend", "index", i, "ids", idsToFetch)
			entities, err := service.GetDetails(idsToFetch)
			if err != nil {
				slog.Error("Failed to get details from a backend", "index", i, "error", err)
				// Continuing for now, but logging the error.
				continue
			}
			slog.Debug("Received details from backend", "index", i, "count", len(entities))

			// Merge entities and check for duplicates
			for _, entity := range entities {
				if existingSource, exists := foundIDs[entity.ID]; exists {
					// Found duplicate ID from different backends
					slog.Warn("Duplicate entity ID found across backends (details)",
						"id", entity.ID,
						"source1", existingSource,
						"source2", fmt.Sprintf("backend %d", i),
					)
					// Skip adding the duplicate, keep the first one found.
				} else {
					allEntities = append(allEntities, entity)
					foundIDs[entity.ID] = fmt.Sprintf("backend %d", i) // Record that this ID was found
				}
			}
		}

		slog.Info("Total entity details collected", "count", len(allEntities))

		// Check if all requested IDs were found (relevant if --id was used)
		if hasIDs && len(foundIDs) != len(idsToFetch) {
			missingIDs := []string{}
			requestedMap := make(map[string]bool)
			for _, id := range idsToFetch { // idsToFetch contains the original --id list if hasIDs is true
				requestedMap[id] = true
			}
			for reqID := range requestedMap {
				if _, found := foundIDs[reqID]; !found {
					missingIDs = append(missingIDs, reqID)
				}
			}
			slog.Warn("Some requested entity IDs were not found", "missing_ids", missingIDs)
			// Do not error out, just return the ones that were found.
		}

		// Output the combined details as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ") // Pretty print
		if err := encoder.Encode(allEntities); err != nil {
			slog.Error("Failed to encode details to JSON", "error", err)
			return fmt.Errorf("failed to encode details to JSON: %w", err)
		}

		return nil
	},
}

func init() {
	// Define the --id flag (can be used multiple times)
	detailsCmd.Flags().StringSliceVar(&entityIDs, "id", []string{}, "Entity ID to retrieve details for (use multiple times for multiple IDs)")
	// Removed MarkFlagRequired("id") - now optional if --filter is used

	// Add the filter flag
	detailsCmd.Flags().StringVar(&filterQuery, "filter", "", "Filter entities using a query. Supported syntax: key:value, key:*, -key:value. Implicit AND. E.g., 'tier:must tag:scope:core -type:recipe'")

	// rootCmd.AddCommand(detailsCmd) // AddCommand is now done in root.go's init
}
