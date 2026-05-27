# yoo - Your Own Orchestrator

A terminal-based task and schedule management tool that helps you keep track of your daily to-dos with a clean TUI interface.

## Overview

`yoo` (Your Own Orchestrator) is a lightweight, fast command-line tool for managing your schedule and tasks. It stores your data in SQLite and provides an intuitive terminal user interface built with Bubble Tea.

## Features

- 📅 **Schedule Management**: View and manage your daily schedule
- 📝 **Notes**: Add tasks, reminders, and actions to your schedule
- 🔍 **Smart Queries**: Query specific dates or default to today
- 💾 **Local Storage**: All data stored locally in SQLite
- ⚡ **Fast & Lightweight**: No external dependencies or cloud services
- 🎯 **Templates** *(Coming Soon)*: Structure complex tasks with inputs, steps, and outputs

## Installation

### Prerequisites

- Go 1.21 or higher
- Git (for building from source)

### Option 1: Install from source

```bash
# Clone the repository
git clone https://github.com/yourusername/your-own-orchestrator.git
cd your-own-orchestrator

# Download dependencies
go mod download

# Build the binary
make build

# The binary will be in bin/yoo
# Optionally, install it to your $GOPATH/bin
make install
```

### Option 2: Build manually

```bash
git clone https://github.com/yourusername/your-own-orchestrator.git
cd your-own-orchestrator
go build -o yoo main.go

# Move to a location in your PATH
sudo mv yoo /usr/local/bin/
```

### Option 3: Install directly with Go

```bash
go install github.com/yourusername/your-own-orchestrator@latest
```

### Verify Installation

```bash
yoo --help
```

## Quick Start

### First Run

On first run, `yoo` will automatically create:
- Database file at `~/.local/share/yoo/yoo.db` (Linux/macOS)
- Config directory at `~/.config/yoo/` (if you create a config file)

### Basic Commands

#### View today's schedule
```bash
yoo schedule
```

#### View a specific date
```bash
yoo schedule --date 2024-01-15
yoo schedule -d 2024-12-25
```

#### Add a note to today
```bash
yoo add "Complete project documentation"
```

#### Add a note to a specific date
```bash
yoo add "Team meeting" --date 2024-01-20
```

#### Add a note with priority and tags
```bash
yoo add "Review PR" --priority high --tags work,urgent
```

#### Using Templates (Coming in v0.2.0+)

Templates help structure complex tasks with procedural steps:

```bash
# Apply to multiple jobs with structured workflow
yoo add "Apply to 10 jobs" --template job-applications \
  --input target_count=10 \
  --input resume=~/Documents/resume.pdf

# See available templates
yoo templates list

# View template details
yoo templates show job-applications
```

See `templates/README.md` and `docs/TEMPLATES_DESIGN.md` for details.

### TUI Keyboard Shortcuts

When in the schedule view:

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Enter` / `Space` | Toggle note completion |
| `a` / `n` | Add new note |
| `d` | Delete selected note |
| `q` / `Ctrl+C` | Quit |

## Configuration

Create a config file at `~/.config/yoo/config.yaml`:

```yaml
database:
  path: "~/.local/share/yoo/yoo.db"

format:
  date: "2006-01-02"  # ISO format
  time: "15:04"       # 24-hour format
```

See `config.example.yaml` for all available options.

### Environment Variables

You can also configure `yoo` using environment variables:

```bash
export YOO_DATABASE_PATH="$HOME/my-yoo.db"
export YOO_FORMAT_DATE="Jan 2, 2006"
```

## Development

### Building from source

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Build binary
make build

# Build for all platforms
make build-all
```

### Run in development mode

```bash
make run
```

### Reset database (for testing)

```bash
make db-reset
```

## Technology Stack

- **Language**: Go
- **Database**: SQLite
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Configuration**: [Viper](https://github.com/spf13/viper)

## Documentation

Full architecture and design documentation follows the [arc42](https://arc42.org/) standard and can be found in the `docs/` directory.

## Project Structure

```
your-own-orchestrator/
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command and initialization
│   ├── schedule.go        # Query schedule command
│   └── add.go             # Add note command
├── internal/              # Private application code
│   ├── database/          # SQLite database layer
│   │   ├── db.go         # Database initialization
│   │   └── note.go       # Note CRUD operations
│   ├── tui/              # Bubble Tea TUI components
│   │   └── schedule.go   # Schedule view model
│   └── config/           # Viper configuration
│       └── config.go     # Config initialization
├── docs/                 # arc42 documentation
│   └── architecture.md   # Full architecture docs
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── Makefile              # Build automation
└── config.example.yaml   # Example configuration
```

## Roadmap

### v0.2.0 - Template System (In Design)
- [ ] Note templates with inputs, steps, and outputs
- [ ] Built-in templates (job applications, research, projects)
- [ ] Template management CLI commands
- [ ] Step tracking and completion
- [ ] Artifact management (inputs/outputs)

See `docs/TEMPLATES_DESIGN.md` for detailed design.

### Future Releases
- Recurring notes/tasks
- Week/month calendar view
- Search and filtering
- Export/import functionality
- Reminders and notifications

## Contributing

Contributions are welcome! Please read the documentation in `docs/` for architecture decisions and coding standards.

## License

[Add your license here]