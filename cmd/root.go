package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "monobuild",
	Short: "A build orchestration tool for Continuous Integration in a monorepo.",
	Long: `Read a graph of dependencies in a monorepo codebase (where separate 
components live side by side) and decide what should be built, given the git 
history.`,
}

type commonOptions struct {
	dependencyFilesGlob string
	scope               string
	topLevel            bool
	printDependencies   bool
	dotFormat           bool
}

var commonOpts commonOptions

func init() {
	rootCmd.PersistentFlags().StringVar(&commonOpts.dependencyFilesGlob, "dependency-files", "**/Dependencies", "Search pattern for dependency files")
	rootCmd.PersistentFlags().StringVar(&commonOpts.scope, "scope", "", "Scope output to a single component and its dependencies")
	rootCmd.PersistentFlags().BoolVar(&commonOpts.topLevel, "top-level", false, "Only list top-level components that nothing depends on")
}

// Execute the CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
