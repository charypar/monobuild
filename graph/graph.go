package graph

import (
	"sort"

	"github.com/charypar/monobuild/set"
)

// Graph is a DAG with string labeled vertices and int colored edges
type Graph struct {
	edges map[string]Edges
}

// New creates a new Graph from a map of the shape
// string: Edge
// where Edge is a struct with a Label and a Colour
func New(graph map[string][]Edge) Graph {
	g := make(map[string]Edges)

	// Copy and normalise the graph - every node needs a key in the map
	for v, es := range graph {
		g[v] = make(Edges, 0, len(es))

		for _, e := range es {
			g[v] = append(g[v], e)

			_, ok := graph[e.Label]
			if !ok {
				g[e.Label] = Edges{}
			}
		}
	}

	return Graph{g}
}

// Vertices returns a full list of vertices in the graph
func (g Graph) Vertices() []string {
	vs := make([]string, 0, len(g.edges))

	for v := range g.edges {
		vs = append(vs, v)
	}

	sort.Strings(vs)
	return vs
}

// Children returns the vertices that are connected to given vertices with an edge
func (g Graph) Children(vertices []string) []string {
	all := set.New([]string{})

	for _, vertex := range vertices {
		grandchildren, found := g.edges[vertex]
		if !found {
			continue
		}

		for _, gc := range grandchildren {
			all.Add(gc.Label)
		}
	}

	result := all.AsStrings()
	sort.Strings(result)

	return result
}

// Descendants returns all the vertices x for which a path to x exists from any of
// the vertices given
func (g Graph) Descendants(vertices []string) []string {
	descendants := set.New(g.Children(vertices))
	discovered := descendants

	for discovered.Size() > 0 {
		grandchildren := set.New(g.Children(discovered.AsStrings()))

		discovered = grandchildren.Without(descendants)
		descendants = descendants.Union(discovered)
	}

	result := descendants.AsStrings()
	sort.Strings(result)

	return result
}

// Reverse returns a new graph with edges reversed
func (g Graph) Reverse() Graph {
	edges := make(map[string]Edges)

	// loop over the map keys deterministically
	sorted := make([]string, 0, len(g.edges))
	for v := range g.edges {
		sorted = append(sorted, v)
	}
	sort.Strings(sorted)

	// here
	for _, v := range sorted {
		_, ok := edges[v]
		if !ok {
			edges[v] = Edges{}
		}

		for _, e := range g.edges[v] {
			_, ok := edges[e.Label]
			if !ok {
				edges[e.Label] = Edges{}
			}

			edges[e.Label] = append(edges[e.Label], Edge{v, e.Colour})
		}
	}

	return Graph{edges}
}

// Subgraph filters the graph to only the nodes listed
func (g Graph) Subgraph(nodes []string) Graph {
	filter := set.New(nodes)
	filtered := make(map[string]Edges, len(g.edges))

	for v, es := range g.edges {
		if !filter.Has(v) {
			continue
		}

		filtered[v] = make(Edges, 0, len(es))
		for _, e := range es {
			if !filter.Has(e.Label) {
				continue
			}

			filtered[v] = append(filtered[v], e)
		}
	}

	return Graph{filtered}
}

// FilterEdges returns a new graph with edges with a colour not present in
// colours removed
func (g Graph) FilterEdges(colours []int) Graph {
	filter := make(map[int]bool, len(colours))
	for _, c := range colours {
		filter[c] = true
	}

	filtered := make(map[string]Edges, len(g.edges))

	for v, es := range g.edges {
		_, ok := filtered[v]
		if !ok {
			filtered[v] = make([]Edge, 0, len(es))
		}

		for _, e := range es {
			_, matches := filter[e.Colour]
			if matches {
				filtered[v] = append(filtered[v], e)
			}
		}
	}

	return Graph{filtered}
}
