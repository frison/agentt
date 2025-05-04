package localfs

import (
	"agentt/internal/config"
	"agentt/internal/content"
	"agentt/internal/guidance/backend"
	"bytes"
	"errors" // Added for error wrapping
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Ensure LocalFilesystemBackend implements the GuidanceBackend interface
var _ backend.GuidanceBackend = (*LocalFilesystemBackend)(nil)

type LocalFilesystemBackend struct {
	configDir       string                        // Absolute path to the directory containing the agentt config file
	settings        config.LocalFSBackendSettings // Specific settings for this backend instance
	entityTypes     map[string]config.EntityType  // Map entity type name -> definition (for required fields)
	store           map[string]*content.Item      // In-memory store: absPath -> Item
	mu              sync.RWMutex                  // Mutex for thread-safe access to the store
	initialScanDone bool                          // Flag to track if initial scan completed
}

// NewLocalFSBackend creates and initializes a new LocalFilesystemBackend.
// It now takes specific LocalFS settings and the directory of the config file.
func NewLocalFSBackend(settings config.LocalFSBackendSettings, configFilePath string, entityTypes []config.EntityType) (*LocalFilesystemBackend, error) {
	slog.Debug("Creating new LocalFS backend", "settings", settings, "configFilePath", configFilePath)

	// Validate settings
	if settings.RootDir == "" {
		// Allow empty rootDir, defaults to config file directory
		slog.Info("LocalFS backend 'rootDir' not set, defaulting to config file directory", "configPath", configFilePath)
		settings.RootDir = "."
	}
	if len(settings.EntityLocations) == 0 {
		return nil, errors.New("localfs backend validation failed: 'entityLocations' must be defined and contain at least one entry")
	}

	// Store absolute path of the config directory for resolving rootDir
	configDir := filepath.Dir(configFilePath)

	// Ensure rootDir itself resolves correctly relative to the config dir
	// Note: filepath.Join cleans the path
	absoluteRootDir := filepath.Join(configDir, settings.RootDir)
	slog.Debug("Resolved absolute root directory for backend", "configDir", configDir, "settingsRootDir", settings.RootDir, "absoluteRootDir", absoluteRootDir)

	// Check if the resolved root directory exists
	if _, err := os.Stat(absoluteRootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("localfs backend validation failed: resolved 'rootDir' does not exist: %s (derived from config dir %s and rootDir setting %s)", absoluteRootDir, configDir, settings.RootDir)
	} else if err != nil {
		return nil, fmt.Errorf("localfs backend validation failed: error checking resolved 'rootDir' %s: %w", absoluteRootDir, err)
	}

	// Convert entity type slice to map for quick lookup
	etMap := make(map[string]config.EntityType)
	for _, et := range entityTypes {
		etMap[et.Name] = et
	}

	// Validate entityLocations keys against defined entity types
	for entityName := range settings.EntityLocations {
		if _, exists := etMap[entityName]; !exists {
			return nil, fmt.Errorf("localfs backend validation failed: 'entityLocations' contains key '%s' which is not a defined entity type in the main config", entityName)
		}
	}

	b := &LocalFilesystemBackend{
		configDir:   configDir,
		settings:    settings,
		entityTypes: etMap,
		store:       make(map[string]*content.Item),
	}

	// Perform initial scan on creation
	if err := b.scanFiles(); err != nil {
		// Log the error but allow backend creation? Or fail? Failing for now.
		slog.Error("Initial file scan failed during LocalFS backend creation", "error", err)
		return nil, fmt.Errorf("initial file scan failed: %w", err)
	}
	b.initialScanDone = true
	slog.Info("LocalFS backend created and initial scan complete")
	return b, nil
}

// GetSummary returns a summary of all valid loaded entities.
func (b *LocalFilesystemBackend) GetSummary() ([]backend.Summary, error) {
	// Ensure initial scan ran if needed (or rescan periodically?)
	// if !b.initialScanDone { ... error or trigger scan ... }
	b.mu.RLock()
	defer b.mu.RUnlock()

	summaries := make([]backend.Summary, 0, len(b.store))
	for _, item := range b.store {
		if item.IsValid { // Only include valid items in summary
			summaries = append(summaries, itemToSummary(item))
		}
	}
	slog.Debug("Retrieved backend summary", "valid_item_count", len(summaries))
	return summaries, nil
}

// GetDetails returns the full details for the requested entity IDs.
func (b *LocalFilesystemBackend) GetDetails(ids []string) ([]backend.Entity, error) {
	// Ensure initial scan ran
	b.mu.RLock()
	defer b.mu.RUnlock()

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	entities := make([]backend.Entity, 0, len(ids))
	foundCount := 0
	for _, item := range b.store {
		if item.IsValid { // Only consider valid items
			if idVal, ok := item.FrontMatter["id"].(string); ok {
				if idMap[idVal] {
					entities = append(entities, itemToEntity(item))
					foundCount++
				}
			}
		}
	}
	slog.Debug("Retrieved backend details", "requested_ids", len(ids), "found_entities", foundCount)
	return entities, nil
}

// scanFiles scans the filesystem based on configured globs and updates the internal store.
// It now resolves paths relative to the backend's configured rootDir.
func (b *LocalFilesystemBackend) scanFiles() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newStore := make(map[string]*content.Item)
	var encounteredError error
	fileCount := 0

	// Resolve the absolute root directory for this backend
	absoluteRootDir := filepath.Join(b.configDir, b.settings.RootDir)
	slog.Debug("Starting file scan", "absoluteRootDir", absoluteRootDir)

	// Iterate through the entity types defined for this backend
	for entityTypeName, globPattern := range b.settings.EntityLocations {
		entityDef, ok := b.entityTypes[entityTypeName]
		if !ok {
			slog.Error("Internal inconsistency: Entity type from entityLocations not found in entityTypes map", "entityType", entityTypeName)
			continue // Should not happen due to validation in NewLocalFSBackend
		}

		// Construct the full glob pattern relative to the absolute root dir
		fullGlobPattern := filepath.Join(absoluteRootDir, globPattern)
		slog.Debug("Scanning for entity type", "entityType", entityTypeName, "fullGlobPattern", fullGlobPattern)

		matches, err := filepath.Glob(fullGlobPattern)
		if err != nil {
			slog.Error("Error evaluating glob pattern", "pattern", fullGlobPattern, "error", err)
			if encounteredError == nil {
				encounteredError = fmt.Errorf("error evaluating glob pattern '%s': %w", fullGlobPattern, err)
			}
			continue // Try next entity type
		}

		for _, matchPath := range matches {
			fileCount++
			// Ensure we have an absolute path for storage and parsing
			absPath, err := filepath.Abs(matchPath)
			if err != nil {
				slog.Warn("Failed to get absolute path for matched file, skipping", "matchPath", matchPath, "error", err)
				if encounteredError == nil {
					encounteredError = fmt.Errorf("failed to get absolute path for '%s': %w", matchPath, err)
				}
				continue
			}

			// Check if it's a directory (filepath.Glob can return directories)
			fileInfo, err := os.Stat(absPath)
			if err != nil {
				slog.Warn("Failed to stat matched file, skipping", "path", absPath, "error", err)
				if encounteredError == nil {
					encounteredError = fmt.Errorf("failed to stat '%s': %w", absPath, err)
				}
				continue
			}
			if fileInfo.IsDir() {
				slog.Debug("Skipping directory matched by glob", "path", absPath)
				continue
			}

			slog.Debug("Attempting to load and parse file", "path", absPath, "entityType", entityTypeName)

			// Load and parse the file using the new function
			item, err := parseGuidanceFile(absPath, entityTypeName, entityDef.RequiredFields)

			if err != nil {
				slog.Warn("Failed to load or parse file", "path", absPath, "error", err)
				// Store invalid item using error details
				newStore[absPath] = &content.Item{
					SourcePath:       absPath,
					EntityType:       entityTypeName,
					IsValid:          false,
					ValidationErrors: []string{fmt.Sprintf("Error parsing file: %v", err)},
				}
			} else {
				// Check for duplicate ID (existing logic - simplified slightly)
				if itemIDStr, ok := item.FrontMatter["id"].(string); ok && itemIDStr != "" {
					duplicateFound := false
					var existingPath string
					for path, existingItem := range newStore {
						if existingID, okID := existingItem.FrontMatter["id"].(string); okID && existingID == itemIDStr {
							duplicateFound = true
							existingPath = path
							break
						}
					}
					if duplicateFound {
						slog.Warn("Duplicate entity ID detected during scan. Check guidance files.",
							"id", itemIDStr,
							"path1", existingPath,
							"path2", item.SourcePath,
						)
						continue // Skip this duplicate (keep first encountered)
					}
				}
				newStore[absPath] = item // Store the successfully parsed item
			}
		}
	}

	// Update the main store
	b.store = newStore
	slog.Info("File scan complete", "files_processed", fileCount, "valid_entities_loaded", len(b.store), "first_error", encounteredError)
	return encounteredError // Return the first error encountered, if any
}

