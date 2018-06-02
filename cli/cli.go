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

// Scope of selection
type Scope struct {
	Scope    string
	TopLevel bool
}

// OutputFormat hold the format of text output
type OutputFormat int

// Text is the standard text format.
// Each line follows this pattern:
// <component>: dependency, dependency, dependency...
var Text OutputFormat = 1

// Dot is the DOT graph language, see https://graphviz.gitlab.io/_pages/doc/info/lang.html
var Dot OutputFormat = 2

// OutputType holds the kind of output to show
type OutputType int

// Schedule is a build schedule output showing build steps and their dependencies
var Schedule OutputType = 1

// Dependencies is the dependency graph showing components and their dependencies
var Dependencies OutputType = 2

// OutputOptions hold all the options that change how the result of a command is shown
// on the command line.
// The options are not always independent, e.g. the Dot format has different output
// for Schedule type and Dependencies type.
type OutputOptions struct {
	Format OutputFormat // Output text format
	Type   OutputType   // Type of output shown
}

// Format output for the command line, filtering nodes only to those in the 'filter' slice.
// Output options can be set using 'opts
func Format(dependencies graph.Graph, schedule graph.Graph, filter []string, opts OutputOptions) string {
	if opts.Format == Dot && opts.Type == Dependencies {
		return dependencies.Dot(filter)
	}

	if opts.Format == Dot {
		return schedule.DotSchedule(filter)
	}

	if opts.Type == Dependencies {
		return dependencies.Text(filter)
	}

	return schedule.Text(filter)
}

// Print is 'monobuild print'
func Print(dependencyFilesGlob string, scope Scope) (graph.Graph, graph.Graph, []string, error) {
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

	if scope.Scope != "" {
		var scoped []string

		// ensure valid scope
		for _, c := range components {
			if c == scope.Scope {
				scoped = []string{scope.Scope}
			}
		}

		if len(scoped) < 1 {
			return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Cannot scope to '%s', not a component", scope.Scope)
		}

		selection = append(dependencies.Descendants(scoped), scoped...)
	}

	if scope.TopLevel {
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
func Diff(dependencyFilesGlob string, mode diff.Mode, scope Scope, includeStrong bool) (graph.Graph, graph.Graph, []string, error) {
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
	changes, err := diff.ChangedFiles(mode)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	changedComponents := manifests.FilterComponents(components, changes)

	// Find impacted components
	dependencies := deps.AsGraph()
	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	impacted := diff.Impacted(changedComponents, dependencies)

	if scope.Scope != "" {
		var scoped []string

		// ensure valid scope
		for _, c := range components {
			if c == scope.Scope {
				scoped = []string{scope.Scope}
			}
		}

		if len(scoped) < 1 {
			return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("Cannot scope to '%s', not a component", scope.Scope)
		}

		scopedAndDeps := append(dependencies.Descendants(scoped), scoped...)
		impacted = set.New(impacted).Intersect(set.New(scopedAndDeps)).AsStrings()
	}

	if scope.TopLevel {
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
