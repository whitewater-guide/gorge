package core

import (
	"fmt"
)

// StringSet is a helper for sets of gauge codes
type StringSet map[string]struct{}

// Contains checks if given code is contained in set
func (set StringSet) Contains(str string) bool {
	_, ok := set[str]
	return ok
}

// Only return the only code of set as string. If set contains not exactly 1 code, it returns error.
func (set StringSet) Only() (string, error) {
	l := len(set)
	if l != 1 {
		return "", fmt.Errorf("string set must contain exactly one item, but received '%s'", set.String())
	}
	for k := range set {
		return k, nil
	}
	return "", fmt.Errorf("string set must not be empty, but received '%s'", set.String())
}

// Slice converts StringSet to slice
func (set StringSet) Slice() []string {
	result := make([]string, len(set))
	i := 0
	for k := range set {
		result[i] = k
		i++
	}
	return result
}

func (set StringSet) String() string {
	result := ""
	for k := range set {
		result = result + "," + k
	}
	if len(result) > 0 {
		result = result[1:]
	}
	return result
}
