package version

import (
	"fmt"
	"strings"
)

func Validate(versionStr string) error {
	if versionStr == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if _, err := Parse(versionStr); err != nil {
		return err
	}

	return nil
}

func IsValidBumpType(bumpType string) bool {
	switch strings.ToLower(bumpType) {
	case "major", "minor", "patch":
		return true
	default:
		return false
	}
}

func NormalizeVersion(versionStr string) (string, error) {
	v, err := Parse(versionStr)
	if err != nil {
		return "", err
	}

	return v.String(), nil
}