use super::{Edge, Graph, Subgraph, Vertex};

impl<V, E> PartialEq for Graph<V, E>
where
    V: Vertex,
    E: Edge,
{
    fn eq(&self, other: &Self) -> bool {
        self.iter()
            .zip(other.iter())
            .all(|(a, b)| a.0.eq(b.0) && a.1.eq(b.1))
    }
}

impl<V, E> PartialEq for Subgraph<'_, V, E>
where
    V: Vertex,
    E: Edge,
{
    fn eq(&self, other: &Self) -> bool {
        self.iter()
            .zip(other.iter())
            .all(|(a, b)| a.0.eq(b.0) && a.1.eq(b.1))
    }
}

impl<V, E> PartialEq<Subgraph<'_, V, E>> for Graph<V, E>
where
    V: Vertex,
    E: Edge,
{
    fn eq(&self, other: &Subgraph<'_, V, E>) -> bool {
        self.iter()
            .zip(other.iter())
            .all(|(a, b)| a.0.eq(b.0) && a.1.eq(b.1))
    }
}

impl<V, E> PartialEq<Graph<V, E>> for Subgraph<'_, V, E>
where
    V: Vertex,
    E: Edge,
{
    fn eq(&self, other: &Graph<V, E>) -> bool {
        self.iter()
            .zip(other.iter())
            .all(|(a, b)| a.0.eq(b.0) && a.1.eq(b.1))
    }
}