// parseGuidanceFile reads a file, separates frontmatter from body, parses YAML,
// validates required fields, and returns a content.Item.
func parseGuidanceFile(absPath string, entityType string, requiredFields []string) (*content.Item, error) {
	fileBytes, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	frontMatterBytes, bodyBytes, err := splitFrontMatter(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to split frontmatter: %w", err)
	}

	fm := make(map[string]interface{})
	if len(frontMatterBytes) > 0 {
		err = yaml.Unmarshal(frontMatterBytes, &fm)
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
		}
	}

	// Perform validation
	isValid := true
	validationErrors := []string{}
	requiredMap := make(map[string]bool)
	for _, f := range requiredFields {
		requiredMap[f] = true
	}

	for reqField := range requiredMap {
		if _, exists := fm[reqField]; !exists {
			isValid = false
			validationErrors = append(validationErrors, fmt.Sprintf("Missing required field: '%s'", reqField))
		}
	}

	// Specifically check for non-empty string ID if required
	if requiredMap["id"] {
		if extractedID, ok := fm["id"].(string); !ok || extractedID == "" {
			isValid = false // Mark invalid if ID is required but missing/not string/empty
			if !contains(validationErrors, "Missing required field: 'id'") {
				validationErrors = append(validationErrors, "Required field 'id' is missing or not a non-empty string")
			}
		}
	}

	// Safely get tier string (important for behaviors)
	tierStr := ""
	if tierVal, ok := fm["tier"].(string); ok {
		tierStr = tierVal
	}

	// Trim leading/trailing whitespace from body
	bodyStr := strings.TrimSpace(string(bodyBytes))

	item := &content.Item{
		SourcePath:       absPath,
		EntityType:       entityType,
		FrontMatter:      fm,
		Body:             bodyStr,
		IsValid:          isValid,
		ValidationErrors: validationErrors,
		LastUpdated:      time.Now(), // Use current time for now
		Tier:             tierStr,
	}

	// If validation failed, log it but return the item anyway
	if !isValid {
		slog.Warn("Guidance file validation failed", "path", absPath, "errors", validationErrors)
	}

	return item, nil // Return the parsed item, validation status is inside the item
}

