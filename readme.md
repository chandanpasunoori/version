# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- Interactive Module Selection: Choose the module interactively.
- Generate Next Version: Automatically generate the next version based on the selected module and release type.
- Create Git Tag: Create a Git tag for the generated version.
- Unique Build Identification: Each build has a unique identifier logged on application start.

## Installation

### Using Go Install
```bash
go install github.com/chandanpasunoori/version@latest
```

### Using Homebrew (macOS)
```bash
brew install chandanpasunoori/tap/version
```

### Manual Installation
Download the latest binary from the [releases page](https://github.com/chandanpasunoori/version/releases) for your platform.

## Building from Source

To build with unique build information:

```bash
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_ID="build-$(date +%s)"
go build -ldflags "-X main.buildTime=$BUILD_TIME -X main.commitHash=$COMMIT_HASH -X main.buildID=$BUILD_ID" -o version .
```

This will inject unique build information that is displayed when the application starts.

## Releases

This project automatically creates cross-platform binary releases using GitHub Actions. When a new version tag is pushed (e.g., `v1.2.3`), the release workflow:

1. Builds binaries for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64) 
   - Windows (amd64, arm64)

2. Creates a GitHub release with all binary artifacts

3. Automatically updates the Homebrew tap at [chandanpasunoori/homebrew-tap](https://github.com/chandanpasunoori/homebrew-tap)

All binaries include the unique build information injected at compile time.

### Setting up Automated Homebrew Releases

To enable automatic updates to the Homebrew tap, you need to set up a GitHub secret:

1. Create a Personal Access Token (PAT) in GitHub with `repo` scope for the `homebrew-tap` repository
2. Add it as a repository secret named `HOMEBREW_TAP_TOKEN` in this repository's settings

The release workflow will automatically update the Homebrew formula when a new tag is pushed.

## Navigate to the project repository

```bash
cd repo
```

### Run the CLI

```bash
version

```

The application will first log its unique build information, then proceed with the normal CLI functionality.

### Git Tag Format

```txt
<moduleName>/<releaseType>/v<major.minor.path> = app/production/v0.1.1
```

### License

This project is licensed under the MIT License - see the LICENSE file for details.
