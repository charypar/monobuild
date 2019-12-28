package manifests

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar"
)

func chdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(fmt.Errorf("Error returning to current directory: %s", err))
	}
}

func Test_Read(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Error finding current directory: %s", err))
	}

	tests := []struct {
		name    string
		cwd     string
		pattern string
		want    []string
		want1   Dependencies
		wantErr bool
	}{
		{
			"reads tests manifests correctly",
			"../test/fixtures/manifests-test",
			"**/Dependencies",
			[]string{
				"app1",
				"app2",
				"app3",
				"app4",
				"libs/lib1",
				"libs/lib2",
				"libs/lib3",
				"stack1",
			},
			Dependencies{deps: map[string][]Dependency{
				"app1":      []Dependency{{"app1", Weak}, {"libs/lib1", Weak}, {"libs/lib2", Weak}},
				"app2":      []Dependency{{"app2", Weak}, {"libs/lib2", Weak}, {"libs/lib3", Weak}},
				"app3":      []Dependency{{"app3", Weak}, {"libs/lib3", Weak}},
				"app4":      []Dependency{{"app4", Weak}},
				"libs/lib1": []Dependency{{"libs/lib1", Weak}, {"libs/lib3", Weak}},
				"libs/lib2": []Dependency{{"libs/lib2", Weak}, {"libs/lib3", Weak}},
				"libs/lib3": []Dependency{{"libs/lib3", Weak}},
				"stack1":    []Dependency{{"stack1", Weak}, {"app1", Strong}, {"app2", Strong}, {"app3", Strong}},
			}},
			false,
		},
		{
			"fails on a bad manifest",
			"../test/fixtures/bad-manifests",
			"**/Dependencies",
			nil,
			Dependencies{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chdir(tt.cwd)
			defer chdir(cwd)

			manifests, err := doublestar.Glob("**/Dependencies")
			if err != nil {
				panic(fmt.Errorf("Error finding dependency manifests: %s", err))
			}

			got, got1, errs := Read(manifests, true)
			if (errs != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			sort.Strings(got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Read() got1 = %#v, want %#v", got1, tt.want1)
			}
		})
	}
}

