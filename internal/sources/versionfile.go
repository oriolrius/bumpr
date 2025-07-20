package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type VersionFileSource struct{}

func NewVersionFileSource() VersionSource {
	return &VersionFileSource{}
}

func (v *VersionFileSource) Name() string {
	return ".version"
}

func (v *VersionFileSource) GetDefaultFileName() string {
	return ".version"
}

func (v *VersionFileSource) Detect(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, ".version"))
	return err == nil
}

func (v *VersionFileSource) GetVersion(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Trim whitespace and newlines
	version := strings.TrimSpace(string(content))
	
	// Take only the first line if multiple lines exist
	if idx := strings.IndexAny(version, "\r\n"); idx != -1 {
		version = version[:idx]
	}

	if version == "" {
		return "", fmt.Errorf("version file is empty")
	}

	return version, nil
}

func (v *VersionFileSource) SetVersion(filePath string, newVersion string) error {
	// Write version with a newline at the end
	content := newVersion + "\n"
	
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}