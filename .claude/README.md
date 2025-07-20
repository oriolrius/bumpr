# Go Release Tool Development Plan

## Project Overview

Create a cross-platform CLI tool in Go that automates the software release process. The tool will handle version bumping and orchestrate git and GitHub CLI operations by executing external commands and displaying them to the user.

## Reference Implementation

The current PowerShell script that this tool will replace:

```powershell
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet('major', 'minor', 'patch')]
    [string]$Type
)

function Get-CurrentVersion {
    $content = Get-Content "pyproject.toml" -Raw
    if ($content -match 'version\s*=\s*["\']([^"\']+)["\']') {
        return $matches[1]
    }
    throw "Could not find version in pyproject.toml"
}

function Set-NewVersion {
    param($newVersion)
    $content = Get-Content "pyproject.toml" -Raw
    $newContent = $content -replace 'version\s*=\s*["\'][^"\']+["\']', "version = `"$newVersion`""
    Set-Content "pyproject.toml" $newContent
}

function Bump-Version {
    param($currentVersion, $bumpType)
    
    $prefix = if ($currentVersion.StartsWith('v')) { 'v' } else { '' }
    $version = $currentVersion.TrimStart('v')
    $parts = $version.Split('.')
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]
    
    switch ($bumpType) {
        'major' { $major++; $minor = 0; $patch = 0 }
        'minor' { $minor++; $patch = 0 }
        'patch' { $patch++ }
    }
    
    return "$prefix$major.$minor.$patch"
}

# Main execution
try {
    $currentVersion = Get-CurrentVersion
    $newVersion = Bump-Version $currentVersion $Type
    
    Write-Host "Current version: $currentVersion"
    Write-Host "New version: $newVersion"
    
    Set-NewVersion $newVersion
    Write-Host "Updated pyproject.toml with new version"
    
    # Git operations
    git add pyproject.toml
    git commit -m "releasing $newVersion"
    Write-Host "Committed: releasing $newVersion"
    
    # Clean up existing tags
    git tag -d $newVersion 2>$null
    git push origin --delete $newVersion 2>$null
    Write-Host "Cleaned up existing tags..."
    
    # Create and push new tag
    git tag -a $newVersion -m "Release: $newVersion"
    git push origin $newVersion
    Write-Host "Created and pushed tag: $newVersion"
    
    Write-Host "✅ Successfully released $newVersion" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "1. Push the commit to origin: git push origin"
    Write-Host "2. Check GitHub Actions for the automated build"
    Write-Host "3. Once build completes, the release will be available at:"
    Write-Host "   https://github.com/USERNAME/REPO/releases/tag/$newVersion"
    
} catch {
    Write-Error "Release failed: $_"
    exit 1
}
```

## Project Goals

- **Primary**: Replace PowerShell release script with cross-platform Go binary
- **Portability**: Single executable for Linux and Windows (amd64)
- **External Dependencies**: Leverage git and GitHub CLI instead of implementing Git operations
- **Transparency**: Show users exactly what commands are being executed
- **Flexibility**: Support multiple version source files
- **Simplicity**: Focus on core functionality, avoid over-engineering

## Technical Requirements

### Core Functionality
1. **Version Management**
   - Parse semantic versions from configurable source files
   - Support pyproject.toml, package.json, and .version files
   - Support major/minor/patch bumping
   - Preserve version prefixes (v1.0.0 vs 1.0.0)
   - Validate version format compliance

2. **File Operations**
   - Read and modify version source files safely
   - Preserve file formatting and comments where possible
   - Support JSON, TOML, and plain text formats

3. **External Tool Integration**
   - Execute git commands and display them to user
   - Execute gh (GitHub CLI) commands and display them to user
   - Validate that required tools are available
   - Capture and display command output

4. **Error Handling**
   - Graceful failure with meaningful messages
   - Pre-flight checks (git/gh availability, clean working directory)
   - Detailed logging and command tracing

## External Dependencies

The tool requires these external tools to be installed and available in PATH:
- **git**: For version control operations
- **gh**: GitHub CLI for GitHub-specific operations (optional, for future enhancements)

## Project Structure

```
release-tool/
├── cmd/
│   └── root.go              # CLI command definitions
├── internal/
│   ├── version/
│   │   ├── parser.go        # Version parsing and manipulation
│   │   ├── bumper.go        # Version bumping logic
│   │   └── validator.go     # Version format validation
│   ├── sources/
│   │   ├── pyproject.go     # pyproject.toml handler
│   │   ├── packagejson.go   # package.json handler
│   │   ├── versionfile.go   # .version file handler
│   │   └── interface.go     # Version source interface
│   ├── external/
│   │   ├── git.go           # Git command execution
│   │   ├── github.go        # GitHub CLI execution
│   │   └── runner.go        # Command runner utilities
│   ├── config/
│   │   ├── settings.go      # Configuration management
│   │   └── detection.go     # Auto-detection logic
│   └── release/
│       └── orchestrator.go  # Main release workflow
├── test/
│   ├── fixtures/           # Test data and mock projects
│   ├── integration/        # Integration tests
│   └── unit/              # Unit tests
├── .github/
│   └── workflows/
│       ├── ci.yml         # Continuous integration
│       └── release.yml    # Release automation
├── scripts/
│   └── install.sh         # Installation helper script
├── docs/
│   ├── installation.md
│   ├── usage.md
│   └── version-sources.md
├── go.mod
├── go.sum
├── main.go               # Application entry point
├── Makefile             # Build automation
└── README.md
```

## Detailed Implementation Specifications

### 1. CLI Interface Design

**Command Structure:**
```bash
# Basic usage
release patch                    # Bump patch version
release minor                    # Bump minor version  
release major                    # Bump major version

