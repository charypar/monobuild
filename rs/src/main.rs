use std::collections::{BTreeMap, HashSet};
use std::fs;
use std::io::{self, BufRead};
use std::path::Path;
use std::process;

use anyhow::{anyhow, Result};
use ignore::WalkBuilder;
use structopt::StructOpt;

mod cli;
mod core;
mod git;
mod graph;
mod read;
mod write;

use crate::{core::Dependency, write::DotFormat};
use cli::{Command, DiffOpts, InputOpts, Opts, OutputOpts};
use git::{Git, Mode};
use graph::{Graph, Subgraph};
use read::Warning;
use write::TextFormat;

fn main() {
    let opts = Opts::from_args();
    let result = match opts.cmd {
        Command::Print(opts) => print(opts.input_opts, opts.output_opts),
        Command::Diff(opts) => diff(&opts),
    };

    match result {
        Ok(out) => {
            println!("{}", out);
            process::exit(0);
        }
        Err(err) => {
            eprintln!("{}", err);
            process::exit(1);
        }
    }
}

// Commands

fn print(input_opts: InputOpts, output_opts: OutputOpts) -> Result<String> {
    let (graph, warnings) = load_graph(&input_opts)?;
    for warning in warnings {
        eprint!("{}", warning);
    }

    if output_opts.full {
        return Ok(write::to_text(&graph, TextFormat::Full).to_string());
    }

    // FIXME this is here purely to star us off with a Subgraph
    let mut graph = graph.filter_vertices(|_| true);

    graph = scope_graph(graph, &output_opts);

    print_output(graph, &output_opts)
}

fn diff(opts: &DiffOpts) -> Result<String> {
    let (graph, warnings) = load_graph(&opts.input_opts)?;
    for warning in warnings {
        eprint!("{}", warning);
    }
    let impact_graph = graph.reverse();

    // Get changes

    let components: Vec<&Path> = graph.vertices().map(|path| Path::new(path)).collect();

    let changed = changed_components(components, &opts)?;
    let affected: HashSet<_> = impact_graph
        .filter_vertices(|v| changed.contains(v))
        .expand()
        .vertices()
        .collect();

    // Scope

    // FIXME this is here purely to star us off with a Subgraph
    let mut graph = graph.filter_vertices(|_| true);
    graph = scope_graph(graph, &opts.output_opts);

    graph = graph.filter_vertices(|v| affected.contains(&v));

    if opts.rebuild_strong {
        graph = graph.expand_via(|e| *e == Dependency::Strong)
    };

    // Output

    if opts.output_opts.full {
        return Ok(write::to_text(&graph, TextFormat::Full).to_string());
    }

    print_output(graph, &opts.output_opts)
}

// Support functions

fn changed_components(components: Vec<&Path>, opts: &DiffOpts) -> Result<HashSet<String>> {
    Ok(match opts.changes {
        cli::Source::Stdin => io::stdin().lock().lines().flatten().collect(),
        cli::Source::Git => {
            let mut git = Git::new(execute);

            let mode = if opts.main_branch {
                Mode::Main(opts.base_branch.clone())
            } else {
                Mode::Feature(opts.base_commit.clone())
            };

            git.diff(mode)?
        }
    }
    .into_iter()
    .filter_map(|file_path| {
        let file_path = Path::new(&file_path);
        for component in &components {
            if file_path.starts_with(component) {
                return component.to_str().map(|s| s.to_owned());
            }
        }

        None
    })
    .collect())
}

fn execute(command: Vec<String>) -> Result<String, String> {
    let prog = &command[0];
    let args = &command[1..];

    let out = process::Command::new(prog)
        .args(args)
        .output()
        .map_err(|e| format!("Git call failed: {}", e))?;

    if out.status.success() {
        std::str::from_utf8(&out.stdout)
            .map(|s| s.to_string())
            .map_err(|e| format!("Could not convert git output to string: {}", e))
    } else {
        let error = std::str::from_utf8(&out.stderr)
            .map_err(|e| format!("Could not convert git output to string: {}", e))?;

        Err(error.to_string())
    }
}

fn load_graph(input_opts: &InputOpts) -> Result<(Graph<String, Dependency>, Vec<Warning>)> {
    if let Some(path) = &input_opts.full_manifest {
        let content = fs::read_to_string(path)?;
        Ok(read::repo_manifest(content))
    } else {
        let manifests = read_manifest_files(&input_opts.dependency_files_glob)?;
        Ok(read::manifests(manifests))
    }
}

fn read_manifest_files(glob: &str) -> Result<BTreeMap<String, String>> {
    let pattern = glob::Pattern::new(&glob).expect("Malformed manifest search pattern"); // FIXME handle this better

    let manifests: BTreeMap<String, String> = WalkBuilder::new("./")
        .hidden(false)
        .build()
        .filter_map(|r| r.ok())
        .filter(|entry| pattern.matches_path(entry.path()))
        .flat_map(|manifest| -> Result<_> {
            let path = manifest.path();
            let component = manifest
                .path()
                .parent()
                .ok_or(anyhow!("cannot find a directory path for: {:?}", path))
                .and_then(|dir| {
                    dir.to_str()
                        .ok_or(anyhow!("Cannot convert path to string: {:?}", dir))
                })
                .map(|component_path| component_path.trim_start_matches("./").to_owned())?;

            let manifest_content = fs::read_to_string(path)?;

            Ok((component, manifest_content))
        })
        .collect();

    Ok(manifests)
}

fn scope_graph<'g>(
    graph: Subgraph<'g, String, Dependency>,
    opts: &OutputOpts,
) -> Subgraph<'g, String, Dependency> {
    let mut graph = graph;

    if let Some(scope) = &opts.scope {
        graph = graph.filter_vertices(|v| v == scope).expand();
    };

    if opts.top_level {
        graph = graph.roots();
    };

    graph
}

fn print_output<'g>(
    graph: Subgraph<'g, String, Dependency>,
    output_opts: &OutputOpts,
) -> Result<String> {
    let mut graph = graph;

    let dot_format = if !output_opts.dependencies {
        graph = graph.filter_edges(|e| *e == Dependency::Strong);

        DotFormat::Schedule
    } else {
        DotFormat::Dependencies
    };

    if output_opts.dot {
        return Ok(write::to_dot(&graph, dot_format).to_string());
    }

    Ok(write::to_text(&graph, TextFormat::Simple).to_string())
}
