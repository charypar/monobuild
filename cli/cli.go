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

func loadManifests(globPattern string) ([]string, graph.Graph, graph.Graph, error) {
	manifestFiles, err := doublestar.Glob(globPattern)
	if err != nil {
		return []string{}, graph.Graph{}, graph.Graph{}, fmt.Errorf("error finding dependency manifests: %s", err)
	}

	// Find components and dependencies
	components, deps, errs := manifests.Read(manifestFiles, false)
	if errs != nil {
		return []string{}, graph.Graph{}, graph.Graph{}, fmt.Errorf("%s", joinErrors("cannot load dependencies:", errs))
	}

	dependencies := deps.AsGraph()
	buildSchedule := dependencies.FilterEdges([]int{graph.Strong})

	return components, dependencies, buildSchedule, nil
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
	components, dependencies, buildSchedule, err := loadManifests(dependencyFilesGlob)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, err
	}

	selection := newFilter(components, []string{})

	if scope.Scope != "" {
		err = selection.scopeTo(scope.Scope, dependencies)
		if err != nil {
			return graph.Graph{}, graph.Graph{}, []string{}, err
		}
	}

	if scope.TopLevel {
		selection.onlyTop(dependencies)
	}

	return dependencies, buildSchedule, selection.AsStrings(), nil
}

// Diff is 'monobuild diff'
func Diff(dependencyFilesGlob string, mode diff.Mode, scope Scope, includeStrong bool) (graph.Graph, graph.Graph, []string, error) {
	components, dependencies, buildSchedule, err := loadManifests(dependencyFilesGlob)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, err
	}

	// Get changed files
	changes, err := diff.ChangedFiles(mode)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Find impacted components
	changedComponents := manifests.FilterComponents(components, changes)
	impacted := diff.Impacted(changedComponents, dependencies)

	// Select what to show

	selection := newFilter(components, impacted)

	if scope.Scope != "" {
		err = selection.scopeTo(scope.Scope, dependencies)
		if err != nil {
			return graph.Graph{}, graph.Graph{}, []string{}, err
		}
	}

	if scope.TopLevel {
		selection.onlyTop(dependencies)
	}

	// needs to come _after_ topLevel!
	if includeStrong {
		selection.addStrong(buildSchedule)
	}

	return dependencies, buildSchedule, selection.AsStrings(), nil
}