# Advanced options
release patch --dry-run          # Preview changes without execution
release minor --source package.json # Specify version source file
release major --verbose          # Show all executed commands
release patch --no-push          # Skip pushing to remote
release --version               # Show tool version
release --help                  # Show help information
```

**Flag Specifications:**
- `--dry-run, -n`: Preview mode showing what would be done
- `--source, -s`: Specify version source file (auto-detect if not provided)
- `--verbose, -v`: Show all executed commands and their output
- `--no-push`: Skip pushing tags to remote repository
- `--no-commit`: Skip committing changes
- `--quiet, -q`: Suppress non-essential output
- `--force, -f`: Skip safety checks and confirmations

### 2. Version Source Handlers

**Common Interface:**
```go
type VersionSource interface {
    Name() string
    Detect(projectPath string) bool
    GetVersion(filePath string) (string, error)
    SetVersion(filePath string, newVersion string) error
    GetDefaultFileName() string
}
```

**PyProject.toml Handler:**
```go
type PyProjectSource struct{}

func (p *PyProjectSource) Name() string {
    return "pyproject.toml"
}

func (p *PyProjectSource) Detect(projectPath string) bool {
    _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml"))
    return err == nil
}

func (p *PyProjectSource) GetVersion(filePath string) (string, error) {
    // Parse TOML and extract version from:
    // version = "1.0.0"
    // OR
    // [tool.poetry]
    // version = "1.0.0"
}

func (p *PyProjectSource) SetVersion(filePath string, newVersion string) error {
    // Update version while preserving file structure and comments
}
```

**Package.json Handler:**
```go
type PackageJsonSource struct{}

func (p *PackageJsonSource) Name() string {
    return "package.json"
}

func (p *PackageJsonSource) GetVersion(filePath string) (string, error) {
    // Parse JSON and extract version field
}

func (p *PackageJsonSource) SetVersion(filePath string, newVersion string) error {
    // Update version while preserving JSON formatting
}
```

**Version File Handler:**
```go
type VersionFileSource struct{}

func (v *VersionFileSource) Name() string {
    return ".version"
}

func (v *VersionFileSource) GetVersion(filePath string) (string, error) {
    // Read version from plain text file (first line)
}

func (v *VersionFileSource) SetVersion(filePath string, newVersion string) error {
    // Write version to plain text file
}
```

### 3. External Command Integration

**Command Runner Interface:**
```go
type CommandRunner interface {
    Run(ctx context.Context, cmd string, args ...string) (*CommandResult, error)
    RunWithOutput(ctx context.Context, cmd string, args ...string) (*CommandResult, error)
}

type CommandResult struct {
    Command  string
    Args     []string
    ExitCode int
    Stdout   string
    Stderr   string
    Duration time.Duration
}

type GitCommands struct {
    runner CommandRunner
    verbose bool
}

