use std::collections::{BTreeMap, BTreeSet};

mod debug;
mod iter;
mod partial_eq;

use self::iter::{GraphIter, Vertices};

pub trait Vertex: PartialEq + Clone {}
impl<T> Vertex for T where T: PartialEq + Clone {}
pub trait Edge: PartialEq + Ord + Copy {}
impl<T> Edge for T where T: PartialEq + Ord + Copy {}

pub struct Graph<V, E>
where
    V: Vertex,
    E: Edge,
{
    vertices: Vec<V>,                 // vertex index -> V
    edges: Vec<BTreeSet<(usize, E)>>, // vertex index -> {(vertex index, E)}
}

impl<V, E> Graph<V, E>
where
    V: Vertex,
    E: Edge,
{
    pub fn new() -> Self {
        Graph {
            vertices: vec![],
            edges: vec![],
        }
    }

    // Iterate over graph or vertices

    pub fn vertices<'g>(&'g self) -> Vertices<'g, '_, V> {
        Vertices {
            iter: self.vertices.iter().enumerate(),
            mask: None,
        }
    }

    pub fn iter<'g>(&'g self) -> GraphIter<'g, '_, V, E> {
        GraphIter {
            graph: self,
            iter: self.vertices.iter().enumerate(),
            masks: None,
        }
    }

    // Scope graph

    pub fn roots<'g>(&'g self) -> Subgraph<'g, V, E> {
        let mut vertex_mask: Vec<_> = std::iter::repeat(true).take(self.vertices.len()).collect();

        // Remove all vertices with an incoming edge
        for edgs in &self.edges {
            for (to, _) in edgs {
                vertex_mask[*to] = false;
            }
        }

        // Remove edges using the mask we just made
        let edge_mask = (0..self.vertices.len())
            .flat_map(|from| {
                self.edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, _)| (from, *to))
            })
            .filter(|(from, to)| vertex_mask[*from] && vertex_mask[*to])
            .collect();

        Subgraph {
            graph: self,
            vertex_mask,
            edge_mask,
        }
    }

    // Removes vertices and associated edges
    pub fn filter_vertices<'g, P>(&'g self, predicate: P) -> Subgraph<'g, V, E>
    where
        P: Fn(&V) -> bool,
    {
        // Remove vertices
        let vertex_mask: Vec<bool> = self.vertices.iter().map(|v| predicate(v)).collect();

        // Remove edges using the mask we just made
        let edge_mask = (0..self.vertices.len())
            .flat_map(|from| {
                self.edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, _)| (from, *to))
            })
            .filter(|(from, to)| vertex_mask[*from] && vertex_mask[*to])
            .collect();

        Subgraph {
            graph: self,
            vertex_mask,
            edge_mask,
        }
    }

    // Removes edges but keeps vertices alone
    pub fn filter_edges<'g, P>(&'g self, predicate: P) -> Subgraph<'g, V, E>
    where
        P: Fn(&E) -> bool,
    {
        let vertex_mask = std::iter::repeat(true).take(self.vertices.len()).collect();
        let edge_mask = (0..self.vertices.len())
            .flat_map(|from| {
                self.edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, label)| (from, *to, label))
            })
            .filter_map(|(from, to, label)| {
                if predicate(label) {
                    Some((from, to))
                } else {
                    None
                }
            })
            .collect();

        Subgraph {
            graph: self,
            vertex_mask,
            edge_mask,
        }
    }

    // Transform graph

    pub fn reverse(&self) -> Self {
        let vertices = self.vertices.clone();
        let mut edges: Vec<_> = std::iter::repeat_with(|| BTreeSet::new())
            .take(vertices.len())
            .collect();

        for (to, edgs) in self.edges.iter().enumerate() {
            for (from, label) in edgs {
                edges[*from].insert((to, *label));
            }
        }

        Graph { vertices, edges }
    }
}

