package xstrings

import (
	"testing"
)

func TestHashStringsDeterminism(t *testing.T) {
	t.Parallel()

	const delimiter = "|"
	strs := []string{
		"a",
		"b",
		"c",
	}
	a := HashStrings(delimiter, strs...)
	b := HashStrings(delimiter, strs...)
	if a != b {
		t.Fatalf("determinism failed: HashStrings(%q, %v) returned %q then %q", delimiter, strs, a, b)
	}
}
