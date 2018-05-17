package graph

import (
	"reflect"
	"testing"
)

func TestGraph_Children(t *testing.T) {
	tests := []struct {
		name     string
		graph    Graph
		vertices []string
		want     []string
	}{
		{
			"returns empty on an empty graph",
			New(map[string][]Edge{}),
			[]string{"foo"},
			[]string{},
		},
		{
			"returns empty for a single node graph",
			New(map[string][]Edge{"foo": Edges{}}),
			[]string{"foo"},
			[]string{},
		},
		{
			"finds a single child of a single vertex",
			New(map[string][]Edge{"foo": Edges{{"bar", 0}}}),
			[]string{"foo"},
			[]string{"bar"},
		},
		{
			"finds multiple children of a single vertex",
			New(map[string][]Edge{"foo": Edges{{"bar", 0}, {"baz", 1}}}),
			[]string{"foo"},
			[]string{"bar", "baz"},
		},
		{
			"finds multiple children of multiple vertices",
			New(map[string][]Edge{"a": Edges{{"b", 0}, {"c", 0}}, "b": Edges{{"c", 1}, {"d", 0}}}),
			[]string{"a", "b"},
			[]string{"b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.graph.Children(tt.vertices)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Graph.Children() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraph_Descendants(t *testing.T) {
	tests := []struct {
		name     string
		graph    Graph
		vertices []string
		want     []string
	}{
		{
			"returns empty set on an empty graph",
			New(map[string][]Edge{}),
			[]string{"foo"},
			[]string{},
		},
		{
			"returns empty set for a single node graph",
			New(map[string][]Edge{"foo": []Edge{}}),
			[]string{"foo"},
			[]string{},
		},
		{
			"finds a single child of a single vertex",
			New(map[string][]Edge{"foo": []Edge{{"bar", 0}}}),
			[]string{"foo"},
			[]string{"bar"},
		},
		{
			"finds multiple children of a single vertex",
			New(map[string][]Edge{"foo": []Edge{{"bar", 0}, {"baz", 0}}}),
			[]string{"foo"},
			[]string{"bar", "baz"},
		},
		{
			"finds all descendants of a single vertex",
			New(map[string][]Edge{"a": []Edge{{"b", 0}, {"c", 1}}, "b": []Edge{{"c", 1}, {"d", 0}}}),
			[]string{"a"},
			[]string{"b", "c", "d"},
		},
		{
			"finds all descendants of a single vertex over several levels",
			New(map[string][]Edge{
				"a": []Edge{{"b", 0}, {"c", 0}},
				"b": []Edge{{"c", 0}, {"d", 1}, {"e", 0}},
				"c": []Edge{{"a", 0}, {"d", 0}},
				"d": []Edge{{"b", 0}, {"f", 0}},
				"g": []Edge{{"a", 0}, {"b", 0}}}),
			[]string{"a"},
			[]string{"a", "b", "c", "d", "e", "f"},
		},
		{
			"finds all descendants of multiple vertices in a complex graph",
			New(map[string][]Edge{
				"a": []Edge{{"d", 0}, {"e", 0}},
				"b": []Edge{{"f", 0}},
				"c": []Edge{{"h", 0}, {"i", 0}},
				"d": []Edge{{"g", 1}},
				"g": []Edge{{"h", 1}},
				"h": []Edge{{"e", 0}},
			}),
			[]string{"a", "b"},
			[]string{"d", "e", "f", "g", "h"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.graph.Descendants(tt.vertices)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Graph.Children() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraph_Reverse(t *testing.T) {
	tests := []struct {
		name  string
		graph Graph
		want  Graph
	}{
		{
			"reverses an empty graph",
			New(map[string][]Edge{}),
			New(map[string][]Edge{}),
		},
		{
			"reverses a single edge",
			New(map[string][]Edge{"a": []Edge{{"b", 0}}}),
			New(map[string][]Edge{"a": []Edge{}, "b": []Edge{{"a", 0}}}),
		},
		{
			"reverses a fan of edges",
			New(map[string][]Edge{"a": []Edge{{"b", 0}, {"c", 0}, {"d", 0}}}),
			New(map[string][]Edge{"a": []Edge{}, "b": []Edge{{"a", 0}}, "c": []Edge{{"a", 0}}, "d": []Edge{{"a", 0}}}),
		},
		{
			"reverses a complex graph",
			New(map[string][]Edge{"a": []Edge{{"b", 0}, {"c", 0}}, "b": []Edge{{"d", 0}}, "c": []Edge{{"d", 0}}}),
			New(map[string][]Edge{"a": []Edge{}, "b": []Edge{{"a", 0}}, "c": []Edge{{"a", 0}}, "d": []Edge{{"b", 0}, {"c", 0}}}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.graph.Reverse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Graph.Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraph_Subgraph(t *testing.T) {
	tests := []struct {
		name  string
		graph Graph
		nodes []string
		want  Graph
	}{
		{
			"works with empty graph",
			New(map[string][]Edge{}),
			[]string{},
			New(map[string][]Edge{}),
		},
		{
			"works with empty selection",
			New(map[string][]Edge{
				"a": []Edge{{"b", 0}, {"c", 0}},
				"b": []Edge{{"c", 0}},
				"c": []Edge{},
			}),
			[]string{},
			New(map[string][]Edge{}),
		},
		{
			"works with a selection",
			New(map[string][]Edge{
				"a": []Edge{{"b", 0}, {"c", 0}},
				"b": []Edge{{"c", 0}},
				"c": []Edge{},
			}),
			[]string{"b", "c"},
			New(map[string][]Edge{
				"b": []Edge{{"c", 0}},
				"c": []Edge{},
			}),
		},
		{
			"works on a larger graph",
			New(map[string][]Edge{
				"a": []Edge{{"b", 0}, {"c", 0}},
				"b": []Edge{{"d", 0}},
				"c": []Edge{{"d", 0}},
				"d": []Edge{{"a", 0}},
			}),
			[]string{"a", "c"},
			New(map[string][]Edge{
				"a": []Edge{{"c", 0}},
				"c": []Edge{},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.graph.Subgraph(tt.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Graph.Subgraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name  string
		edges map[string][]Edge
		want  Graph
	}{
		{
			"creates an empty graph",
			map[string][]Edge{},
			Graph{edges: map[string]Edges{}},
		},
		{
			"normalises the graph adding nodes that don't have dependencies",
			map[string][]Edge{
				"a": []Edge{{"b", 0}},
			},
			Graph{edges: map[string]Edges{
				"a": []Edge{{"b", 0}},
				"b": []Edge{},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.edges); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
