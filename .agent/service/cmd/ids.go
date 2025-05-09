package cmd

import (
	"agentt/internal/filter"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// idsCmd represents the ids command
var idsCmd = &cobra.Command{
	Use:   "ids",
	Short: "Get a list of entity IDs, optionally filtered",
	Long: `Retrieves a JSON array of entity IDs based on the provided filter criteria.
If no filter is provided, all entity IDs are returned.

Example:
  agentt ids --filter "type:behavior tier:must"
  agentt ids -f "tag:core"`,
	RunE: runIds,
}

// Flag for the filter string
var idsFilter string

func init() {
	// Add the required filter flag
	// It reuses the filterQuery variable defined in details.go/root.go
	idsCmd.Flags().StringVarP(&idsFilter, "filter", "f", "", "Filter entities by attributes (e.g., 'type:behavior tier:must tag:scope:core')")
	_ = idsCmd.MarkFlagRequired("filter") // Mark as required for this command

	rootCmd.AddCommand(idsCmd) // Add the command back to rootCmd
}

// runIds executes the ids command logic.
func runIds(cmd *cobra.Command, args []string) error {
	// Get verbosity level from root command's persistent flag
	verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")

	seenSummaryIDs := make(map[string]bool)
	matchingIDs := make(map[string]bool)

	// Always need the backend to get summaries
	backendInstance, _, err := GetBackendAndConfig(verbosity) // Ignore config
	if err != nil {
		return fmt.Errorf("failed to get backend: %w", err)
	}
	if backendInstance == nil {
		return fmt.Errorf("backend instance is nil")
	}

	allSummaries, err := backendInstance.GetSummary()
	if err != nil {
		return fmt.Errorf("failed to fetch summaries: %w", err)
	}

	var parsedFilter filter.FilterNode // Assuming FilterNode is the type returned by parseFilterString
	if idsFilter != "" {
		parsedFilter, err = filter.ParseFilter(idsFilter)
		if err != nil {
			return fmt.Errorf("invalid filter string: %w", err)
		}
	}

	for _, summary := range allSummaries {
		if _, exists := seenSummaryIDs[summary.ID]; exists {
			continue // Avoid duplicates
		}
		seenSummaryIDs[summary.ID] = true

		if parsedFilter == nil || parsedFilter.Evaluate(summary) {
			matchingIDs[summary.ID] = true
		}
	}

	// Convert matching IDs map to slice
	finalIDs := make([]string, 0, len(matchingIDs))
	for id := range matchingIDs {
		finalIDs = append(finalIDs, id)
	}

	// Marshal the IDs to JSON
	jsonData, err := json.MarshalIndent(finalIDs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal IDs to JSON: %w", err)
	}

	// Print the JSON output
	fmt.Println(string(jsonData))
	return nil
}
