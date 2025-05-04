package cmd

// "agentt/internal/config" // Unused after refactor
// "agentt/internal/discovery" // Unused after refactor
// "agentt/internal/store" // Unused after refactor
import (
	// "agentt/internal/guidance/backend" // REMOVED - Unused
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"agentt/internal/guidance/backend"
	"github.com/spf13/cobra"
)

var entityIDs []string

// detailsCmd represents the details command
var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Displays full details for specified entity IDs",
	Long: `Retrieves and outputs the full details (including body content) for one or
more guidance entities specified by their IDs using the --id flag.

Details are fetched from all configured backend(s).
Output is a JSON array containing the full entity details for each ID found.
If an ID is found in multiple backends, a warning is logged, but only the first
instance found is included in the output.`, // Updated long description
	Args: cobra.NoArgs, // Ensure no positional args, IDs must come from flags
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(entityIDs) == 0 {
			return errors.New("at least one --id flag must be provided")
		}

		// Backend initialization is now handled by rootCmd.PersistentPreRunE
		if len(globalBackendService) == 0 {
			return fmt.Errorf("internal error: no backend service available")
		}
		slog.Info("Fetching details from initialized backends", "backend_count", len(globalBackendService), "requested_ids", entityIDs)

		allEntities := make([]backend.Entity, 0, len(entityIDs))
		foundIDs := make(map[string]string) // Map ID -> source backend info

		for i, service := range globalBackendService {
			slog.Debug("Fetching details from backend", "index", i, "ids", entityIDs)
			entities, err := service.GetDetails(entityIDs)
			if err != nil {
				slog.Error("Failed to get details from a backend", "index", i, "error", err)
				// Continuing for now, but logging the error.
				continue
			}
			slog.Debug("Received details from backend", "index", i, "count", len(entities))

			// Merge entities and check for duplicates
			for _, entity := range entities {
				if existingSource, exists := foundIDs[entity.ID]; exists {
					// Found duplicate ID from different backends
					slog.Warn("Duplicate entity ID found across backends (details)",
						"id", entity.ID,
						"source1", existingSource,
						"source2", fmt.Sprintf("backend %d", i),
					)
					// Skip adding the duplicate, keep the first one found.
				} else {
					allEntities = append(allEntities, entity)
					foundIDs[entity.ID] = fmt.Sprintf("backend %d", i) // Record that this ID was found
				}
			}
		}

		slog.Info("Total entity details collected", "count", len(allEntities))

		// Check if all requested IDs were found
		if len(foundIDs) != len(entityIDs) {
			missingIDs := []string{}
			requestedMap := make(map[string]bool)
			for _, id := range entityIDs {
				requestedMap[id] = true
			}
			for reqID := range requestedMap {
				if _, found := foundIDs[reqID]; !found {
					missingIDs = append(missingIDs, reqID)
				}
			}
			slog.Warn("Some requested entity IDs were not found", "missing_ids", missingIDs)
			// Do not error out, just return the ones that were found.
		}

		// Output the combined details as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ") // Pretty print
		if err := encoder.Encode(allEntities); err != nil {
			slog.Error("Failed to encode details to JSON", "error", err)
			return fmt.Errorf("failed to encode details to JSON: %w", err)
		}

		return nil
	},
}

func init() {
	// Define the --id flag (can be used multiple times)
	detailsCmd.Flags().StringSliceVar(&entityIDs, "id", []string{}, "Entity ID to retrieve details for (required, use multiple times for multiple IDs)")
	_ = detailsCmd.MarkFlagRequired("id") // Mark as required

	// rootCmd.AddCommand(detailsCmd) // AddCommand is now done in root.go's init
}
