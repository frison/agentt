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
  agentt details --id bhv-code-style --id rcp-git-commit
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
	backendInstance, _, err := GetBackendAndConfig(verbosity)
	if err != nil {
		return fmt.Errorf("failed to get backend instance: %w", err)
	}
	if backendInstance == nil {
		return fmt.Errorf("internal error: backend instance is nil after initialization")
	}

	var idsToFetch []string
	if len(entityIDs) > 0 {
		idsToFetch = entityIDs
	} else {
		idsToFetch, err = getIDsFromFilter(backendInstance, filterQuery)
		if err != nil {
			return err
		}
	}

	if len(idsToFetch) == 0 {
		return outputEmptyJSON()
	}

	slog.Info("Fetching details from backend", "requested_ids_count", len(idsToFetch))
	allEntities, err := backendInstance.GetDetails(idsToFetch)
	if err != nil {
		slog.Error("Error encountered while fetching details", "error", err)
	}

	if len(entityIDs) > 0 {
		foundIDs := make(map[string]bool)
		for _, entity := range allEntities {
			foundIDs[entity.ID] = true
		}

		var missingIDs []string
		for _, reqID := range idsToFetch {
			if !foundIDs[reqID] {
				missingIDs = append(missingIDs, reqID)
			}
		}
		if len(missingIDs) > 0 {
			slog.Warn("Some requested entity IDs were not found", "missing_ids", missingIDs)
		}
	}

	return outputJSON(allEntities)
}

func init() {
	detailsCmd.Flags().StringSliceVar(&entityIDs, "id", []string{}, "Entity ID to retrieve details for (use multiple times for multiple IDs)")
	detailsCmd.Flags().StringVarP(&filterQuery, "filter", "f", "", "Filter entities using a query (e.g., 'type:behavior tier:must tag:api')")
	rootCmd.AddCommand(detailsCmd)
}
