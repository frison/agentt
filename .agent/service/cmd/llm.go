package cmd

import (
	_ "embed" // Import the embed package for side effects
	// "errors"
	"fmt"
	"log"
	"os"
	"strings"

	// "agentt/internal/config"
	"github.com/spf13/cobra"
)

//go:embed llm_cli_help.txt
var llmHelpTextContent string // Variable to hold embedded content

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Prints guidance on how an LLM agent should use this CLI.",
	Long: `Provides instructions for an LLM on how to interact with the agentt CLI.
It details the flow: fetch summary, identify relevant IDs, fetch details.
Configuration is loaded via --config flag, AGENTT_CONFIG env var, or default search paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Use common setup (Optional for basic help text, but good practice) ---
		// We attempt setup primarily to check for config errors that might prevent usage.
		_, err := setupDiscovery(rootConfigPath)
		if err != nil {
			// We can still show the help text even if the scan fails (e.g., no files found),
			// but not if the config itself couldn't be loaded.
			// Let's check for specific config loading errors.
			if strings.Contains(err.Error(), "configuration error") {
				// If the core config loading failed, return the error.
				return err
			}
			// For other errors (like scan errors), log but continue.
			log.Printf("Warning during setup for llm command: %v. Proceeding with help text.", err)
		}

		// --- Get Path to Agentt Binary (Potentially use configDir from setupRes later if needed) ---
		agenttPath, err := os.Executable()
		if err != nil {
			log.Printf("Warning: failed to get executable path, examples may be incorrect: %v", err)
			agenttPath = "agentt" // Fallback
		}

		// Simple placeholder replacement for now
		output := strings.ReplaceAll(llmHelpTextContent, "{{AGENTT_EXECUTABLE_PATH}}", agenttPath)

		fmt.Println(output)
		return nil
	},
}

func init() {
	// Add llmCmd directly to the rootCmd.
	rootCmd.AddCommand(llmCmd)
}