func (g *GitCommands) Add(files ...string) error {
    args := append([]string{"add"}, files...)
    result, err := g.runner.Run(context.Background(), "git", args...)
    
    if g.verbose {
        fmt.Printf("→ git %s\n", strings.Join(args, " "))
        if result.Stdout != "" {
            fmt.Printf("  %s\n", result.Stdout)
        }
    }
    
    return err
}

func (g *GitCommands) Commit(message string) error {
    args := []string{"commit", "-m", message}
    result, err := g.runner.Run(context.Background(), "git", args...)
    
    if g.verbose {
        fmt.Printf("→ git %s\n", strings.Join(args, " "))
        if result.Stdout != "" {
            fmt.Printf("  %s\n", result.Stdout)
        }
    }
    
    return err
}

func (g *GitCommands) CreateTag(tagName, message string) error {
    args := []string{"tag", "-a", tagName, "-m", message}
    result, err := g.runner.Run(context.Background(), "git", args...)
    
    if g.verbose {
        fmt.Printf("→ git %s\n", strings.Join(args, " "))
    }
    
    return err
}

func (g *GitCommands) PushTag(tagName string) error {
    args := []string{"push", "origin", tagName}
    result, err := g.runner.Run(context.Background(), "git", args...)
    
    if g.verbose {
        fmt.Printf("→ git %s\n", strings.Join(args, " "))
        if result.Stdout != "" {
            fmt.Printf("  %s\n", result.Stdout)
        }
    }
    
    return err
}
```

### 4. Pre-flight Validation

**Dependency Checker:**
```go
type DependencyChecker struct {
    runner CommandRunner
}

func (d *DependencyChecker) CheckGit() error {
    _, err := d.runner.Run(context.Background(), "git", "--version")
    if err != nil {
        return fmt.Errorf("git is not available in PATH. Please install git")
    }
    return nil
}

func (d *DependencyChecker) CheckGitHub() error {
    _, err := d.runner.Run(context.Background(), "gh", "--version")
    if err != nil {
        return fmt.Errorf("gh (GitHub CLI) is not available in PATH. Install from: https://cli.github.com/")
    }
    return nil
}

func (d *DependencyChecker) CheckRepository() error {
    _, err := d.runner.Run(context.Background(), "git", "rev-parse", "--git-dir")
    if err != nil {
        return fmt.Errorf("not in a git repository")
    }
    return nil
}

func (d *DependencyChecker) CheckWorkingDirectory() error {
    result, err := d.runner.RunWithOutput(context.Background(), "git", "status", "--porcelain")
    if err != nil {
        return err
    }
    
    if strings.TrimSpace(result.Stdout) != "" {
        return fmt.Errorf("working directory is not clean. Please commit or stash changes")
    }
    
    return nil
}
```

### 5. Release Orchestration

**Main Release Flow:**
```go
type ReleaseOrchestrator struct {
    versionSources []VersionSource
    gitCommands    *GitCommands
    checker        *DependencyChecker
    config         *Config
}

func (r *ReleaseOrchestrator) Execute(bumpType BumpType, options ReleaseOptions) error {
    // 1. Pre-flight checks
    if err := r.runPreflightChecks(); err != nil {
        return fmt.Errorf("pre-flight check failed: %w", err)
    }
    
    // 2. Detect or use specified version source
    source, sourceFile, err := r.detectVersionSource(options.SourceFile)
    if err != nil {
        return err
    }
    
    fmt.Printf("Using version source: %s\n", sourceFile)
    
    // 3. Get current version
    currentVersion, err := source.GetVersion(sourceFile)
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }
    
    // 4. Calculate new version
    newVersion, err := r.bumpVersion(currentVersion, bumpType)
    if err != nil {
        return err
    }
    
    fmt.Printf("Current version: %s\n", currentVersion)
    fmt.Printf("New version: %s\n", newVersion)
    
    if options.DryRun {
        r.showDryRunCommands(sourceFile, newVersion)
        return nil
    }
    
    // 5. Update version file
    if err := source.SetVersion(sourceFile, newVersion); err != nil {
        return fmt.Errorf("failed to update version: %w", err)
    }
    
    fmt.Printf("Updated %s with new version\n", sourceFile)
    
    // 6. Git operations
    if !options.NoCommit {
        if err := r.gitCommands.Add(sourceFile); err != nil {
            return fmt.Errorf("failed to stage file: %w", err)
        }
        
        commitMessage := fmt.Sprintf("releasing %s", newVersion)
        if err := r.gitCommands.Commit(commitMessage); err != nil {
            return fmt.Errorf("failed to commit: %w", err)
        }
        
        fmt.Printf("Committed: %s\n", commitMessage)
    }
    
    // 7. Tag operations
    if err := r.cleanupExistingTag(newVersion); err != nil {
        fmt.Printf("Warning: failed to cleanup existing tag: %v\n", err)
    }
    
    tagMessage := fmt.Sprintf("Release: %s", newVersion)
    if err := r.gitCommands.CreateTag(newVersion, tagMessage); err != nil {
        return fmt.Errorf("failed to create tag: %w", err)
    }
    
    if !options.NoPush {
        if err := r.gitCommands.PushTag(newVersion); err != nil {
            return fmt.Errorf("failed to push tag: %w", err)
        }
        
        fmt.Printf("Created and pushed tag: %s\n", newVersion)
    }
    
    // 8. Success message and next steps
    r.showSuccessMessage(newVersion, options)
    
    return nil
}
```

### 6. Configuration System

**Configuration Structure:**
```go
type Config struct {
    VersionSource string `yaml:"version_source"` // File path or "auto"
    Git           GitConfig `yaml:"git"`
    Display       DisplayConfig `yaml:"display"`
}

