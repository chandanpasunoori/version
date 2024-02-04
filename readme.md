# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- Interactive Module Selection: Choose the module interactively.
- Generate Next Version: Automatically generate the next version based on the selected module and release type.
- Create Git Tag: Create a Git tag for the generated version.

## Installation

```bash
go install github.com/chandanpasunoori/version@latest
```

## Navigate to the project repository

```bash
cd repo
```

### Run the CLI

```bash
version

```

### Git Tag Format

```txt
<moduleName>/<releaseType>/v<major.minor.path> = app/production/v0.1.1
```

### License

This project is licensed under the MIT License - see the LICENSE file for details.