// splitFrontMatter separates YAML frontmatter (delimited by ---) from the main content.
func splitFrontMatter(data []byte) (frontMatter []byte, body []byte, err error) {
	delimiter := []byte("\n---\n")
	startDelimiter := []byte("---\n")

	if !bytes.HasPrefix(data, startDelimiter) {
		// No frontmatter detected, treat entire content as body
		return nil, data, nil
	}

	// Handle empty frontmatter case (--- followed immediately by ---)
	if bytes.HasPrefix(data[len(startDelimiter):], startDelimiter) {
		frontMatter = nil
		body = data[len(startDelimiter)*2:] // Body starts after the second ---
		return frontMatter, body, nil
	}

	// Find the end delimiter (--- preceded by newline)
	// Start searching after the initial '---'
	endIndex := bytes.Index(data[len(startDelimiter):], delimiter)
	if endIndex == -1 {
		// Found start delimiter but no end delimiter
		return nil, nil, fmt.Errorf("invalid frontmatter format: start delimiter '---' found but no proper end delimiter '\n---' detected")
	}

	// Adjust endIndex to be relative to the original data slice
	endIndex += len(startDelimiter)

	frontMatter = bytes.TrimSpace(data[len(startDelimiter):endIndex]) // Trim whitespace from FM
	body = data[endIndex+len(delimiter):]

	return frontMatter, body, nil
}

// Helper to check slice contains string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// Helper to convert content.Item to backend.Summary
// Now populates fields correctly from FrontMatter
func itemToSummary(item *content.Item) backend.Summary {
	idStr := ""
	if idVal, ok := item.FrontMatter["id"].(string); ok {
		idStr = idVal
	}
	descStr := ""
	if descVal, ok := item.FrontMatter["description"].(string); ok {
		descStr = descVal
	}
	var tags []string
	if tagsVal, ok := item.FrontMatter["tags"].([]interface{}); ok {
		for _, t := range tagsVal {
			if tagStr, okStr := t.(string); okStr {
				tags = append(tags, tagStr)
			}
		}
	} else if tagsStr, ok := item.FrontMatter["tags"].(string); ok && tagsStr != "" {
		// Handle tags specified as a single comma-separated string
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	return backend.Summary{
		ID:          idStr,
		Type:        item.EntityType,
		Tier:        item.Tier, // Tier is now a direct field on content.Item
		Description: descStr,
		Tags:        tags,
	}
}

// Helper to convert content.Item to backend.Entity
// Now populates fields correctly
func itemToEntity(item *content.Item) backend.Entity {
	idStr := ""
	if idVal, ok := item.FrontMatter["id"].(string); ok {
		idStr = idVal
	}
	return backend.Entity{
		ID:              idStr,
		Type:            item.EntityType,
		Tier:            item.Tier, // Use direct field
		Body:            item.Body, // Use direct field
		ResourceLocator: item.SourcePath,
		Metadata:        item.FrontMatter, // Pass the whole map
		LastUpdated:     item.LastUpdated, // Use direct field
	}
}