type GitConfig struct {
    Remote         string `yaml:"remote"`         // Default: "origin"
    RequireClean   bool   `yaml:"require_clean"`  // Default: true
    AutoPush       bool   `yaml:"auto_push"`      // Default: true
    CommitTemplate string `yaml:"commit_template"` // Default: "releasing {version}"
    TagTemplate    string `yaml:"tag_template"`    // Default: "Release: {version}"
}

type DisplayConfig struct {
    Verbose    bool `yaml:"verbose"`     // Show all commands
    ShowSteps  bool `yaml:"show_steps"`  // Show step-by-step progress
    Quiet      bool `yaml:"quiet"`       // Minimal output
}
```

**Configuration File Example (.release.yml):**
```yaml
version_source: "auto"  # or specific file like "pyproject.toml"
git:
  remote: "origin"
  require_clean: true
  auto_push: true
  commit_template: "chore: releasing {version}"
  tag_template: "Release {version}"
display:
  verbose: false
  show_steps: true
  quiet: false
```

### 7. Build and Distribution

**Makefile:**
```makefile
BINARY_NAME := release-tool
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
BUILD_DIR := dist

.PHONY: build clean test lint

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for Linux and Windows
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Install locally
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Development build with race detection
dev:
	go build -race $(LDFLAGS) -o $(BINARY_NAME) .
```

### 8. GitHub Actions Workflows

**CI Workflow (.github/workflows/ci.yml):**
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v4
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: coverage.out
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

  build-test:
    runs-on: ubuntu-latest
    needs: test
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Test cross-platform builds
      run: make build-all
    
    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: binaries
        path: dist/
```

