package cmd

import (
	"agentt/internal/config"                           // Added import for config package
	guidanceBackend "agentt/internal/guidance/backend" // Aliased import
	"fmt"
	"io"
	"log/slog" // To be used later
	"os"       // To be used later
	"strings"

	"github.com/spf13/cobra"
)

// TODO: Define flags for createCmd (e.g., backend, id, title, type, tags, body-from-file, body-stdin)
// var createBackendTarget string
var (
	// entityType is captured as an argument
	entityID            string
	entityTitle         string
	entityTier          string
	entityDescription   string
	entityTags          []string
	bodyFromFile        string
	outputDir           string
	forceOverwrite      bool
	backendTarget       string   // Added for specifying backend
	entityMetadataPairs []string // For -d key=value pairs
)

var createCmd = &cobra.Command{
	Use:   "create <type>",
	Short: "Creates a new guidance entity (behavior or recipe)",
	Long: `Creates a new guidance entity (behavior or recipe).
Uses a specified backend if --backend-target is given, otherwise the default writable backend.
Requires metadata flags (like --id, --title) and body input.`,
	Args: cobra.ExactArgs(1), // Type will be the first argument
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verbosity for logging setup is handled by initializeSharedState called by Get...Backend funcs
		// We still get it here in case we need to make logging decisions specific to this command.
		verbosity, _ := cmd.Root().PersistentFlags().GetCount("verbose")
		_ = verbosity // আপাতত ব্যবহার করা হচ্ছে না (Currently unused, but keep for potential future use)

		var writableBackend guidanceBackend.WritableBackend
		var err error

		if backendTarget != "" {
			slog.Debug("Attempting to get named writable backend", "target", backendTarget)
			writableBackend, err = GetNamedWritableBackend(backendTarget)
		} else {
			slog.Debug("Attempting to get default writable backend")
			writableBackend, err = GetDefaultWritableBackend()
		}

		if err != nil {
			return fmt.Errorf("failed to get a writable backend: %w", err)
		}
		// At this point, writableBackend is guaranteed to be non-nil and a WritableBackend if err is nil.

		// Config is loaded globally by the Get...Backend functions, access if needed for entityTypeConf
		if globalConfig == nil { // Should be populated by initializeSharedState
			return fmt.Errorf("internal error: global config not available after backend initialization")
		}

		entityTypeArg := args[0]

		var entityTypeConf *config.EntityType // Changed to pointer
		validType := false
		for i, et := range globalConfig.EntityTypes {
			if et.Name == entityTypeArg {
				entityTypeConf = &globalConfig.EntityTypes[i] // Store pointer
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("invalid entity type specified: %s. Valid types are: %v", entityTypeArg, getValidEntityTypeNames(globalConfig.EntityTypes))
		}
		if entityTypeConf == nil { // Should not happen if validType is true
			return fmt.Errorf("internal error: entityTypeConf is nil despite validType being true for %s", entityTypeArg)
		}

		// Get entityData map from flags.
		entityData := make(map[string]interface{})
		entityData["id"] = entityID        // From flag, already marked as required
		entityData["type"] = entityTypeArg // From argument

		if entityTitle != "" {
			entityData["title"] = entityTitle
		}
		if entityDescription != "" {
			entityData["description"] = entityDescription
		}
		if entityTier != "" {
			if entityTypeArg == "behavior" {
				if entityTier != "must" && entityTier != "should" {
					return fmt.Errorf("invalid tier '%s' for behavior. Must be 'must' or 'should'", entityTier)
				}
			} else {
				slog.Debug("Tier provided for non-behavior type", "type", entityTypeArg, "tier", entityTier)
			}
			entityData["tier"] = entityTier
		}
		if len(entityTags) > 0 {
			entityData["tags"] = entityTags
		}

		// Process general key-value pairs from --data flags
		for _, pair := range entityMetadataPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 || parts[0] == "" {
				slog.Warn("Skipping invalid --data entry", "entry", pair)
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key == "id" || key == "type" {
				slog.Warn("Skipping --data entry for reserved key", "key", key, "value", value)
				continue
			}
			entityData[key] = value
			slog.Debug("Added/updated entity data from --data flag", "key", key, "value", value)
		}

		// Get body string from flags (--body-from-file, --body-stdin, or piped).
		bodyContent := "" // Default empty body
		if bodyFromFile != "" {
			bodyBytes, err := os.ReadFile(bodyFromFile)
			if err != nil {
				return fmt.Errorf("failed to read body from file '%s': %w", bodyFromFile, err)
			}
			bodyContent = string(bodyBytes)
		} else {
			slog.Debug("No --body-from-file specified, attempting to read from stdin.")
			stdinBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read body from stdin: %w", err)
			}
			if len(stdinBytes) > 0 {
				bodyContent = string(stdinBytes)
				slog.Debug("Read body from stdin", "length", len(bodyContent))
			} else {
				slog.Warn("No body content provided via --body-from-file or stdin.")
			}
		}

		// Validate entityData against entityTypeConf.RequiredFields.
		for _, reqField := range entityTypeConf.RequiredFields {
			if _, exists := entityData[reqField]; !exists {
				// This is a simplified check. A more robust validation might be needed
				// depending on how optional flags map to required fields.
				// For now, relying on backend to do final deep validation.
				return fmt.Errorf("missing required field '%s' for type '%s'. Please provide it via a specific flag or --data %s=<value>", reqField, entityTypeArg, reqField)
			}
		}

		// Call WritableBackend.CreateEntity(entityData, bodyContent).
		slog.Debug("Calling CreateEntity", "entityData", entityData, "bodyLength", len(bodyContent), "force", forceOverwrite)
		if err := writableBackend.CreateEntity(entityData, bodyContent, forceOverwrite); err != nil {
			return fmt.Errorf("failed to create entity: %w", err)
		}

		fmt.Printf("Successfully created entity '%s' of type '%s'.\n", entityID, entityTypeArg)
		return nil
	},
}

