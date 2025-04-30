package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "agentt",
		Short: "Agent Guidance Service and CLI",
		Long: `Agentt provides an HTTP server for agent guidance discovery
and a command-line interface for interacting with the same definitions.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agentt.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add subcommands here (e.g., serverCmd, validateCmd)
	// Example: rootCmd.AddCommand(serverCmd)
	// Example: rootCmd.AddCommand(validateCmd)
}

// Helper function for handling errors
func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
