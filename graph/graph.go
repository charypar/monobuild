package graph

import "github.com/charypar/monobuild/set"

// Graph is a DAG with string labeled vertices
type Graph struct {
	edges map[string]set.Set
}

// New creates a new Graph from a map of vertex to vertex label describing the edges
func New(edges map[string][]string) Graph {
	edgs := make(map[string]set.Set)

	for v, es := range edges {
		edgs[v] = set.New([]string{})

		for _, e := range es {
			edgs[v].Add(e)
		}
	}

	return Graph{edgs}
}

// Children returns the vertices that are connected to given vertices with an edge
func (g Graph) Children(vs set.Set) set.Set {
	all := set.New([]string{})

	for _, vertex := range vs.AsStrings() {
		grandchildren, found := g.edges[vertex]

		if found {
			all = all.Union(grandchildren)
		}
	}

	return all
}

// Descendants returns all the vertices x for which a path to x exists from any of
// the vertices given
func (g Graph) Descendants(vertices set.Set) set.Set {
	descendants := g.Children(vertices)
	discovered := descendants

	for discovered.Size() > 0 {
		grandchildren := g.Children(discovered)
		discovered = grandchildren.Without(descendants)

		descendants = descendants.Union(discovered)
	}

	return descendants
}

// Reverse returns a new graph with edges reversed
func (g Graph) Reverse() Graph {
	edges := make(map[string]set.Set)

	for v, es := range g.edges {
		for _, e := range es.AsStrings() {
			_, ok := edges[e]
			if !ok {
				edges[e] = set.New([]string{})
			}

			edges[e].Add(v)
		}
	}

	return Graph{edges}
}
