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
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// Ensure LocalFilesystemBackend implements the GuidanceBackend interface
var _ backend.GuidanceBackend = (*LocalFilesystemBackend)(nil)
var _ backend.WritableBackend = (*LocalFilesystemBackend)(nil) // Ensure it implements WritableBackend

type LocalFilesystemBackend struct {
	configDir       string                        // Absolute path to the directory containing the agentt config file
	settings        config.LocalFSBackendSettings // Specific settings for this backend instance
	writable        bool                          // Whether this backend instance is writable
	entityTypes     map[string]config.EntityType  // Map entity type name -> definition (for required fields)
	store           map[string]*content.Item      // In-memory store: absPath -> Item
	mu              sync.RWMutex                  // Mutex for thread-safe access to the store
	initialScanDone bool                          // Flag to track if initial scan completed
	absoluteRootDir string                        // Resolved absolute path to the root directory
}

// NewLocalFSBackend creates and initializes a new LocalFilesystemBackend.
// It now takes specific LocalFS settings, the directory of the config file, and writability.
func NewLocalFSBackend(settings config.LocalFSBackendSettings, configFilePath string, entityTypes []config.EntityType, writable bool) (*LocalFilesystemBackend, error) {
	slog.Debug("Creating new LocalFS backend", "settings", settings, "configFilePath", configFilePath, "writable", writable)

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

	var absoluteRootDir string

	// Resolve RootDir:
	// 1. If it starts with "~", expand to user's home directory.
	// 2. If it's an absolute path (e.g., starts with "/" or "C:\"), use it as is.
	// 3. Otherwise, resolve it relative to the config file's directory.
	if strings.HasPrefix(settings.RootDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Propagate error as NewLocalFSBackend can return an error
			return nil, fmt.Errorf("failed to get user home directory to expand rootDir '%s': %w", settings.RootDir, err)
		}
		// Join home directory with the path part after "~"
		// e.g., "~" -> homeDir
		// e.g., "~/some/path" -> homeDir/some/path
		pathRelativeToHome := strings.TrimPrefix(settings.RootDir, "~")
		absoluteRootDir = filepath.Join(homeDir, pathRelativeToHome)
		// Update settings.RootDir to use the expanded path
		settings.RootDir = absoluteRootDir
	} else if filepath.IsAbs(settings.RootDir) {
		// Path is already absolute
		absoluteRootDir = settings.RootDir
	} else {
		// Path is relative to the config directory
		absoluteRootDir = filepath.Join(configDir, settings.RootDir)
	}

	// Clean the path to resolve any ".." or "." components and ensure a canonical representation.
	// filepath.Join typically calls Clean, but an explicit Clean here ensures it for all branches
	// and handles cases like an empty path segment from TrimPrefix or direct absolute path.
	absoluteRootDir = filepath.Clean(absoluteRootDir)

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
		configDir:       configDir,
		settings:        settings,
		writable:        writable,
		entityTypes:     etMap,
		store:           make(map[string]*content.Item),
		absoluteRootDir: absoluteRootDir,
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