**Release Workflow (.github/workflows/release.yml):**
```yaml
name: Release

on:
  push:
    tags: ['v*']

env:
  BINARY_NAME: release-tool

jobs:
  build-and-release:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Get version from tag
      id: version
      run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
    
    - name: Build binaries
      run: |
        mkdir -p dist
        
        # Build for Linux
        GOOS=linux GOARCH=amd64 go build \
          -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
          -o dist/${{ env.BINARY_NAME }}-linux-amd64 .
        
        # Build for Windows
        GOOS=windows GOARCH=amd64 go build \
          -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
          -o dist/${{ env.BINARY_NAME }}-windows-amd64.exe .
    
    - name: Generate checksums
      run: |
        cd dist
        sha256sum * > checksums.txt
        cat checksums.txt
    
    - name: Create release notes
      id: release_notes
      run: |
        echo "RELEASE_NOTES<<EOF" >> $GITHUB_OUTPUT
        echo "## Release ${{ steps.version.outputs.VERSION }}" >> $GITHUB_OUTPUT
        echo "" >> $GITHUB_OUTPUT
        echo "### Downloads" >> $GITHUB_OUTPUT
        echo "- **Linux (amd64)**: [${{ env.BINARY_NAME }}-linux-amd64](https://github.com/${{ github.repository }}/releases/download/${{ steps.version.outputs.VERSION }}/${{ env.BINARY_NAME }}-linux-amd64)" >> $GITHUB_OUTPUT
        echo "- **Windows (amd64)**: [${{ env.BINARY_NAME }}-windows-amd64.exe](https://github.com/${{ github.repository }}/releases/download/${{ steps.version.outputs.VERSION }}/${{ env.BINARY_NAME }}-windows-amd64.exe)" >> $GITHUB_OUTPUT
        echo "" >> $GITHUB_OUTPUT
        echo "### Installation" >> $GITHUB_OUTPUT
        echo "\`\`\`bash" >> $GITHUB_OUTPUT
        echo "# Linux" >> $GITHUB_OUTPUT
        echo "curl -L https://github.com/${{ github.repository }}/releases/download/${{ steps.version.outputs.VERSION }}/${{ env.BINARY_NAME }}-linux-amd64 -o ${{ env.BINARY_NAME }}" >> $GITHUB_OUTPUT
        echo "chmod +x ${{ env.BINARY_NAME }}" >> $GITHUB_OUTPUT
        echo "" >> $GITHUB_OUTPUT
        echo "# Windows (PowerShell)" >> $GITHUB_OUTPUT
        echo "Invoke-WebRequest -Uri 'https://github.com/${{ github.repository }}/releases/download/${{ steps.version.outputs.VERSION }}/${{ env.BINARY_NAME }}-windows-amd64.exe' -OutFile '${{ env.BINARY_NAME }}.exe'" >> $GITHUB_OUTPUT
        echo "\`\`\`" >> $GITHUB_OUTPUT
        echo "EOF" >> $GITHUB_OUTPUT
    
    - name: Create GitHub Release
      uses: softprops/action-gh-release@v1
      with:
        name: Release ${{ steps.version.outputs.VERSION }}
        body: ${{ steps.release_notes.outputs.RELEASE_NOTES }}
        files: |
          dist/${{ env.BINARY_NAME }}-linux-amd64
          dist/${{ env.BINARY_NAME }}-windows-amd64.exe
          dist/checksums.txt
        draft: false
        prerelease: false
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Update latest tag
      run: |
        git tag -f latest
        git push origin latest --force
```

### 9. Dependencies

**go.mod:**
```go
module github.com/username/release-tool

go 1.21

require (
    github.com/spf13/cobra v1.8.0           // CLI framework
    github.com/pelletier/go-toml/v2 v2.1.1  // TOML parsing for pyproject.toml
    gopkg.in/yaml.v3 v3.0.1                 // YAML for configuration
    github.com/stretchr/testify v1.8.4       // Testing framework
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5               // indirect
)
```

### 10. Installation Instructions

**Manual Installation:**

Create an installation script (`scripts/install.sh`):
```bash
#!/bin/bash

set -e

REPO="username/release-tool"
BINARY_NAME="release-tool"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
fi

if [ "$OS" != "linux" ] && [ "$OS" != "windows" ]; then
    echo "Error: Unsupported OS: $OS"
    echo "Supported: linux, windows"
    exit 1
fi

if [ "$ARCH" != "amd64" ]; then
    echo "Error: Unsupported architecture: $ARCH"
    echo "Supported: amd64"
    exit 1
fi

# Construct download URL
BINARY_FILE="$BINARY_NAME-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
    BINARY_FILE="$BINARY_FILE.exe"
fi

# Get latest release
echo "Fetching latest release information..."
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
DOWNLOAD_URL=$(curl -s "$LATEST_URL" | grep "browser_download_url.*$BINARY_FILE" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $BINARY_FILE"
    exit 1
fi

echo "Downloading $BINARY_NAME..."
echo "URL: $DOWNLOAD_URL"

# Download binary
curl -L "$DOWNLOAD_URL" -o "$BINARY_NAME"
chmod +x "$BINARY_NAME"

echo "✅ Successfully downloaded $BINARY_NAME"
echo ""
echo "Installation options:"
echo "1. Move to system PATH:"
echo "   sudo mv $BINARY_NAME /usr/local/bin/"
echo ""
echo "2. Move to project bin/ directory:"
echo "   mkdir -p bin && mv $BINARY_NAME bin/"
echo "   # Add bin/ to your project's .gitignore"
echo ""
echo "3. Keep in current directory:"
echo "   ./$BINARY_NAME --help"
```

