package cmd

import (
	"agentt/internal/filter" // Import the filter package
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

Use the --filter flag to apply specific criteria based on entity attributes.
Examples:
  agentt summary --filter "tier:must"
  agentt summary --filter "tag:scope:core -tag:domain:ai"
  agentt summary --filter "type:recipe priority:*"

The summary includes the entity ID, type, tier (if applicable), description, and tags.
Duplicate entity IDs found across different backends will be noted with a warning.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get verbosity level from root command's persistent flag
		verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")

		// Get backend instance
		backendInstance, _, err := GetBackendAndConfig(verbosity) // Use GetBackendAndConfig, ignore config
		if err != nil {
			return fmt.Errorf("failed to get backend instance: %w", err)
		}
		if backendInstance == nil {
			return fmt.Errorf("internal error: backend instance is nil after initialization")
		}
		slog.Info("Fetching summaries from backend") // Simplified log

		// --- Parse Filter --- START ---
		var parsedFilter filter.FilterNode
		// Use simple assignment `=` for the second time err is assigned
		if filterQuery != "" {
			slog.Debug("Parsing filter query", "query", filterQuery)
			parsedFilter, err = filter.ParseFilter(filterQuery) // Use = instead of :=
			if err != nil {
				slog.Error("Failed to parse filter query", "query", filterQuery, "error", err)
				return fmt.Errorf("failed to parse filter query: %w", err)
			}
			if parsedFilter != nil {
				slog.Info("Applying filter", "parsed", parsedFilter.String())
			}
		} else {
			slog.Debug("No filter query provided")
		}
		// --- Parse Filter --- END ---

		// Slice to hold summaries after filtering
		var filteredSummaries []backend.Summary
		// seenIDs map is handled within MultiBackend now, not needed here.

		// Fetch summaries from the single (potentially aggregate) backend instance
		slog.Debug("Fetching summary from backend instance")
		summaries, err := backendInstance.GetSummary() // Use the instance
		if err != nil {
			slog.Error("Failed to get summary from backend", "error", err)
			// Decide whether to fail or just continue with potentially partial results from MultiBackend
			// Let's proceed and filter whatever we got.
		}
		slog.Debug("Received summaries from backend instance", "count", len(summaries))

		// Apply filter if necessary
		if parsedFilter != nil {
			filteredSummaries = make([]backend.Summary, 0, len(summaries))
			for _, summary := range summaries {
				if parsedFilter.Evaluate(summary) {
					filteredSummaries = append(filteredSummaries, summary)
				}
			}
		} else {
			filteredSummaries = summaries // Use all summaries if no filter
		}

		slog.Info("Total summaries collected after filtering", "count", len(filteredSummaries))

		// Output the combined summary as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")                              // Pretty print
		if err = encoder.Encode(filteredSummaries); err != nil { // Use = here too
			slog.Error("Failed to encode summary to JSON", "error", err)
			return fmt.Errorf("failed to encode summary to JSON: %w", err)
		}

		return nil
	},
}

func init() {
	// Add the filter flag
	summaryCmd.Flags().StringVar(&filterQuery, "filter", "", "Filter entities using a query. Supported syntax: key:value, key:*, -key:value. Implicit AND between terms. E.g., 'tier:must tag:scope:core -type:recipe'")
	rootCmd.AddCommand(summaryCmd) // Add the command back to rootCmd
}
