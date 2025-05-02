package cmd

// "agentt/internal/config" // Unused after refactor
// "agentt/internal/discovery" // Unused after refactor
// "agentt/internal/store" // Unused after refactor
import (
	// "agentt/internal/guidance/backend" // REMOVED - Unused
	"encoding/json"
	"fmt"
	"log/slog"

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
	// Args: cobra.MinimumNArgs(1), // REMOVE: We are using flags, not positional args for IDs
	RunE: func(cmd *cobra.Command, args []string) error {
		// requestedIDs := args // OLD: IDs were passed as arguments
		requestedIDs := detailsIDs // NEW: Use the slice populated by --id flags

		// --- Add check for empty IDs from flags ---
		if len(requestedIDs) == 0 {
			return fmt.Errorf("at least one --id flag must be provided")
		}

		// --- Use common setup ---
		setupRes, err := setupDiscovery(rootConfigPath)
		if err != nil {
			return err // Errors already formatted by helper
		}

		// --- Retrieve Details from Backend ---
		// Use setupRes.Backend instead of setupRes.Store
		// Call GetDetails once with all requested IDs
		entities, err := setupRes.Backend.GetDetails(requestedIDs)
		if err != nil {
			slog.Error("Failed to retrieve details from backend", "error", err)
			return fmt.Errorf("failed to retrieve details: %w", err)
		}
		// Update log message
		slog.Info("Retrieved details from backend", "found_count", len(entities), "requested_count", len(requestedIDs))

		// --- Marshal to JSON ---
		// Use the entities slice directly returned by the backend
		// The backend.Entity struct matches the desired output format
		outputJSON, err := json.MarshalIndent(entities, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal details data to JSON", "error", err)
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
	// Config flag is persistent on root
	// detailsCmd.Flags().StringVarP(&detailsConfigPath, "config", "c", "", "Path to the configuration file (overrides AGENTT_CONFIG env var and default search paths)")
}
