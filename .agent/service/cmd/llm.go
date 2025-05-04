package cmd

import (
	"embed"
	"fmt"
	"log/slog"
	// "agentt/internal/server"

	"github.com/spf13/cobra"
)

//go:embed llm_cli_help.txt
var embeddedHelpText embed.FS

const llmHelpTextPath = "llm_cli_help.txt"

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Provides instructions for LLM interaction (internal use)",
	Long: `Outputs the standard interaction protocol for LLMs interacting with the Agentt service.
This is primarily intended for internal use during development and testing.
It prints the expected API endpoints and interaction flow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Displaying LLM interaction protocol")

		// Read the embedded help text file
		// TODO: Verify embedding and path access are correct
		helpTextBytes, err := embeddedHelpText.ReadFile(llmHelpTextPath)
		if err != nil {
			slog.Error("Failed to read embedded LLM help text", "path", llmHelpTextPath, "error", err)
			return fmt.Errorf("failed to read internal help text: %w", err)
		}

		fmt.Println("```")
		fmt.Println(string(helpTextBytes))
		fmt.Println("```")
		return nil
	},
}

func init() {
	// No specific flags for llm command
}
