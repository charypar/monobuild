package cmd

import (
	"fmt"
	"log"

	"github.com/charypar/monobuild/cli"
	"github.com/charypar/monobuild/diff"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	baseBranch    string
	baseCommit    string
	mainBranch    bool
	rebuildStrong bool
	dotHighlight  bool
}

var diffOpts diffOptions

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

	diffCmd.Flags().StringVar(&diffOpts.baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().StringVar(&diffOpts.baseCommit, "base-commit", "HEAD^1", "Base commit to compare with (useful in main-brahnch mode when using rebase merging)")
	diffCmd.Flags().BoolVar(&diffOpts.mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
	diffCmd.Flags().BoolVar(&diffOpts.rebuildStrong, "rebuild-strong", false, "Include all strong dependencies of affected components")
	diffCmd.Flags().BoolVar(&commonOpts.printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	diffCmd.Flags().BoolVar(&commonOpts.dotFormat, "dot", false, "Print in DOT format for GraphViz")
}

func diffFn(cmd *cobra.Command, args []string) {
	// first we tediously process the CLI flags

	var branchMode diff.BranchMode
	if diffOpts.mainBranch {
		branchMode = diff.Main
	} else {
		branchMode = diff.Feature
	}

	var format cli.OutputFormat
	if commonOpts.dotFormat {
		format = cli.Dot
	} else {
		format = cli.Text
	}

	diffMode := diff.Mode{Mode: branchMode, BaseBranch: diffOpts.baseBranch, BaseCommit: diffOpts.baseCommit}
	scope := cli.Scope{Scope: commonOpts.scope, TopLevel: commonOpts.topLevel}

	var outType cli.OutputType
	if commonOpts.printDependencies {
		outType = cli.Dependencies
	} else {
		outType = cli.Schedule
	}

	outputOpts := cli.OutputOptions{Format: format, Type: outType}

	// run the CLI command

	dependencies, schedule, impacted, err := cli.Diff(commonOpts.dependencyFilesGlob, diffMode, scope, diffOpts.rebuildStrong)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cli.Format(dependencies, schedule, impacted, outputOpts))
}
