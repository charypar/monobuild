package manifests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validDependency(components []string, dependency string) bool {
	for _, c := range components {
		if c == dependency {
			return true
		}
	}

	return false
}

func Read(manifestPaths []string, dependOnSelf bool) ([]string, map[string][]string, []error) {
	dependencies := make(map[string][]string, len(manifestPaths))
	components := make([]string, 0)
	errors := []error{}

	for _, manifest := range manifestPaths {
		file, err := os.Open(manifest)
		if err != nil {
			return nil, nil, []error{fmt.Errorf("cannot open dependency manifest %s: %s", manifest, err)}
		}

		dir, _ := filepath.Split(manifest)
		component := strings.TrimRight(dir, "/")
		components = append(components, component)

		if dependOnSelf {
			dependencies[component] = []string{component}
		} else {
			dependencies[component] = []string{}
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			dep := strings.TrimRight(strings.TrimSpace(scanner.Text()), "/")
			if len(dep) > 0 && dep[0] != '#' {
				dependencies[component] = append(dependencies[component], dep)
			}
		}

		err = scanner.Err()
		if err != nil {
			return nil, nil, []error{fmt.Errorf("cannot read dependency manifest %s: %s", manifest, err)}
		}
	}

	// validate dependencies
	for manifest, deps := range dependencies {
		for _, dep := range deps {
			if !validDependency(components, dep) {
				errors = append(errors, fmt.Errorf("unknown dependency '%s' of '%s'", dep, manifest))
			}
		}
	}

	if len(errors) > 0 {
		return components, dependencies, errors
	}

	return components, dependencies, nil
}
