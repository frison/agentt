package config

import (
	"fmt"
	// "fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper function to create a temporary config directory structure for FindAndLoadConfig tests
func createTestConfigDir(t *testing.T, path string, content string) (cleanupFunc func()) {
	t.Helper()
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory %s: %v", dir, err)
	}
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(filepath.Dir(dir)) // Attempt cleanup on failure
		t.Fatalf("Failed to write test config file %s: %v", path, err)
	}

	cleanupFunc = func() {
		// Find the top-level directory created (e.g., "tmp_test_find_1")
		parts := strings.Split(dir, string(os.PathSeparator))
		if len(parts) > 0 {
			os.RemoveAll(parts[0])
		}
	}
	return cleanupFunc
}

func TestLoadConfig_Valid_V42(t *testing.T) {
	configPath := filepath.Join("testdata", "valid_multi_backend.yaml")

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed for valid v4.2 config: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config for valid input")
	}

	// Basic structural checks
	if cfg.ListenAddress != ":9090" {
		t.Errorf("Expected ListenAddress :9090, got %s", cfg.ListenAddress)
	}
	if len(cfg.EntityTypes) != 2 {
		t.Fatalf("Expected 2 EntityTypes, got %d", len(cfg.EntityTypes))
	}
	if len(cfg.Backends) != 2 {
		t.Fatalf("Expected 2 Backends, got %d", len(cfg.Backends))
	}

	// Check EntityTypes
	expectedEntityType0 := EntityType{
		Name:           "behavior",
		Description:    "Defines rules for agent operation.",
		RequiredFields: []string{"id", "title", "tier", "priority", "description"},
	}
	if !reflect.DeepEqual(cfg.EntityTypes[0], expectedEntityType0) {
		t.Errorf("EntityType[0] mismatch:\nExpected: %+v\nGot:      %+v", expectedEntityType0, cfg.EntityTypes[0])
	}

	// Check Backends
	if cfg.Backends[0].Name != "local_primary" || cfg.Backends[0].Type != "localfs" {
		t.Errorf("Backend[0] expected name 'local_primary', type 'localfs', got name '%s', type '%s'", cfg.Backends[0].Name, cfg.Backends[0].Type)
	}
	if cfg.Backends[1].Name != "shared_behaviors" || cfg.Backends[1].Type != "localfs" {
		t.Errorf("Backend[1] expected name 'shared_behaviors', type 'localfs', got name '%s', type '%s'", cfg.Backends[1].Name, cfg.Backends[1].Type)
	}

	// Check Settings extraction (using helper)
	fsSettings0, err := cfg.Backends[0].GetLocalFSSettings()
	if err != nil {
		t.Fatalf("GetLocalFSSettings failed for backend[0]: %v", err)
	}
	expectedFsSettings0 := LocalFSBackendSettings{
		RootDir: ".",
		EntityLocations: map[string]string{
			"behavior": ".agent/behaviors/**/*.bhv",
			"recipe":   ".agent/recipes/**/*.rcp",
		},
	}
	if !reflect.DeepEqual(fsSettings0, expectedFsSettings0) {
		t.Errorf("LocalFSBackendSettings[0] mismatch:\nExpected: %+v\nGot:      %+v", expectedFsSettings0, fsSettings0)
	}

	fsSettings1, err := cfg.Backends[1].GetLocalFSSettings()
	if err != nil {
		t.Fatalf("GetLocalFSSettings failed for backend[1]: %v", err)
	}
	expectedFsSettings1 := LocalFSBackendSettings{
		RootDir: "../shared/guidance",
		EntityLocations: map[string]string{
			"behavior": "common/behaviors/*.bhv",
		},
	}
	if !reflect.DeepEqual(fsSettings1, expectedFsSettings1) {
		t.Errorf("LocalFSBackendSettings[1] mismatch:\nExpected: %+v\nGot:      %+v", expectedFsSettings1, fsSettings1)
	}

	// Check LoadedFromPath is set (should be absolute)
	if cfg.LoadedFromPath == "" {
		t.Error("Expected LoadedFromPath to be set, but it was empty")
	}
	if !filepath.IsAbs(cfg.LoadedFromPath) {
		t.Errorf("Expected LoadedFromPath to be absolute, got: %s", cfg.LoadedFromPath)
	}
	if !strings.HasSuffix(cfg.LoadedFromPath, configPath) {
		t.Errorf("Expected LoadedFromPath to end with %s, got: %s", configPath, cfg.LoadedFromPath)
	}

}