**Repository Structure Recommendations:**

```
your-project/
├── bin/                    # ✅ Recommended: Project-specific tools
│   ├── release-tool        # Downloaded binary
│   └── .gitignore         # Ignore binaries: *
├── tools/                  # ❌ Alternative: If you prefer
│   └── release-tool
├── scripts/                # ❌ Avoid: Usually for shell scripts
├── .local/                 # ❌ Alternative: Hidden directory
│   └── bin/
│       └── release-tool
└── src/                    # Your project code
```

**Recommended .gitignore entries:**
```gitignore
# Binaries
bin/
tools/release-tool
tools/release-tool.exe

# Or if using bin/ for version control
bin/*
!bin/.gitkeep
!bin/README.md
```

**Usage Instructions:**
```bash
# Download and install
curl -sSL https://raw.githubusercontent.com/username/release-tool/main/scripts/install.sh | bash

# Move to project bin directory (recommended)
mkdir -p bin
mv release-tool bin/

# Add to .gitignore
echo "bin/" >> .gitignore

# Use in project
bin/release-tool patch
bin/release-tool minor --verbose
bin/release-tool major --dry-run
```

## Development Phases

### Phase 1: Core Foundation
- [ ] Project structure setup with Go modules
- [ ] Basic CLI framework with cobra
- [ ] Version parsing and manipulation logic
- [ ] Command runner for external tools
- [ ] Unit tests for version operations

### Phase 2: Version Source Handlers
- [ ] PyProject.toml handler with TOML parsing
- [ ] Package.json handler with JSON manipulation
- [ ] .version file handler for plain text
- [ ] Auto-detection logic for version sources
- [ ] Integration tests for file operations

### Phase 3: External Tool Integration
- [ ] Git command wrappers with output display
- [ ] Dependency checking (git availability)
- [ ] Command execution with proper error handling
- [ ] Verbose mode for command tracing
- [ ] Pre-flight validation checks

### Phase 4: Release Orchestration
- [ ] Main release workflow implementation
- [ ] Dry-run mode for command preview
- [ ] Configuration file support
- [ ] Error handling and rollback capabilities
- [ ] End-to-end integration tests

### Phase 5: Build and CI/CD
- [ ] Cross-platform build setup (Linux/Windows amd64)
- [ ] GitHub Actions workflows for CI and releases
- [ ] Binary distribution via GitHub releases
- [ ] Installation documentation and scripts
- [ ] Release process validation

## Success Criteria

1. **Functionality**: Successfully replaces existing PowerShell script
2. **Portability**: Single binary runs on Linux and Windows (amd64) without dependencies
3. **Transparency**: Users can see exactly what git commands are executed
4. **Reliability**: Comprehensive error handling and pre-flight checks
5. **Usability**: Intuitive CLI with helpful error messages and dry-run mode
6. **Automation**: Fully automated CI/CD pipeline for binary distribution

## Example Usage Scenarios

**Scenario 1: Python project with pyproject.toml**
```bash
$ cd my-python-project
$ bin/release-tool patch --verbose

Using version source: pyproject.toml
Current version: v1.2.3
New version: v1.2.4
Updated pyproject.toml with new version
→ git add pyproject.toml
→ git commit -m "releasing v1.2.4"
[main abc1234] releasing v1.2.4
 1 file changed, 1 insertion(+), 1 deletion(-)
→ git tag -a v1.2.4 -m "Release: v1.2.4"
→ git push origin v1.2.4
To github.com:user/my-python-project.git
 * [new tag]         v1.2.4 -> v1.2.4
✅ Successfully released v1.2.4

Next steps:
1. Push commit: git push origin
2. Check GitHub release: https://github.com/user/my-python-project/releases/tag/v1.2.4
```

**Scenario 2: JavaScript project with package.json**
```bash
$ bin/release-tool minor --source package.json --dry-run

Using version source: package.json
Current version: 2.1.0
New version: 2.2.0

Dry run mode - commands that would be executed:
→ git add package.json
→ git commit -m "releasing 2.2.0"
→ git tag -a 2.2.0 -m "Release: 2.2.0"
→ git push origin 2.2.0

Run without --dry-run to execute these commands.
```

This comprehensive plan provides all the details needed to implement a robust, cross-platform release automation tool that leverages existing git tooling while providing transparency and reliability.
