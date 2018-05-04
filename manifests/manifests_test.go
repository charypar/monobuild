package manifests

import (
	"fmt"
	"os"
	"reflect"
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

			got, got1, errs := Read(manifests, true)
			if (errs != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Read() got1 = %#v, want %#v", got1, tt.want1)
			}
		})
	}
}
