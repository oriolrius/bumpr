package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type GalaxySource struct{}

func NewGalaxySource() VersionSource {
	return &GalaxySource{}
}

func (g *GalaxySource) Name() string {
	return "galaxy.yml"
}

func (g *GalaxySource) GetDefaultFileName() string {
	return "galaxy.yml"
}

func (g *GalaxySource) Detect(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, "galaxy.yml"))
	return err == nil
}

func (g *GalaxySource) GetVersion(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// First try simple regex match to handle quoted strings properly
	re := regexp.MustCompile(`(?m)^version:\s*["']?([^"'\s]+)["']?\s*$`)
	matches := re.FindSubmatch(content)
	if matches != nil {
		return string(matches[1]), nil
	}

	// If regex fails, try YAML parsing
	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Check for version field
	if version, ok := data["version"]; ok {
		switch v := version.(type) {
		case string:
			return v, nil
		case float64:
			// Handle cases where version might be parsed as float
			return fmt.Sprintf("%.1f", v), nil
		case int:
			// Handle cases where version might be parsed as int
			return fmt.Sprintf("%d.0.0", v), nil
		default:
			return "", fmt.Errorf("version field is not a string: %T", version)
		}
	}

	return "", fmt.Errorf("version field not found in galaxy.yml")
}

func (g *GalaxySource) SetVersion(filePath string, newVersion string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	
	// Use regex to replace version while preserving formatting
	// This handles various formats: version: "1.0.0", version: '1.0.0', version: 1.0.0
	re := regexp.MustCompile(`(?m)^(\s*version:\s*)["']?[^"'\s]+["']?(\s*)$`)
	
	if !re.MatchString(contentStr) {
		return fmt.Errorf("could not find version pattern to replace")
	}

	// Preserve quotes if they were present
	versionPattern := regexp.MustCompile(`(?m)^version:\s*(["']?)`)
	matches := versionPattern.FindStringSubmatch(contentStr)
	quote := ""
	if len(matches) > 1 {
		quote = matches[1]
	}

	// Replace the version
	contentStr = re.ReplaceAllString(contentStr, fmt.Sprintf("${1}%s%s%s${2}", quote, newVersion, quote))

	// Write back to file
	if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}