func TestLoadConfig_InvalidCases_V42(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		expectedErr string // Substring of the expected error
	}{
		{
			name:        "Missing File",
			filePath:    filepath.Join("testdata", "non_existent_config.yaml"),
			expectedErr: "failed to read config file",
		},
		{
			name:        "Invalid YAML Syntax",
			filePath:    filepath.Join("testdata", "invalid_syntax.yaml"), // Need to create this file
			expectedErr: "failed to parse config YAML",
		},
		{
			name:        "Missing EntityTypes",
			filePath:    filepath.Join("testdata", "invalid_no_entity_types.yaml"),
			expectedErr: "'entityTypes' field is required",
		},
		{
			name:        "Missing Backends",
			filePath:    filepath.Join("testdata", "invalid_no_backends.yaml"),
			expectedErr: "'backends' field is required",
		},
		{
			name:        "Duplicate EntityType Name",
			filePath:    filepath.Join("testdata", "invalid_duplicate_entity_name.yaml"),
			expectedErr: "duplicate entity type name 'behavior'",
		},
		{
			name:        "Missing EntityType Name",
			filePath:    filepath.Join("testdata", "invalid_missing_entity_name.yaml"), // Need to create
			expectedErr: "entityTypes[0]: 'name' field is required",
		},
		{
			name:        "EntityType Missing 'id' in RequiredFields",
			filePath:    filepath.Join("testdata", "invalid_missing_entity_id_req.yaml"),
			expectedErr: "'requiredFields' must include 'id'",
		},
		{
			name:        "Backend Missing Type",
			filePath:    filepath.Join("testdata", "invalid_missing_backend_type.yaml"),
			expectedErr: "backends[0]: 'type' field is required",
		},
	}

	// Create missing invalid_syntax.yaml and invalid_missing_entity_name.yaml
	_ = os.MkdirAll("testdata", 0755)
	_ = os.WriteFile(filepath.Join("testdata", "invalid_syntax.yaml"), []byte("listenAddress: :8080\nentityTypes: [ { name: behavior, requiredFields: [id] }\nbackends: [{ type: localfs incorrect_indent }]"), 0644)
	_ = os.WriteFile(filepath.Join("testdata", "invalid_missing_entity_name.yaml"), []byte("entityTypes: [{ requiredFields: [id] }]\nbackends: [{ type: localfs, settings: { rootDir: \".\"}}]"), 0644)
	t.Cleanup(func() {
		os.Remove(filepath.Join("testdata", "invalid_syntax.yaml"))
		os.Remove(filepath.Join("testdata", "invalid_missing_entity_name.yaml"))
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadConfig(tc.filePath)
			if err == nil {
				t.Fatalf("LoadConfig succeeded for invalid case '%s', expected error containing '%s'", tc.name, tc.expectedErr)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("LoadConfig error mismatch for '%s':\nExpected to contain: %s\nGot:               %v", tc.name, tc.expectedErr, err)
			}
		})
	}
}

