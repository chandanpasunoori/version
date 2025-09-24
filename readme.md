# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- **Interactive Module Selection**: Choose the module interactively using a beautiful terminal interface powered by bubbletea.
- **Multi-Selection for Modules**: Select multiple modules at once in interactive mode using checkbox-style selection.
- **Multi-Selection for Release Channels**: Select multiple release channels at once in interactive mode using checkbox-style selection.
- **Traditional Text Input**: Fallback to traditional text input when needed.
- **Generate Next Version**: Automatically generate the next version based on the selected module and release type.
- **Create Git Tag**: Create a Git tag for the generated version.
- **Batch Tag Creation**: Create tags for multiple modules and release channels simultaneously.

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
- **Select multiple modules using space bar** (new feature!)
- Navigate through available release channels using arrow keys
- **Select multiple release channels using space bar**
- Press Enter to confirm your selection(s)
- Press 'q' to quit

**Multi-Selection for Modules:**
In interactive mode, you can now select multiple modules at once. Use the space bar to toggle selection of individual modules, and press Enter to confirm. This allows you to create tags for multiple modules simultaneously (e.g., frontend, backend, and api all at once).

**Multi-Selection for Release Channels:**
In interactive mode, you can also select multiple release channels at once. Use the space bar to toggle selection of individual release channels, and press Enter to confirm. This allows you to create tags for multiple releases simultaneously (e.g., dev, staging, and prod all at once).

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

# Interactive mode with multi-selection (select multiple modules and/or release channels)
version -i

# Direct specification (single module and release)
version -m myapp -r production

# Direct specification (multiple modules)
version -m frontend,backend,api -r production

# Direct specification (multiple releases)
version -m myapp -r dev,staging,prod

# Direct specification (multiple modules and releases)
version -m frontend,backend -r dev,staging,prod

# Mixed mode (interactive for missing values)
version -i -m myapp
```

### Git Tag Format

```txt
<moduleName>/<releaseType>/v<major.minor.patch> = app/production/v0.1.1
```

### License

This project is licensed under the MIT License - see the LICENSE file for details.
