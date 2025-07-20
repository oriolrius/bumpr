package version

import (
	"testing"
)

func TestParseBumpType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    BumpType
		wantErr bool
	}{
		{
			name:    "major lowercase",
			input:   "major",
			want:    BumpMajor,
			wantErr: false,
		},
		{
			name:    "major uppercase",
			input:   "MAJOR",
			want:    BumpMajor,
			wantErr: false,
		},
		{
			name:    "minor lowercase",
			input:   "minor",
			want:    BumpMinor,
			wantErr: false,
		},
		{
			name:    "patch lowercase",
			input:   "patch",
			want:    BumpPatch,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBumpType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBumpType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBumpType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBump(t *testing.T) {
	tests := []struct {
		name    string
		version string
		bump    BumpType
		want    string
		wantErr bool
	}{
		{
			name:    "bump patch without prefix",
			version: "1.2.3",
			bump:    BumpPatch,
			want:    "1.2.4",
			wantErr: false,
		},
		{
			name:    "bump minor without prefix",
			version: "1.2.3",
			bump:    BumpMinor,
			want:    "1.3.0",
			wantErr: false,
		},
		{
			name:    "bump major without prefix",
			version: "1.2.3",
			bump:    BumpMajor,
			want:    "2.0.0",
			wantErr: false,
		},
		{
			name:    "bump patch with v prefix",
			version: "v1.2.3",
			bump:    BumpPatch,
			want:    "v1.2.4",
			wantErr: false,
		},
		{
			name:    "bump minor with v prefix",
			version: "v1.2.3",
			bump:    BumpMinor,
			want:    "v1.3.0",
			wantErr: false,
		},
		{
			name:    "bump major with v prefix",
			version: "v1.2.3",
			bump:    BumpMajor,
			want:    "v2.0.0",
			wantErr: false,
		},
		{
			name:    "invalid version format",
			version: "invalid",
			bump:    BumpPatch,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Bump(tt.version, tt.bump)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bump() = %v, want %v", got, tt.want)
			}
		})
	}
}