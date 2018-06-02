package cli

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/diff"
	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
	"github.com/charypar/monobuild/set"
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
func Print(dependencyFilesGlob string, scope string, topLevel bool) (graph.Graph, graph.Graph, []string, error) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Error finding dependency manifests: %s", err)
	}

	components, deps, errs := manifests.Read(paths, false)
	if errs != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	dependencies := deps.AsGraph()
	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	selection := dependencies.Vertices()

	if scope != "" {
		var scoped []string

		// ensure valid scope
		for _, c := range components {
			if c == scope {
				scoped = []string{scope}
			}
		}

		if len(scoped) < 1 {
			return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Cannot scope to '%s', not a component", scope)
		}

		selection = append(dependencies.Descendants(scoped), scoped...)
	}

	if topLevel {
		reverse := dependencies.Reverse()
		vertices := dependencies.Vertices()

		topLevel := make([]string, 0, len(vertices))
		for i := range vertices {
			if len(reverse.Children(vertices[i:i+1])) < 1 {
				topLevel = append(topLevel, vertices[i])
			}
		}

		selection = set.New(selection).Intersect(set.New(topLevel)).AsStrings()
	}

	return dependencies, buildSchedule, selection, nil
}

// Diff is 'monobuild diff'
func Diff(dependencyFilesGlob string, mainBranch bool, baseBranch string, baseCommit string, includeStrong bool, scope string, topLevel bool) (graph.Graph, graph.Graph, []string, error) {
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
	changes, err := diff.ChangedFiles(mainBranch, baseBranch, baseCommit)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Find impacted components
	dependencies := deps.AsGraph()
	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	impacted := diff.Impacted(changedComponents, dependencies)

	if scope != "" {
		var scoped []string

		// ensure valid scope
		for _, c := range components {
			if c == scope {
				scoped = []string{scope}
			}
		}

		if len(scoped) < 1 {
			return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Cannot scope to '%s', not a component", scope)
		}

		scopedAndDeps := append(dependencies.Descendants(scoped), scoped...)
		impacted = set.New(impacted).Intersect(set.New(scopedAndDeps)).AsStrings()
	}

	if topLevel {
		reverse := dependencies.Reverse()
		vertices := dependencies.Vertices()

		topLevel := make([]string, 0, len(vertices))
		for i := range vertices {
			if len(reverse.Children(vertices[i:i+1])) < 1 {
				topLevel = append(topLevel, vertices[i])
			}
		}

		impacted = set.New(impacted).Intersect(set.New(topLevel)).AsStrings()
	}

	// needs to come _after_ topLevel!
	if includeStrong {
		strong := buildSchedule.Descendants(impacted)
		impacted = append(impacted, strong...)
	}

	return dependencies, buildSchedule, impacted, nil
}
