package cmd

// "agentt/internal/config" // Unused after refactor
// "agentt/internal/discovery" // Unused after refactor
// "agentt/internal/store" // Unused after refactor
import (
	"agentt/internal/content"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	detailsIDs []string // Slice to store the IDs passed via flags
	// detailsConfigPath string // REMOVED - Use rootConfigPath from root.go
)

// detailsCmd represents the details command
var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Outputs the full JSON details for specific guidance entities by ID.",
	Long: `Outputs the full JSON details for one or more specific guidance entities, identified by their unique IDs.
Configuration is loaded via --config flag, AGENTT_CONFIG env var, or default search paths.`,
	Args: cobra.MinimumNArgs(1), // Require at least one ID
	RunE: func(cmd *cobra.Command, args []string) error {
		requestedIDs := args // IDs are passed as arguments

		// --- Use common setup ---
		setupRes, err := setupDiscovery(rootConfigPath)
		if err != nil {
			return err // Errors already formatted by helper
		}

		// --- Retrieve Details using GetByID ---
		results := make([]*content.Item, 0, len(requestedIDs))
		for _, requestedID := range requestedIDs {
			item, found := setupRes.Store.GetByID(requestedID)
			if !found {
				log.Printf("Warning: ID '%s' not found in store.", requestedID)
				continue // Skip IDs not found
			}
			// Optionally: Check if the found item is valid if needed, although GetByID fetches directly.
			// if !item.IsValid { ... }
			results = append(results, item)
		}
		log.Printf("Retrieved details for %d/%d requested IDs", len(results), len(requestedIDs))

		// --- Marshal to JSON ---
		// Output the full item details
		outputJSON, err := json.MarshalIndent(results, "", "  ")
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
	// REMOVED config flag definition - Now persistent on root
	// detailsCmd.Flags().StringVarP(&detailsConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")
}
