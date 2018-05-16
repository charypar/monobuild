package graph

import (
	"sort"

	"github.com/charypar/monobuild/set"
)

// Graph is a DAG with string labeled vertices
type Graph struct {
	edges map[string]set.Set
}

// New creates a new Graph from a map of vertex -> vertex label that describes the edges
func New(edges map[string][]string) Graph {
	graph := make(map[string]set.Set)

	for v, es := range edges {
		graph[v] = set.New([]string{})

		for _, e := range es {
			graph[v].Add(e)

			// Normalise the graph - every node is a key in the map
			_, ok := graph[e]
			if !ok {
				graph[e] = set.New([]string{})
			}
		}
	}

	return Graph{graph}
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
		_, ok := edges[v]
		if !ok {
			edges[v] = set.New([]string{})
		}

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

// Subgraph filters the graph to only the nodes listed
func (g Graph) Subgraph(nodes []string) Graph {
	filter := set.New(nodes)
	filtered := make(map[string]set.Set, len(g.edges))

	for v, es := range g.edges {
		if !filter.Has(v) {
			continue
		}

		filtered[v] = set.New([]string{})
		for _, e := range es.AsStrings() {
			if !filter.Has(e) {
				continue
			}

			filtered[v].Add(e)
		}
	}

	return Graph{filtered}
}

// AsStrings returns the graph as a map[string][]string
func (g Graph) AsStrings() map[string][]string {
	result := make(map[string][]string, len(g.edges))

	for v, es := range g.edges {
		sorted := es.AsStrings()
		sort.Strings(sorted)

		result[v] = sorted
	}

	return result
}
