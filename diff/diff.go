package diff

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/manifests"
	"github.com/charypar/monobuild/set"
)

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
	components, dependencies, errs := manifests.Read(manifestPaths, true)
	if errs != nil {
		errstrings := make([]string, len(errs))
		for i, e := range errs {
			errstrings[i] = string(e.Error())
		}

		return []string{}, fmt.Errorf("cannot load dependencies:\n%s", strings.Join(errstrings, "\n"))
	}

	// Get changed files
	changes, err := changedFiles(mainBranch, baseBranch)
	if err != nil {
		return []string{}, fmt.Errorf("cannot find changes: %s", err)
	}

	// Reduce changed files to components
	chgdComponents := changedComponents(components, changes)

	// Construct build graph
	dependencyGraph := graph.New(manifests.Filter(dependencies, 0))
	buildGraph := dependencyGraph.Reverse()

	// Include the dependents
	componentsToBuild := buildGraph.Descendants(set.New(chgdComponents)).AsStrings()

	return componentsToBuild, nil
}
