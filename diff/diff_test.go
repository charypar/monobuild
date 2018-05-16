package diff

import (
	"reflect"
	"testing"

	"github.com/charypar/monobuild/graph"
)

func Test_Impacted(t *testing.T) {
	exampleDependencies := graph.New(map[string][]graph.Edge{
		"a": []graph.Edge{{Label: "b", Colour: graph.Weak}, {Label: "c", Colour: graph.Weak}},
		"b": []graph.Edge{{Label: "c", Colour: graph.Weak}},
		"c": []graph.Edge{},
		"d": []graph.Edge{{Label: "a", Colour: graph.Strong}},
		"e": []graph.Edge{{Label: "a", Colour: graph.Strong}, {Label: "b", Colour: graph.Strong}},
	})

	type args struct {
		changedComponents []string
		dependencies      graph.Graph
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"works with empty changes",
			args{
				[]string{},
				exampleDependencies,
			},
			[]string{},
		},
		{
			"collects affected strong dependencies",
			args{
				[]string{"a"},
				exampleDependencies,
			},
			[]string{"a", "d", "e"},
		},
		{
			"collects all affected dependencies",
			args{
				[]string{"c"},
				exampleDependencies,
			},
			[]string{"a", "b", "c", "d", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Impacted(tt.args.changedComponents, tt.args.dependencies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}
