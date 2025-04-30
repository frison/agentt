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

const defaultSummaryConfigPath = "config.yaml"

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Outputs a JSON summary of all guidance entities (behaviors and recipes).",
	Long: `Outputs a JSON summary of all guidance entities (behaviors and recipes).
This includes minimal information like ID, type, tags, and description,
suitable for initial discovery by an agent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Configuration ---
		// TODO: Add a flag for config path, using default for now
		configPath := defaultSummaryConfigPath
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration from %s: %w", configPath, err)
		}

		// --- Setup Dependencies ---
		guidanceStore := store.NewGuidanceStore()
		// The watcher needs the config path to resolve relative globs correctly
		watcher, err := discovery.NewWatcher(cfg, guidanceStore, configPath)
		if err != nil {
			// Log fatal similar to server? Or return error for CLI?
			// For CLI, returning error is usually better.
			return fmt.Errorf("failed to create discovery watcher: %w", err)
		}

		// --- Load Entities via Initial Scan ---
		err = watcher.InitialScan() // Populates the guidanceStore
		if err != nil {
			// Log potentially non-fatal errors from scan? Or just return?
			// Let's return the error for now, signifies loading failed.
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

	// TODO: Add flag for --config path
	// summaryCmd.Flags().StringVarP(&configPath, "config", "c", defaultSummaryConfigPath, "Path to the configuration file.")
}

// prepareSummary converts a slice of full content Items into ItemSummary structs.
func prepareSummary(items []*content.Item) []content.ItemSummary {
	summaries := make([]content.ItemSummary, 0, len(items))
	for _, item := range items {
		if !item.IsValid {
			continue // Skip invalid items for summary
		}

		// Extract common fields
		id := getStringFromFrontMatter(item.FrontMatter, "id", item.SourcePath) // Use SourcePath as fallback ID?
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

		// >>> Phase 5 TODO: Add ID prefixing here <<<
		// Example (Needs refinement):
		// if item.EntityType == "behavior" {
		// 	id = "bhv-" + id
		// } else if item.EntityType == "recipe" {
		// 	id = "rcp-" + id
		// }

		summaries = append(summaries, content.ItemSummary{
			ID:          id,
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
