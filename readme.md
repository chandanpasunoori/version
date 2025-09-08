# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- **Interactive Module Selection**: Choose the module interactively using a beautiful terminal interface powered by bubbletea.
- **Traditional Text Input**: Fallback to traditional text input when needed.
- **Generate Next Version**: Automatically generate the next version based on the selected module and release type.
- **Create Git Tag**: Create a Git tag for the generated version.

## Installation

```bash
go install github.com/chandanpasunoori/version@latest
```

## Navigate to the project repository

```bash
cd repo
```

## Usage

### Interactive Mode (Recommended)

Use the `-i` flag to enable interactive mode with a beautiful list selection interface:

```bash
version -i
```

This will show a terminal UI where you can:
- Navigate through available modules using arrow keys
- Navigate through available release channels using arrow keys
- Press Enter to select an option
- Press 'q' to quit

### Traditional Mode

Run without the `-i` flag for traditional text input:

```bash
version
```

### Command Line Arguments

- `-m string`: Specify module name directly
- `-r string`: Specify release channel directly  
- `-i`: Enable interactive mode with bubbletea list selection

Examples:
```bash
# Interactive mode
version -i

# Direct specification
version -m myapp -r production

# Mixed mode (interactive for missing values)
version -i -m myapp
```

### Git Tag Format

```txt
<moduleName>/<releaseType>/v<major.minor.patch> = app/production/v0.1.1
```

### License

This project is licensed under the MIT License - see the LICENSE file for details.
