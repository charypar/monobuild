package cmd

import (
	"fmt"

	"github.com/bmatcuk/doublestar"

	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
	"github.com/charypar/monobuild/output"
	"github.com/spf13/cobra"
)

var printDependencies bool
var dotFormat bool

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print the build schedule",
	Long:  `Read the dependency graph from dependency manifests, check all dependencies exist and print the build schedule`,
	Run:   printFn,
}

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.Flags().BoolVar(&printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	printCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
}

func printGraph(dependencies map[string][]manifests.Dependency, buildSchedule map[string][]string, dependencyGraph map[string][]string) {
	if dotFormat && printDependencies {
		fmt.Print(output.Dot(dependencies, dependencyGraph))
		return
	}

	if dotFormat {
		fmt.Print(output.DotSchedule(dependencies, buildSchedule))
		return
	}

	if printDependencies {
		fmt.Print(output.Text(dependencyGraph))
		return
	}

	fmt.Print(output.Text(buildSchedule))
}

func printFn(cmd *cobra.Command, args []string) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	_, dependencies, errs := manifests.Read(paths, false)
	if errs != nil {
		fmt.Print(joinErrors("cannot load dependencies:", errs))
	}

	if errs != nil && dotFormat {
		return
	}

	dependencyGraph := manifests.Filter(dependencies, 0)
	buildSchedule := graph.New(manifests.Filter(dependencies, 2)).Reverse().AsStrings()

	printGraph(dependencies, buildSchedule, dependencyGraph)
}