// CreateEntity creates a new guidance entity in the local filesystem.
func (b *LocalFilesystemBackend) CreateEntity(entityData map[string]interface{}, body string, force bool) error {
	b.mu.Lock() // Exclusive lock for create operation
	defer b.mu.Unlock()

	if !b.writable {
		return fmt.Errorf("localfs backend is not configured as writable")
	}

	// 1. Extract ID and EntityType from entityData
	idVal, ok := entityData["id"]
	if !ok {
		return errors.New("entityData must contain an 'id' field")
	}
	id, ok := idVal.(string)
	if !ok || id == "" {
		return errors.New("'id' field must be a non-empty string")
	}

	typeVal, ok := entityData["type"]
	if !ok {
		return errors.New("entityData must contain a 'type' field")
	}
	entityType, ok := typeVal.(string)
	if !ok || entityType == "" {
		return errors.New("'type' field must be a non-empty string")
	}

	// 2. Check if entityType is configured for this backend
	entityLocationPattern, found := b.settings.EntityLocations[entityType]
	if !found {
		return fmt.Errorf("entity type '%s' is not configured in this backend's entityLocations", entityType)
	}

	// Check if this entity type has a definition (for required fields later)
	entityDef, entityDefExists := b.entityTypes[entityType]
	if !entityDefExists {
		// This should ideally not happen if config validation is thorough
		return fmt.Errorf("internal inconsistency: entity type '%s' configured in entityLocations but not defined in global entityTypes", entityType)
	}

	// 3. Check if an entity with this ID already exists in the store
	for _, item := range b.store {
		if item.IsValid {
			if storedID, ok := item.FrontMatter["id"].(string); ok && storedID == id {
				return fmt.Errorf("entity with ID '%s' already exists at path '%s'", id, item.SourcePath)
			}
		}
	}

	// 4. Determine target file path
	//    We take the directory part of the glob and append "id.ext".
	//    This uses the template directly if available.
	var targetPath string
	tmplData := struct{ ID string }{ID: id}
	var pathBuffer bytes.Buffer
	tmpl, err := template.New("filename").Parse(entityLocationPattern)
	if err != nil {
		return fmt.Errorf("failed to parse filename template '%s': %w", entityLocationPattern, err)
	}
	if err := tmpl.Execute(&pathBuffer, tmplData); err != nil {
		return fmt.Errorf("failed to execute filename template '%s' for ID '%s': %w", entityLocationPattern, id, err)
	}
	relativePathFromRoot := pathBuffer.String()
	targetPath = filepath.Join(b.absoluteRootDir, relativePathFromRoot)
	targetPath = filepath.Clean(targetPath)

	// Security check: ensure the targetPath is still within b.absoluteRootDir
	if !strings.HasPrefix(targetPath, b.absoluteRootDir) {
		slog.Error("Potential path traversal attempt in CreateEntity", "targetPath", targetPath, "rootDir", b.absoluteRootDir)
		return fmt.Errorf("invalid target path: constructed path is outside the backend's root directory")
	}

	slog.Debug("Attempting to create entity file", "path", targetPath, "force", force)

	// 5. Create and write the file
	// Ensure the directory for the target file exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for '%s': %w", targetPath, err)
	}

	// Prepare file content
	fileContentBytes, err := constructFileContent(entityData, body)
	if err != nil {
		return fmt.Errorf("failed to construct file content: %w", err)
	}

	// Open file with O_EXCL to prevent overwriting unless force is true
	fileFlags := os.O_RDWR | os.O_CREATE
	if force {
		fileFlags |= os.O_TRUNC // Truncate if exists, or create if not
		slog.Debug("Force flag is true, using O_TRUNC for file opening")
	} else {
		fileFlags |= os.O_EXCL // Fail if exists
		slog.Debug("Force flag is false, using O_EXCL for file opening")
	}

	file, err := os.OpenFile(targetPath, fileFlags, 0666)
	if err != nil {
		if os.IsExist(err) && !force { // Check !force here explicitly for clarity
			return fmt.Errorf("file already exists at target path '%s'", targetPath)
		} else if os.IsExist(err) && force {
			// This case should be handled by O_TRUNC, but if OpenFile somehow still returns IsExist with O_TRUNC
			// it implies a race or an issue. For robustness, log it but proceed if err is nil later.
			slog.Warn("os.OpenFile with O_TRUNC returned IsExist, proceeding with write if file handle is valid", "path", targetPath, "error", err)
			// We might need to re-open without O_EXCL if O_TRUNC isn't behaving as expected with O_EXCL still present implicitly.
			// Simpler: O_TRUNC implies write, so if it exists, it will be truncated.
			// If it still errors with IsExist, something is very wrong with the flags combination or OS behavior.
			// Let's assume standard behavior: if force is true, O_TRUNC will handle existing files.
			// The error here would be some other OpenFile error.
			return fmt.Errorf("failed to open file '%s' (even with force): %w", targetPath, err)
		}
		return fmt.Errorf("failed to open file '%s': %w", targetPath, err)
	}
	defer file.Close()

	_, err = file.Write(fileContentBytes)
	if err != nil {
		return fmt.Errorf("failed to write file to '%s': %w", targetPath, err)
	}
	slog.Info("Successfully wrote new entity file", "path", targetPath)

	// 9. Add to in-memory b.store
	// Use the existing parseGuidanceFile to load it as a content.Item
	newItem, parseErr := parseGuidanceFile(targetPath, entityType, entityDef.RequiredFields)
	if parseErr != nil {
		// If parsing fails, this is a problem. The file was written but can't be loaded.
		// Log an error. Depending on desired atomicity, might consider deleting the file.
		slog.Error("Failed to parse newly created guidance file, store will be inconsistent until next scan", "path", targetPath, "error", parseErr)
		return fmt.Errorf("failed to parse newly created file '%s' for store update: %w. Backend may be inconsistent.", targetPath, parseErr)
	}
	if !newItem.IsValid {
		slog.Error("Newly created guidance file is invalid after parsing, store will be inconsistent until next scan", "path", targetPath, "validationErrors", newItem.ValidationErrors)
		return fmt.Errorf("newly created file '%s' is invalid: %v. Backend may be inconsistent.", targetPath, newItem.ValidationErrors)
	}

	b.store[targetPath] = newItem // Assuming targetPath is already absolute and clean
	slog.Debug("Added new entity to in-memory store", "path", targetPath, "id", newItem.FrontMatter["id"])

	return nil // Success
}