func TestFindAndLoadConfig(t *testing.T) {
	originalAgenttConfig := os.Getenv("AGENTT_CONFIG")
	os.Unsetenv("AGENTT_CONFIG")
	defer os.Setenv("AGENTT_CONFIG", originalAgenttConfig)

	originalWD, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("Failed to restore original working directory: %v", err)
		}
	}() // Restore original WD after test

	validContent := `
entityTypes:
  - name: test
    requiredFields: [id]
backends:
  - type: localfs
    settings:
      rootDir: "."
      entityLocations:
        test: "*.test"
`
	testCases := []struct {
		name          string
		setupDir      string // Directory relative to temp test root where test runs
		configPath    string // Path relative to temp test root where config is placed
		expectFound   bool
		expectLoadErr bool
	}{
		{
			name:        "Found in current dir",
			setupDir:    "tmp_test_find_1",
			configPath:  filepath.Join("tmp_test_find_1", ConfigDirName, DefaultConfigFileName),
			expectFound: true,
		},
		{
			name:        "Found in parent dir",
			setupDir:    filepath.Join("tmp_test_find_2", "subdir"),
			configPath:  filepath.Join("tmp_test_find_2", ConfigDirName, DefaultConfigFileName),
			expectFound: true,
		},
		{
			name:        "Found in grandparent dir",
			setupDir:    filepath.Join("tmp_test_find_3", "subdir", "subsubdir"),
			configPath:  filepath.Join("tmp_test_find_3", ConfigDirName, DefaultConfigFileName),
			expectFound: true,
		},
		{
			name:        "Not Found within 3 levels",
			setupDir:    filepath.Join("tmp_test_find_4", "a", "b", "c", "d"),
			configPath:  filepath.Join("tmp_test_find_4", ConfigDirName, DefaultConfigFileName),
			expectFound: false,
		},
		{
			name:          "Found but invalid content",
			setupDir:      "tmp_test_find_5",
			configPath:    filepath.Join("tmp_test_find_5", ConfigDirName, DefaultConfigFileName),
			expectFound:   true,
			expectLoadErr: true, // Will find the file but fail validation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the directory to cd into, if it's not just the base temp dir
			// This ensures os.Chdir doesn't fail for multi-level tc.setupDir
			targetSetupDir := filepath.Join(originalWD, tc.setupDir)
			if err := os.MkdirAll(targetSetupDir, 0755); err != nil {
				t.Fatalf("Failed to create setup directory %s: %v", targetSetupDir, err)
			}

			// Create the config file itself using the helper
			var cleanupFunc func()
			configContentToUse := validContent // Default to valid content
			if tc.name == "Found but invalid content" {
				configContentToUse = "invalid yaml content: {{{{ " // Invalid content
			}
			// tc.configPath is relative to the temp test root (e.g., tmp_test_find_1)
			// We need to join it with originalWD to get the true absolute path for creation
			absConfigPathForCreation := filepath.Join(originalWD, tc.configPath)
			cleanupFunc = createTestConfigDir(t, absConfigPathForCreation, configContentToUse)
			defer cleanupFunc()

			if err := os.Chdir(targetSetupDir); err != nil {
				t.Fatalf("Failed to change directory to %s: %v", targetSetupDir, err)
			}

			cfg, loadedPath, err := FindAndLoadConfig("")

			if tc.expectLoadErr {
				if err == nil {
					t.Errorf("Expected a load error for '%s', but got nil error", tc.name)
				}
				if cfg != nil {
					t.Errorf("Expected config to be nil for '%s' due to load error, but got non-nil config", tc.name)
				}
				// Check that loadedPath points to the file we intended to load, even if invalid
				absExpectedConfigPath, _ := filepath.Abs(filepath.Join(originalWD, tc.configPath))
				absLoadedPath, errAbs := filepath.Abs(loadedPath)
				if errAbs != nil {
					t.Errorf("For '%s', failed to make loadedPath '%s' absolute: %v", tc.name, loadedPath, errAbs)
				}
				if absLoadedPath != absExpectedConfigPath {
					t.Errorf("For '%s', expected absolute loadedPath to be '%s' even with load error, got '%s' (from relative '%s')", tc.name, absExpectedConfigPath, absLoadedPath, loadedPath)
				}
				// Optionally, check err content: e.g., if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "validation") etc.
				return // Test case handled
			}

			if !tc.expectFound {
				if err == nil {
					t.Errorf("Expected an error for '%s' (config not found), but got nil error", tc.name)
				} else {
					if !strings.Contains(err.Error(), "not found") {
						t.Errorf("For '%s', expected error to contain 'not found', got: %v", tc.name, err)
					}
				}
				if cfg != nil {
					t.Errorf("Expected config to be nil for '%s' (config not found), but got non-nil config", tc.name)
				}
				if loadedPath != "" {
					t.Errorf("Expected loadedPath to be empty for '%s' (config not found), but got '%s'", tc.name, loadedPath)
				}
				return // Test case handled
			}

			// Case: Config should be found and loaded successfully
			if err != nil {
				t.Fatalf("FindAndLoadConfig failed unexpectedly for '%s': %v", tc.name, err)
			}
			if cfg == nil {
				t.Fatalf("Expected a non-nil config for '%s', but got nil", tc.name)
			}
			if loadedPath == "" {
				t.Fatalf("Expected a non-empty loadedPath for '%s', but got empty string", tc.name)
			}

			// Compare paths after making loadedPath absolute
			absLoadedPath, pathErr := filepath.Abs(loadedPath)
			if pathErr != nil {
				t.Fatalf("For '%s', failed to make loadedPath ('%s') absolute: %v", tc.name, loadedPath, pathErr)
			}

			absExpectedConfigPathForFoundFile, _ := filepath.Abs(filepath.Join(originalWD, tc.configPath))
			if cfg.LoadedFromPath != absExpectedConfigPathForFoundFile {
				t.Errorf("For '%s', cfg.LoadedFromPath (%s) does not match expected absolute config file path (%s)", tc.name, cfg.LoadedFromPath, absExpectedConfigPathForFoundFile)
			}
			// Also check that absLoadedPath (from FindAndLoadConfig direct return) matches the expected config file path.
			// This is important because cfg.LoadedFromPath is set *inside* LoadConfig, while loadedPath is from FindAndLoadConfig.
			if absLoadedPath != absExpectedConfigPathForFoundFile {
				t.Errorf("For '%s', absLoadedPath from FindAndLoadConfig (%s) does not match expected absolute config file path (%s)", tc.name, absLoadedPath, absExpectedConfigPathForFoundFile)
			}

			// Check content (already excluding "Found but invalid content" as it returns early)
			if len(cfg.EntityTypes) == 0 || cfg.EntityTypes[0].Name != "test" {
				t.Errorf("For '%s', config content not as expected, got: %+v", tc.name, cfg)
			}
		})
	}
}

