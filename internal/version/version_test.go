package version

import (
	"regexp"
	"testing"
)

func TestStringIsStableSemanticVersion(t *testing.T) {
	if got := String(); !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(got) {
		t.Fatalf("String() = %q, want stable semantic version", got)
	}
}
