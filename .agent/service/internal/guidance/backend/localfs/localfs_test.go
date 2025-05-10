// This file will contain tests for the localfs backend.
// TODO: Implement tests.

package localfs

import (
	"agentt/internal/config"
	// "agentt/internal/content" // Comment out until used by scan/get tests
	// "agentt/internal/guidance/backend"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	// "time"
)

// TODO: Add tests for NewLocalFSBackend (validation, rootDir resolution)

// TODO: Add tests for scanFiles (path joining, glob matching)

// TODO: Add tests for parseGuidanceFile (valid/invalid files, frontmatter/body splitting, validation)

// TODO: Add tests for GetSummary and GetDetails (after scan)

// --- Test Setup Helpers ---

// --- NewLocalFSBackend Tests ---

func TestNewLocalFSBackend_Validation(t *testing.T) {
	// Define entity types used across tests
	entityTypes := []config.EntityType{
		{Name: "behavior", RequiredFields: []string{"id", "title"}},
		{Name: "recipe", RequiredFields: []string{"id"}},
	}

	// Assume a testdata structure like:
	// testdata/
	// 	 scenario1_valid/config.yaml (rootDir=".")
	// 	 scenario2_rel_root/config_dir/config.yaml (rootDir="../entities_base")
	// 	 entities_base/ (target for scenario2)
	// 	 scenario3_invalid_root/config.yaml (rootDir="./nonexistent")
	// Create necessary directories/files in testdata manually.

	testCases := []struct {
		name          string
		settings      config.LocalFSBackendSettings
		configPath    string // Path relative to package dir, e.g., "testdata/scenario1/config.yaml"
		expectErr     bool
		expectErrCont string
	}{
		{
			name: "Valid config, rootDir='.'",
			settings: config.LocalFSBackendSettings{
				RootDir:         ".",
				EntityLocations: map[string]string{"behavior": "*.bhv"},
			},
			configPath: "testdata/scenario1_valid/config.yaml",
			expectErr:  false,
		},
		{
			name: "Valid config, relative rootDir exists",
			settings: config.LocalFSBackendSettings{
				RootDir:         "../entities_base", // Relative to config file location
				EntityLocations: map[string]string{"recipe": "*.rcp"},
			},
			configPath: "testdata/scenario2_rel_root/config_dir/config.yaml",
			expectErr:  false,
		},
		// {
		// 	name: "Valid config, absolute rootDir exists",
		// 	// Absolute paths are hard to test reliably without complex setup.
		// 	// Consider testing this manually or with integration tests.
		// },
		{
			name: "Invalid config, rootDir does NOT exist",
			settings: config.LocalFSBackendSettings{
				RootDir:         "./nonexistent",
				EntityLocations: map[string]string{"behavior": "*.bhv"},
			},
			configPath:    "testdata/scenario3_invalid_root/config.yaml",
			expectErr:     true,
			expectErrCont: "resolved 'rootDir' does not exist",
		},
		{
			name: "Invalid config, relative rootDir does NOT exist",
			settings: config.LocalFSBackendSettings{
				RootDir:         "../nonexistent_base",
				EntityLocations: map[string]string{"behavior": "*.bhv"},
			},
			configPath:    "testdata/scenario2_rel_root/config_dir/config.yaml",
			expectErr:     true,
			expectErrCont: "resolved 'rootDir' does not exist",
		},
		{
			name: "Invalid config, no entityLocations",
			settings: config.LocalFSBackendSettings{
				RootDir: ".",
				// EntityLocations: map[string]string{}, // Empty map
			},
			configPath:    "testdata/scenario1_valid/config.yaml", // Use a valid config path
			expectErr:     true,
			expectErrCont: "entityLocations' must be defined",
		},
		{
			name: "Invalid config, entityLocations key not in entityTypes",
			settings: config.LocalFSBackendSettings{
				RootDir:         ".",
				EntityLocations: map[string]string{"unknown_type": "*.test"},
			},
			configPath:    "testdata/scenario1_valid/config.yaml", // Use a valid config path
			expectErr:     true,
			expectErrCont: "not a defined entity type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure the config file path exists for setup validation, even if RootDir is invalid later
			// The test assumes testdata and its structure exists.
			absConfigPath, err := filepath.Abs(tc.configPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path for test config %s: %v", tc.configPath, err)
			}
			if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
				t.Fatalf("Test setup error: Config file %s (abs: %s) does not exist in testdata. Please create it.", tc.configPath, absConfigPath)
			}

			_, err = NewLocalFSBackend(tc.settings, absConfigPath, entityTypes, true)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("NewLocalFSBackend succeeded, expected error containing '%s'", tc.expectErrCont)
				}
				if !strings.Contains(err.Error(), tc.expectErrCont) {
					t.Errorf("NewLocalFSBackend error mismatch:\nExpected to contain: %s\nGot:               %v", tc.expectErrCont, err)
				}
			} else {
				if err != nil {
					t.Fatalf("NewLocalFSBackend failed unexpectedly: %v", err)
				}
			}
		})
	}
}

