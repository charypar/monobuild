package cmd

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/manifests"
	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print the dependency graph",
	Long:  `Read the dependency graph from dependency manifests, check all dependencies exist and pretty print it`,
	Run:   printFn,
}

var dotFormat bool

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.Flags().BoolVar(&dotFormat, "dot", false, "Output the dependencies in DOT format for GraphViz")
}

func printFn(cmd *cobra.Command, args []string) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	components, dependencies, errs := manifests.Read(paths, false)
	if errs != nil {
		for _, e := range errs {
			fmt.Printf("Error: %s\n", e)
		}
		fmt.Println("")
	}

	if errs != nil && dotFormat {
		return
	}

	if !dotFormat {
		fmt.Printf("Found %d component(s). Dependency structure:\n\n", len(components))
		for c, d := range dependencies {
			fmt.Printf("%s -> %s\n", c, strings.Join(d, ", "))
		}
	} else {
		fmt.Println("digraph graphname {")

		for c, deps := range dependencies {
			for _, d := range deps {
				fmt.Printf("  %s -> %s\n", c, d)
			}
		}

		fmt.Println("}")
	}
}