// UpdateEntity updates an existing guidance entity in the local filesystem.
func (b *LocalFilesystemBackend) UpdateEntity(entityID string, updatedData map[string]interface{}, updatedBody *string) error {
	b.mu.Lock() // Exclusive lock for update operation
	defer b.mu.Unlock()

	if !b.writable {
		return fmt.Errorf("localfs backend is not configured as writable")
	}

	if updatedData == nil && updatedBody == nil {
		return fmt.Errorf("UpdateEntity requires either updatedData or updatedBody to be non-nil")
	}

	// 1. Find entity by ID in b.store
	var existingItem *content.Item
	var entityPath string

	for path, item := range b.store {
		if item.IsValid {
			if itemID, ok := item.FrontMatter["id"].(string); ok && itemID == entityID {
				existingItem = item
				entityPath = path // path is the absolute path, which is the key in b.store
				break
			}
		}
	}

	if existingItem == nil {
		return fmt.Errorf("entity with ID '%s' not found", entityID)
	}
	slog.Debug("Found entity to update", "id", entityID, "path", entityPath)

	// Prepare the new frontmatter and body
	newFrontMatter := make(map[string]interface{})
	// Copy existing frontmatter to start with
	for k, v := range existingItem.FrontMatter {
		newFrontMatter[k] = v
	}

	// 2. Update frontmatter if updatedData is provided
	for k, v := range updatedData {
		// Prevent changing ID or Type via this update mechanism
		if k == "id" || k == "type" {
			slog.Warn("Attempted to update restricted field, skipping", "field", k, "entityID", entityID)
			continue
		}
		newFrontMatter[k] = v
	}

	// 3. Validate the potentially modified newFrontMatter
	entityDef, entityDefExists := b.entityTypes[existingItem.EntityType]
	if !entityDefExists {
		return fmt.Errorf("internal inconsistency: entity type '%s' for existing entity not defined", existingItem.EntityType)
	}
	for _, reqField := range entityDef.RequiredFields {
		if _, exists := newFrontMatter[reqField]; !exists {
			// This could happen if updatedData REMOVES a required field without replacing it.
			return fmt.Errorf("update would result in missing required field '%s' for type '%s'", reqField, existingItem.EntityType)
		}
	}

	// 4. Determine new body
	newBody := existingItem.Body
	if updatedBody != nil {
		newBody = *updatedBody
	}

	// 5. Construct new file content
	fileBytes, err := constructFileContent(newFrontMatter, newBody)
	if err != nil {
		return fmt.Errorf("failed to construct updated file content for '%s': %w", entityID, err)
	}

	// 6. Write modified content back to file
	// entityPath is already absolute and clean as it came from b.store keys
	if err := os.WriteFile(entityPath, fileBytes, 0644); err != nil {
		return fmt.Errorf("failed to write updated file to '%s': %w", entityPath, err)
	}
	slog.Info("Successfully updated entity file", "path", entityPath, "id", entityID)

	// 7. Update in-memory b.store
	updatedItem, parseErr := parseGuidanceFile(entityPath, existingItem.EntityType, entityDef.RequiredFields)
	if parseErr != nil {
		slog.Error("Failed to parse updated guidance file, store will be inconsistent until next scan", "path", entityPath, "error", parseErr)
		return fmt.Errorf("failed to parse updated file '%s' for store update: %w. Backend may be inconsistent.", entityPath, parseErr)
	}
	if !updatedItem.IsValid {
		slog.Error("Updated guidance file is invalid after parsing, store will be inconsistent until next scan", "path", entityPath, "validationErrors", updatedItem.ValidationErrors)
		return fmt.Errorf("updated file '%s' is invalid: %v. Backend may be inconsistent.", entityPath, updatedItem.ValidationErrors)
	}

	b.store[entityPath] = updatedItem
	slog.Debug("Updated entity in in-memory store", "path", entityPath, "id", entityID)

	return nil // Success
}

