// Package version records versioning information about this module.
package version

import (
	"fmt"
	"strings"
)

const (
	Major      = 0
	Minor      = 30
	Patch      = 0
	PreRelease = "alpha"
)

// String formats the version string for this module in semver format.
//
// Examples:
//	v1.20.1
//	v1.21.0-rc.1
func String() string {
	v := fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
	if PreRelease != "" {
		v += "-" + PreRelease

		var metadata string
		if strings.Contains(PreRelease, "devel") && metadata != "" {
			v += "+" + metadata
		}
	}
	return v
}
