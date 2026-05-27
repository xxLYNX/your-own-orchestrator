# Quick Start Guide - yoo

Get up and running with **yoo** in 5 minutes!

## Installation

### Step 1: Install Go (if not already installed)

Visit https://golang.org/dl/ and install Go 1.21 or higher.

Verify installation:
```bash
go version
```

### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/your-own-orchestrator.git
cd your-own-orchestrator

# Download dependencies
go mod download

# Build
go build -o yoo main.go

# Optional: Move to PATH
sudo mv yoo /usr/local/bin/
```

## First Steps

### 1. Add your first note

```bash
yoo add "Buy groceries"
```

Output:
```
✓ Added note for today: Buy groceries
```

### 2. View today's schedule

```bash
yoo schedule
```

You'll see an interactive TUI showing your notes for today.

### 3. Navigate the TUI

- **↑/↓** or **k/j** - Move cursor up/down
- **Space** or **Enter** - Mark note as complete
- **a** - Add a new note
- **d** - Delete selected note
- **q** - Quit

## Common Tasks

### Add a note to a specific date

```bash
yoo add "Team meeting" --date 2024-02-15
```

### Add a note with priority

```bash
yoo add "Important deadline" --priority high
```

### View a specific date

```bash
yoo schedule --date 2024-02-15
```

### Add a note with tags

```bash
yoo add "Review code" --tags work,urgent --priority high
```

## What's Next?

### Configuration (Optional)

Create `~/.config/yoo/config.yaml`:

```yaml
database:
  path: "~/.local/share/yoo/yoo.db"

format:
  date: "2006-01-02"
  time: "15:04"
```

### Get Help

```bash
yoo --help          # General help
yoo add --help      # Help for add command
yoo schedule --help # Help for schedule command
```

## Example Workflow

```bash
# Start your day
yoo schedule

# Add tasks as they come up
yoo add "Call client about project X"
yoo add "Review PR #123" --tags work
yoo add "Dentist appointment at 3pm" --date 2024-02-20

# Check tomorrow's schedule
yoo schedule --date 2024-02-16

# Add a high-priority task for tomorrow
yoo add "Prepare presentation" --date 2024-02-16 --priority high
```

## Troubleshooting

### "yoo: command not found"

Either:
- Use `./yoo` from the build directory
- Or move yoo to a directory in your PATH
- Or run `make install` to install globally

### Database location

By default, yoo stores data at:
- **Linux/macOS**: `~/.local/share/yoo/yoo.db`
- **Windows**: `%APPDATA%\yoo\yoo.db`

### Reset everything

```bash
rm -rf ~/.local/share/yoo/
rm -rf ~/.config/yoo/
```

Then run yoo again to start fresh.

## Need More Help?

- **Full documentation**: See `README.md`
- **Architecture details**: See `docs/architecture.md`
- **Contributing**: See `CONTRIBUTING.md`

---

**You're all set!** Start organizing your schedule with yoo. 🎉