use std::fmt::Debug;

use super::{Edge, Graph, Subgraph, Vertex};

impl<V, E> Debug for Graph<V, E>
where
    V: Vertex + Debug,
    E: Edge + Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let edges: Vec<_> = self
            .edges
            .iter()
            .enumerate()
            .flat_map(|(fi, es)| {
                es.iter()
                    .map(move |(ti, l)| (self.vertices[fi].clone(), self.vertices[*ti].clone(), l))
            })
            .collect();

        f.debug_struct("Graph")
            .field("vertices", &self.vertices)
            .field("edges", &edges)
            .finish()
    }
}

impl<V, E> Debug for Subgraph<'_, V, E>
where
    V: Vertex + Debug,
    E: Edge + Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let vertices: Vec<_> = self
            .graph
            .vertices
            .iter()
            .enumerate()
            .filter_map(|(i, v)| if self.vertex_mask[i] { Some(v) } else { None })
            .collect();

        let edges: Vec<_> = self
            .graph
            .edges
            .iter()
            .enumerate()
            .flat_map(|(fi, es)| {
                es.iter().filter_map(move |(ti, l)| {
                    if self.edge_mask.contains(&(fi, *ti)) {
                        Some((
                            self.graph.vertices[fi].clone(),
                            self.graph.vertices[*ti].clone(),
                            l,
                        ))
                    } else {
                        None
                    }
                })
            })
            .collect();

        f.debug_struct("Subgraph")
            .field("vertices", &vertices)
            .field("edges", &edges)
            .finish()
    }
}
