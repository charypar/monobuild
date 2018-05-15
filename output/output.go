package output

import (
	"fmt"
	"strings"

	"github.com/charypar/monobuild/manifests"
)

func Text(graph map[string][]string) string {
	var result string

	for c, d := range graph {
		result += fmt.Sprintf("%s: %s\n", c, strings.Join(d, ", "))
	}

	return result
}

func Dot(dependencies map[string][]manifests.Dependency, graph map[string][]string) string {
	result := fmt.Sprintln("digraph dependencies {")

	for c, deps := range dependencies {
		for _, d := range deps {
			var format string

			if d.Kind == manifests.Weak {
				format = " [style=dashed]"
			}

			result += fmt.Sprintf("  \"%s\" -> \"%s\"%s\n", c, d.Name, format)
		}
	}

	return result + "}\n"
}

func DotSchedule(dependencies map[string][]manifests.Dependency, graph map[string][]string) string {
	result := fmt.Sprintln("digraph schedule {\n  rankdir=\"LR\"\n  node [shape=box]")

	for c, deps := range graph {
		if len(deps) < 1 {
			result += fmt.Sprintf("  \"%s\"\n", c)
		}

		for _, d := range deps {
			result += fmt.Sprintf("  \"%s\" -> \"%s\"\n", c, d)
		}
	}

	return result + "}\n"
}
