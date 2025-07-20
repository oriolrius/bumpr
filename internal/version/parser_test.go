package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "version without prefix",
			input: "1.2.3",
			want: &Version{
				Prefix: "",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			wantErr: false,
		},
		{
			name:  "version with v prefix",
			input: "v1.2.3",
			want: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			wantErr: false,
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want: &Version{
				Prefix: "",
				Major:  0,
				Minor:  0,
				Patch:  0,
			},
			wantErr: false,
		},
		{
			name:    "invalid format - missing patch",
			input:   "1.2",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - extra components",
			input:   "1.2.3.4",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   "1.2.x",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !versionsEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name: "without prefix",
			version: &Version{
				Prefix: "",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			want: "1.2.3",
		},
		{
			name: "with v prefix",
			version: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			want: "v1.2.3",
		},
		{
			name: "zero version",
			version: &Version{
				Prefix: "",
				Major:  0,
				Minor:  0,
				Patch:  0,
			},
			want: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Bump(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		bumpFunc func(*Version)
		want     *Version
	}{
		{
			name: "bump patch",
			version: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			bumpFunc: (*Version).BumpPatch,
			want: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  4,
			},
		},
		{
			name: "bump minor",
			version: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			bumpFunc: (*Version).BumpMinor,
			want: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  3,
				Patch:  0,
			},
		},
		{
			name: "bump major",
			version: &Version{
				Prefix: "v",
				Major:  1,
				Minor:  2,
				Patch:  3,
			},
			bumpFunc: (*Version).BumpMajor,
			want: &Version{
				Prefix: "v",
				Major:  2,
				Minor:  0,
				Patch:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Prefix: tt.version.Prefix,
				Major:  tt.version.Major,
				Minor:  tt.version.Minor,
				Patch:  tt.version.Patch,
			}
			tt.bumpFunc(v)
			if !versionsEqual(v, tt.want) {
				t.Errorf("After bump = %v, want %v", v, tt.want)
			}
		})
	}
}

func versionsEqual(v1, v2 *Version) bool {
	if v1 == nil || v2 == nil {
		return v1 == v2
	}
	return v1.Prefix == v2.Prefix &&
		v1.Major == v2.Major &&
		v1.Minor == v2.Minor &&
		v1.Patch == v2.Patch
}