package version

import (
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
}

var versionRegex = regexp.MustCompile(`^(v)?(\d+)\.(\d+)\.(\d+)$`)

func Parse(versionStr string) (*Version, error) {
	matches := versionRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, _ := strconv.Atoi(matches[2])
	minor, _ := strconv.Atoi(matches[3])
	patch, _ := strconv.Atoi(matches[4])

	return &Version{
		Prefix: matches[1],
		Major:  major,
		Minor:  minor,
		Patch:  patch,
	}, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%s%d.%d.%d", v.Prefix, v.Major, v.Minor, v.Patch)
}

func (v *Version) BumpMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
}

func (v *Version) BumpMinor() {
	v.Minor++
	v.Patch = 0
}

func (v *Version) BumpPatch() {
	v.Patch++
}

func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

func (v *Version) IsPreRelease() bool {
	return false
}

func (v *Version) WithoutPrefix() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}