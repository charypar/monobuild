package diff

import (
	"reflect"
	"testing"
)

func TestGraph_Children(t *testing.T) {
	tests := []struct {
		name     string
		graph    Graph
		vertices Vertices
		want     Vertices
	}{
		{
			"fails on an empty graph",
			NewGraph(map[string][]string{}),
			vertices([]string{"foo"}),
			vertices([]string{}),
		},
		{
			"returns empty set for a single node graph",
			NewGraph(map[string][]string{"foo": []string{}}),
			vertices([]string{"foo"}),
			vertices([]string{}),
		},
		{
			"finds a single child of a single vertex",
			NewGraph(map[string][]string{"foo": []string{"bar"}}),
			vertices([]string{"foo"}),
			vertices([]string{"bar"}),
		},
		{
			"finds multiple children of a single vertex",
			NewGraph(map[string][]string{"foo": []string{"bar", "baz"}}),
			vertices([]string{"foo"}),
			vertices([]string{"bar", "baz"}),
		},
		{
			"finds multiple children of multiple vertices",
			NewGraph(map[string][]string{"a": []string{"b", "c"}, "b": []string{"c", "d"}}),
			vertices([]string{"a", "b"}),
			vertices([]string{"b", "c", "d"}),
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
		vertices Vertices
		want     Vertices
	}{
		{
			"returns empty set on an empty graph",
			NewGraph(map[string][]string{}),
			vertices([]string{"foo"}),
			vertices([]string{}),
		},
		{
			"returns empty set for a single node graph",
			NewGraph(map[string][]string{"foo": []string{}}),
			vertices([]string{"foo"}),
			vertices([]string{}),
		},
		{
			"finds a single child of a single vertex",
			NewGraph(map[string][]string{"foo": []string{"bar"}}),
			vertices([]string{"foo"}),
			vertices([]string{"bar"}),
		},
		{
			"finds multiple children of a single vertex",
			NewGraph(map[string][]string{"foo": []string{"bar", "baz"}}),
			vertices([]string{"foo"}),
			vertices([]string{"bar", "baz"}),
		},
		{
			"finds all descendants of a single vertex",
			NewGraph(map[string][]string{"a": []string{"b", "c"}, "b": []string{"c", "d"}}),
			vertices([]string{"a"}),
			vertices([]string{"b", "c", "d"}),
		},
		{
			"finds all descendants of a single vertex over several levels",
			NewGraph(map[string][]string{
				"a": []string{"b", "c"},
				"b": []string{"c", "d", "e"},
				"c": []string{"a", "d"},
				"d": []string{"b", "f"},
				"g": []string{"a", "b"}}),
			vertices([]string{"a"}),
			vertices([]string{"a", "b", "c", "d", "e", "f"}),
		},
		{
			"finds all descendants of multiple vertices in a complex graph",
			NewGraph(map[string][]string{
				"a": []string{"d", "e"},
				"b": []string{"f"},
				"c": []string{"h", "i"},
				"d": []string{"g"},
				"g": []string{"h"},
				"h": []string{"e"},
			}),
			vertices([]string{"a", "b"}),
			vertices([]string{"d", "e", "f", "g", "h"}),
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
			NewGraph(map[string][]string{}),
			NewGraph(map[string][]string{}),
		},
		{
			"reverses a single edge",
			NewGraph(map[string][]string{"a": []string{"b"}}),
			NewGraph(map[string][]string{"b": []string{"a"}}),
		},
		{
			"reverses a fan of edges",
			NewGraph(map[string][]string{"a": []string{"b", "c", "d"}}),
			NewGraph(map[string][]string{"b": []string{"a"}, "c": []string{"a"}, "d": []string{"a"}}),
		},
		{
			"reverses a complex graph",
			NewGraph(map[string][]string{"a": []string{"b", "c"}, "b": []string{"d"}, "c": []string{"d"}}),
			NewGraph(map[string][]string{"b": []string{"a"}, "c": []string{"a"}, "d": []string{"b", "c"}}),
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
