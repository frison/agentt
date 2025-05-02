package cmd

import (
	// "agentt/internal/config" // Unused after refactor
	"agentt/internal/content" // Import content package
	// "agentt/internal/discovery" // Unused after refactor
	// "agentt/internal/store" // Unused after refactor
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
// summaryConfigPath string // REMOVED - Use rootConfigPath from root.go
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Outputs a JSON summary of all guidance entities (behaviors and recipes).",
	Long: `Outputs a JSON summary of all guidance entities (behaviors and recipes).
This includes minimal information like ID, type, tags, and description,
suitable for initial discovery by an agent.
Configuration is loaded via --config flag, AGENTT_CONFIG env var, or default search paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Use common setup ---
		setupRes, err := setupDiscovery(rootConfigPath)
		if err != nil {
			return err // Errors already formatted by helper
		}

		// --- Retrieve All Items from Store ---
		allItems := setupRes.Store.GetAll()
		// Use slog.Info
		slog.Info("Retrieved items from store", "count", len(allItems))

		// --- Prepare Summary Data ---
		summaries := prepareSummary(allItems)

		// --- Marshal to JSON ---
		outputJSON, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal summary to JSON: %w", err)
		}

		// --- Print JSON to stdout ---
		fmt.Println(string(outputJSON))

		return nil // Return nil on success
	},
}

func init() {
	// Add summaryCmd directly to the rootCmd.
	rootCmd.AddCommand(summaryCmd)

	// Config flag is now persistent on root command
}

// prepareSummary converts a slice of full content Items into ItemSummary structs.
// It requires items to have an explicit, non-empty "id" field in their frontmatter.
// Items without a valid "id" are skipped, and a warning is logged.
func prepareSummary(items []*content.Item) []content.ItemSummary {
	summaries := make([]content.ItemSummary, 0, len(items))
	for _, item := range items {
		if !item.IsValid {
			continue // Skip invalid items for summary
		}

		// --- CHANGE: Require explicit 'id' in frontmatter ---
		var itemID string
		if item.FrontMatter != nil {
			if idVal, ok := item.FrontMatter["id"].(string); ok && idVal != "" {
				itemID = idVal // Use the explicit ID
			} else {
				slog.Warn("Skipping item for summary: missing or invalid 'id' field", "path", item.SourcePath) // Use slog.Warn
				continue                                                                                       // Skip item if 'id' is missing or not a non-empty string
			}
		} else {
			slog.Warn("Skipping item for summary: missing frontmatter (required for 'id' field)", "path", item.SourcePath) // Use slog.Warn
			continue                                                                                                       // Skip item if frontmatter is missing
		}
		// --- END CHANGE ---

		// Description and Tags logic remains the same
		description := getStringFromFrontMatter(item.FrontMatter, "description", "")
		var tags []string
		if tagsInterface, ok := item.FrontMatter["tags"].([]interface{}); ok {
			tags = make([]string, 0, len(tagsInterface))
			for _, t := range tagsInterface {
				if tagStr, okStr := t.(string); okStr {
					tags = append(tags, tagStr)
				}
			}
		}

		summaries = append(summaries, content.ItemSummary{
			ID:          itemID, // Use the validated explicit ID
			Type:        item.EntityType,
			Tier:        item.Tier, // Will be empty if not a behavior
			Tags:        tags,
			Description: description,
		})
	}
	return summaries
}

// Helper function to safely get string from frontmatter map
func getStringFromFrontMatter(fm map[string]interface{}, key string, defaultValue string) string {
	if val, ok := fm[key]; ok {
		if strVal, okStr := val.(string); okStr {
			return strVal
		}
	}
	return defaultValue
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