// scanFiles scans the filesystem based on configured globs and updates the internal store.
func (b *LocalFilesystemBackend) scanFiles() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newStore := make(map[string]*content.Item)
	var encounteredError error
	fileCount := 0

	slog.Debug("Starting file scan", "absoluteRootDir", b.absoluteRootDir)

	// Iterate through the entity types defined for this backend
	for entityTypeName, globPattern := range b.settings.EntityLocations {
		entityDef, ok := b.entityTypes[entityTypeName]
		if !ok {
			slog.Error("Internal inconsistency: Entity type from entityLocations not found in entityTypes map", "entityType", entityTypeName)
			continue // Should not happen due to validation in NewLocalFSBackend
		}

		// Construct the full glob pattern relative to the absolute root dir
		fullGlobPattern := filepath.Join(b.absoluteRootDir, globPattern)
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

			// Load and parse the file
			item, err := parseGuidanceFile(absPath, entityTypeName, entityDef.RequiredFields)
			if err != nil {
				slog.Error("Failed to parse guidance file", "path", absPath, "error", err)
				if encounteredError == nil {
					encounteredError = fmt.Errorf("failed to parse guidance file '%s': %w", absPath, err)
				}
				continue
			}

			// Store the item in our new map
			newStore[absPath] = item
		}
	}

	// Update the store with our new data
	b.store = newStore
	slog.Info("File scan complete", "total_files", fileCount)

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

// helper function to construct file content from frontmatter and body
func constructFileContent(frontMatterData map[string]interface{}, body string) ([]byte, error) {
	fmBytes, err := yaml.Marshal(frontMatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter to YAML: %w", err)
	}

	var buffer bytes.Buffer

	// Write starting delimiter
	buffer.WriteString("---\n")

	// Write frontmatter bytes
	if len(fmBytes) > 0 { // Ensure yaml.Marshal didn't return empty (e.g. for nil map)
		buffer.Write(fmBytes)
		// Ensure frontmatter ends with a newline before the next delimiter
		if fmBytes[len(fmBytes)-1] != '\n' {
			buffer.WriteString("\n")
		}
	}

	// Write ending delimiter
	buffer.WriteString("---\n")

	// Write body if it's not empty
	trimmedBody := strings.TrimSpace(body)
	if trimmedBody != "" {
		buffer.WriteString(trimmedBody)
		buffer.WriteString("\n") // Ensure body ends with a newline
	}

	return buffer.Bytes(), nil
}
