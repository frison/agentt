package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	// Import the shared types from the internal package
	"nhi/basetools/pkg/manifesttypes"
)

// ManifestHandler processes manifest.yml files for both human and machine consumption
type ManifestHandler struct {
	ManifestPath string
	OutputFormat string
	OutputPath   string
	Command      string // "show" or "verify"
}

// VerificationError represents an error found during state verification
type VerificationError struct {
	Type        string // "environment", "input_path", "output_path", "permission"
	Name        string // Name of the variable/path
	Description string // Description of the error
}

// --- Shared Structs are now imported from nhi/basetools/pkg/manifesttypes ---

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{
		ManifestPath: os.Getenv("MANIFEST_PATH"),
		OutputFormat: os.Getenv("OUTPUT_FORMAT"),
		OutputPath:   os.Getenv("OUTPUT_PATH"),
		Command:      os.Getenv("COMMAND"),
	}
}

// ... (rest of the functions remain the same, using manifesttypes.Manifest etc.) ...

func (h *ManifestHandler) Process() error {
	// Default to manifest.yml in the current directory if not specified
	if h.ManifestPath == "" {
		h.ManifestPath = "manifest.yml"
	}

	// Read manifest file
	data, err := os.ReadFile(h.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var manifest manifesttypes.Manifest // Use imported type
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Process based on command
	switch strings.ToLower(h.Command) {
	case "verify":
		return h.verifyState(manifest)
	default: // "show" is the default command
		return h.showManifest(manifest)
	}
}

func (h *ManifestHandler) showManifest(m manifesttypes.Manifest) error {
	// Process based on output format
	switch strings.ToLower(h.OutputFormat) {
	case "human":
		return h.outputHuman(m)
	case "nhi":
		return h.outputNHI(m)
	case "json":
		return h.outputJSON(m)
	default:
		return fmt.Errorf("unsupported output format: %s", h.OutputFormat)
	}
}

func (h *ManifestHandler) verifyState(m manifesttypes.Manifest) error {
	var errors []VerificationError

	// Verify environment variables
	errors = append(errors, h.verifyEnvironment(m.Environment)...)

	// Verify input paths
	errors = append(errors, h.verifyInputPaths(m.InputPaths)...)

	// Verify output paths and permissions
	outputErrors := h.verifyOutputs(m.Stdout, m.OutputPaths)
	errors = append(errors, outputErrors...)

	// If there are errors, format and output them
	if len(errors) > 0 {
		return h.outputVerificationErrors(errors)
	}

	fmt.Println("✓ All manifest requirements satisfied")
	return nil
}

func (h *ManifestHandler) verifyEnvironment(envVars map[string]manifesttypes.InputSpec) []VerificationError {
	var errors []VerificationError

	for name, spec := range envVars {
		value := os.Getenv(name)

		if spec.Required && value == "" {
			errors = append(errors, VerificationError{
				Type:        "environment",
				Name:        name,
				Description: "Required environment variable is not set",
			})
			continue
		}

		if value != "" && spec.Pattern != "" {
			matched, err := filepath.Match(spec.Pattern, value)
			if err != nil {
				errors = append(errors, VerificationError{
					Type:        "environment",
					Name:        name,
					Description: fmt.Sprintf("Invalid pattern in manifest: %v", err),
				})
			} else if !matched {
				errors = append(errors, VerificationError{
					Type:        "environment",
					Name:        name,
					Description: fmt.Sprintf("Value does not match required pattern: %s", spec.Pattern),
				})
			}
		}
	}

	return errors
}

func (h *ManifestHandler) verifyInputPaths(paths map[string]manifesttypes.PathSpec) []VerificationError {
	var errors []VerificationError

	for name, spec := range paths {
		if !spec.Required {
			continue
		}

		matches, err := filepath.Glob(name)
		if err != nil {
			errors = append(errors, VerificationError{
				Type:        "input_path",
				Name:        name,
				Description: fmt.Sprintf("Invalid path pattern: %v", err),
			})
			continue
		}

		if len(matches) == 0 {
			errors = append(errors, VerificationError{
				Type:        "input_path",
				Name:        name,
				Description: "Required input path not found",
			})
		}
	}

	return errors
}

func (h *ManifestHandler) verifyOutputs(stdout *manifesttypes.PathSpec, outputs map[string]manifesttypes.PathSpec) []VerificationError {
	var errors []VerificationError

	// If there are any output paths, verify CALLING_UID/GID are set
	if len(outputs) > 0 {
		if os.Getenv("CALLING_UID") == "" || os.Getenv("CALLING_GID") == "" {
			errors = append(errors, VerificationError{
				Type:        "permission",
				Name:        "CALLING_UID/GID",
				Description: "CALLING_UID and CALLING_GID must be set when output paths are specified",
			})
		}
	}

	// Verify output paths exist and are writable
	for path := range outputs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			errors = append(errors, VerificationError{
				Type:        "output_path",
				Name:        path,
				Description: fmt.Sprintf("Cannot create output directory: %v", err),
			})
		}
	}

	return errors
}

