package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	// Import the shared types from the internal package
	"nhi/basetools/pkg/manifesttypes"
)

// Mandated manifest path
const manifestPath = "/manifest.yml"

// Base path for mounted inputs/outputs
const appIOBasePath = "/app"

// --- Helper Logic ---

func main() {
	// --- Logger Setup ---
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// --- Flag Definition ---
	showHelpShort := flag.Bool("h", false, "Show help message")
	showHelpLong := flag.Bool("help", false, "Show help message")
	flag.Parse() // Parse command-line flags

	// --- Early Exits: SHOW_MANIFEST or Help Flags ---

	// Check for SHOW_MANIFEST first
	if strings.ToLower(os.Getenv("SHOW_MANIFEST")) == "true" {
		showManifest()
		os.Exit(0)
	}

	// Check if help was requested via flags
	if *showHelpShort || *showHelpLong {
		// Try to read manifest for more informative help
		manifestData, err := os.ReadFile(manifestPath)
		var m manifesttypes.Manifest
		if err == nil {
			_ = yaml.Unmarshal(manifestData, &m) // Ignore parsing errors for help display
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Could not read manifest %s for help: %v\n", manifestPath, err)
		}
		printUsage(m, nil) // Pass empty slice for required envs when just showing help
		os.Exit(0)
	}

	// --- Regular Execution Logic --- //

	// --- Get Command Args ---
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No command provided to the entrypoint helper.")
		printUsage(manifesttypes.Manifest{}, nil)
		os.Exit(1)
	}
	targetCmdArgs := flag.Args()
	targetCmdPath := targetCmdArgs[0]

	// Read and parse manifest (required for validation and env export)
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading manifest %s: %v\n", manifestPath, err)
		// Still attempt execution if manifest is unreadable, as per original logic
		executeCommand(logger, targetCmdPath, targetCmdArgs, os.Environ()) // Pass original env
		return
	}

	var m manifesttypes.Manifest
	if err := yaml.Unmarshal(manifestData, &m); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not parse manifest %s: %v\n", manifestPath, err)
		// Still attempt execution if manifest is unparseable
		executeCommand(logger, targetCmdPath, targetCmdArgs, os.Environ()) // Pass original env
		return
	}

	// Prepare environment variables
	envVars := os.Environ() // Start with current environment
	exportedEnvVars := []string{} // Track vars added by helper
	validatedInputPaths := make(map[string]string)
	validatedOutputPaths := make(map[string]string)

	// --- Validate Inputs/Outputs and Prepare Env Vars (using parsed manifest 'm') ---
	logger.Info("Validating manifest inputs...")
	for name := range m.InputPaths {
		inputPath := filepath.Join(appIOBasePath, "input_"+name)
		logger.Info("Checking input", "name", name, "path", inputPath)
		_, err := os.Stat(inputPath) // Validation still happens as the process user
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: Required input '%s' not found at expected path: %s\n", name, inputPath)
			} else {
				fmt.Fprintf(os.Stderr, "Error checking input path %s for '%s': %v\n", inputPath, name, err)
			}
			os.Exit(1)
		}
		envVarName := "INPUT_" + strings.ToUpper(name)
		envVar := fmt.Sprintf("%s=%s", envVarName, inputPath)
		exportedEnvVars = append(exportedEnvVars, envVar)
		validatedInputPaths[envVarName] = inputPath
	}
	logger.Info("Validating manifest outputs...")
	for name := range m.OutputPaths {
		outputPath := filepath.Join(appIOBasePath, "output_"+name)
		logger.Info("Checking output", "name", name, "path", outputPath)
		info, err := os.Stat(outputPath) // Validation still happens as the process user
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: Required output directory '%s' not found at expected path: %s\n", name, outputPath)
			} else {
				fmt.Fprintf(os.Stderr, "Error checking output path %s for '%s': %v\n", outputPath, name, err)
			}
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: Output path '%s' (%s) is not a directory.\n", name, outputPath)
			os.Exit(1)
		}
		// Perform write test (should work if --user flag was correct)
		tempFile, err := os.CreateTemp(outputPath, ".writetest-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Output directory '%s' (%s) is not writable by current user (UID: %d, GID: %d): %v\n", name, outputPath, os.Geteuid(), os.Getegid(), err)
			os.Exit(1)
		}
		tempFile.Close()
		os.Remove(tempFile.Name())
		envVarName := "OUTPUT_" + strings.ToUpper(name)
		envVar := fmt.Sprintf("%s=%s", envVarName, outputPath)
		exportedEnvVars = append(exportedEnvVars, envVar)
		validatedOutputPaths[envVarName] = outputPath
	}
	logger.Info("Exporting derived environment variables:")
	for k, v := range validatedInputPaths {
		logger.Info("  Exporting", "var", k, "value", v)
	}
	for k, v := range validatedOutputPaths {
		logger.Info("  Exporting", "var", k, "value", v)
	}

	// --- Check Required Environment Variables (from manifest 'environment' section) ---
	missingRequired := false
	requiredInputsDesc := []string{}
	for name, spec := range m.Environment {
		if spec.Required {
			val := os.Getenv(name) // Check original env
			if val == "" {
				missingRequired = true
				desc := fmt.Sprintf("  - %s: %s", name, spec.Description)
				requiredInputsDesc = append(requiredInputsDesc, desc)
			}
		}
	}

	if missingRequired {
		fmt.Fprintln(os.Stderr, "\nError: Missing required environment variables.")
		printUsage(m, requiredInputsDesc) // Print usage with specific missing vars
		os.Exit(1)
	}

	// --- Execute Command --- //
	logger.Info("Executing command", "cmd", targetCmdArgs)
	// Combine initial env with helper-exported vars
	finalEnv := append(envVars, exportedEnvVars...)
	executeCommand(logger, targetCmdPath, targetCmdArgs, finalEnv)
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

