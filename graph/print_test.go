package graph

import (
	"testing"
)

var exampleDependencies = New(map[string][]Edge{
	"a": []Edge{{Label: "b", Colour: Dashed}, {Label: "c", Colour: Dashed}},
	"b": []Edge{{Label: "c", Colour: Dashed}},
	"c": []Edge{},
	"d": []Edge{{Label: "a", Colour: Solid}},
	"e": []Edge{{Label: "a", Colour: Solid}, {Label: "b", Colour: Solid}},
})

func TestText(t *testing.T) {
	tests := []struct {
		name      string
		graph     Graph
		selection []string
		want      string
	}{
		{
			"prints an empty graph",
			exampleDependencies,
			[]string{},
			"",
		},
		{
			"prints a single node",
			exampleDependencies,
			[]string{"a"},
			"a: \n",
		},
		{
			"prints a single edge",
			exampleDependencies,
			[]string{"a", "b"},
			"a: b\nb: \n",
		},
		{
			"prints a fan",
			exampleDependencies,
			[]string{"a", "b", "c"},
			"a: b, c\nb: c\nc: \n",
		},
		{
			"prints a graph",
			exampleDependencies,
			[]string{"a", "b", "c", "d"},
			"a: b, c\nb: c\nc: \nd: a\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.graph.Text(tt.selection); got != tt.want {
				t.Errorf("Text() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDot(t *testing.T) {
	tests := []struct {
		name      string
		graph     Graph
		selection []string
		want      string
	}{
		{
			"prints an empty graph",
			exampleDependencies,
			[]string{},
			`digraph dependencies {
}
`,
		},
		{
			"prints a single node",
			exampleDependencies,
			[]string{"a"},
			`digraph dependencies {
  node "a"
}
`,
		},
		{
			"prints a single edge",
			exampleDependencies,
			[]string{"a", "b"},
			`digraph dependencies {
  "a" -> "b" [style=dashed]
  node "b"
}
`,
		},
		{
			"prints a single strong edge",
			exampleDependencies,
			[]string{"a", "d"},
			`digraph dependencies {
  node "a"
  "d" -> "a"
}
`,
		},
		{
			"prints a graph",
			exampleDependencies,
			[]string{"a", "b", "c", "d"},
			`digraph dependencies {
  "a" -> "b" [style=dashed]
  "a" -> "c" [style=dashed]
  "b" -> "c" [style=dashed]
  node "c"
  "d" -> "a"
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.graph.Dot(tt.selection); got != tt.want {
				t.Errorf("Dot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDotSchedule(t *testing.T) {
	tests := []struct {
		name      string
		graph     Graph
		selection []string
		want      string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.graph.DotSchedule(tt.selection); got != tt.want {
				t.Errorf("DotSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}
