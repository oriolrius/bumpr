# bumpr

A cross-platform CLI tool for automated software releases that handles version bumping and orchestrates git operations.

## Features

- 🚀 Automated version bumping (major/minor/patch)
- 📄 Multiple version source support:
  - `pyproject.toml` (Python projects)
  - `package.json` (Node.js projects)
  - `galaxy.yml` (Ansible roles and collections)
  - `.version` (plain text files)
- 🔍 Auto-detection of version source files
- 🏷️ Git tag creation and pushing
- 🔎 Pre-flight validation checks
- 🌐 Cross-platform (Linux/Windows)
- 🎯 Dry-run mode for safe testing
- 📢 Verbose mode for debugging

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/oriol/bumpr/releases).

### Build from Source

Requires Go 1.23 or later.

```bash
git clone https://github.com/oriol/bumpr.git
cd bumpr
make build
sudo make install
```

### Install to Project

For project-specific installation:

```bash
# Create bin directory in your project
mkdir -p bin

# Download or build bumpr
# ... download bumpr to current directory ...

# Move to project bin
mv bumpr bin/

# Add to .gitignore
echo "bin/" >> .gitignore
```

## Usage

### Basic Commands

```bash
# Bump patch version (1.0.0 → 1.0.1)
bumpr patch

# Bump minor version (1.0.0 → 1.1.0)
bumpr minor

# Bump major version (1.0.0 → 2.0.0)
bumpr major
```

### Options

```bash
# Preview changes without executing
bumpr patch --dry-run

# Specify version source file
bumpr minor --source package.json

# Show all executed commands
bumpr major --verbose

# Skip pushing to remote
bumpr patch --no-push

# Skip committing changes
bumpr patch --no-commit

# Suppress non-essential output
bumpr patch --quiet

# Skip safety checks
bumpr patch --force
```

### Version Command

```bash
# Show version information
bumpr version
```

## Version Source Files

### pyproject.toml

```toml
[project]
version = "1.0.0"

# or

[tool.poetry]
version = "1.0.0"
```

### package.json

```json
{
  "name": "my-project",
  "version": "1.0.0"
}
```

### .version

```
1.0.0
```

### galaxy.yml

```yaml
namespace: my_namespace
name: my_collection
version: "1.0.0"
readme: README.md
authors:
  - Your Name
```

## How It Works

1. **Pre-flight Checks**: Validates git is available, repository exists, and working directory is clean
2. **Version Detection**: Auto-detects or uses specified version source file
3. **Version Bumping**: Increments version according to semver rules
4. **File Update**: Updates the version in the source file
5. **Git Operations**: 
   - Stages the updated file
   - Creates a commit with message "releasing X.Y.Z"
   - Creates an annotated tag
   - Pushes the tag to origin

## Example Workflow

```bash
$ bumpr patch --verbose
🚀 Starting release process...

🔍 Running pre-flight checks...
✅ Pre-flight checks passed

📄 Using version source: pyproject.toml
📊 Current version: 1.2.3
📈 New version: 1.2.4

✅ Updated pyproject.toml with new version
→ git add pyproject.toml
→ git commit -m "releasing 1.2.4"
💾 Committed: releasing 1.2.4
🏷️  Created tag: 1.2.4
→ git push origin 1.2.4
📤 Pushed tag: 1.2.4

✅ Successfully released 1.2.4

Next steps:
1. Push the commit to origin: git push origin
2. Check GitHub Actions for the automated build
3. Once build completes, the release will be available at:
   https://github.com/USERNAME/REPO/releases/tag/1.2.4
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Development build with race detection
make dev
```

### Testing

```bash
# Run tests
make test

# Run linter
make lint
```

## License

MIT