// --- parseGuidanceFile Tests ---

func TestParseGuidanceFile(t *testing.T) {
	// Assumes a testdata structure like:
	// testdata/
	//   parsing/
	//     valid_fm.bhv
	//     valid_no_fm.txt
	//     valid_empty_fm.rcp
	//     invalid_yaml.bhv
	//     missing_req_field.bhv
	//     empty_id.bhv
	//     numeric_id.bhv
	// Create these files in testdata manually with appropriate content.

	testCases := []struct {
		name                string
		filePath            string // Relative path within testdata, e.g., "parsing/valid_fm.bhv"
		entityType          string
		requiredFields      []string
		expectErr           bool // Expect error during parsing itself
		expectValid         bool // Expect item.IsValid to be true
		expectedFrontMatter map[string]interface{}
		expectedBody        string
		expectedTier        string
		expectedErrCont     string // Substring for validation errors
	}{
		{
			name:     "Valid file with frontmatter",
			filePath: "testdata/parsing/valid_fm.bhv",
			// fileContent: "---\nid: bhv1\ntitle: Behavior 1\ntier: must\n---\nBody content.", // Content should be in the file
			entityType:          "behavior",
			requiredFields:      []string{"id", "title"},
			expectErr:           false,
			expectValid:         true,
			expectedFrontMatter: map[string]interface{}{"id": "bhv1", "title": "Behavior 1", "tier": "must"}, // Example expected
			expectedBody:        "Body content.",                                                             // Example expected
			expectedTier:        "must",                                                                      // Example expected
		},
		{
			name:     "Valid file no frontmatter",
			filePath: "testdata/parsing/valid_no_fm.txt",
			// fileContent: "Just body content.",
			entityType:          "generic",
			requiredFields:      []string{}, // No fields required
			expectErr:           false,
			expectValid:         true,
			expectedFrontMatter: map[string]interface{}{},
			expectedBody:        "Just body content.",
		},
		{
			name:     "Valid file empty frontmatter",
			filePath: "testdata/parsing/valid_empty_fm.rcp",
			// fileContent: "---\n---\nRecipe body.",
			entityType:          "recipe",
			requiredFields:      []string{}, // No fields required
			expectErr:           false,
			expectValid:         true,
			expectedFrontMatter: map[string]interface{}{},
			expectedBody:        "Recipe body.",
		},
		{
			name:     "Invalid YAML syntax in frontmatter",
			filePath: "testdata/parsing/invalid_yaml.bhv",
			// fileContent: "---\nid: bhv_bad\n  bad_indent: true\n---\nBody.",
			entityType:      "behavior",
			requiredFields:  []string{"id"},
			expectErr:       true, // Expect parsing error
			expectedErrCont: "failed to parse YAML",
		},
		{
			name:     "Missing required field",
			filePath: "testdata/parsing/missing_req_field.bhv",
			// fileContent: "---\ntitle: Missing ID\ntier: should\n---\nBody.",
			entityType:          "behavior",
			requiredFields:      []string{"id", "title"},
			expectErr:           false,                                                           // Parsing succeeds
			expectValid:         false,                                                           // But validation fails
			expectedFrontMatter: map[string]interface{}{"title": "Missing ID", "tier": "should"}, // Expect parsed FM
			expectedBody:        "Body.",
			expectedTier:        "should", // Expect parsed tier
			expectedErrCont:     "Missing required field: 'id'",
		},
		{
			name:     "Required field 'id' present but empty string",
			filePath: "testdata/parsing/empty_id.bhv",
			// fileContent: "---\nid: \"\"\ntitle: Empty ID\n---\nBody.",
			entityType:     "behavior",
			requiredFields: []string{"id", "title"},
			expectErr:      false,
			expectValid:    false,
			// expectedFrontMatter: map[string]interface{}{"id": "", "title": "Empty ID"},
			// expectedBody: "Body.",
			expectedErrCont: "Required field 'id' is missing or not a non-empty string",
		},
		{
			name:     "Required field 'id' not a string",
			filePath: "testdata/parsing/numeric_id.bhv",
			// fileContent: "---\nid: 123\ntitle: Numeric ID\n---\nBody.",
			entityType:     "behavior",
			requiredFields: []string{"id", "title"},
			expectErr:      false,
			expectValid:    false,
			// expectedFrontMatter: map[string]interface{}{"id": 123, "title": "Numeric ID"},
			// expectedBody: "Body.",
			expectedErrCont: "Required field 'id' is missing or not a non-empty string",
		},
		{
			name:     "File not found",
			filePath: "testdata/parsing/nonexistent.dat",
			// fileContent: "", // Content doesn't matter
			entityType:      "any",
			requiredFields:  []string{},
			expectErr:       true, // Expect read error
			expectedErrCont: "failed to read file",
		},
	}

	// tempDir := t.TempDir() // No longer needed
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use absolute path based on current dir + relative testdata path
			absFilePath, err := filepath.Abs(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to get absolute path for test file %s: %v", tc.filePath, err)
			}

			// Check if file exists before attempting parse (except for not found test)
			if tc.name != "File not found" {
				if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
					t.Fatalf("Test setup error: Test file %s (abs: %s) does not exist in testdata. Please create it.", tc.filePath, absFilePath)
				}
			}

			item, err := parseGuidanceFile(absFilePath, tc.entityType, tc.requiredFields)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("parseGuidanceFile succeeded, expected error containing '%s'", tc.expectedErrCont)
				}
				if !strings.Contains(err.Error(), tc.expectedErrCont) {
					t.Errorf("parseGuidanceFile error mismatch:\nExpected to contain: %s\nGot:               %v", tc.expectedErrCont, err)
				}
				return // Don't check item if error was expected
			}

			// No parsing error expected, check item details
			if err != nil {
				t.Fatalf("parseGuidanceFile failed unexpectedly: %v", err)
			}
			if item == nil {
				t.Fatal("parseGuidanceFile returned nil item unexpectedly")
			}

			if item.IsValid != tc.expectValid {
				t.Errorf("Expected item.IsValid to be %t, but got %t. Errors: %v", tc.expectValid, item.IsValid, item.ValidationErrors)
			}

			if !tc.expectValid && len(item.ValidationErrors) == 0 {
				t.Error("Expected validation errors when item.IsValid is false, but got none")
			}

			if !tc.expectValid && tc.expectedErrCont != "" {
				foundErr := false
				for _, vErr := range item.ValidationErrors {
					if strings.Contains(vErr, tc.expectedErrCont) {
						foundErr = true
						break
					}
				}
				if !foundErr {
					t.Errorf("Expected validation errors to contain '%s', but got: %v", tc.expectedErrCont, item.ValidationErrors)
				}
			}

			// Only check content equality if expecting valid result or specific invalid state
			if tc.expectValid || tc.expectedFrontMatter != nil || tc.expectedBody != "" {
				if !reflect.DeepEqual(item.FrontMatter, tc.expectedFrontMatter) {
					t.Errorf("FrontMatter mismatch:\nExpected: %+v\nGot:      %+v", tc.expectedFrontMatter, item.FrontMatter)
				}

				if item.Body != tc.expectedBody {
					t.Errorf("Body mismatch:\nExpected: %s\nGot:      %s", tc.expectedBody, item.Body)
				}
			}

			if item.EntityType != tc.entityType {
				t.Errorf("EntityType mismatch: Expected %s, Got %s", tc.entityType, item.EntityType)
			}

			if item.Tier != tc.expectedTier {
				t.Errorf("Tier mismatch: Expected %s, Got %s", tc.expectedTier, item.Tier)
			}
		})
	}
}