func TestGetLocalFSSettings(t *testing.T) {
	testCases := []struct {
		name        string
		backendSpec BackendSpec
		expectErr   bool
		expectedCfg LocalFSBackendSettings
	}{
		{
			name: "Valid LocalFS settings",
			backendSpec: BackendSpec{
				Type: "localfs",
				Settings: map[string]interface{}{
					"rootDir": "./data",
					"entityLocations": map[string]string{
						"behavior": "beh/*.bhv",
					},
				},
			},
			expectErr: false,
			expectedCfg: LocalFSBackendSettings{
				RootDir: "./data",
				EntityLocations: map[string]string{
					"behavior": "beh/*.bhv",
				},
			},
		},
		{
			name: "Missing rootDir (should be valid, defaults later)",
			backendSpec: BackendSpec{
				Type: "localfs",
				Settings: map[string]interface{}{
					"entityLocations": map[string]string{"recipe": "rec/*.rcp"},
				},
			},
			expectErr: false,
			expectedCfg: LocalFSBackendSettings{
				RootDir:         "", // Expected to be empty after extraction
				EntityLocations: map[string]string{"recipe": "rec/*.rcp"},
			},
		},
		{
			name: "Missing entityLocations",
			backendSpec: BackendSpec{
				Type: "localfs",
				Settings: map[string]interface{}{
					"rootDir": ".",
				},
			},
			expectErr: false, // Extraction doesn't fail, validation happens later
			expectedCfg: LocalFSBackendSettings{
				RootDir:         ".",
				EntityLocations: nil, // Expect nil map
			},
		},
		{
			name: "Incorrect type for rootDir",
			backendSpec: BackendSpec{
				Type: "localfs",
				Settings: map[string]interface{}{
					"rootDir":         123,
					"entityLocations": map[string]string{"behavior": "*.bhv"},
				},
			},
			expectErr: true, // Expect unmarshal error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings, err := tc.backendSpec.GetLocalFSSettings()

			if tc.expectErr {
				if err == nil {
					t.Error("Expected an error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if !reflect.DeepEqual(settings, tc.expectedCfg) {
					t.Errorf("Settings mismatch:\nExpected: %+v\nGot:      %+v", tc.expectedCfg, settings)
				}
			}
		})
	}
}

