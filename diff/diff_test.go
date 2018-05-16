package diff

import (
	"reflect"
	"testing"

	"github.com/charypar/monobuild/manifests"
)

func TestBuildSchedule(t *testing.T) {
	exampleDependencies := map[string][]manifests.Dependency{
		"a": []manifests.Dependency{{Name: "b", Kind: manifests.Weak}, {Name: "c", Kind: manifests.Weak}},
		"b": []manifests.Dependency{{Name: "c", Kind: manifests.Weak}},
		"c": []manifests.Dependency{},
		"d": []manifests.Dependency{{Name: "a", Kind: manifests.Strong}},
		"e": []manifests.Dependency{{Name: "a", Kind: manifests.Strong}, {Name: "b", Kind: manifests.Strong}},
	}

	type args struct {
		changedComponents []string
		dependencies      map[string][]manifests.Dependency
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			"works with empty changes",
			args{
				[]string{""},
				exampleDependencies,
			},
			map[string][]string{},
		},
		{
			"collects affected strong dependencies",
			args{
				[]string{"a"},
				exampleDependencies,
			},
			map[string][]string{
				"a": []string{},
				"d": []string{"a"},
				"e": []string{"a"},
			},
		},
		{
			"collects all affected dependencies, keeps strong",
			args{
				[]string{"c"},
				exampleDependencies,
			},
			map[string][]string{
				"a": []string{},
				"b": []string{},
				"c": []string{},
				"d": []string{"a"},
				"e": []string{"a", "b"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildSchedule(tt.args.changedComponents, tt.args.dependencies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDependencies(t *testing.T) {
	exampleDependencies := map[string][]manifests.Dependency{
		"a": []manifests.Dependency{{Name: "b", Kind: manifests.Weak}, {Name: "c", Kind: manifests.Weak}},
		"b": []manifests.Dependency{{Name: "c", Kind: manifests.Weak}},
		"c": []manifests.Dependency{},
		"d": []manifests.Dependency{{Name: "a", Kind: manifests.Strong}},
		"e": []manifests.Dependency{{Name: "a", Kind: manifests.Strong}, {Name: "b", Kind: manifests.Strong}},
	}

	type args struct {
		changedComponents []string
		dependencies      map[string][]manifests.Dependency
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			"works with empty changes",
			args{
				[]string{""},
				exampleDependencies,
			},
			map[string][]string{},
		},
		{
			"collects affected strong dependencies",
			args{
				[]string{"a"},
				exampleDependencies,
			},
			map[string][]string{
				"a": []string{},
				"d": []string{"a"},
				"e": []string{"a"},
			},
		},
		{
			"collects all affected dependencies",
			args{
				[]string{"c"},
				exampleDependencies,
			},
			map[string][]string{
				"a": []string{"b", "c"},
				"b": []string{"c"},
				"c": []string{},
				"d": []string{"a"},
				"e": []string{"a", "b"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dependencies(tt.args.changedComponents, tt.args.dependencies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}
