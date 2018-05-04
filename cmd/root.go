package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "monobuild",
	Short: "A build orchestration tool for Continuous Integration in a monorepo.",
	Long: `Monobuild is a simple tool that understands a graph of dependencies in
a monorepo codebase (where separate components live side by side) and
based on it can decide what should be built, given a set of changes.`,
}

var dependencyFilesGlob string

func init() {
	rootCmd.PersistentFlags().StringVar(&dependencyFilesGlob, "dependency-files", "**/Dependencies", "Search pattern for dependency files")
}

// Execute the CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
