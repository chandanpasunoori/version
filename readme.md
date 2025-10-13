# Version

Welcome to the Tag Generator CLI. This CLI helps you generate and manage version tags for your Git repository.

## Features

- **Interactive Module Selection**: Choose the module interactively using a beautiful terminal interface powered by bubbletea.
- **Commit Selection**: Tag specific commits by hash or select from the last 5 commits in interactive mode.
- **Multi-Selection for Release Channels**: Select multiple release channels at once in interactive mode using checkbox-style selection.
- **Traditional Text Input**: Fallback to traditional text input when needed.
- **Generate Next Version**: Automatically generate the next version based on the selected module and release type.
- **Create Git Tag**: Create a Git tag for the generated version.
- **Batch Tag Creation**: Create tags for multiple release channels simultaneously.

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
- **Select commit to tag** - Choose between current commit (HEAD) or select from the last 5 commits
- Navigate through available release channels using arrow keys
- **Select multiple release channels using space bar** (new feature!)
- Press Enter to confirm your selection(s)
- Press 'q' to quit

**Commit Selection:**
In interactive mode, you can choose which commit to tag:
- **Current commit (HEAD)**: Default option, tags the latest commit
- **Select from last 5 commits**: View and select from a list of the most recent 5 commits with their messages

**Multi-Selection for Release Channels:**
In interactive mode, you can now select multiple release channels at once. Use the space bar to toggle selection of individual release channels, and press Enter to confirm. This allows you to create tags for multiple releases simultaneously (e.g., dev, staging, and prod all at once).

### Traditional Mode

Run without the `-i` flag for traditional text input:

```bash
version
```

### Command Line Arguments

- `-m string`: Specify module name directly
- `-r string`: Specify release channel directly  
- `-i`: Enable interactive mode with bubbletea list selection
- `-c string`: Specify commit hash (short or full) to tag, defaults to HEAD if not specified

Examples:
```bash
# Interactive mode
version -i

# Interactive mode with multi-selection (select multiple release channels)
version -i

# Direct specification (single release)
version -m myapp -r production

# Direct specification (multiple releases)
version -m myapp -r dev,staging,prod

# Tag a specific commit (using commit hash)
version -m myapp -r production -c abc1234

# Tag a specific commit with full hash
version -m myapp -r production -c abc1234567890abcdef1234567890abcdef1234

# Mixed mode (interactive for missing values)
version -i -m myapp
```

### Git Tag Format

```txt
<moduleName>/<releaseType>/v<major.minor.patch> = app/production/v0.1.1
```

### License

This project is licensed under the MIT License - see the LICENSE file for details.
