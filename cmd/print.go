package cmd

import (
	"fmt"
	"log"

	"github.com/charypar/monobuild/cli"
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

func printFn(cmd *cobra.Command, args []string) {
	dependencies, schedule, impacted, err := cli.Print(dependencyFilesGlob, dotFormat, printDependencies)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cli.Format(dependencies, schedule, impacted, dotFormat, printDependencies))
}
