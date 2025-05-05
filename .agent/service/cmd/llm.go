package cmd

import (
	_ "embed" // Use blank import for go:embed
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

// Embed the help text directly into a string.
//
//go:embed llm_cli_help.txt
var llmCliHelpContent string

// const llmHelpTextPath = "llm_cli_help.txt" // REMOVED: No longer needed

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Provides instructions for LLM interaction (internal use)",
	Long: `Outputs the standard interaction protocol for LLMs interacting with the Agentt service.
This is primarily intended for internal use during development and testing.
It prints the expected API endpoints and interaction flow.`,
	RunE: runLLM,
}

func init() {
	// No specific flags for llm command
}

// runLLM executes the llm command logic.
func runLLM(cmd *cobra.Command, args []string) error {
	// This command only prints static, embedded help text from llm_cli_help.txt.
	// Backend initialization is not required.
	slog.Debug("Executing llm command to print embedded CLI help")

	// Print the directly embedded string content.
	fmt.Print(llmCliHelpContent)
	return nil
}