// Construct a graph from adjacency list
impl<Adj, Edg, V, E> From<Adj> for Graph<V, E>
where
    Adj: IntoIterator<Item = (V, Edg)>,
    Edg: IntoIterator<Item = (V, E)>,
    V: Vertex + Ord,
    E: Edge,
{
    fn from(graph: Adj) -> Self {
        // Make a local copy to allow multiple passes
        let graph: Vec<(V, BTreeSet<_>)> = graph
            .into_iter()
            .map(|(v, es)| (v, es.into_iter().collect()))
            .collect();

        // Find and insert all the vertices, sorted using Ord
        let mut vertices: BTreeSet<V> = BTreeSet::new();
        for (from, edg) in &graph {
            vertices.insert(from.clone());

            for (to, _) in edg.into_iter() {
                vertices.insert(to.clone());
            }
        }

        // Index and reverse index them
        let vertices: Vec<_> = vertices.into_iter().collect();
        let vertex_index: BTreeMap<V, usize> = vertices
            .iter()
            .enumerate()
            .map(|(i, v)| (v.clone(), i))
            .collect();

        // Allocate edge mapping
        let mut edges: Vec<_> = std::iter::repeat_with(|| BTreeSet::new())
            .take(vertices.len())
            .collect();

        // Insert all the edges
        for (from, edg) in graph {
            let from_idx = *vertex_index.get(&from).expect("a vertex to exist");
            let edges_from = edges.get_mut(from_idx).expect("an edge set to exist");

            for (to, label) in edg {
                let to_idx = *vertex_index.get(&to).expect("a vertex to exist");

                edges_from.insert((to_idx, label));
            }
        }

        Graph { vertices, edges }
    }
}

pub struct Subgraph<'g, V, E>
where
    V: Vertex + 'g,
    E: Edge,
{
    graph: &'g Graph<V, E>,
    vertex_mask: Vec<bool>,
    edge_mask: BTreeSet<(usize, usize)>,
}

