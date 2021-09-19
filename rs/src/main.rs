use std::{collections::BTreeMap, fs, io, process};

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

fn print(input_opts: InputOpts, output_opts: OutputOpts) -> io::Result<String> {
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

fn diff(opts: DiffOpts) -> io::Result<String> {
    let (graph, warnings) = load_graph(opts.input_opts)?;
    for warning in warnings {
        eprint!("{}", warning);
    }

    let mut git = Git::new(|cmd| {
        let prog = &cmd[0];
        let args = &cmd[1..];

        let out = process::Command::new(prog).args(args).output()?;

        if out.status.success() {
            std::str::from_utf8(&out.stdout)
                .map(|s| s.to_string())
                .map_err(|_| {
                    io::Error::new(
                        io::ErrorKind::Other, // FIXME this is ugly, use Anyhow
                        format!("Output not utf8 {:?}", &out.stdout),
                    )
                })
        } else {
            let err = std::str::from_utf8(&out.stdout)
                .map(|s| s.to_string())
                .map_err(|_| {
                    io::Error::new(
                        io::ErrorKind::Other, // FIXME this is ugly, use Anyhow
                        format!(
                            "Command execution failed with non-utf8 output: {:?}",
                            &out.stdout
                        ),
                    )
                })?;

            Err(io::Error::new(
                io::ErrorKind::Other, // FIXME this is ugly, use Anyhow
                format!("Command execution failed: {}", err),
            ))
        }
    });

    let mode = if opts.main_branch {
        Mode::Main(opts.base_branch)
    } else {
        Mode::Feature(opts.base_commit)
    };

    let changed = git.diff(mode).expect("Fix this error handling AAAH");

    Ok(changed.join("\n"))
}

// Input processing

fn load_graph(input_opts: InputOpts) -> io::Result<(Graph<String, Dependency>, Vec<Warning>)> {
    if let Some(path) = input_opts.full_manifest {
        let content = fs::read_to_string(path)?;
        Ok(read::repo_manifest(content))
    } else {
        let manifests = read_manifest_files(input_opts.dependency_files_glob)?;
        Ok(read::manifests(manifests))
    }
}

fn read_manifest_files(glob: String) -> io::Result<BTreeMap<String, String>> {
    let pattern = glob::Pattern::new(&glob).expect("Malformed manifest search pattern"); // FIXME handle this better

    let manifests: BTreeMap<String, String> = WalkBuilder::new("./")
        .hidden(false)
        .build()
        .filter_map(|r| r.ok())
        .filter(|entry| pattern.matches_path(entry.path()))
        .flat_map(|manifest| -> io::Result<(String, String)> {
            let path = manifest.path();
            let component = path
                .parent()
                .expect("cannot find a manifests directory path")
                .to_str()
                .expect("Cannot convert path to string")
                .trim_start_matches("./")
                .to_owned();

            let manifest_content = fs::read_to_string(path)?;

            Ok((component, manifest_content))
        })
        .collect();

    Ok(manifests)
}
