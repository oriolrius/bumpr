package sources

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGalaxySource_GetVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "version with double quotes",
			content: `namespace: my_namespace
name: my_collection
version: "1.2.3"
readme: README.md
authors:
  - Your Name
`,
			want:    "1.2.3",
			wantErr: false,
		},
		{
			name: "version with single quotes",
			content: `namespace: my_namespace
name: my_collection
version: '2.0.0'
readme: README.md
`,
			want:    "2.0.0",
			wantErr: false,
		},
		{
			name: "version without quotes",
			content: `namespace: my_namespace
name: my_collection
version: 0.1.0
readme: README.md
`,
			want:    "0.1.0",
			wantErr: false,
		},
		{
			name: "version with v prefix",
			content: `namespace: my_namespace
name: my_collection
version: "v1.0.0"
readme: README.md
`,
			want:    "v1.0.0",
			wantErr: false,
		},
		{
			name: "no version field",
			content: `namespace: my_namespace
name: my_collection
readme: README.md
`,
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid yaml",
			content: `namespace: my_namespace
name: my_collection
version: "1.0.0
readme: README.md
`,
			want:    "1.0.0",
			wantErr: false, // Regex should still find it
		},
	}

	g := NewGalaxySource()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "galaxy.yml")
			
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			got, err := g.GetVersion(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalaxySource_SetVersion(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		newVersion string
		wantContains string
	}{
		{
			name: "update version with double quotes",
			content: `namespace: my_namespace
name: my_collection
version: "1.2.3"
readme: README.md
`,
			newVersion:   "2.0.0",
			wantContains: `version: "2.0.0"`,
		},
		{
			name: "update version with single quotes",
			content: `namespace: my_namespace
name: my_collection
version: '1.2.3'
readme: README.md
`,
			newVersion:   "2.0.0",
			wantContains: `version: '2.0.0'`,
		},
		{
			name: "update version without quotes",
			content: `namespace: my_namespace
name: my_collection
version: 1.2.3
readme: README.md
`,
			newVersion:   "2.0.0",
			wantContains: `version: 2.0.0`,
		},
		{
			name: "update version with spaces",
			content: `namespace: my_namespace
name: my_collection
version:   "1.2.3"  
readme: README.md
`,
			newVersion:   "2.0.0",
			wantContains: `version:   "2.0.0"`,
		},
	}

	g := NewGalaxySource()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "galaxy.yml")
			
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			if err := g.SetVersion(tmpFile, tt.newVersion); err != nil {
				t.Errorf("SetVersion() error = %v", err)
				return
			}

			// Read the file and check if it contains the expected version
			updatedContent, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("failed to read updated file: %v", err)
			}

			if !contains(string(updatedContent), tt.wantContains) {
				t.Errorf("SetVersion() file does not contain %q\nGot:\n%s", tt.wantContains, string(updatedContent))
			}

			// Verify we can read the new version
			gotVersion, err := g.GetVersion(tmpFile)
			if err != nil {
				t.Errorf("GetVersion() after SetVersion() error = %v", err)
			}
			if gotVersion != tt.newVersion {
				t.Errorf("GetVersion() after SetVersion() = %v, want %v", gotVersion, tt.newVersion)
			}
		})
	}
}

func TestGalaxySource_Detect(t *testing.T) {
	g := NewGalaxySource()

	// Test when galaxy.yml exists
	tmpDir := t.TempDir()
	galaxyFile := filepath.Join(tmpDir, "galaxy.yml")
	
	if err := os.WriteFile(galaxyFile, []byte("version: 1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create galaxy.yml: %v", err)
	}

	if !g.Detect(tmpDir) {
		t.Error("Detect() = false, want true when galaxy.yml exists")
	}

	// Test when galaxy.yml doesn't exist
	emptyDir := t.TempDir()
	if g.Detect(emptyDir) {
		t.Error("Detect() = true, want false when galaxy.yml doesn't exist")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}