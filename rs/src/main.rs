use std::collections::{BTreeMap, BTreeSet};
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
use graph::Graph;
use read::Warning;
use write::TextFormat;

fn main() {
    let opts = Opts::from_args();
    let result = match opts.cmd {
        Command::Print(opts) => print(opts.input_opts, opts.output_opts),
        Command::Diff(opts) => diff(opts),
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

fn print(input_opts: InputOpts, output_opts: OutputOpts) -> Result<String> {
    let (graph, warnings) = load_graph(input_opts)?;
    for warning in warnings {
        eprint!("{}", warning);
    }

    if output_opts.full {
        return Ok(format!("{}", write::to_text(&graph, TextFormat::Full)));
    }

    let graph = if output_opts.top_level {
        let roots = graph.roots();

        graph.filter(|v| roots.contains(v), |_| true)
    } else {
        graph
    };

    let (graph, dot_format) = if !output_opts.dependencies {
        (
            graph.filter(|_| true, |d| d == &Dependency::Strong),
            DotFormat::Schedule,
        )
    } else {
        (graph, DotFormat::Dependencies)
    };

    let graph = if let Some(scope) = output_opts.scope {
        let root = vec![scope.clone()].into_iter().collect();
        let mut tree_vertices = graph.descendants(&root);
        tree_vertices.insert(&scope);

        graph.filter(|v| tree_vertices.contains(v), |_| true)
    } else {
        graph
    };

    if output_opts.dot {
        return Ok(format!("{}", write::to_dot(&graph, dot_format)));
    }

    Ok(format!("{}", write::to_text(&graph, TextFormat::Simple)))
}

fn diff(opts: DiffOpts) -> Result<String> {
    let (graph, warnings) = load_graph(opts.input_opts)?;
    for warning in warnings {
        eprint!("{}", warning);
    }

    let components: Vec<&Path> = graph
        .vertices()
        .into_iter()
        .map(|path| Path::new(path))
        .collect();

    // Get changed files

    let changed: BTreeSet<String> = match opts.changes {
        cli::Source::Stdin => io::stdin().lock().lines().flatten().collect(),
        cli::Source::Git => {
            let mut git = Git::new(execute);

            let mode = if opts.main_branch {
                Mode::Main(opts.base_branch)
            } else {
                Mode::Feature(opts.base_commit)
            };

            git.diff(mode)?
        }
    }
    .into_iter()
    .filter_map(|file_path| {
        // TODO Filter down to changed components
        let file_path = Path::new(&file_path);
        for component in &components {
            if file_path.starts_with(component) {
                return component.to_str().map(|s| s.to_owned());
            }
        }

        None
    })
    .collect();

    // Get impaced subgraph
    let impact_graph = graph.reverse();
    let affected = impact_graph.descendants(&changed);
    let graph = graph.filter(
        |v| affected.contains(v) || changed.contains(v),
        |e| *e == Dependency::Strong,
    );

    Ok(format!("{}", write::to_text(&graph, TextFormat::Simple)))
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
        std::str::from_utf8(&out.stdout)
            .map(|s| s.to_string())
            .map_err(|e| format!("Could not convert git output to string: {}", e))
    }
}

// Input processing

fn load_graph(input_opts: InputOpts) -> Result<(Graph<String, Dependency>, Vec<Warning>)> {
    if let Some(path) = input_opts.full_manifest {
        let content = fs::read_to_string(path)?;
        Ok(read::repo_manifest(content))
    } else {
        let manifests = read_manifest_files(input_opts.dependency_files_glob)?;
        Ok(read::manifests(manifests))
    }
}

fn read_manifest_files(glob: String) -> Result<BTreeMap<String, String>> {
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
