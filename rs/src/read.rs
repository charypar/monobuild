use std::collections::{BTreeMap, BTreeSet};
use std::fmt::Display;

use crate::{
    core::Dependency,
    graph::{Edge, Graph},
};

#[derive(PartialEq, Debug)]
pub enum Warning {
    Unknown(String, String),      // dependent, dependency
    BadLineFormat(usize, String), // line number, line
}

impl Display for Warning {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Warning::Unknown(of, what) => write!(f, "Unknown dependency {} of {}.", what, of),
            Warning::BadLineFormat(l, line) => write!(f, "Bad line format: {}: '{}' expected 'component: dependency, dependency, dependency, ...", l, line),
        }
    }
}

fn manifest(manifest: &str) -> BTreeSet<Edge<String, Dependency>> {
    manifest
        .lines()
        .flat_map(|l| match l.trim().trim_end_matches("/") {
            "" => None,
            d if d.starts_with("#") => None,
            d if d.starts_with("!") => Some(Edge::new(d[1..].to_string(), Dependency::Strong)),
            d => Some(Edge::new(d.to_string(), Dependency::Weak)),
        })
        .collect()
}

// TODO consider building a safer mutable graph API for adding vertices and edges while keeping the graph
// normalised instead of building graphs up manually

pub fn manifests(manifests: BTreeMap<String, String>) -> (Graph<String, Dependency>, Vec<Warning>) {
    let components: BTreeSet<_> = manifests.keys().collect();

    let mut edges = BTreeMap::new();
    let mut warnings = Vec::new();

    for (c, m) in &manifests {
        let (deps, ws): (BTreeSet<_>, BTreeSet<_>) = manifest(m)
            .into_iter()
            .partition(|e| components.contains(&e.to));

        let mut warns = ws
            .into_iter()
            .map(|e| Warning::Unknown(c.clone(), e.to))
            .collect();

        warnings.append(&mut warns);

        if deps.is_empty() {
            edges.insert(c.clone(), BTreeSet::new());
        } else {
            edges.insert(c.clone(), deps);
        }
    }

    (Graph { edges }, warnings)
}

pub fn repo_manifest(manifest: String) -> (Graph<String, Dependency>, Vec<Warning>) {
    let mut graph = BTreeMap::new();
    let mut warnings = Vec::new();

    let lines = manifest
        .lines()
        .map(|l| l.trim())
        .filter(|l| !l.starts_with("#") && *l != "")
        .enumerate();

    for (i, line) in lines {
        if let Some((c, ds)) = line.split_once(":") {
            let component = c.trim().to_owned();
            let mut dependencies: BTreeSet<_> = ds
                .trim()
                .split(",")
                .map(|d| d.trim())
                .filter(|d| *d != "")
                .map(|d| {
                    if d.starts_with("!") {
                        Edge::new(d[1..].to_owned(), Dependency::Strong)
                    } else {
                        Edge::new(d.to_owned(), Dependency::Weak)
                    }
                })
                .collect();

            for d in &dependencies {
                graph.entry(d.to.clone()).or_insert_with(|| BTreeSet::new());
            }

            graph
                .entry(component)
                .or_insert_with(|| BTreeSet::new())
                .append(&mut dependencies);
        } else {
            warnings.push(Warning::BadLineFormat(i, line.to_owned()));
        }
    }

    (Graph { edges: graph }, warnings)
}

#[cfg(test)]
mod test {
    mod manifest {
        use super::super::*;
        use std::collections::BTreeSet;

