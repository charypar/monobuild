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

	if !dotFormat {
		var g map[string][]string

		if printDependencies {
			g = dependencyGraph
		} else {
			g = buildSchedule
		}

		for c, d := range g {
			fmt.Printf("%s: %s\n", c, strings.Join(d, ", "))
		}
		return
	}

	fmt.Println("digraph graphname {")

	if printDependencies {
		for c, deps := range dependencies {
			for _, d := range deps {
				var format string

				if d.Kind == manifests.Strong {
					format = ""
				} else {
					format = " [style=dashed]"
				}

				fmt.Printf("  \"%s\" -> \"%s\"%s\n", c, d.Name, format)
			}
		}
	} else {
		fmt.Println("  rankdir=\"LR\"")
		fmt.Println("  node [shape=box]")

		for c, deps := range buildSchedule {
			if len(deps) < 1 {
				fmt.Printf("  \"%s\"\n", c)
			}

			for _, d := range deps {
				fmt.Printf("  \"%s\" -> \"%s\"\n", c, d)
			}
		}
	}

	fmt.Println("}")
}
