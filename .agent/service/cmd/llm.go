package cmd

import (
	"fmt"
	_ "embed" // Import the embed package for side effects

	"github.com/spf13/cobra"
)

//go:embed llm_cli_help.txt
var llmHelpTextContent string // Variable to hold embedded content

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Prints guidance on how an LLM agent should use this CLI.",
	Long: `Prints a detailed explanation of the standard interaction flow
for an LLM agent using the 'agentt summary' and 'agentt details' commands
to retrieve guidance information embedded within the binary.`,
	Run: func(cmd *cobra.Command, args []string) { // Changed back to Run, no error expected
		fmt.Println(llmHelpTextContent)
	},
}

func init() {
	// Add llmCmd directly to the rootCmd.
	rootCmd.AddCommand(llmCmd)
}
