package cmd

import (
	"fmt"
	"log"

	"github.com/charypar/monobuild/cli"
	"github.com/spf13/cobra"
)

var baseBranch string
var mainBranch bool
var dotHighlight bool

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Build schedule for components affected by git changes",
	Long: `Create a build schedule based on git history and dependency graph.
Each line in the output is a component and its dependencies. 
The format of each line is:

<component>: <dependency>, <dependency>, <dependency>, ...

Diff can output either the build schedule (using only strong dependencies) or 
the original dependeny graph (using all dependencies).`,
	Run: diffFn,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().BoolVar(&mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
	diffCmd.Flags().BoolVar(&printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	diffCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
}

func diffFn(cmd *cobra.Command, args []string) {
	dependencies, schedule, impacted, err := cli.Diff(dependencyFilesGlob, mainBranch, baseBranch, dotFormat, printDependencies)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cli.Format(dependencies, schedule, impacted, dotFormat, printDependencies))
}
