package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pelletier/go-toml/v2"
)

type PyProjectSource struct{}

func NewPyProjectSource() VersionSource {
	return &PyProjectSource{}
}

func (p *PyProjectSource) Name() string {
	return "pyproject.toml"
}

func (p *PyProjectSource) GetDefaultFileName() string {
	return "pyproject.toml"
}

func (p *PyProjectSource) Detect(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, "pyproject.toml"))
	return err == nil
}

func (p *PyProjectSource) GetVersion(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// First try simple regex match for common patterns
	patterns := []string{
		`version\s*=\s*["']([^"']+)["']`,
		`version\s*=\s*{[^}]*default\s*=\s*["']([^"']+)["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindSubmatch(content)
		if matches != nil {
			return string(matches[1]), nil
		}
	}

	// If regex fails, try TOML parsing
	var data map[string]interface{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return "", fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Check top-level version
	if version, ok := data["version"].(string); ok {
		return version, nil
	}

	// Check [project] section
	if project, ok := data["project"].(map[string]interface{}); ok {
		if version, ok := project["version"].(string); ok {
			return version, nil
		}
	}

	// Check [tool.poetry] section
	if tool, ok := data["tool"].(map[string]interface{}); ok {
		if poetry, ok := tool["poetry"].(map[string]interface{}); ok {
			if version, ok := poetry["version"].(string); ok {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version not found in pyproject.toml")
}

func (p *PyProjectSource) SetVersion(filePath string, newVersion string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	
	// Simple regex replacement to preserve formatting
	patterns := []struct {
		pattern     string
		replacement string
	}{
		{
			pattern:     `(version\s*=\s*["'])([^"']+)(["'])`,
			replacement: "${1}" + newVersion + "${3}",
		},
		{
			pattern:     `(version\s*=\s*{[^}]*default\s*=\s*["'])([^"']+)(["'][^}]*})`,
			replacement: "${1}" + newVersion + "${3}",
		},
	}

	replaced := false
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		if re.MatchString(contentStr) {
			contentStr = re.ReplaceAllString(contentStr, p.replacement)
			replaced = true
			break
		}
	}

	if !replaced {
		return fmt.Errorf("could not find version pattern to replace")
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}