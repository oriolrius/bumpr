package version

import (
	"fmt"
	"strings"
)

type BumpType string

const (
	BumpMajor BumpType = "major"
	BumpMinor BumpType = "minor"
	BumpPatch BumpType = "patch"
)

func ParseBumpType(s string) (BumpType, error) {
	switch strings.ToLower(s) {
	case "major":
		return BumpMajor, nil
	case "minor":
		return BumpMinor, nil
	case "patch":
		return BumpPatch, nil
	default:
		return "", fmt.Errorf("invalid bump type: %s", s)
	}
}

func Bump(currentVersion string, bumpType BumpType) (string, error) {
	v, err := Parse(currentVersion)
	if err != nil {
		return "", err
	}

	switch bumpType {
	case BumpMajor:
		v.BumpMajor()
	case BumpMinor:
		v.BumpMinor()
	case BumpPatch:
		v.BumpPatch()
	default:
		return "", fmt.Errorf("invalid bump type: %s", bumpType)
	}

	return v.String(), nil
}

func BumpVersion(v *Version, bumpType BumpType) *Version {
	newVersion := &Version{
		Prefix: v.Prefix,
		Major:  v.Major,
		Minor:  v.Minor,
		Patch:  v.Patch,
	}

	switch bumpType {
	case BumpMajor:
		newVersion.BumpMajor()
	case BumpMinor:
		newVersion.BumpMinor()
	case BumpPatch:
		newVersion.BumpPatch()
	}

	return newVersion
}