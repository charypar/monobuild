package diff

// Vertices is a set of vertices
type Vertices struct {
	members map[string]bool
}

func vertices(names []string) Vertices {
	result := make(map[string]bool, len(names))

	for _, n := range names {
		result[n] = true
	}

	return Vertices{result}
}

func (v Vertices) add(vertex string) {
	v.members[vertex] = true
}

func (v Vertices) has(vertex string) bool {
	_, ok := v.members[vertex]

	return ok
}

func (v Vertices) remove(vertex string) {
	delete(v.members, vertex)
}

func (v Vertices) without(other Vertices) Vertices {
	result := v

	for o := range other.members {
		result.remove(o)
	}

	return result
}

func (v Vertices) union(other Vertices) Vertices {
	result := v

	for o := range other.members {
		result.add(o)
	}

	return result
}

// AsStrings returns the vertices as a string slice
func (v Vertices) AsStrings() []string {
	vertices := make([]string, 0, len(v.members))

	for k := range v.members {
		vertices = append(vertices, k)
	}

	return vertices
}

// Graph is a DAG with string labeled vertices
type Graph struct {
	edges map[string]Vertices
}

// New creates a new Graph from a map of vertex to vertex label describing the edges
func NewGraph(edges map[string][]string) Graph {
	edgs := make(map[string]Vertices)

	for v, es := range edges {
		edgs[v] = vertices([]string{})

		for _, e := range es {
			edgs[v].add(e)
		}
	}

	return Graph{edgs}
}

// Children returns the vertices that are connected to given vertices with an edge
func (g Graph) Children(vs Vertices) Vertices {
	all := vertices([]string{})

	for vertex := range vs.members {
		grandchildren, found := g.edges[vertex]

		if found {
			all = all.union(grandchildren)
		}
	}

	return all
}

// Descendants returns all the vertices x for which a path to x exists from any of
// the vertices given
func (g Graph) Descendants(vertices Vertices) Vertices {
	descendants := g.Children(vertices)
	discovered := descendants

	for len(discovered.members) > 0 {
		grandchildren := g.Children(discovered)
		discovered = grandchildren.without(descendants)

		descendants = descendants.union(discovered)
	}

	return descendants
}

func (g Graph) Reverse() Graph {
	edges := make(map[string]Vertices)

	for v, es := range g.edges {
		for e := range es.members {
			_, ok := edges[e]
			if !ok {
				edges[e] = vertices([]string{})
			}

			edges[e].add(v)
		}
	}

	return Graph{edges}
}
