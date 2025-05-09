package cmd

import (
	_ "embed" // Use blank import for go:embed
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

//go:embed llm_cli_help.txt
var llmCliHelpContent string

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Provides instructions for LLM interaction (internal use)",
	Long: `Outputs the standard interaction protocol for LLMs interacting with the Agentt service.
This is primarily intended for internal use during development and testing.
It prints the expected API endpoints and interaction flow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Executing llm command to print embedded CLI help")
		fmt.Print(llmCliHelpContent)
		return nil
	},
}

func init() {
	// No specific flags for llm command
}
