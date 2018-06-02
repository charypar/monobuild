package diff

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/charypar/monobuild/graph"
)

func diffBase(mainBranch bool, baseBranch string, baseCommit string) (string, error) {
	if mainBranch {
		if baseCommit != "" {
			return baseCommit, nil
		}

		return "HEAD^1", nil
	}

	gitMergeBase := exec.Command("git", "merge-base", baseBranch, "HEAD")
	mergeBase, err := gitMergeBase.Output()
	if err != nil {
		return "", fmt.Errorf("cannot find merge base with branch '%s': %s", baseBranch, err.(*exec.ExitError).Stderr)
	}

	return strings.TrimRight(string(mergeBase), "\n"), nil
}

// ChangedFiles uses git to determine the list of files that changed for
// the current revision.
// It can operate in a normal (branch) mode, where it compares to a 'baseBranch'
// or a 'mainBranch' mode, where it compares to the previous revision or a 'baseCommit'
func ChangedFiles(mainBranch bool, baseBranch string, baseCommit string) ([]string, error) {
	base, err := diffBase(mainBranch, baseBranch, baseCommit)
	if err != nil {
		return []string{}, err
	}

	gitDiff := exec.Command("git", "diff", "--no-commit-id", "--name-only", "-r", base)

	gitOut, err := gitDiff.Output()
	if err != nil {
		return []string{}, fmt.Errorf("cannot find changed files:\n%s", err.(*exec.ExitError).Stderr)
	}

	changed := strings.Split(strings.TrimRight(string(gitOut), "\n"), "\n")
	return changed, nil
}

// Impacted calculates the list of changes impacted by a change
func Impacted(changedComponents []string, dependencies graph.Graph) []string {
	impactGraph := dependencies.Reverse()
	impacted := impactGraph.Descendants(changedComponents)

	result := append(impacted, changedComponents...)
	sort.Strings(result)

	return result
}
