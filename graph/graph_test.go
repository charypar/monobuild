package graph

import (
	"reflect"
	"testing"

	"github.com/charypar/monobuild/set"
)

func TestGraph_Children(t *testing.T) {
	tests := []struct {
		name     string
		graph    Graph
		vertices set.Set
		want     set.Set
	}{
		{
			"fails on an empty graph",
			New(map[string][]string{}),
			set.New([]string{"foo"}),
			set.New([]string{}),
		},
		{
			"returns empty set for a single node graph",
			New(map[string][]string{"foo": []string{}}),
			set.New([]string{"foo"}),
			set.New([]string{}),
		},
		{
			"finds a single child of a single vertex",
			New(map[string][]string{"foo": []string{"bar"}}),
			set.New([]string{"foo"}),
			set.New([]string{"bar"}),
		},
		{
			"finds multiple children of a single vertex",
			New(map[string][]string{"foo": []string{"bar", "baz"}}),
			set.New([]string{"foo"}),
			set.New([]string{"bar", "baz"}),
		},
		{
			"finds multiple children of multiple vertices",
			New(map[string][]string{"a": []string{"b", "c"}, "b": []string{"c", "d"}}),
			set.New([]string{"a", "b"}),
			set.New([]string{"b", "c", "d"}),
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
		vertices set.Set
		want     set.Set
	}{
		{
			"returns empty set on an empty graph",
			New(map[string][]string{}),
			set.New([]string{"foo"}),
			set.New([]string{}),
		},
		{
			"returns empty set for a single node graph",
			New(map[string][]string{"foo": []string{}}),
			set.New([]string{"foo"}),
			set.New([]string{}),
		},
		{
			"finds a single child of a single vertex",
			New(map[string][]string{"foo": []string{"bar"}}),
			set.New([]string{"foo"}),
			set.New([]string{"bar"}),
		},
		{
			"finds multiple children of a single vertex",
			New(map[string][]string{"foo": []string{"bar", "baz"}}),
			set.New([]string{"foo"}),
			set.New([]string{"bar", "baz"}),
		},
		{
			"finds all descendants of a single vertex",
			New(map[string][]string{"a": []string{"b", "c"}, "b": []string{"c", "d"}}),
			set.New([]string{"a"}),
			set.New([]string{"b", "c", "d"}),
		},
		{
			"finds all descendants of a single vertex over several levels",
			New(map[string][]string{
				"a": []string{"b", "c"},
				"b": []string{"c", "d", "e"},
				"c": []string{"a", "d"},
				"d": []string{"b", "f"},
				"g": []string{"a", "b"}}),
			set.New([]string{"a"}),
			set.New([]string{"a", "b", "c", "d", "e", "f"}),
		},
		{
			"finds all descendants of multiple vertices in a complex graph",
			New(map[string][]string{
				"a": []string{"d", "e"},
				"b": []string{"f"},
				"c": []string{"h", "i"},
				"d": []string{"g"},
				"g": []string{"h"},
				"h": []string{"e"},
			}),
			set.New([]string{"a", "b"}),
			set.New([]string{"d", "e", "f", "g", "h"}),
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
			New(map[string][]string{}),
			New(map[string][]string{}),
		},
		{
			"reverses a single edge",
			New(map[string][]string{"a": []string{"b"}}),
			New(map[string][]string{"b": []string{"a"}}),
		},
		{
			"reverses a fan of edges",
			New(map[string][]string{"a": []string{"b", "c", "d"}}),
			New(map[string][]string{"b": []string{"a"}, "c": []string{"a"}, "d": []string{"a"}}),
		},
		{
			"reverses a complex graph",
			New(map[string][]string{"a": []string{"b", "c"}, "b": []string{"d"}, "c": []string{"d"}}),
			New(map[string][]string{"b": []string{"a"}, "c": []string{"a"}, "d": []string{"b", "c"}}),
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
