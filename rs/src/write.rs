use std::fmt::Display;

use crate::core::Dependency;

pub enum TextFormat {
    Simple,
    Full,
}

pub struct Text<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It>,
    V: Clone + PartialEq + 'g,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    graph: G,
    format: TextFormat,
}

pub enum DotFormat {
    Dependencies,
    Schedule,
}

pub struct Dot<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It>,
    V: Clone + PartialEq + 'g,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    graph: G,
    format: DotFormat,
}

pub fn to_text<'g, G, V, It, InnIt>(graph: G, format: TextFormat) -> Text<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It> + Clone,
    V: Clone + PartialEq + 'g,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    Text { graph, format }
}

pub fn to_dot<'g, G, V, It, InnIt>(graph: G, format: DotFormat) -> Dot<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It> + Clone,
    V: Clone + Ord,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    Dot { graph, format }
}

impl<'g, G, V, It, InnIt> Display for Text<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It> + Clone,
    V: Clone + Display + PartialEq,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        for (component, dependencies) in self.graph.clone() {
            let joined = dependencies
                .map(|(d, kind)| match self.format {
                    TextFormat::Full if kind == Dependency::Strong => format!("!{}", d),
                    _ => format!("{}", d),
                })
                .collect::<Vec<_>>()
                .join(", ");

            writeln!(f, "{}: {}", component, joined)?;
        }

        Ok(())
    }
}

impl<'g, G, V, It, InnIt> Display for Dot<'g, G, V, It, InnIt>
where
    G: IntoIterator<IntoIter = It> + Clone,
    V: Clone + Ord + Display,
    It: Iterator<Item = (&'g V, InnIt)>,
    InnIt: Iterator<Item = (&'g V, Dependency)>,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        // FIXME check schedule prints correctly (may need reversing)
        match self.format {
            DotFormat::Dependencies => write!(f, "digraph dependencies {{\n")?,
            DotFormat::Schedule => write!(
                f,
                "digraph schedule {{\n  randir=\"LR\"\n  node [shape=box]\n"
            )?,
        }

        for (cmp, dependencies) in self.graph.clone() {
            let mut deps = dependencies.peekable();
            if let None = deps.peek() {
                write!(f, "  \"{}\"\n", cmp)?;
                continue;
            }

            for (dep, kind) in deps {
                match kind {
                    Dependency::Weak => write!(f, "  \"{}\" -> \"{}\" [style=dashed]\n", cmp, dep)?,
                    Dependency::Strong => write!(f, "  \"{}\" -> \"{}\"\n", cmp, dep)?,
                }
            }
        }

        write!(f, "}}\n")?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::core::Dependency;
    use crate::graph::Graph;

    mod text {
        use super::super::{to_text, TextFormat::Full, TextFormat::Simple};
        use super::*;

        #[test]
        fn empty_graph() {
            let graph = example();
            let filtered = graph.filter_vertices(|_| false);

            let actual = format!("{}", to_text(&filtered, Simple));
            let expected = "";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| *v == "a".to_string());

            let actual = format!("{}", to_text(&filtered, Simple));
            let expected = "a: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_edge() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b"].contains(&v.as_str()));

            let actual = format!("{}", to_text(&filtered, Simple));
            let expected = "a: b\nb: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn edge_fan() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b", "c"].contains(&v.as_str()));

            let actual = format!("{}", to_text(&filtered, Simple));
            let expected = "a: b, c\nb: c\nc: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn graph() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b", "c", "d"].contains(&v.as_str()));

            let actual = format!("{}", to_text(&filtered, Simple));
            let expected = "a: b, c\nb: c\nc: \nd: a\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn full_format() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b", "c", "d"].contains(&v.as_str()));

            let actual = format!("{}", to_text(&filtered, Full));
            let expected = "a: b, c\nb: c\nc: \nd: !a\n";

            assert_eq!(actual, expected);
        }
    }

    mod dot {
        use super::super::{to_dot, DotFormat::Dependencies, DotFormat::Schedule};
        use super::*;

        #[test]
        fn empty_graph() {
            let graph = example();
            let filtered = graph.filter_vertices(|_| false);

            let actual = format!("{}", to_dot(&filtered, Dependencies));
            let expected = "digraph dependencies {\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| *v == "a".to_string());

            let actual = format!("{}", to_dot(&filtered, Dependencies));
            let expected = "digraph dependencies {\n  \"a\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_edge() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b"].contains(&v.as_str()));

            let actual = format!("{}", to_dot(&filtered, Dependencies));
            let expected = "digraph dependencies {\n  \"a\" -> \"b\" [style=dashed]\n  \"b\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_strong_edge() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "d"].contains(&v.as_str()));

            let actual = format!("{}", to_dot(&filtered, Dependencies));
            let expected = "digraph dependencies {\n  \"a\"\n  \"d\" -> \"a\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn graph() {
            let graph = example();
            let filtered = graph.filter_vertices(|v| ["a", "b", "c", "d"].contains(&v.as_str()));

            let actual = format!("{}", to_dot(&filtered, Dependencies));
            let expected = "digraph dependencies {\n  \
                            \"a\" -> \"b\" [style=dashed]\n  \
                            \"a\" -> \"c\" [style=dashed]\n  \
                            \"b\" -> \"c\" [style=dashed]\n  \
                            \"c\"\n  \
                            \"d\" -> \"a\"\n\
                            }\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn schedule() {
            let graph = example().reverse();
            let filtered = graph.filter_edges(|c| *c == Dependency::Strong);

            let actual = format!("{}", to_dot(&filtered, Schedule));
            let expected = "digraph schedule {\n  \
                            randir=\"LR\"\n  \
                            node [shape=box]\n  \
                            \"a\" -> \"d\"\n  \
                            \"a\" -> \"e\"\n  \
                            \"b\" -> \"e\"\n  \
                            \"c\"\n  \
                            \"d\"\n  \
                            \"e\"\n\
                            }\n";

            assert_eq!(actual, expected);
        }
    }

    // Fixture

    fn example() -> Graph<String, Dependency> {
        Graph::from([
            (
                "a".into(),
                vec![
                    ("b".into(), Dependency::Weak),
                    ("c".into(), Dependency::Weak),
                ],
            ),
            ("b".into(), vec![("c".into(), Dependency::Weak)]),
            ("c".into(), vec![]),
            ("d".into(), vec![("a".into(), Dependency::Strong)]),
            (
                "e".into(),
                vec![
                    ("a".into(), Dependency::Strong),
                    ("b".into(), Dependency::Strong),
                ],
            ),
        ])
    }
}
