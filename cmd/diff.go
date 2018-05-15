package cmd

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/diff"
	"github.com/charypar/monobuild/manifests"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "List changed components",
	Long:  `List changed components based on git history and dependency graph`,
	Run:   diffFn,
}

var baseBranch string
var mainBranch bool
var dotHighlight bool

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().BoolVar(&mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
	diffCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
	diffCmd.Flags().BoolVar(&dotHighlight, "dot-highlight", false, "Print in DOT format highlighting changed nodes rather than omitting the unchanged ones")
}

func diffFn(cmd *cobra.Command, args []string) {
	manifestFiles, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	// Get changed files
	changes, err := diff.ChangedFiles(mainBranch, baseBranch)
	if err != nil {
		panic(fmt.Errorf("cannot find changes: %s", err))
	}

	// Find components and dependency manifests
	components, dependencies, errs := manifests.Read(manifestFiles, true)
	if errs != nil {
		errstrings := make([]string, len(errs))
		for i, e := range errs {
			errstrings[i] = string(e.Error())
		}

		panic(fmt.Errorf("cannot load dependencies:\n%s", strings.Join(errstrings, "\n")))
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Calculate build schedule
	buildSchedule, err := diff.Diff(changedComponents, dependencies, baseBranch, mainBranch)
	if err != nil {
		panic(fmt.Errorf("Error finding out changed components: %s", err))
	}

	if !dotFormat {
		for c, d := range buildSchedule {
			fmt.Printf("%s: %s\n", c, strings.Join(d, ", "))
		}

		return
	}

	fmt.Println("digraph graphname {")
	fmt.Println("  rankdir=\"LR\"")
	fmt.Println("  node [shape=box]")

	for c, deps := range buildSchedule {
		if len(deps) < 1 {
			fmt.Printf("  \"%s\"", c)
		}

		for _, d := range deps {
			fmt.Printf("  \"%s\" -> \"%s\"\n", c, d)
		}
	}

	fmt.Println("}")
}
