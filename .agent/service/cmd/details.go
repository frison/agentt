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

	// "agentt/internal/guidance/backend" // REMOVED: Unused
	"github.com/spf13/cobra"
)

var (
	entityIDs   []string
	filterQuery string // Filter query string
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
		// Get verbosity level from root command's persistent flag
		verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")

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

		// --- Get Backend (Initialized by Root PersistentPreRunE) --- START ---
		// Call GetBackendAndConfig again here to get the instance for this command's execution.
		backendInstance, _, err := GetBackendAndConfig(verbosity) // Ignore config here
		if err != nil {
			return fmt.Errorf("failed to get backend instance: %w", err)
		}
		if backendInstance == nil {
			return fmt.Errorf("internal error: backend instance is nil after initialization")
		}
		// --- Get Backend --- END ---

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
				slog.Warn("Empty or trivial filter parsed, fetching all details.")
			}

			// Fetch all summaries first using the obtained backend instance
			matchingIDs := make(map[string]bool)
			seenSummaryIDs := make(map[string]string) // Keep track of where ID was first seen
			slog.Debug("Fetching all summaries to apply filter")

			allSummaries, err := backendInstance.GetSummary()
			if err != nil {
				// Log the error, but potentially continue if MultiBackend returned partial results?
				// For now, fail if summaries can't be retrieved.
				slog.Error("Failed to get summaries from backend (for filter)", "error", err)
				return fmt.Errorf("failed to fetch summaries for filtering: %w", err)
			}

			for _, summary := range allSummaries {
				// Check for duplicates across backends (handled by MultiBackend logging, just need to avoid processing twice here)
				if _, exists := seenSummaryIDs[summary.ID]; exists {
					continue
				}
				// NOTE: MultiBackend.GetSummary already logs duplicate warnings. This map is just to ensure we only evaluate the filter once per unique ID summary.
				seenSummaryIDs[summary.ID] = "seen" // Mark as processed in this filter stage

				if parsedFilter == nil || parsedFilter.Evaluate(summary) {
					matchingIDs[summary.ID] = true
				}
			} // End loop through summaries

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
		} // --- Handle Filter Query --- END ---

		// --- Handle Case: No IDs from flags, no filter provided --- START ---
		// This case should now be implicitly handled: if !hasIDs and !hasFilter, error is returned at the start.
		// If hasFilter but no matches, idsToFetch is empty and we return above.
		// If hasIDs, idsToFetch is populated initially.
		// Therefore, we only need to check if idsToFetch is empty *after* potential filtering.
		if len(idsToFetch) == 0 {
			// This should only happen if --id flags were provided but were empty, or filter matched nothing.
			slog.Info("No entity IDs specified or found matching filter.")
			fmt.Println("[]") // Output empty JSON array
			return nil
		}
		// --- Handle Case: No IDs from flags, no filter provided --- END ---

		slog.Info("Fetching details from backend", "requested_ids_count", len(idsToFetch))

		// Fetch details using the SAME backend instance obtained earlier
		allEntities, err := backendInstance.GetDetails(idsToFetch)
		if err != nil {
			// Log the error, but MultiBackend might return partial results
			slog.Error("Error encountered while fetching details from backend", "error", err)
			// Decide if we should still attempt to output partial results or fail hard.
			// Let's try outputting what we got.
			// return fmt.Errorf("failed to fetch details: %w", err) // Option to fail hard
		}

		slog.Info("Total entity details collected", "count", len(allEntities))

		// Check if all requested IDs were found (relevant if --id was used)
		if hasIDs {
			foundIDsMap := make(map[string]bool)
			for _, entity := range allEntities {
				foundIDsMap[entity.ID] = true
			}
			missingIDs := []string{}
			for _, reqID := range idsToFetch { // idsToFetch contains original --id list here
				if !foundIDsMap[reqID] {
					missingIDs = append(missingIDs, reqID)
				}
			}
			if len(missingIDs) > 0 {
				slog.Warn("Some requested entity IDs were not found", "missing_ids", missingIDs)
			}
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

	// Add the filter flag
	detailsCmd.Flags().StringVarP(&filterQuery, "filter", "f", "", "Filter entities using a query (e.g., 'type:behavior tier:must tag:api')")

	// CliCmd.AddCommand(detailsCmd) // Remove from CliCmd
	rootCmd.AddCommand(detailsCmd) // Add the command back to rootCmd
}
