package cli

import (
	"fmt"

	"github.com/charypar/monobuild/graph"
	"github.com/charypar/monobuild/set"
)

type filter struct {
	components set.Set
	filtered   set.Set
}

func newFilter(components []string, filtered []string) filter {
	componentsSet := set.New(components)
	var filteredSet set.Set
	if len(filtered) < 1 {
		filteredSet = componentsSet
	} else {
		filteredSet = set.New(filtered)
	}

	return filter{components: componentsSet, filtered: filteredSet}
}

func (f *filter) scopeTo(component string, dependencies graph.Graph) error {
	if !f.components.Has(component) {
		return fmt.Errorf("cannot scope to '%s', not a component", component)
	}

	scoped := set.New(append(dependencies.Descendants([]string{component}), component))
	f.filtered = f.filtered.Intersect(scoped)

	return nil
}

func (f *filter) onlyTop(dependencies graph.Graph) {
	// FIXME this algorithm probably belongs to graph
	reverse := dependencies.Reverse()
	vertices := dependencies.Vertices()

	topLevel := make([]string, 0, len(vertices))
	for i := range vertices {
		if len(reverse.Children(vertices[i:i+1])) < 1 {
			topLevel = append(topLevel, vertices[i])
		}
	}

	f.filtered = f.filtered.Intersect(set.New(topLevel))
}

func (f *filter) addStrong(buildSchedule graph.Graph) {
	strong := buildSchedule.Descendants(f.filtered.AsStrings())

	f.filtered = f.filtered.Union(set.New(strong))
}

func (f *filter) AsStrings() []string {
	return f.filtered.AsStrings()
}
