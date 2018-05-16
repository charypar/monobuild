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

func ChangedFiles(mainBranch bool, baseBranch string) ([]string, error) {
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

// BuildSchedule calculates the build schedule based on the dependency graph and
// the list of components that changed
func BuildSchedule(changedComponents []string, dependencies map[string][]manifests.Dependency) map[string][]string {
	impactGraph := graph.New(manifests.Filter(dependencies, 0)).Reverse()
	impacted := impactGraph.Descendants(set.New(changedComponents)).AsStrings()
	impacted = append(impacted, changedComponents...)

	// Construct build schedule
	fullBuildSchedule := graph.New(manifests.Filter(dependencies, manifests.Strong))
	buildSchedule := fullBuildSchedule.Subgraph(impacted).AsStrings()
	// Select

	return buildSchedule
}

// Dependencies calculates the dependency graph off the affected components
// based on the full dependency graph and the components that changed
func Dependencies(changedComponents []string, dependencies map[string][]manifests.Dependency) map[string][]string {
	dependencyGraph := graph.New(manifests.Filter(dependencies, 0))
	impactGraph := dependencyGraph.Reverse()

	impacted := impactGraph.Descendants(set.New(changedComponents)).AsStrings()
	impacted = append(impacted, changedComponents...)

	impactedDependencies := dependencyGraph.Subgraph(impacted).AsStrings()

	return impactedDependencies
}
