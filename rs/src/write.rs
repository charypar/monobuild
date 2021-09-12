use std::fmt::Display;

use crate::core::Dependency;
use crate::graph::Graph;

pub enum TextFormat {
    Simple,
    Full,
}

pub struct Text<'a, V>
where
    V: Clone + Ord,
{
    graph: &'a Graph<V, Dependency>,
    format: TextFormat,
}

pub enum DotFormat {
    Dependencies,
    Schedule,
}

pub struct Dot<'a, V>
where
    V: Clone + Ord,
{
    graph: &'a Graph<V, Dependency>,
    format: DotFormat,
}

pub fn to_text<V>(graph: &Graph<V, Dependency>, format: TextFormat) -> Text<V>
where
    V: Clone + Ord,
{
    Text { graph, format }
}

pub fn to_dot<V>(graph: &Graph<V, Dependency>, format: DotFormat) -> Dot<V>
where
    V: Clone + Ord,
{
    Dot { graph, format }
}

impl<'a, V> Display for Text<'a, V>
where
    V: Clone + Ord + Display,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        for (v, es) in &self.graph.edges {
            let edges = es
                .iter()
                .map(|e| match self.format {
                    TextFormat::Full if e.color == Dependency::Strong => format!("!{}", e.to),
                    _ => format!("{}", e.to),
                })
                .collect::<Vec<_>>()
                .join(", ");

            write!(f, "{}: {}", v, edges)?;

            write!(f, "\n")?;
        }

        Ok(())
    }
}

impl<'a, V> Display for Dot<'a, V>
where
    V: Clone + Ord + Display,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self.format {
            DotFormat::Dependencies => write!(f, "digraph dependencies {{\n")?,
            DotFormat::Schedule => write!(
                f,
                "digraph schedule {{\n  randir=\"LR\"\n  node [shape=box]\n"
            )?,
        }

        for (v, es) in &self.graph.edges {
            if es.is_empty() {
                write!(f, "  \"{}\"\n", v)?;
            }

            for e in es {
                match e.color {
                    Dependency::Weak => write!(f, "  \"{}\" -> \"{}\" [style=dashed]\n", v, e.to)?,
                    Dependency::Strong => write!(f, "  \"{}\" -> \"{}\"\n", v, e.to)?,
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
    use crate::graph::{Edge, Graph};

    mod text {
        use super::super::{to_text, TextFormat::Full, TextFormat::Simple};
        use super::*;

        #[test]
        fn empty_graph() {
            let graph = example().filter(|_| false, |_| true);

            let actual = format!("{}", to_text(&graph, Simple));
            let expected = "";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex() {
            let graph = example().filter(|v| *v == "a".to_string(), |_| true);

            let actual = format!("{}", to_text(&graph, Simple));
            let expected = "a: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_edge() {
            let graph = example().filter(|v| ["a", "b"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_text(&graph, Simple));
            let expected = "a: b\nb: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn edge_fan() {
            let graph = example().filter(|v| ["a", "b", "c"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_text(&graph, Simple));
            let expected = "a: b, c\nb: c\nc: \n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn graph() {
            let graph = example().filter(|v| ["a", "b", "c", "d"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_text(&graph, Simple));
            let expected = "a: b, c\nb: c\nc: \nd: a\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn full_format() {
            let graph = example().filter(|v| ["a", "b", "c", "d"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_text(&graph, Full));
            let expected = "a: b, c\nb: c\nc: \nd: !a\n";

            assert_eq!(actual, expected);
        }
    }

    mod dot {
        use super::super::{to_dot, DotFormat::Dependencies, DotFormat::Schedule};
        use super::*;

        #[test]
        fn empty_graph() {
            let graph = example().filter(|_| false, |_| true);

            let actual = format!("{}", to_dot(&graph, Dependencies));
            let expected = "digraph dependencies {\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex() {
            let graph = example().filter(|v| *v == "a".to_string(), |_| true);

            let actual = format!("{}", to_dot(&graph, Dependencies));
            let expected = "digraph dependencies {\n  \"a\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_edge() {
            let graph = example().filter(|v| ["a", "b"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_dot(&graph, Dependencies));
            let expected = "digraph dependencies {\n  \"a\" -> \"b\" [style=dashed]\n  \"b\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_strong_edge() {
            let graph = example().filter(|v| ["a", "d"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_dot(&graph, Dependencies));
            let expected = "digraph dependencies {\n  \"a\"\n  \"d\" -> \"a\"\n}\n";

            assert_eq!(actual, expected);
        }

        #[test]
        fn graph() {
            let graph = example().filter(|v| ["a", "b", "c", "d"].contains(&v.as_str()), |_| true);

            let actual = format!("{}", to_dot(&graph, Dependencies));
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
            let graph = example()
                .filter(|_| true, |c| *c == Dependency::Strong)
                .reverse();

            let actual = format!("{}", to_dot(&graph, Schedule));
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
        Graph::new(vec![
            (
                "a".into(),
                vec![
                    Edge::new("b".into(), Dependency::Weak),
                    Edge::new("c".into(), Dependency::Weak),
                ],
            ),
            ("b".into(), vec![Edge::new("c".into(), Dependency::Weak)]),
            ("c".into(), vec![]),
            ("d".into(), vec![Edge::new("a".into(), Dependency::Strong)]),
            (
                "e".into(),
                vec![
                    Edge::new("a".into(), Dependency::Strong),
                    Edge::new("b".into(), Dependency::Strong),
                ],
            ),
        ])
    }
}
