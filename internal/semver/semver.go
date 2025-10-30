// Package semver provides semantic versioning
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
}

// Parse parses a semantic version string
func Parse(version string) (*Version, error) {
	version = strings.TrimPrefix(version, "v")
	
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	v := &Version{}
	
	// Parse major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}
	v.Major = major
	
	// Parse minor version
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	v.Minor = minor
	
	// Parse patch and pre-release
	patchAndPre := strings.Split(parts[2], "-")
	if len(patchAndPre) > 1 {
		patch, err := strconv.Atoi(patchAndPre[0])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", patchAndPre[0])
		}
		v.Patch = patch
		v.Pre = strings.Join(patchAndPre[1:], "-")
	} else {
		v.Patch = 0
		v.Pre = ""
	}
	
	return v, nil
}