use structopt::StructOpt;

#[derive(Debug)]
enum Source {
    Stdin,
    Git,
}

/// A build orchestration tool for Continuous Integration in a monorepo.
///
/// Read a graph of dependencies in a monorepo codebase (where separate
/// components live side by side) and decide what should be built, given the git
/// history.
#[derive(StructOpt, Debug)]
#[structopt(name = "monobuild", rename_all = "kebab", no_version)]
pub struct Opts {
    #[structopt(subcommand, no_version)]
    pub cmd: Command,
}

#[derive(StructOpt, Debug)]
pub struct InputOpts {
    /// Search pattern for depenency files
    #[structopt(long = "dependency-files", default_value = "**/Dependencies")]
    pub dependency_files_glob: String,

    /// Full manifest file (as produced by 'print --full')
    #[structopt(short = "f", long = "file")]
    pub full_manifest: Option<String>,
}

#[derive(StructOpt, Debug)]
pub struct OutputOpts {
    /// Ouput the dependencies, not the build schedule
    #[structopt(long)]
    pub dependencies: bool,

    /// Print in DOT format for GraphViz
    #[structopt(long)]
    pub dot: bool,

    /// Print the full dependency graph including strengths
    #[structopt(long)]
    pub full: bool,

    /// Scope output to single component and its dependencies
    #[structopt(long)]
    pub scope: Option<String>,

    /// Only list top-level components that nothing depends on
    #[structopt(long)]
    pub top_level: bool,
}

#[derive(StructOpt, Debug)]
pub struct DiffOpts {
    /// "Base branch to use o comparison"
    #[structopt(long, default_value = "master")]
    base_branch: String,

    /// Base commit to compare with (useful in main-branch mode when using rebase merging)
    #[structopt(long, default_value = "HEAD^1")]
    base_commit: String,

    /// Run in main branch mode (i.e. only compare with parent commit)
    #[structopt(long)]
    main_branch: bool,

    /// Include all strong dependencies of affected components
    #[structopt(long)]
    rebuild_strong: bool,

    // FIXME this seems really hacky, there's got to be a better way
    /// Read changed files from STDIN
    #[structopt(name = "-", default_value = "", parse(from_str = parse_stdin))]
    changes: Source,

    #[structopt(flatten)]
    input_opts: InputOpts,
    #[structopt(flatten)]
    output_opts: OutputOpts,
}

#[derive(StructOpt, Debug)]
pub struct PrintOpts {
    #[structopt(flatten)]
    pub input_opts: InputOpts,
    #[structopt(flatten)]
    pub output_opts: OutputOpts,
}

#[derive(StructOpt, Debug)]
pub enum Command {
    /// Build schedule for components affected by git changes
    ///
    /// Create a build schedule based on git history and dependency graph.
    /// Each line in the output is a component and its dependencies.
    /// The format of each line is:
    ///
    /// <component>: <dependency>, <dependency>, <dependency>, ...
    ///
    /// Diff can output either the build schedule (using only strong dependencies) or
    /// the original dependeny graph (using all dependencies).
    ///
    /// By default changed files are determined from the local git repository.
    /// Optionally, they can be provided externaly from stdin, by adding a hypen (-) after
    /// the diff command.
    #[structopt(no_version)]
    Diff(DiffOpts),
    /// Print the full build schedule or dependency graph
    ///
    /// Print the full build schedule or dependency graph based on the manifest files.
    /// The format of each line is:
    ///
    /// <component>: <dependency>, <dependency>, <dependency>, ...
    ///
    /// Diff can output either the build schedule (using only strong dependencies) or
    /// the original dependeny graph (using all dependencies).
    #[structopt(no_version)]
    Print(PrintOpts),
}

fn parse_stdin(src: &str) -> Source {
    if src == "-" {
        Source::Stdin
    } else {
        Source::Git
    }
}
