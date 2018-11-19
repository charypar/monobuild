package diff

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/charypar/monobuild/graph"
)

// BranchMode is the diff mode based on the kind of branch, either Feature or Main
type BranchMode int

// Feature is a feature branch mode
var Feature BranchMode = 1

// Main is a main branch mode
var Main BranchMode = 2

// Mode holds options for the Diff command
type Mode struct {
	Mode       BranchMode
	BaseBranch string
	BaseCommit string
}

func diffBase(mode Mode) (string, error) {
	if mode.Mode == Main {
		return mode.BaseCommit, nil
	}

	gitMergeBase := exec.Command("git", "merge-base", mode.BaseBranch, "HEAD")
	mergeBase, err := gitMergeBase.Output()
	if err != nil {
		return "", fmt.Errorf("cannot find merge base with branch '%s': %s", mode.BaseBranch, err.(*exec.ExitError).Stderr)
	}

	return strings.TrimRight(string(mergeBase), "\n"), nil
}

// ChangedFiles uses git to determine the list of files that changed for
// the current revision.
// It can operate in a normal (branch) mode, where it compares to a 'baseBranch'
// or a 'mainBranch' mode, where it compares to the previous revision or a 'baseCommit'
func ChangedFiles(mode Mode) ([]string, error) {
	base, err := diffBase(mode)
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
