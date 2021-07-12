package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/charypar/monobuild/cli"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	baseBranch    string
	baseCommit    string
	githubMatrix  bool
	mainBranch    bool
	rebuildStrong bool
	dotHighlight  bool
}

var diffOpts diffOptions

var diffCmd = &cobra.Command{
	Use:   "diff [-]",
	Short: "Build schedule for components affected by git changes",
	Long: `Create a build schedule based on git history and dependency graph.
Each line in the output is a component and its dependencies. 
The format of each line is:

<component>: <dependency>, <dependency>, <dependency>, ...

Diff can output either the build schedule (using only strong dependencies) or 
the original dependeny graph (using all dependencies).

By default changed files are determined from the local git repository. 
Optionally, they can be provided externaly from stdin, by adding a hypen (-) after
the diff command.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("Too many arguments")
		}
		if len(args) == 1 && args[0] != "-" {
			return fmt.Errorf("Invalid first argument: %s, only \"-\" is allowed", args[0])
		}

		return nil
	},
	Run: diffFn,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&diffOpts.baseBranch, "base-branch", "master", "Base branch to use for comparison")
	diffCmd.Flags().StringVar(&diffOpts.baseCommit, "base-commit", "HEAD^1", "Base commit to compare with (useful in main-branch mode when using rebase merging)")
	diffCmd.Flags().BoolVar(&diffOpts.githubMatrix, "github-matrix", false, "Output a list that can be used as a Github Actions build matrix (`[a,b,c]`).")
	diffCmd.Flags().BoolVar(&diffOpts.mainBranch, "main-branch", false, "Run in main branch mode (i.e. only compare with parent commit)")
	diffCmd.Flags().BoolVar(&diffOpts.rebuildStrong, "rebuild-strong", false, "Include all strong dependencies of affected components")
	diffCmd.Flags().BoolVar(&commonOpts.printDependencies, "dependencies", false, "Ouput the dependencies, not the build schedule")
	diffCmd.Flags().BoolVar(&commonOpts.dotFormat, "dot", false, "Print in DOT format for GraphViz")
	diffCmd.Flags().BoolVar(&commonOpts.printFull, "full", false, "Print the full dependency graph including strengths")
}

func diffFn(cmd *cobra.Command, args []string) {
	// first we tediously process the CLI flags
	var branchMode cli.DiffMode
	changedFiles := []string{}

	if len(args) > 0 && args[0] == "-" {
		branchMode = cli.Direct

		// Read stdin into []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			changedFiles = append(changedFiles, scanner.Text())
		}

	} else if diffOpts.mainBranch {
		branchMode = cli.MainBranch
	} else {
		branchMode = cli.FeatureBranch
	}

	var format cli.OutputFormat
	if commonOpts.dotFormat {
		format = cli.Dot
	} else if diffOpts.githubMatrix {
		format = cli.GithubMatrix
	} else {
		format = cli.Text
	}

	diffContext := cli.DiffContext{
		Mode:         branchMode,
		BaseBranch:   diffOpts.baseBranch,
		BaseCommit:   diffOpts.baseCommit,
		ChangedFiles: changedFiles,
	}
	scope := cli.Scope{Scope: commonOpts.scope, TopLevel: commonOpts.topLevel}

	var outType cli.OutputType
	if commonOpts.printFull {
		outType = cli.Full
	} else if commonOpts.printDependencies {
		outType = cli.Dependencies
	} else {
		outType = cli.Schedule
	}

	outputOpts := cli.OutputOptions{Format: format, Type: outType}

	repoManifest := ""
	if len(commonOpts.repoManifestFile) > 0 {
		bytes, err := ioutil.ReadFile(commonOpts.repoManifestFile)
		if err != nil {
			log.Fatal(err)
		}

		repoManifest = string(bytes)
	}

	// run the CLI command
	dependencies, schedule, impacted, err := cli.Diff(commonOpts.dependencyFilesGlob, diffContext, scope, diffOpts.rebuildStrong, repoManifest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cli.Format(dependencies, schedule, impacted, outputOpts))
}