// TestFindAndLoadConfig_Success tests finding and loading a valid config file.
func TestFindAndLoadConfig_Success(t *testing.T) {
	originalAgenttConfig := os.Getenv("AGENTT_CONFIG")
	os.Unsetenv("AGENTT_CONFIG")
	defer os.Setenv("AGENTT_CONFIG", originalAgenttConfig)

	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "testconfig")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create .agent/service dir in temp: %v", err)
	}
	actualConfigPath := filepath.Join(configDir, DefaultConfigFileName)
	content := []byte(`
listenAddress: ":1234"
entityTypes:
  - name: test
    requiredFields: [id]
backends:
  - type: localfs
    name: default
    settings:
      rootDir: "."
      entityLocations:
        test: "*.test"
`) // Ensure this content matches expectations below
	if err := os.WriteFile(actualConfigPath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	cfg, loadedPath, err := FindAndLoadConfig(actualConfigPath) // Pass the explicit path

	if err != nil {
		t.Fatalf("FindAndLoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("FindAndLoadConfig returned nil config")
	}
	if loadedPath == "" {
		t.Fatal("FindAndLoadConfig returned empty loadedFilePath")
	}
	absExpectedPath, _ := filepath.Abs(actualConfigPath)
	if loadedPath != absExpectedPath {
		t.Errorf("Expected loadedFilePath to be %s, got %s", absExpectedPath, loadedPath)
	}
	if cfg.ListenAddress != ":1234" {
		t.Errorf("Expected listenAddress :1234, got %s", cfg.ListenAddress)
	}
}

// TestFindAndLoadConfig_NotFound tests the case where the config file is not found.
func TestFindAndLoadConfig_NotFound(t *testing.T) {
	originalAgenttConfig := os.Getenv("AGENTT_CONFIG")
	os.Unsetenv("AGENTT_CONFIG")
	defer os.Setenv("AGENTT_CONFIG", originalAgenttConfig)

	// Create a temporary directory and change to it
	// Ensure no config file exists in search paths
	tempDir, err := os.MkdirTemp("", "testconfignotfound")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWD, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("Failed to restore original working directory: %v", err)
		}
	}() // Restore original WD after test

	// Test with no flag (should not find any config)
	_, _, err = FindAndLoadConfig("")
	if err == nil {
		t.Fatal("FindAndLoadConfig succeeded when no config file should be found")
	}
	if !strings.Contains(err.Error(), "configuration file not found") {
		t.Errorf("Expected error to contain 'configuration file not found', got: %v", err)
	}

	// Test with flag pointing to non-existent file
	nonExistentPath := filepath.Join(tempDir, "non_existent_config.yaml")
	_, _, err = FindAndLoadConfig(nonExistentPath)
	if err == nil {
		t.Fatalf("FindAndLoadConfig with flag for non-existent file succeeded")
	}
	// Error message for specific file not found should be different
	if !strings.Contains(err.Error(), fmt.Sprintf("failed to load configuration from %s", nonExistentPath)) {
		t.Errorf("Expected error for specific non-existent file, got: %v", err)
	}
}

func TestFindAndLoadConfig_EnvVar(t *testing.T) {
	// This test specifically tests AGENTT_CONFIG, so we don't unset it here.
	// Instead, we set it to a specific test value and then restore.
	originalAgenttConfig := os.Getenv("AGENTT_CONFIG")
	// Defer restoration of original AGENTT_CONFIG
	defer os.Setenv("AGENTT_CONFIG", originalAgenttConfig)

	tempDir := t.TempDir()

	envConfigFilePath := filepath.Join(tempDir, "env_config.yaml")
	content := []byte(`
listenAddress: ":7777"
entityTypes: [{name: testenv, requiredFields: [id]}]
backends: [{type: localfs, name: env_backend, settings: {rootDir: "."}}]
`)
	if err := os.WriteFile(envConfigFilePath, content, 0644); err != nil {
		t.Fatalf("Failed to write env config file: %v", err)
	}

	// Set AGENTT_CONFIG environment variable to point to our test file
	if err := os.Setenv("AGENTT_CONFIG", envConfigFilePath); err != nil {
		t.Fatalf("Failed to set AGENTT_CONFIG: %v", err)
	}

	// Test 1: AGENTT_CONFIG is set, flag is empty
	cfg, loadedPath, err := FindAndLoadConfig("")
	if err != nil {
		t.Fatalf("FindAndLoadConfig with env var failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("FindAndLoadConfig with env var returned nil config")
	}
	if loadedPath != envConfigFilePath {
		t.Errorf("Expected loadedPath to be %s (from env var), got %s", envConfigFilePath, loadedPath)
	}
	if cfg.ListenAddress != ":7777" {
		t.Errorf("Expected listenAddress :7777 from env var config, got %s", cfg.ListenAddress)
	}

	// Test 2: Flag overrides env var
	flagConfigFilePath := filepath.Join(tempDir, "flag_override_config.yaml")
	flagConfigContent := `
listenAddress: ":8888"
entityTypes: [{name: "flag_test", requiredFields: ["id"]}]
backends: [{name: "flag_backend", type: "localfs", settings: {rootDir: "."}}]
`
	if err := os.WriteFile(flagConfigFilePath, []byte(flagConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write flag override config file: %v", err)
	}

	cfgOverride, loadedPathOverride, errOverride := FindAndLoadConfig(flagConfigFilePath)
	if errOverride != nil {
		t.Fatalf("FindAndLoadConfig with flag (overriding env) failed: %v", errOverride)
	}
	if cfgOverride == nil {
		t.Fatal("FindAndLoadConfig with flag (overriding env) returned nil config")
	}
	if loadedPathOverride != flagConfigFilePath {
		t.Errorf("Expected loadedPath to be %s (from flag), got %s", flagConfigFilePath, loadedPathOverride)
	}
	if cfgOverride.ListenAddress != ":8888" {
		t.Errorf("Expected listenAddress :8888 from flag override, got %s", cfgOverride.ListenAddress)
	}
}
