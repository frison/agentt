package localfs

import (
	"agentt/internal/config"
	"agentt/internal/guidance/backend"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var frontMatterSeparator = []byte("---")

// Error for duplicate IDs found during initialization.
var ErrDuplicateID = errors.New("duplicate entity ID detected")

// LocalFilesystemBackend implements the GuidanceBackend interface
// for loading guidance entities from the local filesystem.
type LocalFilesystemBackend struct {
	// Configuration (set during Initialize)
	rootDir           string
	behaviorGlob      string
	recipeGlob        string
	entityTypeDefs    map[string]config.EntityTypeDefinition // Store definitions by name
	behaviorDef       config.EntityTypeDefinition
	recipeDef         config.EntityTypeDefinition
	requireExplicitID bool // Flag based on config/convention

	mu          sync.RWMutex
	entities    map[string]backend.Entity // Store entities by ID
	summaries   []backend.Summary
	initialized bool
}

// NewLocalFilesystemBackend creates a new instance of LocalFilesystemBackend.
func NewLocalFilesystemBackend() *LocalFilesystemBackend {
	return &LocalFilesystemBackend{
		entities:       make(map[string]backend.Entity),
		entityTypeDefs: make(map[string]config.EntityTypeDefinition),
	}
}

// Initialize configures and loads data for the filesystem backend.
// It walks the filesystem based on configured globs, parses files,
// validates them, and stores them in memory.
// Returns ErrDuplicateID if multiple files define the same entity ID.
func (b *LocalFilesystemBackend) Initialize(configMap map[string]interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return fmt.Errorf("backend already initialized")
	}

	// --- Configuration Extraction ---
	if err := b.extractConfig(configMap); err != nil {
		return fmt.Errorf("failed to extract configuration: %w", err)
	}

	slog.Info("Initializing LocalFilesystemBackend", "rootDir", b.rootDir, "behaviors", b.behaviorGlob, "recipes", b.recipeGlob)

	// --- File Scanning and Parsing Logic ---
	foundFiles := make(map[string]string) // file path -> entity type ("behavior", "recipe")

	findFiles := func(glob string, entityType string) error {
		if glob == "" {
			slog.Debug("Skipping file scan: glob pattern is empty", "entityType", entityType)
			return nil
		}
		pattern := filepath.Join(b.rootDir, glob)
		slog.Debug("Evaluating glob pattern", "pattern", pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			slog.Warn("Error evaluating glob pattern", "pattern", pattern, "error", err)
			return fmt.Errorf("error evaluating glob pattern '%s': %w", pattern, err)
		}
		slog.Debug("Glob pattern matched files", "pattern", pattern, "count", len(matches), "matches", matches)
		for _, match := range matches {
			absPath, _ := filepath.Abs(match)
			slog.Debug("Processing matched file", "match", match, "absPath", absPath)
			if _, exists := foundFiles[absPath]; !exists {
				fi, err := os.Stat(absPath)
				if err == nil && !fi.IsDir() {
					slog.Debug("Adding file to process list", "path", absPath, "type", entityType)
					foundFiles[absPath] = entityType
				} else if err != nil {
					slog.Warn("Failed to stat matched file, skipping", "path", absPath, "error", err)
				} else {
					slog.Debug("Skipping directory matched by glob", "path", absPath)
				}
			} else {
				slog.Warn("File matched by multiple globs, using first type encountered", "path", absPath, "type", foundFiles[absPath])
			}
		}
		return nil
	}

	// Scan for behaviors and recipes
	if err := findFiles(b.behaviorGlob, "behavior"); err != nil {
		// Non-fatal, maybe only recipes were intended
		slog.Warn("Could not scan for behaviors", "error", err)
	}
	if err := findFiles(b.recipeGlob, "recipe"); err != nil {
		// Non-fatal, maybe only behaviors were intended
		slog.Warn("Could not scan for recipes", "error", err)
	}

	slog.Info("Initial file scan found potential entities", "count", len(foundFiles))

	// --- Parse and Load Files Concurrently ---
	var wg sync.WaitGroup
	parseResults := make(chan parseResult, len(foundFiles))

	for absPath, entityType := range foundFiles {
		wg.Add(1)
		go func(p string, et string) {
			defer wg.Done()
			entityDef := b.entityTypeDefs[et]
			slog.Debug("Parsing file", "path", p, "type", et)
			parsedItem, parseErr := b.parseAndValidateFile(p, et, entityDef)
			if parseErr != nil {
				slog.Warn("File parsing failed", "path", p, "error", parseErr)
			} else if parsedItem == nil {
				slog.Debug("File parsing returned nil item without error", "path", p)
			} else {
				slog.Debug("File parsing succeeded", "path", p, "id", parsedItem.ID)
			}
			parseResults <- parseResult{item: parsedItem, err: parseErr, path: p}
		}(absPath, entityType)
	}

	wg.Wait()
	close(parseResults)

	// --- Process Results and Populate Store ---
	tempEntities := make(map[string]backend.Entity)
	tempSummaries := make([]backend.Summary, 0)

	for result := range parseResults {
		slog.Debug("Processing parse result", "path", result.path, "hasError", result.err != nil, "hasItem", result.item != nil)
		if result.err != nil {
			slog.Error("Skipping file due to parse/validation error", "path", result.path, "error", result.err)
			continue
		}
		if result.item == nil {
			slog.Warn("Skipping result with nil item but no error", "path", result.path)
			continue
		}

		// Check for duplicate IDs
		if _, exists := tempEntities[result.item.ID]; exists {
			slog.Error("Duplicate entity ID detected", "id", result.item.ID, "path1", tempEntities[result.item.ID].ResourceLocator, "path2", result.path)
			return fmt.Errorf("%w: ID '%s' defined in multiple files (%s, %s)",
				ErrDuplicateID, result.item.ID, tempEntities[result.item.ID].ResourceLocator, result.path)
		}

		tempEntities[result.item.ID] = *result.item
		tempSummaries = append(tempSummaries, entityToSummary(result.item))

		slog.Debug("Successfully processed and stored entity", "id", result.item.ID, "path", result.path)
	}

	b.entities = tempEntities
	b.summaries = tempSummaries
	b.initialized = true
	slog.Info("LocalFilesystemBackend initialized successfully", "loaded_entities", len(b.entities))

	return nil
}

