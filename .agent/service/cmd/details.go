package cmd

import (
	"agentt/internal/config"
	"agentt/internal/content"
	"agentt/internal/discovery"
	"agentt/internal/store"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	detailsIDs []string // Slice to store the IDs passed via flags
)

// detailsCmd represents the details command
var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Outputs the full JSON details for specific guidance entities by ID.",
	Long: `Outputs the full JSON details for one or more specified guidance entities (behaviors or recipes).
Provide the entity IDs using the --id flag (can be repeated). IDs should match those
returned by the 'summary' command (including prefixes like 'bhv-' or 'rcp-' once implemented).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(detailsIDs) == 0 {
			return fmt.Errorf("at least one --id flag must be provided")
		}

		// --- Configuration ---
		// TODO: Use flag for config path if added to summaryCmd
		configPath := "config.yaml" // Reusing const from summary.go for consistency (MANUALLY UPDATED)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration from %s: %w", configPath, err)
		}

		// --- Setup Dependencies & Load ---
		guidanceStore := store.NewGuidanceStore()
		watcher, err := discovery.NewWatcher(cfg, guidanceStore, configPath)
		if err != nil {
			return fmt.Errorf("failed to create discovery watcher: %w", err)
		}
		err = watcher.InitialScan() // Populates the guidanceStore
		if err != nil {
			log.Printf("Warning during initial scan: %v", err)
			return fmt.Errorf("error during initial scan of guidance files: %w", err)
		}

		// --- Retrieve Specific Items by ID ---
		foundItems := make([]*content.Item, 0, len(detailsIDs))

		// Option 1: Iterate and check ID manually (requires ID extraction logic)
		// This approach mirrors prepareSummary logic and handles fallback/prefixing later.
		allItems := guidanceStore.GetAll()
		requestedIDMap := make(map[string]bool)
		for _, id := range detailsIDs {
			requestedIDMap[id] = true
		}

		for _, item := range allItems {
			if !item.IsValid {
				continue
			}
			// Extract ID from frontmatter (using helper from summary.go logic)
			// >>> Phase 5 TODO: Apply prefixing *before* lookup OR handle it here <<<
			itemID := getStringFromFrontMatter(item.FrontMatter, "id", item.SourcePath) // Fallback needed?

			if requestedIDMap[itemID] {
				foundItems = append(foundItems, item)
			}
		}

		// Option 2: Use store.Query (if ID is reliably in frontmatter["id"])
		// Requires ensuring the store's query logic handles the 'id' key correctly.
		// for _, requestedID := range detailsIDs {
		// 	// >>> Phase 5 TODO: Ensure requestedID includes prefix if needed by query <<<
		// 	results := guidanceStore.Query(map[string]interface{}{"id": requestedID})
		// 	if len(results) > 0 {
		// 		foundItems = append(foundItems, results[0]) // Assume ID is unique
		// 	}
		// }

		// --- Marshal to JSON ---
		// Output the full item details
		outputJSON, err := json.MarshalIndent(foundItems, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal details to JSON: %w", err)
		}

		// --- Print JSON to stdout ---
		fmt.Println(string(outputJSON))

		return nil // Return nil on success
	},
}

func init() {
	// Add detailsCmd directly to the rootCmd.
	rootCmd.AddCommand(detailsCmd)

	// Define the repeatable --id flag
	detailsCmd.Flags().StringSliceVar(&detailsIDs, "id", []string{}, "Entity ID to get details for (repeatable)")
	// Mark the flag as required? Cobra doesn't easily support *requiring* repeatable flags.
	// We added a check in RunE instead.
}
