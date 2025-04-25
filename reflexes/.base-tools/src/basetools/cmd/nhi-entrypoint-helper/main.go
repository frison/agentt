package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"

	// Import the shared types from the internal package
	"nhi/basetools/pkg/manifesttypes"
)

// Mandated manifest path
const manifestPath = "/manifest.yml"

// --- Helper Logic ---

func main() {
	// Check for SHOW_MANIFEST first
	if strings.ToLower(os.Getenv("SHOW_MANIFEST")) == "true" {
		showManifest()
		os.Exit(0)
	}

	// os.Args[0] is the helper itself.
	// os.Args[1:] is the command and its arguments to execute.
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: No command provided to the entrypoint helper.")
		os.Exit(1)
	}

	targetCmdArgs := os.Args[1:]
	targetCmdPath := targetCmdArgs[0]

	// Determine manifest path (Removed AGENTT_MANIFEST logic)
	// manifestPath := os.Getenv("AGENTT_MANIFEST")
	// if manifestPath == "" {
	// 	manifestPath = defaultManifestPath
	// }

	// Check if manifest exists
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		// Always print the specific error before executing command
		fmt.Fprintf(os.Stderr, "Error reading manifest %s: %v\n", manifestPath, err) // Print the actual error

		// Manifest doesn't exist or is unreadable, just execute the command
		// The original distinction based on IsNotExist is removed for clarity here
		// if os.IsNotExist(err) {
		// 	executeCommand(targetCmdPath, targetCmdArgs)
		// } else {
		// 	fmt.Fprintf(os.Stderr, "Warning: Could not read manifest %s: %v\n", manifestPath, err)
		// 	executeCommand(targetCmdPath, targetCmdArgs)
		// }
		executeCommand(targetCmdPath, targetCmdArgs) // Always try to execute after logging error
		return // executeCommand exits or replaces the process
	}

	// Parse manifest
	var m manifesttypes.Manifest // Use imported type
	if err := yaml.Unmarshal(manifestData, &m); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not parse manifest %s: %v\n", manifestPath, err)
		executeCommand(targetCmdPath, targetCmdArgs)
		return
	}

	// --- Check if help is needed ---
	// Help is needed if:
	// 1. The manifest defines required environment variables that are NOT set.
	// 2. OR (optional refinement): No arguments were passed to the underlying command (len(targetCmdArgs) == 1).
	// For simplicity, we'll start by checking only for missing required env vars.

	missingRequired := false
	requiredInputsDesc := []string{}
	for name, spec := range m.Environment { // spec is already manifesttypes.InputSpec
		if spec.Required {
			val := os.Getenv(name)
			if val == "" {
				missingRequired = true
				desc := fmt.Sprintf("  - %s: %s", name, spec.Description)
				requiredInputsDesc = append(requiredInputsDesc, desc)
			}
		}
	}

	// Show help if required env vars are missing
	if missingRequired {
		printUsage(m, requiredInputsDesc) // Pass imported type
		os.Exit(1)
	}

	// --- Help not needed, execute the command ---
	executeCommand(targetCmdPath, targetCmdArgs)
}

// New function to handle showing the manifest
func showManifest() {
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		// If manifest doesn't exist or is unreadable when asked to show it, print error to stderr and exit non-zero
		fmt.Fprintf(os.Stderr, "Error reading manifest %s: %v\n", manifestPath, err)
		os.Exit(1)
	}
	// Print raw manifest content to stdout
	fmt.Print(string(manifestData))
}

func printUsage(m manifesttypes.Manifest, requiredInputs []string) { // Use imported type
	fmt.Fprintln(os.Stderr, "Usage information for this reflex:")
	fmt.Fprintln(os.Stderr, "----------------------------------")
	if m.Description != "" {
		// Indent description
		for _, line := range strings.Split(m.Description, "\n") {
			fmt.Fprintf(os.Stderr, "  %s\n", line)
		}
	}
	fmt.Fprintln(os.Stderr, "")
	if len(requiredInputs) > 0 {
		fmt.Fprintln(os.Stderr, "Required environment variables (check manifest.yml for full details):")
		for _, desc := range requiredInputs {
			fmt.Fprintln(os.Stderr, desc)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Consult manifest.yml for input details.")
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Provide necessary environment variables and/or command-line arguments when running the container.")
}

// executeCommand replaces the current process with the target command.
// It looks up the command in PATH.
func executeCommand(cmdPath string, cmdArgs []string) {
	// Find the absolute path of the command
	resolvedPath, err := exec.LookPath(cmdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find command '%s' in PATH: %v\n", cmdPath, err)
		os.Exit(127) // Command not found exit code
	}

	// Use syscall.Exec to replace the current process.
	// Note: This is generally preferred for entrypoints over os/exec.Command().Run()
	// as it avoids leaving the helper process running.
	// However, direct syscall usage can be less portable.
	// os/exec.Command is simpler if direct syscall isn't strictly needed.
	// Let's use os/exec for broader compatibility first.

	cmd := exec.Command(resolvedPath, cmdArgs[1:]...) // Pass args *after* the command name
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command, replacing this process is harder with os/exec
	// without resorting to syscall.Exec on specific OSes.
	// Running it as a subprocess is usually acceptable.
	err = cmd.Run()

	// Handle exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error executing command '%s': %v\n", cmdPath, err)
			os.Exit(1) // Generic error exit code
		}
	}
	os.Exit(0) // Success
}