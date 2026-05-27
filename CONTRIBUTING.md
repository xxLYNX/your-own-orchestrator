# Contributing to yoo

Thank you for your interest in contributing to **yoo** (Your Own Orchestrator)! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please be respectful and constructive in all interactions.

**Expected Behavior:**
- Be respectful of differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

- **Go**: Version 1.21 or higher
- **Git**: For version control
- **Make**: For build automation (optional but recommended)
- **A terminal emulator**: To test the TUI

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/your-own-orchestrator.git
   cd your-own-orchestrator
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/your-own-orchestrator.git
   ```

## Development Setup

### Install Dependencies

```bash
make deps
```

Or manually:
```bash
go mod download
go mod tidy
```

### Build the Project

```bash
make build
```

The binary will be created in `bin/yoo`.

### Run the Application

```bash
make run
```

Or directly:
```bash
./bin/yoo
```

### Development Workflow

For rapid development with auto-reload (requires [air](https://github.com/cosmtrek/air)):

```bash
go install github.com/cosmtrek/air@latest
make dev
```

## Project Structure

```
your-own-orchestrator/
├── cmd/                    # Command-line interface (Cobra commands)
│   ├── root.go            # Root command and initialization
│   ├── schedule.go        # Schedule query command
│   └── add.go             # Add note command
├── internal/              # Private application code
│   ├── config/            # Configuration management (Viper)
│   ├── database/          # Database layer (SQLite)
│   │   ├── db.go         # Connection and initialization
│   │   └── note.go       # Note model and operations
│   └── tui/              # Terminal UI (Bubble Tea)
│       └── schedule.go   # Schedule view model
├── docs/                  # Documentation (arc42)
│   └── architecture.md   # Architecture documentation
├── main.go               # Application entry point
├── go.mod                # Go module definition
└── Makefile              # Build automation
```

## Coding Standards

### Go Style

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Use `go vet` for static analysis (run `make vet`)
- Use meaningful variable and function names
- Write idiomatic Go code

### Code Organization

- **Keep functions small**: Aim for functions under 50 lines
- **Single Responsibility**: Each function/package should do one thing well
- **Error Handling**: Always check and handle errors explicitly
- **Comments**: Document exported functions, types, and complex logic

### Naming Conventions

- **Packages**: Short, lowercase, single-word names (e.g., `database`, `config`)
- **Files**: Lowercase with underscores if needed (e.g., `note.go`, `db_test.go`)
- **Interfaces**: Use `-er` suffix (e.g., `Reader`, `Writer`)
- **Constants**: Use camelCase or PascalCase based on visibility

### Example Code Style

```go
// Good: Clear, concise, well-documented
// GetNotesByDate retrieves all notes scheduled for a specific date.
// It returns notes sorted by scheduled time and priority.
func GetNotesByDate(db *sql.DB, date time.Time) ([]*Note, error) {
    startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
    endOfDay := startOfDay.Add(24 * time.Hour)
    
    query := `
        SELECT id, title, description, scheduled_at, status, priority, created_at, updated_at
        FROM notes
        WHERE scheduled_at >= ? AND scheduled_at < ?
        ORDER BY scheduled_at ASC, priority DESC
    `
    
    rows, err := db.Query(query, startOfDay, endOfDay)
    if err != nil {
        return nil, fmt.Errorf("query notes by date: %w", err)
    }
    defer rows.Close()
    
    // ... rest of implementation
}
```

## Making Changes

### Branch Naming

Create a descriptive branch name:
- `feature/add-recurring-notes` - For new features
- `fix/database-connection-leak` - For bug fixes
- `docs/update-architecture` - For documentation
- `refactor/simplify-tui-model` - For refactoring

### Commit Messages

Write clear, descriptive commit messages:

```
Short (50 chars or less) summary

More detailed explanatory text, if necessary. Wrap it to about 72
characters. The blank line separating the summary from the body is
critical.

