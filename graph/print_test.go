package graph

import (
	"testing"
)

var exampleDependencies = New(map[string][]Edge{
	"a": []Edge{{Label: "c", Colour: Weak}, {Label: "b", Colour: Weak}},
	"b": []Edge{{Label: "c", Colour: Weak}},
	"c": []Edge{},
	"d": []Edge{{Label: "a", Colour: Strong}},
	"e": []Edge{{Label: "b", Colour: Strong}, {Label: "a", Colour: Strong}},
})

func TestText(t *testing.T) {
	tests := []struct {
		name      string
		graph     Graph
		selection []string
		showType  bool
		want      string
	}{
		{
			"prints an empty graph",
			exampleDependencies,
			[]string{},
			false,
			"",
		},
		{
			"prints a single node",
			exampleDependencies,
			[]string{"a"},
			false,
			"a: \n",
		},
		{
			"prints a single edge",
			exampleDependencies,
			[]string{"a", "b"},
			false,
			"a: b\nb: \n",
		},
		{
			"prints a fan",
			exampleDependencies,
			[]string{"a", "b", "c"},
			false,
			"a: b, c\nb: c\nc: \n",
		},
		{
			"prints a graph",
			exampleDependencies,
			[]string{"a", "b", "c", "d"},
			false,
			"a: b, c\nb: c\nc: \nd: a\n",
		},
		{
			"prints a graph with strength shown",
			exampleDependencies,
			[]string{"a", "b", "c", "d"},
			true,
			"a: b, c\nb: c\nc: \nd: !a\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.graph.Text(tt.selection, tt.showType); got != tt.want {
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
  "a"
}
`,
		},
		{
			"prints a single edge",
			exampleDependencies,
			[]string{"a", "b"},
			`digraph dependencies {
  "a" -> "b" [style=dashed]
  "b"
}
`,
		},
		{
			"prints a single strong edge",
			exampleDependencies,
			[]string{"a", "d"},
			`digraph dependencies {
  "a"
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
  "c"
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
