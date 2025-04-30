package cmd

import (
	"agentt/internal/config"
	"agentt/internal/content" // Import content package
	"agentt/internal/discovery"
	"agentt/internal/store"
	"encoding/json"
	"fmt"
	"log"

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
		// --- Configuration ---
		// Use rootConfigPath directly from root.go
		cfg, loadedPath, err := config.FindAndLoadConfig(rootConfigPath)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		log.Printf("Using configuration file: %s", loadedPath) // Log which config was used

		// --- Setup Dependencies ---
		guidanceStore := store.NewGuidanceStore()
		// Pass the *loaded* config path to the watcher for correct relative glob resolution
		watcher, err := discovery.NewWatcher(cfg, guidanceStore, loadedPath)
		if err != nil {
			return fmt.Errorf("failed to create discovery watcher: %w", err)
		}

		// --- Load Entities via Initial Scan ---
		err = watcher.InitialScan() // Populates the guidanceStore
		if err != nil {
			log.Printf("Warning during initial scan: %v", err) // Log it as well
			return fmt.Errorf("error during initial scan of guidance files: %w", err)
		}

		// --- Retrieve All Items from Store ---
		allItems := guidanceStore.GetAll()

		// --- Prepare Summary Data ---
		summaryData := prepareSummary(allItems)

		// --- Marshal to JSON ---
		outputJSON, err := json.MarshalIndent(summaryData, "", "  ")
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

	// REMOVED flag definition - Now persistent on root
	// summaryCmd.Flags().StringVarP(&summaryConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")
}

// prepareSummary converts a slice of full content Items into ItemSummary structs.
func prepareSummary(items []*content.Item) []content.ItemSummary {
	summaries := make([]content.ItemSummary, 0, len(items))
	for _, item := range items {
		if !item.IsValid {
			continue // Skip invalid items for summary
		}

		// Use the new utility function to get the prefixed ID
		prefixedID, err := content.GetPrefixedID(item)
		if err != nil {
			log.Printf("Warning: Skipping item for summary: %v", err)
			continue
		}

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
			ID:          prefixedID, // Use the prefixed ID
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
