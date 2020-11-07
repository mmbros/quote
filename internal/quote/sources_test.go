package quote

import (
	"testing"
)

func TestSources(t *testing.T) {
	// check the strings are alphabetically sorted
	names := Sources()
	for j := 1; j < len(names); j++ {
		if names[j] < names[j-1] {
			t.Fatalf("Names are not sorted! %v", names)
		}
	}
}