// Helper function to get valid entity type names for error messages
func getValidEntityTypeNames(entityTypes []config.EntityType) []string {
	names := make([]string, len(entityTypes))
	for i, et := range entityTypes {
		names[i] = et.Name
	}
	return names
}

func init() {
	// TODO: Add flags to createCmd here.
	// createCmd.Flags().StringVar(&createBackendTarget, "backend", "", "Name or index of the target writable backend (required if multiple writable backends exist)")
	createCmd.Flags().StringVar(&backendTarget, "backend-target", "", "Name of the target backend to use (if multiple are configured and writable)")

	createCmd.Flags().StringVar(&entityID, "id", "", "Unique ID of the new entity (required)")
	if err := createCmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Errorf("failed to mark 'id' flag as required: %w", err))
	}

	createCmd.Flags().StringVar(&entityTitle, "title", "", "Title of the entity")
	createCmd.Flags().StringVar(&entityTier, "tier", "", "Tier of the behavior (e.g., 'must', 'should')")
	createCmd.Flags().StringVar(&entityDescription, "desc", "", "Short description of the entity")
	createCmd.Flags().StringSliceVarP(&entityTags, "tags", "t", []string{}, "Comma-separated tags for the entity (e.g., tag1,tag2)")
	createCmd.Flags().StringSliceVarP(&entityMetadataPairs, "data", "d", []string{}, "Arbitrary key=value data for entity frontmatter (can be repeated)")
	createCmd.Flags().StringVar(&bodyFromFile, "body-from-file", "", "Path to a file containing the body content for the entity")
	createCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Directory to create the entity file in (defaults to backend configuration)")
	createCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing entity file if it exists")

	// createCmd.Flags().StringVar(&createEntityID, "id", "", "ID of the new entity (required)")
	// createCmd.MarkFlagRequired("id") // Example
	// ... other flags ...

	rootCmd.AddCommand(createCmd)
}
