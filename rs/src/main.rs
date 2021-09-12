use structopt::StructOpt;

mod cli;
mod core;
mod graph;
mod read;
mod write;

fn main() {
    let opt = cli::Opts::from_args();

    println!("{:?}", opt);
}