// parseResult holds the outcome of parsing a single file.
type parseResult struct {
	item *backend.Entity
	err  error
	path string
}

// extractConfig parses the map and sets struct fields.
func (b *LocalFilesystemBackend) extractConfig(configMap map[string]interface{}) error {
	var ok bool
	b.rootDir, ok = configMap["rootDir"].(string)
	if !ok || b.rootDir == "" {
		return fmt.Errorf("missing or invalid 'rootDir' in backend configMap")
	}
	// --- REMOVED filepath.Abs call ---
	// rootDir provided by common_setup should already be absolute/correct.
	// absRootDir, err := filepath.Abs(b.rootDir)
	// if err != nil {
	// 	return fmt.Errorf("failed to get absolute path for rootDir '%s': %w", b.rootDir, err)
	// }
	// b.rootDir = absRootDir

	// --- Extract Globs from EntityTypeDefinitions ---
	b.behaviorGlob = "" // Reset globs
	b.recipeGlob = ""

	if defsVal, exists := configMap["entityTypes"]; exists {
		if defsSlice, okSlice := defsVal.([]config.EntityTypeDefinition); okSlice {
			b.entityTypeDefs = make(map[string]config.EntityTypeDefinition) // Re-initialize map
			for _, def := range defsSlice {
				b.entityTypeDefs[def.Name] = def
				if def.Name == "behavior" {
					b.behaviorDef = def
					b.behaviorGlob = def.PathGlob // CORRECT: Extract from definition
					// slog.Debug("Extracted behavior glob", "glob", b.behaviorGlob) // DEBUG REMOVED
				}
				if def.Name == "recipe" {
					b.recipeDef = def
					b.recipeGlob = def.PathGlob // CORRECT: Extract from definition
					// slog.Debug("Extracted recipe glob", "glob", b.recipeGlob) // DEBUG REMOVED
				}
			}
		} else {
			return fmt.Errorf("invalid format for 'entityTypes' in configMap, expected []config.EntityTypeDefinition")
		}
	} else {
		return fmt.Errorf("missing 'entityTypes' definitions in configMap")
	}

	// Verify that we actually found the globs if types were expected
	if b.behaviorGlob == "" {
		slog.Warn("Behavior glob pattern is empty after config extraction.")
	}
	if b.recipeGlob == "" {
		slog.Warn("Recipe glob pattern is empty after config extraction.")
	}

	// Check if explicit ID is required
	if reqIDVal, exists := configMap["requireExplicitID"]; exists {
		if reqIDBool, okBool := reqIDVal.(bool); okBool {
			b.requireExplicitID = reqIDBool
			// slog.Debug("Extracted requireExplicitID setting", "value", b.requireExplicitID) // DEBUG REMOVED
		} else {
			slog.Warn("Invalid format for 'requireExplicitID' in configMap, expected bool")
		}
	}

	return nil
}

