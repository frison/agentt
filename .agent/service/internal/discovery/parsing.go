package discovery

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/content"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var frontMatterSeparator = []byte("---")

// parseFile reads, parses, and validates a single guidance file based on an entity type definition.
func parseFile(filePath string, entityDef config.EntityTypeDefinition) (*content.Item, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	fileData, err := os.ReadFile(tabsPath)
	if err != nil {
		// Handle cases where file might disappear between discovery and read
		if os.IsNotExist(err) {
			return nil, nil // Not an error for the service, file is just gone
		}
		return nil, fmt.Errorf("failed to read file %s: %w", absPath, err)
	}

	item := &content.Item{
		EntityType:  entityDef.Name,
		SourcePath:  absPath,
		FrontMatter: make(map[string]interface{}),
		IsValid:     true, // Assume valid initially
		LastUpdated: time.Now().UTC(),
	}

	parts := bytes.SplitN(fileData, frontMatterSeparator, 3)

	if len(parts) >= 3 && len(bytes.TrimSpace(parts[0])) == 0 { // Found frontmatter: --- YAML --- BODY
		yamlData := parts[1]
		item.Body = string(bytes.TrimSpace(parts[2]))

		err = yaml.Unmarshal(yamlData, &item.FrontMatter)
		if err != nil {
			item.IsValid = false
			item.ValidationErrors = append(item.ValidationErrors, fmt.Sprintf("YAML parsing error: %v", err))
			// Don't return error, just mark as invalid and proceed to check required fields if possible
		}
	} else { // No valid frontmatter detected or file only contains frontmatter
		item.IsValid = false
		item.ValidationErrors = append(item.ValidationErrors, "No valid YAML frontmatter detected (must start with ---, end with ---, and have content)")
		// Treat the whole file as body if no separator? Or require frontmatter?
		// Forcing frontmatter seems reasonable for this use case.
		item.Body = string(bytes.TrimSpace(fileData)) // Store body anyway, though item is invalid
	}

	// Validation: Check for required frontmatter keys
	for _, requiredKey := range entityDef.RequiredFrontMatter {
		if _, ok := item.FrontMatter[requiredKey]; !ok {
			item.IsValid = false
			item.ValidationErrors = append(item.ValidationErrors, fmt.Sprintf("Missing required frontmatter key: '%s'", requiredKey))
		}
	}

	// Metadata Inference (Example: Behavior Tier)
	if item.EntityType == "behavior" {
		dir := filepath.Dir(item.SourcePath)
		parentDir := filepath.Base(dir)
		if parentDir == "must" || parentDir == "should" {
			item.Tier = parentDir
		} else {
			// Optionally add a validation error if tier can't be inferred but is expected
			// item.IsValid = false
			// item.ValidationErrors = append(item.ValidationErrors, "Could not infer behavior tier ('must' or 'should') from path")
		}
	}

	// Potentially add more validation/inference steps here

	return item, nil // Return the item, even if invalid (caller decides what to do)
}