// TODO: Add tests for scanFiles (requires more setup with config and file structure)

// TODO: Add tests for GetSummary and GetDetails (requires scan first)

// --- TestLocalFSBackend_InitErrors ---
// This test checks validation within NewLocalFSBackend *itself*,
// not the full config loading process.
func TestLocalFSBackend_InitErrors(t *testing.T) {
	// No real config file needed here, just a placeholder path for dir resolution
	tempDir := t.TempDir()
	dummyConfigPath := filepath.Join(tempDir, "dummy_config.yaml")

	entityTypes := []config.EntityType{
		{Name: "behavior", RequiredFields: []string{"id"}},
	}

	tests := []struct {
		name        string
		settings    config.LocalFSBackendSettings // Use settings directly
		expectError string                        // Substring expected in error
	}{
		{
			name: "Missing entityLocations (validation inside New)",
			// RootDir defaults to "." relative to dummyConfigPath if empty, which is tempDir (exists)
			settings:    config.LocalFSBackendSettings{RootDir: "."},
			expectError: "entityLocations' must be defined", // Error from NewLocalFSBackend validation
		},
		{
			name:        "RootDir resolves to non-existent path",
			settings:    config.LocalFSBackendSettings{RootDir: "nonexistent_subdir", EntityLocations: map[string]string{"behavior": "*.bhv"}},
			expectError: "resolved 'rootDir' does not exist", // Error from NewLocalFSBackend validation
		},
		{
			name:        "entityLocations for unknown type (validation inside New)",
			settings:    config.LocalFSBackendSettings{RootDir: ".", EntityLocations: map[string]string{"unknown": "*.test"}},
			expectError: "not a defined entity type", // Error from NewLocalFSBackend validation
		},
		// Note: Missing RootDir itself is *not* an error here if it defaults to an existing dir,
		// but missing entityLocations *is* an error in NewLocalFSBackend.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a dummy file at dummyConfigPath so filepath.Dir works
			if err := os.WriteFile(dummyConfigPath, []byte{}, 0644); err != nil {
				t.Fatalf("Failed to create dummy config file: %v", err)
			}

			_, err := NewLocalFSBackend(tc.settings, dummyConfigPath, entityTypes, true)
			if tc.expectError != "" {
				if err == nil {
					t.Fatal("Expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectError) {
					t.Errorf("Expected error message to contain '%s', got: %v", tc.expectError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
