package discoverytypes

// DiscoveredReflex represents the structured information for a single discovered reflex.
// It's designed to be marshalled into JSON format.
type DiscoveredReflex struct {
	Path        string                 `json:"path"`                  // Relative path to the reflex directory from the reflexes root
	Name        string                 `json:"name"`                  // Name of the reflex from its manifest
	Description string                 `json:"description"`           // Description from its manifest
	Inputs      map[string]interface{} `json:"inputs,omitempty"`    // Input paths defined in the manifest (using interface{} for flexibility)
	Outputs     map[string]interface{} `json:"outputs,omitempty"`   // Output paths defined in the manifest (using interface{})
}