impl<'g, V, E> Subgraph<'g, V, E>
where
    V: Vertex + 'g,
    E: Edge,
{
    // Inspect graph

    pub fn iter<'s>(&'s self) -> GraphIter<'g, 's, V, E> {
        GraphIter {
            graph: self.graph,
            iter: self.graph.vertices.iter().enumerate(),
            masks: Some((&self.vertex_mask, &self.edge_mask)),
        }
    }

    pub fn vertices<'s>(&'s self) -> Vertices<'g, 's, V> {
        Vertices {
            iter: self.graph.vertices.iter().enumerate(),
            mask: Some(&self.vertex_mask),
        }
    }

    pub fn roots(&self) -> Subgraph<'g, V, E> {
        let mut vertex_mask: Vec<_> = self.vertex_mask.clone();

        // Remove all vertices with an incoming edge
        for (from, edgs) in self.graph.edges.iter().enumerate() {
            for (to, _) in edgs {
                if self.edge_mask.contains(&(from, *to)) {
                    vertex_mask[*to] = false;
                }
            }
        }

        // Remove edges using the mask we just made, respecting the original edge mask
        let edge_mask = (0..self.graph.vertices.len())
            .flat_map(|from| {
                self.graph
                    .edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, _)| (from, *to))
            })
            .filter(|(from, to)| {
                vertex_mask[*from] && vertex_mask[*to] && self.edge_mask.contains(&(*from, *to))
            })
            .collect();

        Subgraph {
            graph: self.graph,
            vertex_mask,
            edge_mask,
        }
    }

    // Removes vertices and associated edges
    pub fn filter_vertices<P>(&self, predicate: P) -> Subgraph<'g, V, E>
    where
        P: Fn(&V) -> bool,
    {
        // Remove vertices
        let vertex_mask: Vec<bool> = self
            .graph
            .vertices
            .iter()
            .enumerate()
            .map(|(i, v)| self.vertex_mask[i] && predicate(v))
            .collect();

        // Remove edges using the mask we just made
        let edge_mask = self
            .graph
            .vertices
            .iter()
            .enumerate()
            .flat_map(|(from, _)| {
                self.graph
                    .edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, _)| (from, *to))
            })
            .filter(|(from, to)| {
                vertex_mask[*from] && vertex_mask[*to] && self.edge_mask.contains(&(*from, *to))
            })
            .collect();

        Subgraph {
            graph: self.graph,
            vertex_mask,
            edge_mask,
        }
    }

    // Removes edges but keeps vertices alone
    pub fn filter_edges<P>(&self, predicate: P) -> Subgraph<'g, V, E>
    where
        P: Fn(&E) -> bool,
    {
        let edge_mask = (0..self.vertex_mask.len())
            .flat_map(|from| {
                self.graph
                    .edges
                    .get(from)
                    .expect("edges to exist")
                    .iter()
                    .map(move |(to, label)| (from, *to, label))
            })
            .filter_map(|(from, to, label)| {
                if predicate(label) && self.edge_mask.contains(&(from, to)) {
                    Some((from, to))
                } else {
                    None
                }
            })
            .collect();

        Subgraph {
            graph: self.graph,
            vertex_mask: self.vertex_mask.clone(),
            edge_mask,
        }
    }

    // Expand subgraph along original edges, applying a predicate to decide whether to follow an edge
    pub fn expand_via<P>(&self, predicate: P) -> Subgraph<'g, V, E>
    where
        P: Fn(&E) -> bool,
    {
        let mut vertex_mask = self.vertex_mask.clone();
        let mut edge_mask = self.edge_mask.clone();

        // front is the set of vertices we're attempting to expand the graph from
        // via edges that are *not* in the edge mask (because those are already included)
        // and that pass the predicate test
        let mut front: Vec<usize> = vertex_mask
            .iter()
            .enumerate()
            .filter_map(|(idx, inc)| if *inc { Some(idx) } else { None })
            .collect();

        while front.len() > 0 {
            front = front
                .iter()
                .flat_map(|from| {
                    let edges = self.graph.edges.get(*from).expect("edges to exist");

                    edges.iter().map(move |(to, label)| (*from, *to, label))
                })
                .filter_map(|(from, to, label)| {
                    if !self.edge_mask.contains(&(from, to)) && predicate(label) {
                        // expand the graph
                        vertex_mask[to] = true;
                        edge_mask.insert((from, to));

                        // and add the destination vertec to the next iteration front
                        Some(to)
                    } else {
                        None
                    }
                })
                .collect::<Vec<_>>();
        }

        Subgraph {
            graph: self.graph,
            vertex_mask,
            edge_mask,
        }
    }

    pub fn expand(&self) -> Self {
        self.expand_via(|_| true)
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn new_normalizes_graph() {
        let graph = Graph::<usize, usize>::from([(1, [(2, 0), (3, 0)])]);

        let expected = vec![1, 2, 3];
        let actual = graph.vertices().cloned().collect::<Vec<_>>();

        assert_eq!(actual, expected);
    }

    mod filter {
        use super::super::*;

        #[test]
        fn filters_an_empty_graph() {
            let graph = Graph::<usize, usize>::new();

            let expected = Graph::<usize, usize>::new();
            let actual_v = graph.filter_vertices(|_v| true);
            let actual_e = graph.filter_edges(|_e| true);

            assert_eq!(actual_v, expected);
            assert_eq!(actual_e, expected);
        }

        #[test]
        fn filters_a_vertex_from_a_graph() {
            let graph = Graph::from([(1, [(2, 0), (3, 0)])]);

            let expected = Graph::from([(1, [(3, 0)])]);
            let actual = graph.filter_vertices(|v| *v != 2);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_an_edge_from_a_graph() {
            let graph = Graph::from([(1, vec![(2, 0), (3, 1)]), (2, vec![(3, 0)])]);

            let expected = Graph::from([(1, [(2, 0)]), (2, [(3, 0)])]);
            let actual = graph.filter_edges(|e| *e != 1);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_a_vertex_from_a_subgraph() {
            let graph = Graph::from([(1, [(2, 0), (3, 0)])]);

            let expected = Graph::from([(1, [])]);
            let actual = graph
                .filter_vertices(|v| *v != 2)
                .filter_vertices(|v| *v != 3);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_an_edge_from_a_subgraph() {
            let graph = Graph::from([(1, vec![(2, 0), (3, 1)]), (2, vec![(3, 0), (1, 2)])]);

            let expected = Graph::from([(1, [(2, 0)]), (2, [(3, 0)])]);
            let actual = graph.filter_edges(|e| *e != 1).filter_edges(|e| *e != 2);

            assert_eq!(actual, expected);
        }

        #[test]
        fn filters_an_edge_from_a_vertex_filtered_subgraph() {
            let graph = Graph::from([(1, vec![(2, 0), (3, 1)]), (2, vec![(3, 0), (1, 2)])]);

            let expected = Graph::from([(1, vec![(2, 0)]), (2, vec![])]);
            let actual = graph.filter_vertices(|v| *v != 3).filter_edges(|e| *e != 2);

            assert_eq!(actual, expected);
        }
    }

    mod expand {
        use super::super::*;

        #[test]
        fn empty_graph() {
            let graph = Graph::<usize, usize>::new();

            let actual = graph.filter_vertices(|v| *v == 1).expand();
            let expected = Graph::<usize, usize>::new();

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_vertex_graph() {
            let graph = Graph::<_, usize>::from([(1, [])]);

            let actual = graph.filter_vertices(|v| *v == 1).expand();
            let expected = Graph::<_, usize>::from([(1, [])]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn single_child_of_one_vertex() {
            let graph = Graph::from([(1, [(2, 0)]), (3, [(2, 0)])]);

            let actual = graph.filter_vertices(|v| *v == 1).expand();
            let expected = Graph::from([(1, [(2, 0)])]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn multiple_children_of_one_vertex() {
            let graph = Graph::from([(1, [(2, 0), (3, 0)])]);

            let actual = &graph.filter_vertices(|v| *v == 1).expand();
            let expected = &graph;

            assert_eq!(actual, expected);
        }

        #[test]
        fn all_descendants_of_one_vertex() {
            let graph = Graph::from([(1, [(2, 0), (3, 0)]), (2, [(3, 0), (4, 0)])]);

            let actual = &graph.filter_vertices(|v| *v == 1).expand();
            let expected = &graph;

            assert_eq!(actual, expected);
        }

        #[test]
        fn all_descendants_of_multiple_vertices() {
            let graph = Graph::from([
                (1, vec![(4, 0), (5, 0)]),
                (2, vec![(6, 0)]),
                (3, vec![(8, 0), (9, 0)]),
                (4, vec![(7, 0)]),
                (7, vec![(8, 0)]),
                (8, vec![(5, 0)]),
            ]);

            let actual = graph.filter_vertices(|v| *v == 1 || *v == 2).expand();
            let expected = Graph::from([
                (1, vec![(4, 0), (5, 0)]),
                (2, vec![(6, 0)]),
                (4, vec![(7, 0)]),
                (7, vec![(8, 0)]),
                (8, vec![(5, 0)]),
            ]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn all_descendants_via_edges_of_a_color() {
            let graph = Graph::from([
                (1, vec![(4, 1), (5, 0)]),
                (2, vec![(6, 1)]),
                (3, vec![(8, 0), (9, 0)]),
                (4, vec![(7, 1)]),
                (7, vec![(8, 0)]),
                (8, vec![(5, 0)]),
            ]);

            let actual = graph
                .filter_vertices(|v| *v == 1 || *v == 2 || *v == 5)
                .expand_via(|e| *e == 1);
            let expected = Graph::from([
                (1, vec![(4, 1), (5, 0)]),
                (2, vec![(6, 1)]),
                (4, vec![(7, 1)]),
                (5, vec![]),
                (7, vec![]),
            ]);

            assert_eq!(actual, expected);
        }
    }

    mod reverse {
        use super::super::*;

        #[test]
        fn reverses_empty_graph() {
            let graph = Graph::<usize, usize>::new();

            let expected = Graph::<usize, usize>::new();
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_single_edge() {
            let graph = Graph::<usize, usize>::from([(1, [(2, 0)])]);

            let expected = Graph::<usize, usize>::from([(2, [(1, 0)])]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_a_fan() {
            let graph = Graph::<usize, usize>::from([(1, [(2, 0), (3, 1), (4, 0)])]);

            let expected =
                Graph::<usize, usize>::from([(2, [(1, 0)]), (3, [(1, 1)]), (4, [(1, 0)])]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }

        #[test]
        fn reverses_a_complex_graph() {
            let graph = Graph::<usize, usize>::from([
                (1, vec![(2, 0), (3, 1)]),
                (2, vec![(3, 0)]),
                (3, vec![(4, 0)]),
            ]);

            let expected = Graph::<usize, usize>::from([
                (2, vec![(1, 0)]),
                (3, vec![(1, 1), (2, 0)]),
                (4, vec![(3, 0)]),
            ]);
            let actual = graph.reverse();

            assert_eq!(actual, expected);
        }
    }

    mod roots {
        use super::super::*;

        #[test]
        fn of_an_empty_graph() {
            let graph = Graph::<usize, usize>::new();

            let actual = graph.roots();
            let expected = Graph::<usize, usize>::new();

            assert_eq!(actual, expected);
        }

        #[test]
        fn of_a_single_edge() {
            let graph = Graph::from([(1, [(2, 0)])]);

            let actual = graph.roots();
            let expected = Graph::from([(1, [])]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn of_a_complex_graph() {
            let graph = Graph::from([
                (1, vec![(2, 0), (3, 0)]),
                (2, vec![(3, 0)]),
                (4, vec![(5, 0)]),
                (5, vec![(6, 0)]),
                (7, vec![(2, 0)]),
            ]);

            let actual = graph.roots();
            let expected = Graph::from([(1, vec![]), (4, vec![]), (7, vec![])]);

            assert_eq!(actual, expected);
        }

        #[test]
        fn of_a_subgraph() {
            let graph = Graph::from([
                (1, vec![(2, 0), (3, 1)]),
                (2, vec![(3, 1)]),
                (4, vec![(5, 0)]),
                (5, vec![(6, 0)]),
                (7, vec![(2, 0)]),
            ]);

            let actual = graph
                .filter_edges(|e| *e != 1)
                .filter_vertices(|v| *v != 7)
                .roots();
            let expected = Graph::from([(1, vec![]), (3, vec![]), (4, vec![])]);

            assert_eq!(actual, expected);
        }
    }
}
