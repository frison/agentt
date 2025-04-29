package main

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/content"
	"agent-guidance-service/internal/discovery"
	"agent-guidance-service/internal/store"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Define command structure
const (
	cmdDiscover    = "discover"
	cmdEntityTypes = "entity-types"
	cmdValidate    = "validate"
)

// validationResult holds simplified validation info for output.
type validationResult struct {
	SourcePath       string   `json:"sourcePath"`
	EntityType       string   `json:"entityType"`
	IsValid          bool     `json:"isValid"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

func main() {
	// --- Common Flags ---
	configPath := flag.String("config", ".agent/service/config.yaml", "Path to the configuration file.")
	outputFormat := flag.String("format", "json", "Output format (json, text). Default: json")
	flag.Usage = printUsage // Custom usage function for AI friendliness
	flag.Parse()

	// --- Load Config ---
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration from %s: %v\n", *configPath, err)
		os.Exit(1)
	}

	// --- Command Dispatching ---
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	remainingArgs := args[1:]

	switch command {
	case cmdDiscover:
		executeDiscover(cfg, *outputFormat, remainingArgs, *configPath)
	case cmdEntityTypes:
		executeEntityTypes(cfg, *outputFormat, remainingArgs)
	case cmdValidate:
		executeValidate(cfg, *outputFormat, remainingArgs, *configPath)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// --- Command Implementations ---

func executeDiscover(cfg *config.ServiceConfig, format string, args []string, configPath string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: %s command requires entity type (e.g., behavior, recipe)\n\n", cmdDiscover)
		printUsage()
		os.Exit(1)
	}
	entityType := args[0]
	filterArgs := args[1:] // Treat remaining args as key=value filters

	// Validate entity type
	validEntityType := false
	for _, et := range cfg.EntityTypes {
		if et.Name == entityType {
			validEntityType = true
			break
		}
	}
	if !validEntityType {
		fmt.Fprintf(os.Stderr, "Error: Unknown entity type '%s'\n", entityType)
		os.Exit(1)
	}

	// Build filters (key=value)
	filters := make(map[string]interface{})
	filters["entityType"] = entityType
	for _, filterArg := range filterArgs {
		parts := strings.SplitN(filterArg, "=", 2)
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Ignoring invalid filter argument '%s' (must be key=value)\n", filterArg)
		}
	}

	// Perform scan & query
	results, err := scanAndQuery(cfg, filters, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during discovery: %v\n", err)
		os.Exit(1)
	}

	printOutput(cmdDiscover, results, format)
}

func executeEntityTypes(cfg *config.ServiceConfig, format string, args []string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s command takes no arguments\n\n", cmdEntityTypes)
		printUsage()
		os.Exit(1)
	}
	printOutput(cmdEntityTypes, cfg.EntityTypes, format)
}

func executeValidate(cfg *config.ServiceConfig, format string, args []string, configPath string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s command takes no arguments\n\n", cmdValidate)
		printUsage()
		os.Exit(1)
	}

	// Perform scan, return all items (including invalid)
	allitems, err := scanAndQuery(cfg, nil, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during validation scan: %v\n", err)
		os.Exit(1)
	}

	// Collect validation results (uses package-level struct now)
	validationResults := make([]validationResult, 0, len(allitems))
	for _, item := range allitems {
		validationResults = append(validationResults, validationResult{
			SourcePath:       item.SourcePath,
			EntityType:       item.EntityType,
			IsValid:          item.IsValid,
			ValidationErrors: item.ValidationErrors,
		})
	}

	printOutput(cmdValidate, validationResults, format)
}

// --- Helper Functions ---

// scanAndQuery performs an initial scan and optional query.
// If filters is nil or empty, returns all scanned items.
func scanAndQuery(cfg *config.ServiceConfig, filters map[string]interface{}, configPath string) ([]*content.Item, error) {
	guidanceStore := store.NewGuidanceStore()
	// Use discovery package directly for one-off scan
	// Create a temporary watcher just for the scan logic (or refactor scan logic out)
	wchr, err := discovery.NewWatcher(cfg, guidanceStore, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize scanner: %w", err)
	}

	err = wchr.InitialScan() // This populates the store
	if err != nil {
		// Log? Return error? For CLI, maybe return error is better.
		return nil, fmt.Errorf("error during initial scan: %w", err)
	}

	if filters == nil || len(filters) == 0 {
		return guidanceStore.GetAll(), nil
	}
	return guidanceStore.Query(filters, nil), nil
}

// printOutput prints data in the specified format, tailored by command.
func printOutput(command string, data interface{}, format string) {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	case "text":
		printTextOutput(command, data)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown output format '%s'\n", format)
		os.Exit(1)
	}
}

// printTextOutput handles formatting for the 'text' output option.
func printTextOutput(command string, data interface{}) {
	switch command {
	case cmdEntityTypes:
		if types, ok := data.([]config.EntityTypeDefinition); ok {
			fmt.Println("Configured Entity Types:")
			for _, et := range types {
				fmt.Printf("  - Name: %s\n", et.Name)
				fmt.Printf("    Description: %s\n", et.Description)
				fmt.Printf("    PathGlob: %s\n", et.PathGlob)
				fmt.Printf("    Required: %v\n", et.RequiredFrontMatter)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unexpected data type for text output of %s\n", command)
		}
	case cmdValidate:
		// Uses the simplified validationResult struct defined in executeValidate
		if results, ok := data.([]validationResult); ok {
			fmt.Println("Validation Results:")
			for _, res := range results {
				status := "[VALID]"
				if !res.IsValid {
					status = "[INVALID]"
				}
				fmt.Printf("%s %s (%s)\n", status, res.SourcePath, res.EntityType)
				if !res.IsValid {
					for _, errMsg := range res.ValidationErrors {
						fmt.Printf("    - Error: %s\n", errMsg)
					}
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unexpected data type for text output of %s\n", command)
		}
	case cmdDiscover:
		if items, ok := data.([]*content.Item); ok {
			if len(items) == 0 {
				fmt.Println("No matching items found.")
				return
			}
			fmt.Printf("Discovered %d item(s):\n", len(items))
			for _, item := range items {
				title, _ := item.FrontMatter["title"].(string)
				id, _ := item.FrontMatter["id"].(string)
				identifier := title // Default to title
				if item.EntityType == "recipe" && id != "" {
					identifier = id
				}
				tierInfo := ""
				if item.Tier != "" {
					tierInfo = fmt.Sprintf(" [%s]", item.Tier)
				}
				fmt.Printf("- %s%s: %s\n", identifier, tierInfo, item.SourcePath)
				// Optionally print more frontmatter details here
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unexpected data type for text output of %s\n", command)
		}
	default:
		// Fallback for any unknown command passed (shouldn't happen)
		fmt.Fprintf(os.Stderr, "Warning: Unknown command '%s' for text output. Defaulting to basic print.\n", command)
		fmt.Printf("%v\n", data)
	}
}

// printUsage prints CLI usage information in a structured way.
func printUsage() {
	// Use backticks for multi-line strings
	fmt.Fprintf(os.Stderr, `== Agent Guidance CLI ==

Purpose: Interact with agent guidance definitions (behaviors, recipes, etc.).

Usage:
  agent-guidance-cli [global options] <command> [command options]

Global Options:
`) // End first part
	flag.PrintDefaults() // Prints flags defined above main
	fmt.Fprintf(os.Stderr, `
Commands:
`) // Start next part
	fmt.Fprintf(os.Stderr, "  %-15s Discover guidance items by type and filters.\n", cmdDiscover)
	fmt.Fprintf(os.Stderr, "                  Args: <entityType> [filterKey=filterValue...]\n")
	fmt.Fprintf(os.Stderr, "                  Example: discover behavior tier=must tag=core\n")
	fmt.Fprintf(os.Stderr, "  %-15s List configured entity types.\n", cmdEntityTypes)
	fmt.Fprintf(os.Stderr, "  %-15s Scan all configured paths and report validation status.\n", cmdValidate)
	fmt.Fprintf(os.Stderr, `
Output Format (--format):
`) // Start next part
	fmt.Fprintf(os.Stderr, "  json: Output results as a JSON object/array (Default).\n")
	fmt.Fprintf(os.Stderr, "  text: Output results in a basic text format (Structure varies by command).\n")
}
