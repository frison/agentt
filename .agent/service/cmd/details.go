package cmd

import (
	"agentt/internal/filter"
	"agentt/internal/guidance/backend"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	entityIDs   []string
	filterQuery string
)

var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Displays full details for specified entity IDs or entities matching a filter",
	Long: `Retrieves and outputs the full details (including body content) for guidance entities.

You can specify entities either by providing one or more --id flags OR by providing
a --filter query.

If using --filter, the command first finds summaries matching the filter across all
backends, then retrieves details for those specific IDs.
Examples:
  agentt details --id code-style --id git-commit
  agentt details --filter "tier:must tag:scope:core"

Details are fetched from all configured backend(s).
Output is a JSON array containing the full entity details for each ID found.`,
	Args: cobra.NoArgs,
	RunE: runDetails,
}

func getIDsFromFilter(backend backend.GuidanceBackend, filterStr string) ([]string, error) {
	parsedFilter, err := filter.ParseFilter(filterStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter query: %w", err)
	}

	summaries, err := backend.GetSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summaries for filtering: %w", err)
	}

	matchingIDs := make(map[string]bool)
	seenIDs := make(map[string]bool)

	for _, summary := range summaries {
		if seenIDs[summary.ID] {
			continue
		}
		seenIDs[summary.ID] = true

		if parsedFilter == nil || parsedFilter.Evaluate(summary) {
			matchingIDs[summary.ID] = true
		}
	}

	result := make([]string, 0, len(matchingIDs))
	for id := range matchingIDs {
		result = append(result, id)
	}
	return result, nil
}

func outputEmptyJSON() error {
	fmt.Println("[]")
	return nil
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func validateFlags() error {
	hasIDs := len(entityIDs) > 0
	hasFilter := filterQuery != ""

	if !hasIDs && !hasFilter {
		return errors.New("either one or more --id flags or a --filter flag must be provided")
	}
	if hasIDs && hasFilter {
		return errors.New("cannot use both --id flags and --filter flag simultaneously")
	}
	return nil
}

func runDetails(cmd *cobra.Command, args []string) error {
	if err := validateFlags(); err != nil {
		return err
	}

	verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")

	var idsToFetch []string
	var err error // Declare err here

	// backendInstance will be fetched after determining if we even need it for filtering
	var backendInstance backend.GuidanceBackend

	if len(entityIDs) > 0 { // --id flags were used
		idsToFetch = entityIDs
	} else if filterQuery != "" { // --filter flag was used
		// Need the backend to apply the filter against summaries
		fetchedBackend, _, fetchErr := GetMultiBackendAndConfig(verbosity)
		if fetchErr != nil {
			return fmt.Errorf("failed to get backend instance for filtering: %w", fetchErr)
		}
		backendInstance = fetchedBackend // Assign to the broader scope variable
		idsToFetch, err = getIDsFromFilter(backendInstance, filterQuery)
		if err != nil {
			return err // Error from getIDsFromFilter
		}
	} else {
		// This case should be caught by validateFlags, but as a safeguard:
		return errors.New("no entity IDs or filter query provided")
	}

	if len(idsToFetch) == 0 {
		// If --id was used but resulted in no IDs (e.g. empty slice due to some prior logic error, though unlikely here)
		// or if filter yielded no IDs.
		slog.Info("No entity IDs to fetch details for (either none specified, or filter yielded no results).")
		return outputEmptyJSON()
	}

	// Ensure backendInstance is fetched if it wasn't already for filtering
	if backendInstance == nil {
		var fetchErr error
		backendInstance, _, fetchErr = GetMultiBackendAndConfig(verbosity)
		if fetchErr != nil {
			return fmt.Errorf("failed to get backend instance: %w", fetchErr)
		}
	}

	slog.Info("Fetching details from backend", "requested_ids_count", len(idsToFetch))
	allEntities, err := backendInstance.GetDetails(idsToFetch)
	// Do not return immediately on error from GetDetails; we might have partial results or want to log missing IDs.
	if err != nil {
		slog.Error("Error encountered while fetching details from backend", "error", err)
		// Depending on desired strictness, we could return err here.
		// For now, proceed to check for missing IDs even if some error occurred during fetch,
		// as some backends in a MultiBackend might succeed while others fail.
	}

	// Check for missing IDs only if --id flags were the source
	if len(entityIDs) > 0 {
		foundIDs := make(map[string]bool)
		for _, entity := range allEntities { // allEntities might be incomplete if GetDetails errored
			foundIDs[entity.ID] = true
		}

		var missingIDs []string
		// Iterate over the original entityIDs requested by flag, not potentially filtered idsToFetch
		for _, reqID := range entityIDs {
			if !foundIDs[reqID] {
				missingIDs = append(missingIDs, reqID)
			}
		}
		if len(missingIDs) > 0 {
			slog.Warn("Some requested entity IDs were not found", "missing_ids", missingIDs)
			// Potentially return an error or specific exit code if IDs are explicitly requested and not found
		}
	}

	return outputJSON(allEntities)
}

func init() {
	detailsCmd.Flags().StringSliceVar(&entityIDs, "id", []string{}, "Entity ID to retrieve details for (use multiple times for multiple IDs)")
	detailsCmd.Flags().StringVarP(&filterQuery, "filter", "f", "", "Filter entities using a query (e.g., 'type:behavior tier:must tag:api')")
	rootCmd.AddCommand(detailsCmd)
}