func printUsage(m manifesttypes.Manifest, requiredEnvVars []string) {
	fmt.Fprintln(os.Stderr, "Usage: <docker run options> <image> [-h|--help] <command> [args...]")
	fmt.Fprintln(os.Stderr, "-----------------------------------------------------------------")
	if m.Description != "" {
		fmt.Fprintln(os.Stderr, "Description:")
		for _, line := range strings.Split(m.Description, "\n") {
			fmt.Fprintf(os.Stderr, "  %s\n", line)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	// Print required env vars if provided (means we are exiting due to missing vars)
	if len(requiredEnvVars) > 0 {
		fmt.Fprintln(os.Stderr, "Required Environment Variables (must be set via -e or similar):")
		for _, desc := range requiredEnvVars {
			fmt.Fprintln(os.Stderr, desc)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	// Print general environment variable info
	fmt.Fprintln(os.Stderr, "Optional Environment Variables:")
	fmt.Fprintln(os.Stderr, "  SHOW_MANIFEST=true: Print the raw manifest.yml content to stdout and exit.")
	fmt.Fprintln(os.Stderr, "                      Example: docker run --rm -e SHOW_MANIFEST=true <image>")
	fmt.Fprintln(os.Stderr, "  (Consult manifest.yml for other environment variables used by the reflex)")
	fmt.Fprintln(os.Stderr, "")

	// Print expected mount points based on manifest
	if len(m.InputPaths) > 0 || len(m.OutputPaths) > 0 {
		fmt.Fprintln(os.Stderr, "Expected Mount Points (must be provided via -v or similar):")
	}
	if len(m.InputPaths) > 0 {
		fmt.Fprintln(os.Stderr, "  Inputs (mounted read-only):")
		for name, spec := range m.InputPaths {
			fmt.Fprintf(os.Stderr, "    -v /host/path/to/%s:/app/input_%s:ro  (%s)\n", name, name, spec.Description)
		}
	}
	if len(m.OutputPaths) > 0 {
		fmt.Fprintln(os.Stderr, "  Outputs (mounted read-write):")
		for name, spec := range m.OutputPaths {
			fmt.Fprintf(os.Stderr, "    -v /host/path/to/%s:/app/output_%s  (%s)\n", name, name, spec.Description)
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Arguments:")
	fmt.Fprintln(os.Stderr, "  <command> [args...] : The command and arguments the reflex should execute.")
}

// shellEscape wraps a string in single quotes, escaping any existing single quotes
// suitable for use within sh -c "..."
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// executeCommand uses sh -c to ensure environment propagation.
func executeCommand(logger *slog.Logger, cmdPath string, cmdArgs []string, envVars []string) {
	// Verify the target script exists (as the process user)
	resolvedPath, err := exec.LookPath(cmdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find command '%s' in PATH: %v\n", cmdPath, err)
		os.Exit(127)
	}

	// --- Build the shell command string ---

	// 1. Prepare export statements
	var exports []string
	// NOTE: We iterate over *all* envVars now (initial + exported)
	// This ensures PATH etc. are set correctly for the shell itself.
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			// Export format: export VAR='escaped_value'
			exports = append(exports, fmt.Sprintf("export %s=%s", parts[0], shellEscape(parts[1])))
		}
	}
	exportString := ""
	if len(exports) > 0 {
		// Join with semicolons, add trailing semicolon and space for separation
		exportString = strings.Join(exports, "; ") + "; "
	}

	// 2. Prepare the exec command and arguments
	var commandParts []string
	commandParts = append(commandParts, "exec") // Use exec to replace the shell
	commandParts = append(commandParts, shellEscape(resolvedPath))
	for _, arg := range cmdArgs[1:] {
		commandParts = append(commandParts, shellEscape(arg))
	}
	commandString := strings.Join(commandParts, " ")

	// 3. Combine exports and command
	fullCommand := exportString + commandString

	// Log the command to be executed (consider DEBUG level if too verbose)
	logger.Info("Executing via sh -c", "command", fullCommand)

	// --- Execute the command string using sh ---
	cmd := exec.Command("/bin/sh", "-c", fullCommand)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.Env is not needed as exports are part of the command string

	err = cmd.Run()

	// --- Handle exit code ---
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error executing command '%s' via sh: %v\n", cmdPath, err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}