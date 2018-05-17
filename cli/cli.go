package cli

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/diff"
	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
)

func joinErrors(message string, errors []error) error {
	errstrings := make([]string, len(errors))
	for i, e := range errors {
		errstrings[i] = string(e.Error())
	}

	return fmt.Errorf("%s\n%s", message, strings.Join(errstrings, "\n"))
}

func printGraph(dependencies graph.Graph, schedule graph.Graph, impacted []string, dotFormat bool, printDependencies bool) string {
	if dotFormat && printDependencies {
		return dependencies.Dot(impacted)
	}

	if dotFormat {
		return schedule.DotSchedule(impacted)
	}

	if printDependencies {
		return dependencies.Text(impacted)
	}

	return schedule.Text(impacted)
}

// Print is 'monobuild print'
func Print(dependencyFilesGlob string, dotFormat bool, printDependencies bool) (string, error) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		return "", fmt.Errorf("Error finding dependency manifests: %s", err)
	}

	_, deps, errs := manifests.Read(paths, false)
	if errs != nil {
		return "", fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	dependencies := deps.AsGraph()
	selection := dependencies.Vertices() // everything

	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	return printGraph(dependencies, buildSchedule, selection, dotFormat, printDependencies), nil
}

// Diff is 'monobuild diff'
func Diff(dependencyFilesGlob string, mainBranch bool, baseBranch string, dotFormat bool, printDependencies bool) (string, error) {
	manifestFiles, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		return "", fmt.Errorf("error finding dependency manifests: %s", err)
	}

	// Find components and dependency manifests
	components, deps, errs := manifests.Read(manifestFiles, false)
	if errs != nil {
		return "", fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	// Get changed files
	changes, err := diff.ChangedFiles(mainBranch, baseBranch)
	if err != nil {
		return "", fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Find impacted components
	dependencies := deps.AsGraph()
	impacted := diff.Impacted(changedComponents, dependencies)

	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	return printGraph(dependencies, buildSchedule, impacted, dotFormat, printDependencies), nil
}