func (h *ManifestHandler) outputVerificationErrors(errors []VerificationError) error {
	// Group errors by type
	grouped := make(map[string][]VerificationError)
	for _, err := range errors {
		grouped[err.Type] = append(grouped[err.Type], err)
	}

	// Format error output
	var output strings.Builder
	output.WriteString("❌ Manifest requirements not satisfied:\n\n")

	for _, errType := range []string{"environment", "input_path", "output_path", "permission"} {
		if errs, ok := grouped[errType]; ok {
			switch errType {
			case "environment":
				output.WriteString("Environment Variables:\n")
			case "input_path":
				output.WriteString("Input Paths:\n")
			case "output_path":
				output.WriteString("Output Paths:\n")
			case "permission":
				output.WriteString("Permissions:\n")
			}

			for _, err := range errs {
				output.WriteString(fmt.Sprintf("  - %s: %s\n", err.Name, err.Description))
			}
			output.WriteString("\n")
		}
	}

	return h.writeOutput(output.String())
}

func (h *ManifestHandler) outputHuman(m manifesttypes.Manifest) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s (v%s)\n\n", m.Name, m.Version))
	sb.WriteString(m.Description + "\n\n")

	// Inputs
	sb.WriteString("## Inputs\n\n")
	if len(m.Environment) > 0 {
		sb.WriteString("### Environment Variables\n")
		for name, spec := range m.Environment {
			req := ""
			if spec.Required {
				req = " (Required)"
			}
			sb.WriteString(fmt.Sprintf("- %s%s: %s\n", name, req, spec.Description))
		}
		sb.WriteString("\n")
	}

	if len(m.InputPaths) > 0 {
		sb.WriteString("### Input Paths\n")
		for path, spec := range m.InputPaths {
			req := ""
			if spec.Required {
				req = " (Required)"
			}
			details := []string{spec.Type}
			if spec.Format != "" {
				details = append(details, spec.Format)
			}
			if spec.Pattern != "" {
				details = append(details, fmt.Sprintf("pattern: %s", spec.Pattern))
			}
			sb.WriteString(fmt.Sprintf("- %s%s (%s): %s\n",
				path, req, strings.Join(details, ", "), spec.Description))
		}
		sb.WriteString("\n")
	}

	// Outputs
	sb.WriteString("## Outputs\n\n")
	if m.Stdout != nil {
		sb.WriteString("### Standard Output\n")
		sb.WriteString(fmt.Sprintf("Format: %s\n", m.Stdout.Type))
		if m.Stdout.Description != "" {
			sb.WriteString(m.Stdout.Description + "\n")
		}
		sb.WriteString("\n")
	}

	if len(m.OutputPaths) > 0 {
		sb.WriteString("### Output Paths\n")
		for path, spec := range m.OutputPaths {
			details := []string{spec.Type}
			if spec.Format != "" {
				details = append(details, spec.Format)
			}
			if spec.Pattern != "" {
				details = append(details, fmt.Sprintf("pattern: %s", spec.Pattern))
			}
			sb.WriteString(fmt.Sprintf("- %s (%s): %s\n",
				path, strings.Join(details, ", "), spec.Description))
		}
	}

	return h.writeOutput(sb.String())
}

func (h *ManifestHandler) outputNHI(m manifesttypes.Manifest) error {
	// Output just the NHI-compatible specification section
	nhiSpec := struct {
		Environment map[string]manifesttypes.InputSpec `json:"environment"`
		InputPaths  map[string]manifesttypes.PathSpec `json:"input_paths,omitempty"`
		Stdout     *manifesttypes.PathSpec           `json:"stdout,omitempty"`
		OutputPaths map[string]manifesttypes.PathSpec `json:"output_paths,omitempty"`
	}{
		Environment: m.Environment,
		InputPaths:  m.InputPaths,
		Stdout:     m.Stdout,
		OutputPaths: m.OutputPaths,
	}

	data, err := yaml.Marshal(nhiSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal NHI spec: %w", err)
	}

	return h.writeOutput(string(data))
}

func (h *ManifestHandler) outputJSON(m manifesttypes.Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return h.writeOutput(string(data))
}

func (h *ManifestHandler) writeOutput(content string) error {
	if h.OutputPath == "" || h.OutputPath == "-" {
		_, err := fmt.Print(content)
		return err
	}

	if err := os.MkdirAll(filepath.Dir(h.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(h.OutputPath, []byte(content), 0644)
}

func main() {
	handler := NewManifestHandler()
	if err := handler.Process(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}