use std::{
    cmp,
    collections::{BTreeMap, BTreeSet},
};

#[derive(PartialEq, Eq, Clone, Debug)]
pub struct Edge<V, C>
where
    V: Clone + PartialEq + Ord,
    C: Clone + Eq,
{
    pub to: V,
    pub color: C,
}

impl<V, C> Edge<V, C>
where
    V: Clone + PartialEq + Ord,
    C: Clone + Eq,
{
    pub fn new(to: V, color: C) -> Self {
        Self { to, color }
    }
}

impl<V, C> cmp::PartialOrd for Edge<V, C>
where
    V: Clone + Ord,
    C: Clone + Eq,
{
    fn partial_cmp(&self, other: &Self) -> Option<cmp::Ordering> {
        Some(self.to.cmp(&other.to))
    }
}

impl<V, C> cmp::Ord for Edge<V, C>
where
    V: Clone + Ord,
    C: Clone + Eq,
{
    fn cmp(&self, other: &Self) -> cmp::Ordering {
        self.to.cmp(&other.to)
    }
}

#[derive(PartialEq, Debug)]
pub struct Graph<V, C>
where
    V: Clone + Ord,
    C: Clone + Eq,
{
    pub edges: BTreeMap<V, BTreeSet<Edge<V, C>>>,
}

