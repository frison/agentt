package cmd

import (
	// "agentt/internal/config" // Unused after refactor
	// "agentt/internal/content" // REMOVED
	// "agentt/internal/discovery" // Unused after refactor
	// "agentt/internal/store" // Unused after refactor
	"encoding/json"
	"fmt"
	"log/slog"

	// "agentt/internal/guidance/backend" // REMOVED - Unused
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

		// --- Retrieve All Summaries from Backend ---
		summaries, err := setupRes.Backend.GetSummary()
		if err != nil {
			slog.Error("Failed to retrieve summaries from backend", "error", err)
			return fmt.Errorf("failed to retrieve summaries: %w", err)
		}
		slog.Info("Retrieved summaries from backend", "count", len(summaries))

		// --- Marshal to JSON ---
		outputJSON, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal summary data to JSON", "error", err)
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
