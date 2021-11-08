use std::collections::{btree_set, BTreeSet};
use std::{iter, slice};

use super::{Edge, Graph, Subgraph, Vertex};

// Iterator over the graph adjacency lists
pub struct GraphIter<'g, 's, V, E>
where
    V: Vertex,
    E: Edge,
    's: 'g,
{
    pub(super) graph: &'g Graph<V, E>,
    pub(super) iter: iter::Enumerate<slice::Iter<'g, V>>,
    pub(super) masks: Option<(&'s Vec<bool>, &'s BTreeSet<(usize, usize)>)>,
}

impl<'g, 's, V, E> Iterator for GraphIter<'g, 's, V, E>
where
    V: Vertex,
    E: Edge,
    's: 'g,
{
    type Item = (&'g V, Edges<'g, V, E>);

    fn next<'i>(&'i mut self) -> Option<Self::Item> {
        let edge_mask = self.masks.map(|m| m.1);

        match self.masks {
            Some((vec_mask, _)) => self
                .iter
                .find(|(i, _)| *vec_mask.get(*i).expect("Vector mask size is wrong")),
            None => self.iter.next(),
        }
        .map(|(i, v)| (v, Edges::new(self.graph, i, edge_mask)))
    }
}

// Iterator over edges originating in a vertex
pub struct Edges<'g, V, E>
where
    V: Vertex,
    E: Edge,
{
    vertex: usize,
    graph: &'g Graph<V, E>,
    mask: Option<&'g BTreeSet<(usize, usize)>>,
    iter: btree_set::Iter<'g, (usize, E)>,
}

impl<'g, V, E> Edges<'g, V, E>
where
    V: Vertex,
    E: Edge,
{
    fn new(
        graph: &'g Graph<V, E>,
        index: usize,
        mask: Option<&'g BTreeSet<(usize, usize)>>,
    ) -> Self {
        Self {
            vertex: index,
            graph,
            iter: graph.edges[index].iter(),
            mask,
        }
    }
}

impl<'g, V, E> Iterator for Edges<'g, V, E>
where
    V: Vertex,
    E: Edge,
{
    type Item = (&'g V, E);

    fn next(&mut self) -> Option<Self::Item> {
        match self.mask {
            Some(mask) => {
                let vertex = self.vertex;
                self.iter.find(|(dest, _)| mask.contains(&(vertex, *dest)))
            }
            None => self.iter.next(),
        }
        .map(|(vid, e)| {
            (
                self.graph
                    .vertices
                    .get(*vid)
                    .expect("vertex not found in graph<"),
                *e,
            )
        })
    }
}

// Iterator over the vertices of the graph
pub struct Vertices<'g, 's, V>
where
    V: 'g,
{
    pub(super) iter: std::iter::Enumerate<std::slice::Iter<'g, V>>,
    pub(super) mask: Option<&'s Vec<bool>>,
}

impl<'g, 's, V> Iterator for Vertices<'g, 's, V> {
    type Item = &'g V;

    fn next(&mut self) -> Option<Self::Item> {
        match self.mask {
            Some(mask) => self.iter.find(|(i, _)| mask[*i]).map(|(_, v)| v),
            None => self.iter.next().map(|(_, v)| v),
        }
    }
}

// Converting into iterator

impl<'g, V, E> IntoIterator for &'g Graph<V, E>
where
    V: Vertex,
    E: Edge,
{
    type Item = (&'g V, Edges<'g, V, E>);
    type IntoIter = GraphIter<'g, 'g, V, E>;

    fn into_iter(self) -> Self::IntoIter {
        self.iter()
    }
}

impl<'g, V, E> IntoIterator for &'g Subgraph<'g, V, E>
where
    V: Vertex,
    E: Edge,
{
    type Item = (&'g V, Edges<'g, V, E>);
    type IntoIter = GraphIter<'g, 'g, V, E>;

    fn into_iter(self) -> Self::IntoIter {
        self.iter()
    }
}