        #[test]
        fn empty() {
            let text = "\n   \n \n   # comment   \n";

            let actual = manifest(text);
            let expected = BTreeSet::new();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_dependency() {
            let text = "\n   other/\n \n   \n";

            let actual = manifest(text);
            let expected = vec![Edge::new("other".to_string(), Dependency::Weak)]
                .into_iter()
                .collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_strong_dependency() {
            let text = "\n   !other\n \n   \n";

            let actual = manifest(text);
            let expected = vec![Edge::new("other".to_string(), Dependency::Strong)]
                .into_iter()
                .collect();

            assert_eq!(actual, expected)
        }

        #[test]
        fn full() {
            let text = "\n  some  \n !other\n \none/more  \n  # comment  \n";

            let actual = manifest(text);
            let expected = vec![
                Edge::new("some".to_string(), Dependency::Weak),
                Edge::new("other".to_string(), Dependency::Strong),
                Edge::new("one/more".to_string(), Dependency::Weak),
            ]
            .into_iter()
            .collect();

            assert_eq!(actual, expected)
        }
    }

    mod manifests {
        use crate::graph::Graph;
        use std::collections::BTreeMap;

        use super::super::*;

        #[test]
        fn no_manifests() {
            let files = BTreeMap::new();

            let actual = manifests(files);
            let expected = (Graph::new(vec![]), Vec::new());

            assert_eq!(actual, expected);
        }

        #[test]
        fn ignore_unknown_components() {
            let files = vec![("foo".to_string(), "\n bar\n".to_string())]
                .into_iter()
                .collect();

            let actual = manifests(files);
            let expected = (
                Graph::new(vec![("foo".into(), vec![])]),
                vec![Warning::Unknown("foo".to_string(), "bar".to_string())],
            );

            assert_eq!(actual, expected);
        }

        #[test]
        fn two_manifests() {
            let files = vec![
                ("foo".to_string(), "\n bar\nbaz".to_string()),
                ("bar".to_string(), "\n baz\n".to_string()),
            ]
            .into_iter()
            .collect();

            let actual = manifests(files);
            let expected = (
                Graph::new(vec![
                    (
                        "foo".into(),
                        vec![Edge::new("bar".into(), Dependency::Weak)],
                    ),
                    ("bar".into(), vec![]),
                ]),
                vec![
                    Warning::Unknown("bar".to_string(), "baz".to_string()),
                    Warning::Unknown("foo".to_string(), "baz".to_string()),
                ],
            );

            assert_eq!(actual, expected);
        }

        #[test]
        fn complex_manifests() {
            let files = vec![
                ("app1".to_string(), "\nlibs/lib1\nlibs/lib2/".to_string()),
                ("app2".to_string(), "\nlibs/lib2\n\n\nlibs/lib3".to_string()),
                ("app3".to_string(), "\n\nlibs/lib3".to_string()),
                ("app4".to_string(), "\n\n# yo".to_string()),
                ("libs/lib1".to_string(), "\n libs/lib3\n".to_string()),
                ("libs/lib2".to_string(), "\n libs/lib3\n".to_string()),
                ("libs/lib3".to_string(), "".to_string()),
                (
                    "stack1".to_string(),
                    "# frontend\n!app1\n\n# backend\n!app2\n!app3".to_string(),
                ),
            ]
            .into_iter()
            .collect();

            let actual = manifests(files);
            let expected = (
                Graph::new(vec![
                    (
                        "app1".into(),
                        vec![
                            Edge::new("libs/lib1".into(), Dependency::Weak),
                            Edge::new("libs/lib2".into(), Dependency::Weak),
                        ],
                    ),
                    (
                        "app2".into(),
                        vec![
                            Edge::new("libs/lib2".into(), Dependency::Weak),
                            Edge::new("libs/lib3".into(), Dependency::Weak),
                        ],
                    ),
                    (
                        "app3".into(),
                        vec![Edge::new("libs/lib3".into(), Dependency::Weak)],
                    ),
                    ("app4".into(), vec![]),
                    (
                        "libs/lib1".into(),
                        vec![Edge::new("libs/lib3".into(), Dependency::Weak)],
                    ),
                    (
                        "libs/lib2".into(),
                        vec![Edge::new("libs/lib3".into(), Dependency::Weak)],
                    ),
                    ("libs/lib3".into(), vec![]),
                    (
                        "stack1".into(),
                        vec![
                            Edge::new("app1".into(), Dependency::Strong),
                            Edge::new("app2".into(), Dependency::Strong),
                            Edge::new("app3".into(), Dependency::Strong),
                        ],
                    ),
                ]),
                vec![],
            );

            assert_eq!(actual, expected);
        }
    }

    mod repo_manifest {
        use super::super::*;
        use crate::graph::Graph;

        #[test]
        fn empty_manifest() {
            let manifest = "".into();

            let (actual, _) = repo_manifest(manifest);
            let expected = Graph::new(vec![]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_component() {
            let manifest = "lib1:".into();

            let (actual, _) = repo_manifest(manifest);
            let expected = Graph::new(vec![("lib1".into(), vec![])]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn component_with_depednency() {
            let manifest = "lib1: lib2\nlib2:".into();

            let (actual, _) = repo_manifest(manifest);
            let expected = Graph::new(vec![
                (
                    "lib1".into(),
                    vec![Edge::new("lib2".into(), Dependency::Weak)],
                ),
                ("lib2".into(), vec![]),
            ]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn component_with_mutlitple_depednencies() {
            let manifest = "lib1: lib2, lib3\nlib2: \nlib3: ".into();

            let (actual, _) = repo_manifest(manifest);
            let expected = Graph::new(vec![(
                "lib1".into(),
                vec![
                    Edge::new("lib2".into(), Dependency::Weak),
                    Edge::new("lib3".into(), Dependency::Weak),
                ],
            )]); // Note that 'new' normalises the graph!

            assert_eq!(actual, expected);
        }

        #[test]
        fn component_with_unlisted_dependency() {
            let manifest = "lib1: lib2, lib3\n".into();

            let (actual, _) = repo_manifest(manifest);
            let expected = Graph::new(vec![(
                "lib1".into(),
                vec![
                    Edge::new("lib2".into(), Dependency::Weak),
                    Edge::new("lib3".into(), Dependency::Weak),
                ],
            )]); // Note that 'new' normalises the graph!

            assert_eq!(actual, expected);
        }

        #[test]
        fn complex_manifest() {
            let manifest = "# comment\napp1: lib1, lib2, lib3\napp2: \nlib1: \nlib2: lib3\nlib3: \n\nstack1: !app1, !app2".to_owned();

            let (actual, ws) = repo_manifest(manifest);
            let expected = Graph::new(vec![
                (
                    "app1".into(),
                    vec![
                        Edge::new("lib1".into(), Dependency::Weak),
                        Edge::new("lib2".into(), Dependency::Weak),
                        Edge::new("lib3".into(), Dependency::Weak),
                    ],
                ),
                (
                    "lib2".into(),
                    vec![Edge::new("lib3".into(), Dependency::Weak)],
                ),
                (
                    "stack1".into(),
                    vec![
                        Edge::new("app1".into(), Dependency::Strong),
                        Edge::new("app2".into(), Dependency::Strong),
                    ],
                ),
            ]);

            assert_eq!(actual, expected);
            assert_eq!(ws, vec![]);
        }
    }
}
