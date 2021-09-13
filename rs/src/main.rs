use std::{collections::BTreeMap, fs, io, process};

use ignore::WalkBuilder;
use structopt::StructOpt;

mod cli;
mod core;
mod graph;
mod read;
mod write;

use cli::{Command, DiffOpts, InputOpts, Opts, OutputOpts};

use crate::write::TextFormat;
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
    let dependencies = if let Some(path) = input_opts.full_manifest {
        let content = fs::read_to_string(path)?;
        read::repo_manifest(content)
    } else {
        let manifests = read_manifest_files(input_opts.dependency_files_glob)?;
        let (deps, warnings) = read::manifests(manifests);

        // print warnings
        for warning in warnings {
            eprint!("{}", warning);
        }

        deps
    };

    // TODO Scope

    // Print in format
    let format = if output_opts.full {
        TextFormat::Full
    } else {
        TextFormat::Simple
    };
    let out = write::to_text(&dependencies, format);

    Ok(format!("{}", out))
}

fn diff(opts: DiffOpts) -> io::Result<String> {
    todo!()
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
