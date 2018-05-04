package diff

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

func readManifests(manifestPaths []string) ([]string, map[string][]string, error) {
	dependencies := make(map[string][]string, len(manifestPaths))
	components := make([]string, 0)

	for _, manifest := range manifestPaths {
		file, err := os.Open(manifest)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot open dependency manifest %s: %s", manifest, err)
		}

		dir, _ := filepath.Split(manifest)
		component := strings.TrimRight(dir, "/")
		components = append(components, component)

		// assume self-dependency
		dependencies[component] = []string{component}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			dep := strings.TrimRight(strings.TrimSpace(scanner.Text()), "/")
			if len(dep) > 0 && dep[0] != '#' {
				dependencies[component] = append(dependencies[component], dep)
			}
		}

		err = scanner.Err()
		if err != nil {
			return nil, nil, fmt.Errorf("cannot read dependency manifest %s: %s", manifest, err)
		}
	}

	// validate dependencies
	for manifest, deps := range dependencies {
		for _, dep := range deps {
			if !validDependency(components, dep) {
				return nil, nil, fmt.Errorf("unknown dependency '%s' of '%s'", dep, manifest)
			}
		}
	}

	return components, dependencies, nil
}

func diffBase(mainBranch bool, baseBranch string) (string, error) {
	if mainBranch {
		return "HEAD^1", nil
	}

	gitMergeBase := exec.Command("git", "merge-base", baseBranch, "HEAD")
	mergeBase, err := gitMergeBase.Output()
	if err != nil {
		return "", fmt.Errorf("cannot find merge base with branch '%s': %s", baseBranch, err)
	}

	return strings.TrimRight(string(mergeBase), "\n"), nil
}

func changedFiles(mainBranch bool, baseBranch string) ([]string, error) {
	base, err := diffBase(mainBranch, baseBranch)
	if err != nil {
		return []string{}, err
	}

	gitDiff := exec.Command("git", "diff", "--no-commit-id", "--name-only", "-r", base)

	gitOut, err := gitDiff.Output()
	if err != nil {
		return []string{}, fmt.Errorf("cannot find changed files: %s", err)
	}

	changed := strings.Split(strings.TrimRight(string(gitOut), "\n"), "\n")
	return changed, nil
}

func changedComponents(components []string, changedFiles []string) []string {
	changedComponents := []string{}

	for _, component := range components {
		for _, change := range changedFiles {
			if strings.HasPrefix(change, component) {
				changedComponents = append(changedComponents, component)
				break
			}
		}
	}

	return changedComponents
}

// Diff calculates the list of paths that need to be built based on the list of
func Diff(manifestPaths []string, baseBranch string, mainBranch bool) ([]string, error) {
	// Find components and dependency manifests
	components, dependencies, err := readManifests(manifestPaths)
	if err != nil {
		return []string{}, fmt.Errorf("cannot load dependencies: %s", err)
	}

	// Get changed files
	changes, err := changedFiles(mainBranch, baseBranch)
	if err != nil {
		return []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	chgdComponents := changedComponents(components, changes)

	// Construct build graph
	dependencyGraph := NewGraph(dependencies)
	buildGraph := dependencyGraph.Reverse()

	// Include the dependents
	componentsToBuild := buildGraph.Descendants(vertices(chgdComponents)).AsStrings()

	return componentsToBuild, nil
}
