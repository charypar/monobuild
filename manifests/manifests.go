package manifests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charypar/monobuild/graph"
)

// Kind of the dependency (enum)
type Kind int

// Weak dependency is one, changes to which cause rebuild of its dependents,
// but dependents builds can run in parallel with its build
var Weak Kind = 1

// Strong dependency is one, changes to which cause rebuild of its dependentes,
// and the dependent builds can only be run when its build successfully finishes
var Strong Kind = 2

// Dependency holds information about the dependency relationship of one
// component with another. Dependencies hav a string name and a kind, which
// can be Weak or Strong
type Dependency struct {
	Name string
	Kind Kind
}

// Dependencies holds a collection of dependencies as a map from a component name
// to a list of Dependency instances
type Dependencies struct {
	deps map[string][]Dependency
}

func validDependency(components []string, dependency Dependency) bool {
	for _, c := range components {
		if c == dependency.Name {
			return true
		}
	}

	return false
}

func readDependency(line string) (Dependency, error) {
	dep := strings.TrimRight(strings.TrimSpace(line), "/")

	// blank line or comment
	if len(dep) < 1 || dep[0] == '#' {
		return Dependency{}, nil
	}

	// strong dependency
	if dep[0] == '!' {
		return Dependency{dep[1:], Strong}, nil
	}

	return Dependency{dep, Weak}, nil
}

// ReadManifest reads a single manifest file and returns the dependency list
// or validation errors
func ReadManifest(path string) (string, []Dependency, []error) {
	dependencies := make([]Dependency, 0)
	errors := make([]error, 0)

	file, err := os.Open(path)
	if err != nil {
		return "", nil, []error{fmt.Errorf("cannot open dependency manifest %s: %s", path, err)}
	}

	dir, _ := filepath.Split(path)
	component := strings.TrimRight(dir, "/")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		dep, err := readDependency(scanner.Text())

		if err != nil {
			errors = append(errors, err)
			continue
		}

		// comment or blank line
		if dep.Name == "" {
			continue
		}

		dependencies = append(dependencies, dep)
	}

	err = scanner.Err()
	if err != nil {
		return "", nil, []error{fmt.Errorf("cannot read dependency manifest %s: %s", path, err)}
	}

	return component, dependencies, nil
}

// Read manifests at manifestPaths and return a graph of dependencies
func Read(manifestPaths []string, dependOnSelf bool) ([]string, Dependencies, []error) {
	dependencies := make(map[string][]Dependency, len(manifestPaths))
	components := make([]string, 0)
	errors := []error{}

	for _, manifest := range manifestPaths {
		component, deps, err := ReadManifest(manifest)
		if err != nil {
			errors = append(errors, err...)
			continue
		}

		if dependOnSelf {
			deps = append([]Dependency{Dependency{component, Weak}}, deps...)
		}

		components = append(components, component)
		dependencies[component] = deps
	}

	// validate dependencies
	for manifest, deps := range dependencies {
		for _, dep := range deps {
			if !validDependency(components, dep) {
				errors = append(errors, fmt.Errorf("unknown dependency '%s' of '%s'", dep.Name, manifest))
			}
		}
	}

	if len(errors) > 0 {
		return nil, Dependencies{}, errors
	}

	return components, Dependencies{dependencies}, nil
}

// ReadRepoManifest reads a full repository manifest as produced by monobuild print --full
func ReadRepoManifest(manifest string, dependOnSelf bool) ([]string, Dependencies, []error) {
	lines := strings.Split(manifest, "\n")
	dependencies := make(map[string][]Dependency, len(lines))
	components := make([]string, 0, len(lines))
	errors := []error{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 1 || line[0] == '#' { // skip blank lines and comments
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			errors = append(errors, fmt.Errorf("bad line format: '%s' expected 'componnennt: dependency, dependency, ...'", line))
			continue
		}

		component := strings.TrimSpace(parts[0])
		components = append(components, component)

		dependencies[component] = []Dependency{}

		for _, d := range strings.Split(parts[1], ",") {
			if len(strings.TrimSpace(d)) < 1 {
				continue
			}

			dep, err := readDependency(d)
			if err != nil {
				errors = append(errors, fmt.Errorf("malformed dependency: %s", d))
				continue
			}

			dependencies[component] = append(dependencies[component], dep)
		}

		if dependOnSelf {
			dependencies[component] = append([]Dependency{Dependency{component, Weak}}, dependencies[component]...)
		}
	}

	// validate dependencies
	for manifest, deps := range dependencies {
		for _, dep := range deps {
			if !validDependency(components, dep) {
				errors = append(errors, fmt.Errorf("unknown dependency '%s' of '%s'", dep.Name, manifest))
			}
		}
	}

	if len(errors) > 0 {
		return nil, Dependencies{dependencies}, errors
	}

	return components, Dependencies{dependencies}, nil

}

// FilterComponents filters a list of files to components
func FilterComponents(components []string, changedFiles []string) []string {
	changedComponents := []string{}

	for _, component := range components {
		for _, change := range changedFiles {
			if strings.HasPrefix(change, component+"/") {
				changedComponents = append(changedComponents, component)
				break
			}
		}
	}

	return changedComponents
}

// AsGraph returns the dependencies as a graph.Graph
func (d Dependencies) AsGraph() graph.Graph {
	result := make(map[string][]graph.Edge, len(d.deps))

	for c, ds := range d.deps {
		result[c] = make([]graph.Edge, 0, len(ds))

		for _, d := range ds {
			// FIXME there should really be a mapping of kind to colour
			result[c] = append(result[c], graph.Edge{d.Name, int(d.Kind)})
		}
	}

	return graph.New(result)
}
