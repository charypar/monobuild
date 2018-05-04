package set

// Set is a set of strings
type Set struct {
	members map[string]bool
}

// New creates a new set from a slice
func New(names []string) Set {
	result := make(map[string]bool, len(names))

	for _, n := range names {
		result[n] = true
	}

	return Set{result}
}

// Add a string to a set
func (s Set) Add(label string) {
	s.members[label] = true
}

// Has checks existence of a member
func (s Set) Has(label string) bool {
	_, ok := s.members[label]

	return ok
}

// Remove removes a member from the set
func (s Set) Remove(label string) {
	delete(s.members, label)
}

// Size returns the cardinality (number of members) of the set
func (s Set) Size() int {
	return len(s.members)
}

// Without removes all members of the other set from the set
func (s Set) Without(other Set) Set {
	result := s

	for o := range other.members {
		result.Remove(o)
	}

	return result
}

// Union adds all the members in the other set to the set
func (s Set) Union(other Set) Set {
	result := s

	for o := range other.members {
		result.Add(o)
	}

	return result
}

// AsStrings returns the Set as a string slice
func (s Set) AsStrings() []string {
	Set := make([]string, 0, len(s.members))

	for k := range s.members {
		Set = append(Set, k)
	}

	return Set
}