impl<V, C> Graph<V, C>
where
    V: Clone + Ord,
    C: Clone + Eq,
{
    // Constructs a new graph from an adjacency list normalizing the graph.
    pub fn new(graph: Vec<(V, Vec<Edge<V, C>>)>) -> Self {
        let mut edges = BTreeMap::new();

        for entry in graph {
            // Ensure each vertex has a key in the edges map
            for edge in &entry.1 {
                edges
                    .entry(edge.to.clone())
                    .or_insert_with(|| BTreeSet::new());
            }

            // Insert all the edges
            let vertex = edges.entry(entry.0).or_insert_with(|| BTreeSet::new());
            for edge in entry.1 {
                vertex.insert(edge);
            }
        }

        Self { edges }
    }

    pub fn vertices(&self) -> BTreeSet<&V> {
        self.edges.keys().collect()
    }

    pub fn children(&self, vertices: &BTreeSet<V>) -> BTreeSet<&V> {
        vertices
            .iter()
            .flat_map(|v| self.edges.get(v))
            .flatten()
            .map(|e| &e.to)
            .collect()
    }

    pub fn descendants(&self, vertices: &BTreeSet<V>) -> BTreeSet<&V> {
        let mut result = self.children(&vertices);
        // newly discovered vertices
        let mut front = result.clone();

        loop {
            // FIXME refactor descendants in terms of children without introducing cloning
            front = front
                .iter()
                .flat_map(|v| self.edges.get(v))
                .flatten()
                .map(|e| &e.to)
                .filter(|v| !result.contains(v))
                .collect();

            if front.is_empty() {
                break;
            }

            // add discovered vertices to the results
            for vertex in &front {
                result.insert(vertex);
            }
        }

        result
    }

    pub fn reverse(&self) -> Graph<V, C> {
        let mut edges = BTreeMap::new();

        for (v, es) in &self.edges {
            edges.entry(v.clone()).or_insert_with(|| BTreeSet::new());

            for e in es {
                edges
                    .entry(e.to.clone())
                    .or_insert_with(|| BTreeSet::new())
                    .insert(Edge::new(v.clone(), e.color.clone()));
            }
        }

        Graph { edges }
    }

    pub fn filter<VP, EP>(&self, vertex_predicate: VP, edge_predicate: EP) -> Graph<V, C>
    where
        VP: Fn(&V) -> bool,
        EP: Fn(&C) -> bool,
    {
        let mut edges = BTreeMap::new();

        for (v, es) in &self.edges {
            if vertex_predicate(&v) {
                edges.entry(v.clone()).or_insert_with(|| BTreeSet::new());

                for e in es {
                    if vertex_predicate(&e.to) {
                        let entry = edges.entry(v.clone()).or_insert_with(|| BTreeSet::new());

                        if edge_predicate(&e.color) {
                            entry.insert(e.clone());
                        }
                    }
                }
            }
        }

        Graph { edges }
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn new_normalizes_graph() {
        let graph = Graph::new(vec![(1, vec![Edge::new(2, 0), Edge::new(3, 0)])]);

        let expected = vec![&1, &2, &3].into_iter().collect();
        let actual = graph.vertices();

        assert_eq!(actual, expected);
    }

    mod children {
        use super::super::*;

        #[test]
        fn empty_graph() {
            let graph = Graph::<usize, usize>::new(vec![]);
            let query = vec![1].into_iter().collect();

            let actual = graph.children(&query);
            let expected = vec![].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex_graph() {
            let graph = Graph::<_, usize>::new(vec![(1, vec![])]);
            let query = vec![1].into_iter().collect();

            let actual = graph.children(&query);
            let expected = vec![].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_child_of_one_vertex() {
            let graph = Graph::new(vec![(1, vec![Edge::new(2, 0)])]);
            let query = vec![1].into_iter().collect();

            let actual = graph.children(&query);
            let expected = vec![&2].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn multiple_children_of_one_vertex() {
            let graph = Graph::new(vec![(1, vec![Edge::new(2, 0), Edge::new(3, 0)])]);

            let query = vec![1].into_iter().collect();

            let actual = graph.children(&query);
            let expected = vec![&2, &3].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn multiple_children_of_multiple_vertices() {
            let graph = Graph::new(vec![
                (1, vec![Edge::new(2, 0), Edge::new(3, 0)]),
                (2, vec![Edge::new(3, 0), Edge::new(4, 0)]),
            ]);

            let query = vec![1, 2].into_iter().collect();

            let actual = graph.children(&query);
            let expected = vec![&2, &3, &4].into_iter().collect();

            assert_eq!(actual, expected);
        }
    }

    mod descendants {
        use super::super::*;

        #[test]
        fn empty_graph() {
            let graph = Graph::<usize, usize>::new(vec![]);
            let query = vec![1].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex_graph() {
            let graph = Graph::<_, usize>::new(vec![(1, vec![])]);
            let query = vec![1].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_child_of_one_vertex() {
            let graph = Graph::new(vec![(1, vec![Edge::new(2, 0)])]);
            let query = vec![1].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![&2].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn multiple_children_of_one_vertex() {
            let graph = Graph::new(vec![(1, vec![Edge::new(2, 0), Edge::new(3, 0)])]);
            let query = vec![1].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![&2, &3].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn all_descendants_of_one_vertex() {
            let graph = Graph::new(vec![
                (1, vec![Edge::new(2, 0), Edge::new(3, 0)]),
                (2, vec![Edge::new(3, 0), Edge::new(4, 0)]),
            ]);

            let query = vec![1].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![&2, &3, &4].into_iter().collect();

            assert_eq!(actual, expected);
        }

        #[test]
        fn all_descendants_of_multiple_vertices() {
            let graph = Graph::new(vec![
                (1, vec![Edge::new(4, 0), Edge::new(5, 0)]),
                (2, vec![Edge::new(6, 0)]),
                (3, vec![Edge::new(8, 0), Edge::new(9, 0)]),
                (4, vec![Edge::new(7, 0)]),
                (7, vec![Edge::new(8, 0)]),
                (8, vec![Edge::new(5, 0)]),
            ]);

            let query = vec![1, 2].into_iter().collect();

            let actual = graph.descendants(&query);
            let expected = vec![&4, &5, &6, &7, &8].into_iter().collect();

            assert_eq!(actual, expected);
        }
    }

    mod reverse {
        use super::super::*;

        #[test]
        fn reverses_empty_graph() {
            let graph = Graph::<usize, usize>::new(vec![]);

            let expected = Graph::<usize, usize>::new(vec![]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_single_edge() {
            let graph = Graph::<usize, usize>::new(vec![(1, vec![Edge::new(2, 0)])]);

            let expected = Graph::<usize, usize>::new(vec![(2, vec![Edge::new(1, 0)])]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_a_fan() {
            let graph = Graph::<usize, usize>::new(vec![(
                1,
                vec![Edge::new(2, 0), Edge::new(3, 1), Edge::new(4, 0)],
            )]);

            let expected = Graph::<usize, usize>::new(vec![
                (2, vec![Edge::new(1, 0)]),
                (3, vec![Edge::new(1, 1)]),
                (4, vec![Edge::new(1, 0)]),
            ]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_a_complex_graph() {
            let graph = Graph::<usize, usize>::new(vec![
                (1, vec![Edge::new(2, 0), Edge::new(3, 1)]),
                (2, vec![Edge::new(3, 0)]),
                (3, vec![Edge::new(4, 0)]),
            ]);

            let expected = Graph::<usize, usize>::new(vec![
                (2, vec![Edge::new(1, 0)]),
                (3, vec![Edge::new(1, 1), Edge::new(2, 0)]),
                (4, vec![Edge::new(3, 0)]),
            ]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }
    }

    mod filter {
        use super::super::*;

        #[test]
        fn filters_an_empty_graph() {
            let graph = Graph::<usize, usize>::new(vec![]);

            let expected = Graph::<usize, usize>::new(vec![]);
            let actual = graph.filter(|_v| true, |_c| true);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_a_vertex_from_a_graph() {
            let graph = Graph::new(vec![(1, vec![Edge::new(2, 0), Edge::new(3, 0)])]);

            let expected = Graph::new(vec![(1, vec![Edge::new(3, 0)])]);
            let actual = graph.filter(|v| *v != 2, |_| true);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_an_edge_from_a_graph() {
            let graph = Graph::new(vec![
                (1, vec![Edge::new(2, 0), Edge::new(3, 1)]),
                (2, vec![Edge::new(3, 0)]),
            ]);

            let expected = Graph::new(vec![(1, vec![Edge::new(2, 0)]), (2, vec![Edge::new(3, 0)])]);
            let actual = graph.filter(|_| true, |c| *c != 1);

            assert_eq!(actual, expected);
        }
    }
}
