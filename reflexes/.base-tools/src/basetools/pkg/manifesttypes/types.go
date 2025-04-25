package manifesttypes

// --- Structs shared between manifest tools ---

// InputSpec represents a generic input specification
type InputSpec struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
	Pattern     string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
}

// PathSpec represents a file, directory, or glob pattern specification
type PathSpec struct {
	Type        string      `yaml:"type" json:"type"` // "file", "directory", or "glob"
	Description string      `yaml:"description" json:"description"`
	Required    bool        `yaml:"required" json:"required"`
	Pattern     string      `yaml:"pattern,omitempty" json:"pattern,omitempty"` // Glob pattern or file naming pattern
	Format      string      `yaml:"format,omitempty" json:"format,omitempty"`   // Expected content format
	Schema      interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`   // Optional schema for validation
}

// Manifest represents the structure of a reflex manifest
type Manifest struct {
	Name        string              `yaml:"name" json:"name"`
	Version     string              `yaml:"version" json:"version"`
	Description string              `yaml:"description" json:"description"`
	Environment map[string]InputSpec `yaml:"environment" json:"environment"`
	InputPaths  map[string]PathSpec `yaml:"input_paths,omitempty" json:"input_paths,omitempty"`
	Stdout      *PathSpec           `yaml:"stdout,omitempty" json:"stdout,omitempty"`
	OutputPaths map[string]PathSpec `yaml:"output_paths,omitempty" json:"output_paths,omitempty"`
}