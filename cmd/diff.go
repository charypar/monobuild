package cmd

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/charypar/monobuild/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "List changed components",
	Long:  `List changed components based on git history and dependency graph`,
	Run:   diffFn,
}

var baseBranch string
var mainBranch bool

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().BoolVar(&mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
}

func diffFn(cmd *cobra.Command, args []string) {
	paths, err := doublestar.Glob(dependencyFilesGlob)
	if err != nil {
		panic(fmt.Errorf("Error finding dependency manifests: %s", err))
	}

	changedPaths, err := diff.Diff(paths, baseBranch, mainBranch)
	if err != nil {
		panic(fmt.Errorf("Error finding out changed components: %s", err))
	}

	fmt.Println(strings.Join(changedPaths, "\n"))
}
