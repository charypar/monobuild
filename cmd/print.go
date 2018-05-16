package cmd

import (
	"fmt"

	"github.com/bmatcuk/doublestar"

	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
	"github.com/spf13/cobra"
)

var printDependencies bool
var dotFormat bool

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print the full build schedule or dependency graph",
	Long: `Print the full build schedule or dependency graph based on the manifest files.
The format of each line is:

<component>: <dependency>, <dependency>, <dependency>, ...

Diff can output either the build schedule (using only strong dependencies) or 
the original dependeny graph (using all dependencies).`,
	Run: printFn,
}

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.Flags().BoolVar(&printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	printCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
}

func printGraph(dependencies graph.Graph, schedule graph.Graph, impacted []string) {
	if dotFormat && printDependencies {
		fmt.Print(dependencies.Dot(impacted))
		return
	}

	if dotFormat {
		fmt.Print(schedule.DotSchedule(impacted))
		return
	}

	if printDependencies {
		fmt.Print(dependencies.Text(impacted))
		return
	}

	fmt.Print(schedule.Text(impacted))
}

func printFn(cmd *cobra.Command, args []string) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	_, deps, errs := manifests.Read(paths, false)
	if errs != nil {
		fmt.Print(joinErrors("cannot load dependencies:", errs))
	}

	if errs != nil && dotFormat {
		return
	}

	// this is somewhat redundant in the print case
	dependencies := deps.AsGraph()
	selection := dependencies.Vertices() // everything

	buildSchedule := dependencies.FilterEdges([]int{2})

	printGraph(dependencies, buildSchedule, selection)
}
