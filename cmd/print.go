package cmd

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"

	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print the build schedule",
	Long:  `Read the dependency graph from dependency manifests, check all dependencies exist and print the build schedule`,
	Run:   printFn,
}

var printDependencies bool
var dotFormat bool

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.Flags().BoolVar(&printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	printCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
}

func printText(graph map[string][]string) string {
	var result string

	for c, d := range graph {
		result += fmt.Sprintf("%s: %s\n", c, strings.Join(d, ", "))
	}

	return result
}

func printDotGraph(dependencies map[string][]manifests.Dependency, graph map[string][]string) string {
	result := fmt.Sprintln("digraph dependencies {")

	for c, deps := range dependencies {
		for _, d := range deps {
			var format string

			if d.Kind == manifests.Weak {
				format = " [style=dashed]"
			}

			result += fmt.Sprintf("  \"%s\" -> \"%s\"%s\n", c, d.Name, format)
		}
	}

	return result + "}\n"
}

func printDotSchedule(dependencies map[string][]manifests.Dependency, graph map[string][]string) string {
	result := fmt.Sprintln("digraph schedule {\n  rankdir=\"LR\"\n  node [shape=box]")

	for c, deps := range graph {
		if len(deps) < 1 {
			result += fmt.Sprintf("  \"%s\"\n", c)
		}

		for _, d := range deps {
			result += fmt.Sprintf("  \"%s\" -> \"%s\"\n", c, d)
		}
	}

	return result + "}\n"
}

func printGraph(dependencies map[string][]manifests.Dependency, buildSchedule map[string][]string, dependencyGraph map[string][]string) {
	if dotFormat && printDependencies {
		fmt.Print(printDotGraph(dependencies, dependencyGraph))
		return
	}

	if dotFormat {
		fmt.Print(printDotSchedule(dependencies, buildSchedule))
		return
	}

	if printDependencies {
		fmt.Print(printText(dependencyGraph))
		return
	}

	fmt.Print(printText(buildSchedule))
}

func printFn(cmd *cobra.Command, args []string) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	_, dependencies, errs := manifests.Read(paths, false)
	if errs != nil {
		for _, e := range errs {
			fmt.Printf("Error: %s\n", e)
		}
		fmt.Println("")
	}

	if errs != nil && dotFormat {
		return
	}

	dependencyGraph := manifests.Filter(dependencies, 0)
	buildSchedule := graph.New(manifests.Filter(dependencies, 2)).Reverse().AsStrings()

	printGraph(dependencies, buildSchedule, dependencyGraph)
}
