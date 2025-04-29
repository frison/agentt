package main

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/content"
	"agent-guidance-service/internal/discovery"
	"agent-guidance-service/internal/store"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Define command structure
const (
	cmdDiscover    = "discover"
	cmdEntityTypes = "entity-types"
	cmdValidate    = "validate"
)

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
		executeDiscover(cfg, *outputFormat, remainingArgs)
	case cmdEntityTypes:
		executeEntityTypes(cfg, *outputFormat, remainingArgs)
	case cmdValidate:
		executeValidate(cfg, *outputFormat, remainingArgs)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// --- Command Implementations ---

func executeDiscover(cfg *config.ServiceConfig, format string, args []string) {
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
	results, err := scanAndQuery(cfg, filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during discovery: %v\n", err)
		os.Exit(1)
	}

	printOutput(results, format)
}

func executeEntityTypes(cfg *config.ServiceConfig, format string, args []string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s command takes no arguments\n\n", cmdEntityTypes)
		printUsage()
		os.Exit(1)
	}
	printOutput(cfg.EntityTypes, format)
}

func executeValidate(cfg *config.ServiceConfig, format string, args []string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s command takes no arguments\n\n", cmdValidate)
		printUsage()
		os.Exit(1)
	}

	// Perform scan, return all items (including invalid)
	allItems, err := scanAndQuery(cfg, nil) // No filters, scan everything
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during validation scan: %v\n", err)
		os.Exit(1)
	}

	// Collect validation results
	validationResults := make([]map[string]interface{}, 0, len(allItems))
	for _, item := range allItems {
		validationResults = append(validationResults, map[string]interface{}{
			"sourcePath":       item.SourcePath,
			"entityType":       item.EntityType,
			"isValid":          item.IsValid,
			"validationErrors": item.ValidationErrors,
		})
	}

	printOutput(validationResults, format)
}

// --- Helper Functions ---

// scanAndQuery performs an initial scan and optional query.
// If filters is nil or empty, returns all scanned items.
func scanAndQuery(cfg *config.ServiceConfig, filters map[string]interface{}) ([]*content.Item, error) {
	guidanceStore := store.NewGuidanceStore()
	// Use discovery package directly for one-off scan
	// Create a temporary watcher just for the scan logic (or refactor scan logic out)
	wchr, err := discovery.NewWatcher(cfg, guidanceStore)
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
	return guidanceStore.Query(filters), nil
}

// printOutput prints data in the specified format.
func printOutput(data interface{}, format string) {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	case "text":
		// Basic text output - needs refinement for different data types
		fmt.Printf("%v\n", data) // TODO: Improve text formatting
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown output format '%s'\n", format)
		os.Exit(1)
	}
}

// printUsage prints CLI usage information in a structured way.
func printUsage() {
	fmt.Fprintf(os.Stderr, "== Agent Guidance CLI ==\n\n")
	fmt.Fprintf(os.Stderr, "Purpose: Interact with agent guidance definitions (behaviors, recipes, etc.).\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  agent-guidance-cli [global options] <command> [command options]\n\n")
	fmt.Fprintf(os.Stderr, "Global Options:\n")
	flag.PrintDefaults() // Prints flags defined above main
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  %-15s Discover guidance items by type and filters.\n", cmdDiscover)
	fmt.Fprintf(os.Stderr, "                  Args: <entityType> [filterKey=filterValue...]\n")
	fmt.Fprintf(os.Stderr, "                  Example: discover behavior tier=must tag=core\n")
	fmt.Fprintf(os.Stderr, "  %-15s List configured entity types.\n", cmdEntityTypes)
	fmt.Fprintf(os.Stderr, "  %-15s Scan all configured paths and report validation status.\n", cmdValidate)
	fmt.Fprintf(os.Stderr, "\nOutput Format (--format):
")
	fmt.Fprintf(os.Stderr, "  json: Output results as a JSON object/array (Default).
")
	fmt.Fprintf(os.Stderr, "  text: Output results in a basic text format (Structure varies by command).
")
}