package diff

import (
	"reflect"
	"testing"
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
