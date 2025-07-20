package sources

import (
	"fmt"
	"os"
	"path/filepath"
)

type Detector struct {
	sources []VersionSource
}

func NewDetector() *Detector {
	return &Detector{
		sources: []VersionSource{
			NewPyProjectSource(),
			NewPackageJsonSource(),
			NewVersionFileSource(),
		},
	}
}

func (d *Detector) DetectSource(projectPath string) (VersionSource, string, error) {
	for _, source := range d.sources {
		if source.Detect(projectPath) {
			filePath := filepath.Join(projectPath, source.GetDefaultFileName())
			return source, filePath, nil
		}
	}

	return nil, "", fmt.Errorf("no version source file found in project")
}

func (d *Detector) GetSourceByFile(filePath string) (VersionSource, error) {
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	fileName := filepath.Base(filePath)
	
	for _, source := range d.sources {
		if fileName == source.GetDefaultFileName() {
			return source, nil
		}
	}

	// Try to infer from extension
	switch filepath.Ext(fileName) {
	case ".toml":
		if fileName == "pyproject.toml" {
			return NewPyProjectSource(), nil
		}
	case ".json":
		if fileName == "package.json" {
			return NewPackageJsonSource(), nil
		}
	}

	// Default to version file for any other file
	return NewVersionFileSource(), nil
}

func (d *Detector) ListAvailableSources() []string {
	names := make([]string, len(d.sources))
	for i, source := range d.sources {
		names[i] = source.Name()
	}
	return names
}