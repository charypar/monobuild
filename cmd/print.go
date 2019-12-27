package cmd

import (
	"fmt"
	"log"

	"github.com/charypar/monobuild/cli"
	"github.com/spf13/cobra"
)

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

	printCmd.Flags().BoolVar(&commonOpts.printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	printCmd.Flags().BoolVar(&commonOpts.dotFormat, "dot", false, "Print in DOT format for GraphViz")
	printCmd.Flags().BoolVar(&commonOpts.printFull, "full", false, "Print the full dependency graph including strengths")

}

func printFn(cmd *cobra.Command, args []string) {
	// first we tediously process the CLI flags

	var format cli.OutputFormat
	if commonOpts.dotFormat {
		format = cli.Dot
	} else {
		format = cli.Text
	}

	scope := cli.Scope{Scope: commonOpts.scope, TopLevel: commonOpts.topLevel}

	var outType cli.OutputType
	if commonOpts.printFull {
		outType = cli.Full
	} else if commonOpts.printDependencies {
		outType = cli.Dependencies
	} else {
		outType = cli.Schedule
	}

	outputOpts := cli.OutputOptions{Format: format, Type: outType}

	// then we run the CLI

	dependencies, schedule, impacted, err := cli.Print(commonOpts.dependencyFilesGlob, scope)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cli.Format(dependencies, schedule, impacted, outputOpts))
}
