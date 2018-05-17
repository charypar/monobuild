package graph

// Edge represents a coloured edge of a directed graph
type Edge struct {
	Label  string
	Colour int
}

// Edges is an alias for []Edge with methods that allow
// the same operations as set.Set
type Edges []Edge

// Without removes all members of the other Edges from the Edges
func (edges Edges) Without(other Edges) Edges {
	result := make(Edges, 0, len(edges))

Outer:
	for _, e := range edges {
		for _, o := range other {
			if e.Label == o.Label {
				continue Outer // skip e
			}
		}

		result = append(result, e)
	}

	return result
}

// Union adds all the members in the other Edges to the Edges
func (edges Edges) Union(other Edges) Edges {
	result := edges

Outer:
	for _, o := range other {
		for _, e := range edges {
			if o.Label == e.Label {
				continue Outer
			}
		}

		result = append(result, o)
	}

	return result
}

// AsStrings returns the target vertices as a string slice
func (edges Edges) AsStrings() []string {
	result := make([]string, 0, len(edges))

	for _, k := range edges {
		result = append(result, k.Label)
	}

	return result
}