- Bullet points are okay
- Use present tense ("Add feature" not "Added feature")
- Reference issues and pull requests when relevant

Fixes #123
```

**Good examples:**
- `Add support for recurring notes`
- `Fix database connection leak in schedule view`
- `Refactor TUI model to use interfaces`
- `Update architecture docs with ADR-005`

**Bad examples:**
- `fixed bug`
- `update`
- `WIP`

### Keep Changes Focused

- One logical change per commit
- One feature/fix per pull request
- Avoid mixing refactoring with new features

## Testing

### Writing Tests

We aim for >70% test coverage. Write tests for:
- Database operations (unit tests)
- Business logic (unit tests)
- Commands (integration tests)

### Test File Naming

- Place test files next to the code they test
- Use `_test.go` suffix (e.g., `note_test.go`)

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for a specific package
go test ./internal/database/...

# Run a specific test
go test -run TestGetNotesByDate ./internal/database/
```

### Example Test

```go
func TestGetNotesByDate(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
    
    // Create test data
    note := &Note{
        Title:       "Test Note",
        ScheduledAt: testDate,
        Status:      "pending",
    }
    err := CreateNote(db, note)
    require.NoError(t, err)
    
    // Test
    notes, err := GetNotesByDate(db, testDate)
    require.NoError(t, err)
    assert.Len(t, notes, 1)
    assert.Equal(t, "Test Note", notes[0].Title)
}
```

## Documentation

### Code Documentation

- Document all exported functions, types, and constants
- Use complete sentences
- Include examples for complex APIs

### Architecture Documentation

We follow the [arc42](https://arc42.org/) standard. When making significant architectural changes:

1. Update `docs/architecture.md`
2. Add an Architecture Decision Record (ADR) if appropriate
3. Update diagrams if needed

### User Documentation

Update `README.md` if your changes affect:
- Installation instructions
- Usage examples
- Configuration options
- Command-line flags

## Submitting Changes

### Before Submitting

1. **Format your code**: `make fmt`
2. **Run linters**: `make vet`
3. **Run tests**: `make test`
4. **Update documentation**: If you changed APIs or added features
5. **Test manually**: Run the application and verify your changes work

### Pull Request Process

1. **Update your fork**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes**:
   ```bash
   git push origin your-branch-name
   ```

3. **Create a pull request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Reference to related issues
   - Screenshots for UI changes

4. **Pull Request Template**:
   ```markdown
   ## Description
   Brief description of the changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## How Has This Been Tested?
   Describe the tests you ran
   
   ## Checklist
   - [ ] My code follows the style guidelines
   - [ ] I have performed a self-review
   - [ ] I have commented my code, particularly in hard-to-understand areas
   - [ ] I have updated the documentation
   - [ ] My changes generate no new warnings
   - [ ] I have added tests that prove my fix/feature works
   - [ ] New and existing unit tests pass locally
   ```

5. **Respond to feedback**: Be open to suggestions and make requested changes promptly

### Review Process

- Maintainers will review your PR within a few days
- Address feedback and push updates
- Once approved, a maintainer will merge your PR

## Development Tips

### Database Development

When working with the database:
- Use migrations for schema changes
- Test with both empty and populated databases
- Reset the test database: `make db-reset`

### TUI Development

When working on the TUI:
- Test in multiple terminal emulators
- Consider different terminal sizes
- Handle edge cases (empty lists, long text, etc.)

### Debugging

Use `fmt.Printf` for quick debugging, but remove before committing:
```go
// Temporary debugging - REMOVE BEFORE COMMIT
fmt.Printf("DEBUG: notes count = %d\n", len(notes))
```

For persistent debugging, use proper logging:
```go
log.Printf("loaded %d notes for date %s", len(notes), date.Format("2006-01-02"))
```

## Questions?

If you have questions:
1. Check the [architecture documentation](docs/architecture.md)
2. Search existing issues
3. Open a new issue with the "question" label

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to yoo! 🎉