// parseAndValidateFile reads, parses, and validates a single guidance file.
// Adapts logic from discovery/parsing.go/ParseFile
func (b *LocalFilesystemBackend) parseAndValidateFile(absPath string, entityType string, entityDef config.EntityTypeDefinition) (*backend.Entity, error) {
	fileData, err := os.ReadFile(absPath)
	if err != nil {
		// Handle cases where file might disappear between discovery and read
		if os.IsNotExist(err) {
			return nil, nil // Not a fatal error for Initialize
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	meta := make(map[string]interface{})
	body := ""
	validationErrors := []string{}
	isValid := true

	parts := bytes.SplitN(fileData, frontMatterSeparator, 3)

	if len(parts) >= 3 && len(bytes.TrimSpace(parts[0])) == 0 { // Found frontmatter
		yamlData := parts[1]
		body = string(bytes.TrimSpace(parts[2]))

		err = yaml.Unmarshal(yamlData, &meta)
		if err != nil {
			isValid = false
			validationErrors = append(validationErrors, fmt.Sprintf("YAML parsing error: %v", err))
			// Continue validation even if YAML is broken, might catch missing ID
		}
	} else { // No valid frontmatter
		isValid = false
		validationErrors = append(validationErrors, "No valid YAML frontmatter detected")
		body = string(bytes.TrimSpace(fileData)) // Store body anyway
	}

	// --- Validation and Field Extraction ---

	// ID (Mandatory, must be extracted first)
	itemID := ""
	if idVal, ok := meta["id"].(string); ok && idVal != "" {
		itemID = idVal
	} else {
		if b.requireExplicitID { // Only enforce if required by config
			isValid = false
			validationErrors = append(validationErrors, "Missing required frontmatter key: 'id'")
			// Return error here? Or let Initialize handle duplicates? Let Initialize handle it.
		}
		// If ID is not required or missing, how to identify? Use path?
		// For now, if ID is missing and required, it's invalid.
		// If ID is missing and *not* required, we need a fallback ID strategy (e.g., hash of path/content) - NOT IMPLEMENTED YET.
		// Forcing explicit ID seems safest based on Plan 0.1.
		if itemID == "" {
			// Can't proceed without an ID to store the entity
			return nil, fmt.Errorf("missing required 'id' field in frontmatter")
		}
	}

	// Check other required frontmatter keys defined in config
	for _, requiredKey := range entityDef.RequiredFrontMatter {
		if requiredKey == "id" {
			continue
		} // Already checked
		if _, ok := meta[requiredKey]; !ok {
			isValid = false
			validationErrors = append(validationErrors, fmt.Sprintf("Missing required frontmatter key: '%s'", requiredKey))
		}
	}

	// Tier Inference (for behaviors)
	tier := ""
	if entityType == "behavior" {
		dir := filepath.Dir(absPath)
		parentDir := filepath.Base(dir)
		if parentDir == "must" || parentDir == "should" {
			tier = parentDir
		} else {
			// Mark as invalid if tier cannot be inferred and is expected?
			if _, tierRequired := meta["tier"]; !tierRequired { // Check if explicitly set
				isValid = false
				validationErrors = append(validationErrors, "Could not infer behavior tier ('must' or 'should') from path, and 'tier' not set in frontmatter")
			}
		}
		// Allow explicit frontmatter 'tier' to override inferred?
		if fmTier, ok := meta["tier"].(string); ok && (fmTier == "must" || fmTier == "should") {
			tier = fmTier
		}
	}

	// Last Updated Time (from file stat)
	lastUpdated := time.Now().UTC()
	fi, statErr := os.Stat(absPath)
	if statErr == nil {
		lastUpdated = fi.ModTime().UTC()
	}

	// Log validation errors if any
	if !isValid {
		slog.Warn("Invalid content detected in file", "path", absPath, "errors", validationErrors)
		// Decide whether to return the invalid entity or nil. Let's return it marked invalid.
	}

	entity := &backend.Entity{
		ID:              itemID,
		Type:            entityType,
		Tier:            tier, // Will be empty if not a behavior or not inferred
		Body:            body,
		ResourceLocator: absPath,
		Metadata:        meta, // Store the whole frontmatter map
		LastUpdated:     lastUpdated,
		// Add IsValid and ValidationErrors? The interface doesn't have them.
		// For now, we only store valid entities in the map returned by Initialize.
		// If we need to surface invalid items later, the interface needs adjustment.
	}

	// Only return successfully parsed and *valid* entities from this function
	// to be added to the main map in Initialize.
	if !isValid {
		// Return nil, error? Or just nil? Let's return nil, error will be logged by caller.
		// Return nil, nil - let Initialize log the warning and skip.
		return nil, fmt.Errorf("file content is invalid: %v", validationErrors)
	}

	return entity, nil
}

// GetSummary returns the summaries of all loaded entities.
func (b *LocalFilesystemBackend) GetSummary() ([]backend.Summary, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("backend not initialized")
	}

	// Return a copy to prevent external modification
	summariesCopy := make([]backend.Summary, len(b.summaries))
	copy(summariesCopy, b.summaries)
	return summariesCopy, nil
}

// GetDetails returns the details for the specified entity IDs.
func (b *LocalFilesystemBackend) GetDetails(ids []string) ([]backend.Entity, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("backend not initialized")
	}

	foundEntities := make([]backend.Entity, 0, len(ids))
	for _, id := range ids {
		if entity, ok := b.entities[id]; ok {
			foundEntities = append(foundEntities, entity)
		} else {
			slog.Warn("Entity ID not found in local filesystem backend", "id", id)
		}
	}

	return foundEntities, nil
}

// Helper function to convert Entity to Summary
func entityToSummary(e *backend.Entity) backend.Summary {
	desc := ""
	if descVal, ok := e.Metadata["description"].(string); ok {
		desc = descVal
	}
	tags := []string{}
	if tagsVal, ok := e.Metadata["tags"].([]interface{}); ok {
		for _, t := range tagsVal {
			if tagStr, okStr := t.(string); okStr {
				tags = append(tags, tagStr)
			}
		}
	}

	return backend.Summary{
		ID:          e.ID,
		Type:        e.Type,
		Tier:        e.Tier,
		Tags:        tags,
		Description: desc,
	}
}