func Test_ReadManifest(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []Dependency
		want2 []error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := ReadManifest(tt.args.path)
			if got != tt.want {
				t.Errorf("ReadManifest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ReadManifest() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("ReadManifest() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_FilterComponents(t *testing.T) {
	type args struct {
		components   []string
		changedFiles []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"works with nothing",
			args{[]string{}, []string{}},
			[]string{},
		},
		{
			"works with no components",
			args{[]string{}, []string{"some/file/there.txt"}},
			[]string{},
		},
		{
			"works with no changes",
			args{[]string{"component/one"}, []string{}},
			[]string{},
		},
		{
			"finds a changed component",
			args{[]string{"component/one", "another"}, []string{"component/one/file/two.txt"}},
			[]string{"component/one"},
		},
		{
			"handles multiple files in a component",
			args{
				[]string{"component/one", "another"},
				[]string{"component/one/file/one.txt", "component/one/file/two.txt"},
			},
			[]string{"component/one"},
		},
		{
			"only matches full component name",
			args{
				[]string{"a/component", "a/component-v2"},
				[]string{"a/component-v2/file.txt"},
			},
			[]string{"a/component-v2"},
		},
		{
			"changes outside of components result in no changes",
			args{
				[]string{"component/one", "component/two", "something-else"},
				[]string{".github/CODEOWNERS"},
			},
			[]string{},
		},
		{
			"handles a complex case correctly",
			args{
				[]string{
					"stack",
					"application-one",
					"application-two",
					"libraries/one",
					"libraries/two",
				},
				[]string{
					"stack/config.json",
					"application-one/src/public/index.js",
					"libraries/two/src/index.go",
				},
			},
			[]string{
				"stack",
				"application-one",
				"libraries/two",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterComponents(tt.args.components, tt.args.changedFiles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("changedComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func joinErrors(message string, errors []error) error {
	errstrings := make([]string, len(errors))
	for i, e := range errors {
		errstrings[i] = string(e.Error())
	}

	return fmt.Errorf("%s\n%s", message, strings.Join(errstrings, "\n"))
}

func TestReadRepoManifest(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		want     []string
		want1    Dependencies
		wantErrs bool
		errors   []error
	}{
		// TODO: Add test cases.
		{
			"Empty manifest",
			"",
			[]string{},
			Dependencies{deps: map[string][]Dependency{}},
			false,
			nil,
		},
		{
			"Single component",
			"lib1:",
			[]string{
				"lib1",
			},
			Dependencies{deps: map[string][]Dependency{
				"lib1": []Dependency{{"lib1", Weak}},
			}},
			false,
			nil,
		},
		{
			"Component with a dependency",
			"lib1: lib2\nlib2: ",
			[]string{"lib1", "lib2"},
			Dependencies{deps: map[string][]Dependency{
				"lib1": []Dependency{{"lib1", Weak}, {"lib2", Weak}},
				"lib2": []Dependency{{"lib2", Weak}},
			}},
			false,
			nil,
		},
		{
			"Component with multiple dependencies",
			"lib1: lib2, lib3\nlib2: \nlib3: ",
			[]string{"lib1", "lib2", "lib3"},
			Dependencies{deps: map[string][]Dependency{
				"lib1": []Dependency{{"lib1", Weak}, {"lib2", Weak}, {"lib3", Weak}},
				"lib2": []Dependency{{"lib2", Weak}},
				"lib3": []Dependency{{"lib3", Weak}},
			}},
			false,
			nil,
		},
		{
			"Complex manifest",
			"# comment\napp1: lib1, lib2, lib3\napp2: \nlib1: \nlib2: lib3\nlib3: \n\nstack1: !app1, !app2",
			[]string{"app1", "app2", "lib1", "lib2", "lib3", "stack1"},
			Dependencies{deps: map[string][]Dependency{
				"app1":   []Dependency{{"app1", Weak}, {"lib1", Weak}, {"lib2", Weak}, {"lib3", Weak}},
				"app2":   []Dependency{{"app2", Weak}},
				"lib1":   []Dependency{{"lib1", Weak}},
				"lib2":   []Dependency{{"lib2", Weak}, {"lib3", Weak}},
				"lib3":   []Dependency{{"lib3", Weak}},
				"stack1": []Dependency{{"stack1", Weak}, {"app1", Strong}, {"app2", Strong}},
			}},
			false,
			nil,
		},
		{
			"Malformed line manifest",
			"# comment\napp1: lib1, lib2, lib3\nWHAT\napp2: \nlib1: \nlib2: lib3\nlib3: \n\nstack1: !app1, !app2",
			nil,
			Dependencies{deps: map[string][]Dependency{
				"app1":   []Dependency{{"app1", Weak}, {"lib1", Weak}, {"lib2", Weak}, {"lib3", Weak}},
				"app2":   []Dependency{{"app2", Weak}},
				"lib1":   []Dependency{{"lib1", Weak}},
				"lib2":   []Dependency{{"lib2", Weak}, {"lib3", Weak}},
				"lib3":   []Dependency{{"lib3", Weak}},
				"stack1": []Dependency{{"stack1", Weak}, {"app1", Strong}, {"app2", Strong}},
			}},
			true,
			[]error{fmt.Errorf("bad line format: 'WHAT' expected 'componnennt: dependency, dependency, ...'")},
		},
		{
			"Incomplete manifest",
			"# comment\napp1: lib1, lib2, lib3, unknown\n\napp2: \nlib1: \nlib2: lib3\nlib3: \n\nstack1: !app1, !app2",
			nil,
			Dependencies{deps: map[string][]Dependency{
				"app1":   []Dependency{{"app1", Weak}, {"lib1", Weak}, {"lib2", Weak}, {"lib3", Weak}, {"unknown", Weak}},
				"app2":   []Dependency{{"app2", Weak}},
				"lib1":   []Dependency{{"lib1", Weak}},
				"lib2":   []Dependency{{"lib2", Weak}, {"lib3", Weak}},
				"lib3":   []Dependency{{"lib3", Weak}},
				"stack1": []Dependency{{"stack1", Weak}, {"app1", Strong}, {"app2", Strong}},
			}},
			true,
			[]error{fmt.Errorf("unknown dependency 'unknown' of 'app1'")},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, gotErr := ReadRepoManifest(tt.manifest, true)

			if !tt.wantErrs && gotErr != nil {
				t.Errorf("ReadRepoManifest() received errors %#v", gotErr)
			}

			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadRepoManifest() got = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ReadRepoManifest() got1 = %#v, want %#v", got1, tt.want1)
			}
			if !reflect.DeepEqual(joinErrors("", gotErr), joinErrors("", tt.errors)) {
				t.Errorf("ReadRepoManifest() gotErr = %d, want %d", joinErrors("", gotErr), joinErrors("", tt.errors))
			}
		})
	}
}
