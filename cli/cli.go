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

func Format(dependencies graph.Graph, schedule graph.Graph, impacted []string, dotFormat bool, printDependencies bool) string {
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
func Print(dependencyFilesGlob string, dotFormat bool, printDependencies bool) (graph.Graph, graph.Graph, []string, error) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Error finding dependency manifests: %s", err)
	}

	_, deps, errs := manifests.Read(paths, false)
	if errs != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	dependencies := deps.AsGraph()
	selection := dependencies.Vertices() // everything

	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	return dependencies, buildSchedule, selection, nil
}

// Diff is 'monobuild diff'
func Diff(dependencyFilesGlob string, mainBranch bool, baseBranch string, includeStrong bool, dotFormat bool, printDependencies bool) (graph.Graph, graph.Graph, []string, error) {
	manifestFiles, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("error finding dependency manifests: %s", err)
	}

	// Find components and dependency manifests
	components, deps, errs := manifests.Read(manifestFiles, false)
	if errs != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	// Get changed files
	changes, err := diff.ChangedFiles(mainBranch, baseBranch)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Find impacted components
	dependencies := deps.AsGraph()
	impacted := diff.Impacted(changedComponents, dependencies)

	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	if includeStrong {
		strong := buildSchedule.Descendants(impacted)
		impacted = append(impacted, strong...)
	}

	return dependencies, buildSchedule, impacted, nil
}
