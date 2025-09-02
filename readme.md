# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- Interactive Module Selection: Choose the module interactively.
- Generate Next Version: Automatically generate the next version based on the selected module and release type.
- Create Git Tag: Create a Git tag for the generated version.
- Unique Build Identification: Each build has a unique identifier logged on application start.

## Installation

```bash
go install github.com/chandanpasunoori/version@latest
```

## Building from Source

To build with unique build information:

```bash
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_ID="build-$(date +%s)"
go build -ldflags "-X main.buildTime=$BUILD_TIME -X main.commitHash=$COMMIT_HASH -X main.buildID=$BUILD_ID" -o version .
```

This will inject unique build information that is displayed when the application starts.

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
