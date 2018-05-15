package cmd

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/diff"
	"github.com/charypar/monobuild/manifests"
	"github.com/spf13/cobra"
)

var baseBranch string
var mainBranch bool
var dotHighlight bool

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "List changed components",
	Long:  `List changed components based on git history and dependency graph`,
	Run:   diffFn,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().BoolVar(&mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
	diffCmd.Flags().BoolVar(&printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	diffCmd.Flags().BoolVar(&dotFormat, "dot", false, "Print in DOT format for GraphViz")
	diffCmd.Flags().BoolVar(&dotHighlight, "dot-highlight", false, "Print in DOT format highlighting changed nodes rather than omitting the unchanged ones")
}

func joinErrors(message string, errors []error) error {
	errstrings := make([]string, len(errors))
	for i, e := range errors {
		errstrings[i] = string(e.Error())
	}

	return fmt.Errorf("%s\n%s", message, strings.Join(errstrings, "\n"))
}

func diffFn(cmd *cobra.Command, args []string) {
	manifestFiles, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("error finding dependency manifests: %s", err))
	}

	// Find components and dependency manifests
	components, dependencies, errs := manifests.Read(manifestFiles, false)
	if errs != nil {
		panic(joinErrors("cannot load dependencies:", errs))
	}

	// Get changed files
	changes, err := diff.ChangedFiles(mainBranch, baseBranch)
	if err != nil {
		panic(fmt.Errorf("cannot find changes: %s", err))
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Calculate build schedule
	buildSchedule := diff.BuildSchedule(changedComponents, dependencies)
	dependencyGraph := diff.Dependencies(changedComponents, dependencies)

	printGraph(dependencies, buildSchedule, dependencyGraph)
}
