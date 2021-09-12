use std::collections::BTreeSet;

use crate::graph::Graph;

// FIXME this belongs elsewhere
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Dependency {
    Weak,
    Strong,
}

fn impacted(
    dependencies: &Graph<String, Dependency>,
    changed: &BTreeSet<String>,
) -> BTreeSet<String> {
    let impact_graph = dependencies.reverse();

    let mut impacted: BTreeSet<_> = impact_graph
        .descendants(changed)
        .into_iter()
        .map(|v| v.clone())
        .collect();

    for ch in changed {
        impacted.insert(ch.clone());
    }

    impacted
}
