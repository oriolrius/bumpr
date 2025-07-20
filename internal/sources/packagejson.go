package sources

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PackageJsonSource struct{}

func NewPackageJsonSource() VersionSource {
	return &PackageJsonSource{}
}

func (p *PackageJsonSource) Name() string {
	return "package.json"
}

func (p *PackageJsonSource) GetDefaultFileName() string {
	return "package.json"
}

func (p *PackageJsonSource) Detect(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, "package.json"))
	return err == nil
}

func (p *PackageJsonSource) GetVersion(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	version, ok := data["version"].(string)
	if !ok {
		return "", fmt.Errorf("version field not found or not a string")
	}

	return version, nil
}

func (p *PackageJsonSource) SetVersion(filePath string, newVersion string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.UseNumber()
	
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	data["version"] = newVersion

	// Marshal with indentation to preserve formatting
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Add newline at end of file
	output = append(output, '\n')

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}