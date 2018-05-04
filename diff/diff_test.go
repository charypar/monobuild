package diff

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/bmatcuk/doublestar"
)

func Test_changedComponents(t *testing.T) {
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
			if got := changedComponents(tt.args.components, tt.args.changedFiles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("changedComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func chdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(fmt.Errorf("Error returning to current directory: %s", err))
	}
}

func Test_readManifests(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Error finding current directory: %s", err))
	}

	tests := []struct {
		name    string
		cwd     string
		pattern string
		want    []string
		want1   map[string][]string
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
			map[string][]string{
				"app1":      []string{"app1", "libs/lib1", "libs/lib2"},
				"app2":      []string{"app2", "libs/lib2", "libs/lib3"},
				"app3":      []string{"app3", "libs/lib3"},
				"app4":      []string{"app4"},
				"libs/lib1": []string{"libs/lib1", "libs/lib3"},
				"libs/lib2": []string{"libs/lib2", "libs/lib3"},
				"libs/lib3": []string{"libs/lib3"},
				"stack1":    []string{"stack1", "app1", "app2", "app3"},
			},
			false,
		},
		{
			"fails on a bad manifest",
			"../test/fixtures/bad-manifests",
			"**/Dependencies",
			nil,
			nil,
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

			got, got1, err := readManifests(manifests)
			if (err != nil) != tt.wantErr {
				t.Errorf("readManifests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readManifests() got = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("readManifests() got1 = %#v, want %#v", got1, tt.want1)
			}
		})
	}
}
