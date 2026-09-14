// Package version exposes the application's canonical semantic version.
package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var embedded string

// String returns the release version without a leading "v".
func String() string {
	return strings.TrimSpace(embedded)
}
