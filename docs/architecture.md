# yoo - Architecture Documentation (arc42)

## 1. Introduction and Goals

### 1.1 Requirements Overview

**yoo** (Your Own Orchestrator) is a terminal-based task and schedule management tool designed to help users track their daily to-dos, tasks, reminders, and actions. The primary interface is a Terminal User Interface (TUI).

#### Core Features
- Query schedule (defaults to current day, supports date specification)
- Add notes (tasks/actions/reminders) to the schedule
- View and interact with schedule data via TUI
- Persistent local storage using SQLite

### 1.2 Quality Goals

| Priority | Quality Goal | Scenario |
|----------|-------------|----------|
| 1 | Usability | Users can quickly view today's schedule and add notes with minimal keystrokes |
| 2 | Performance | TUI responds instantly (<100ms) to user interactions |
| 3 | Reliability | Data integrity maintained; no data loss on crashes |
| 4 | Maintainability | Clear separation of concerns; easy to add new commands/features |
| 5 | Portability | Runs on Linux, macOS, and Windows |

### 1.3 Stakeholders

| Role | Goals | Expectations |
|------|-------|--------------|
| End Users | Efficiently manage daily tasks | Fast, intuitive TUI; reliable data storage |
| Developers | Maintainable codebase | Clean architecture; good documentation |
| Contributors | Easy onboarding | Clear structure; idiomatic Go code |

## 2. Constraints

### 2.1 Technical Constraints

| Constraint | Description |
|------------|-------------|
| Programming Language | Go 1.21+ |
| Database | SQLite (local, file-based) |
| TUI Framework | Bubble Tea |
| CLI Framework | Cobra |
| Configuration | Viper |
| Platform | Cross-platform (Linux, macOS, Windows) |

### 2.2 Organizational Constraints

| Constraint | Description |
|------------|-------------|
| Open Source | MIT or similar permissive license |
| Documentation Standard | arc42 template |
| Single Binary | No external dependencies at runtime |

## 3. Context and Scope

### 3.1 Business Context

```
┌─────────┐
│  User   │
└────┬────┘
     │ Commands (yoo, yoo add, yoo schedule)
     ▼
┌─────────────────┐
│   yoo CLI/TUI   │
└─────────────────┘
     │ SQL queries
     ▼
┌─────────────────┐
│  SQLite DB File │
│  (~/.yoo.db)    │
└─────────────────┘
```

**External Entities:**
- **User**: Interacts via terminal commands and TUI
- **SQLite Database**: Persistent storage on local filesystem

### 3.2 Technical Context

```
┌────────────────────────────────────────┐
│           Terminal Emulator            │
└────────────────┬───────────────────────┘
                 │ STDIN/STDOUT
┌────────────────▼───────────────────────┐
│          Cobra CLI Framework           │
│  ┌──────────────────────────────────┐  │
│  │   Command Routing & Parsing      │  │
│  └────────┬────────────┬────────────┘  │
│           │            │                │
│      ┌────▼────┐  ┌───▼─────┐         │
│      │  Query  │  │   Add   │         │
│      │ Command │  │ Command │         │
│      └────┬────┘  └───┬─────┘         │
└───────────┼───────────┼────────────────┘
            │           │
┌───────────▼───────────▼────────────────┐
│        Bubble Tea TUI Layer            │
│  ┌──────────────────────────────────┐  │
│  │  Model-Update-View Pattern       │  │
│  │  - Schedule View                 │  │
│  │  - Add Note Form                 │  │
│  └────────────┬─────────────────────┘  │
└───────────────┼────────────────────────┘
                │
┌───────────────▼────────────────────────┐
│      Database Abstraction Layer        │
│  ┌──────────────────────────────────┐  │
│  │  - Note Repository               │  │
│  │  - Schedule Queries              │  │
│  │  - Migrations                    │  │
│  └────────────┬─────────────────────┘  │
└───────────────┼────────────────────────┘
                │
┌───────────────▼────────────────────────┐
│     SQLite Database (modernc.org)      │
└────────────────────────────────────────┘
```

## 4. Solution Strategy

### 4.1 Technology Decisions

| Decision | Rationale |
|----------|-----------|
| **Go** | Cross-platform, single binary, excellent CLI tooling |
| **SQLite** | Lightweight, serverless, perfect for local-first applications |
| **Cobra** | Industry standard for Go CLI applications; excellent flag/command handling |
| **Bubble Tea** | Modern, composable TUI framework with clean architecture |
| **Viper** | Seamless integration with Cobra; supports multiple config formats |
| **modernc.org/sqlite** | Pure Go SQLite driver; no CGo dependencies |

### 4.2 Architectural Patterns

- **Layered Architecture**: Clear separation between CLI, TUI, business logic, and data access
- **Repository Pattern**: Database operations abstracted behind interfaces
- **MVC (Model-View-Controller)**: Bubble Tea's Model-Update-View pattern for TUI
- **Dependency Injection**: Database connections passed to handlers/models

### 4.3 Key Design Decisions

1. **Local-First**: All data stored locally; no cloud dependencies
2. **Date-Centric**: Notes are organized by date (schedule-based)
3. **Interactive TUI**: Primary interface is full-screen TUI, not just CLI output
4. **Configuration Flexibility**: Support config files and environment variables via Viper

## 5. Building Block View

### 5.1 Level 1: System Overview

```
┌─────────────────────────────────────────────┐
│              yoo Application                 │
│                                              │
│  ┌─────────┐  ┌─────────┐  ┌────────────┐  │
│  │   CLI   │  │   TUI   │  │  Database  │  │
│  │  Layer  │─▶│  Layer  │─▶│   Layer    │  │
│  └─────────┘  └─────────┘  └────────────┘  │
│       │                            ▲         │
│       └────────────────────────────┘         │
│         (Direct DB access for simple ops)    │
└─────────────────────────────────────────────┘
```

### 5.2 Level 2: Component Details

#### 5.2.1 CLI Layer (`cmd/`)
- **root.go**: Main command configuration, global flags, initialization
- **query.go**: Schedule query command implementation
- **add.go**: Note addition command implementation

**Responsibilities:**
- Parse command-line arguments and flags
- Initialize application configuration
- Route to appropriate TUI or direct operations
- Handle errors and exit codes

#### 5.2.2 TUI Layer (`internal/tui/`)
- **schedule.go**: Schedule view model and UI logic

**Responsibilities:**
- Render schedule views
- Handle keyboard input
- Update model state
- Coordinate with database layer

**Bubble Tea Components:**
- **Model**: Application state (current notes, selected date, UI state)
- **Update**: Handle messages (key presses, database responses)
- **View**: Render UI to terminal

#### 5.2.3 Database Layer (`internal/database/`)
- **db.go**: Database connection, initialization, migrations
- **note.go**: Note model and CRUD operations

**Responsibilities:**
- Manage SQLite connection lifecycle
- Execute queries and transactions
- Handle database migrations
- Provide repository interface for notes

#### 5.2.4 Configuration Layer (`internal/config/`)
- **config.go**: Viper configuration setup

**Responsibilities:**
- Load configuration from files and environment
- Provide default values
- Expose configuration to other components

### 5.3 Data Model

```sql
-- Notes Table
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    scheduled_at DATETIME NOT NULL,
    status TEXT DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_notes_scheduled_at ON notes(scheduled_at);
CREATE INDEX idx_notes_status ON notes(status);
```

## 6. Runtime View

### 6.1 Scenario: Launch TUI and View Today's Schedule

```
User                CLI               TUI              Database
 │                   │                 │                  │
 │─── yoo ─────────▶│                 │                  │
 │                   │                 │                  │
 │                   │─ Initialize ───▶│                  │
 │                   │                 │                  │
 │                   │                 │── Connect ──────▶│
 │                   │                 │                  │
 │                   │                 │◀─ DB Ready ──────│
 │                   │                 │                  │
 │                   │                 │── Query Today ──▶│
 │                   │                 │                  │
 │                   │                 │◀─ Notes ─────────│
 │                   │                 │                  │
 │                   │                 │─ Render UI ─────▶│
 │◀─ Display ───────────────────────────                  │
 │                   │                 │                  │
 │─── arrow keys ───▶│────────────────▶│                  │
 │                   │                 │─ Update View ───▶│
 │◀─ Updated UI ─────────────────────────                 │
```

### 6.2 Scenario: Add a Note

```
User                CLI               TUI              Database
 │                   │                 │                  │
 │─ yoo add "..." ──▶│                 │                  │
 │                   │                 │                  │
 │                   │─ Parse Args ───▶│                  │
 │                   │                 │                  │
 │                   │                 │── INSERT ────────▶│
 │                   │                 │                  │
 │                   │                 │◀─ Success ───────│
 │                   │                 │                  │
 │◀─ Confirmation ───│                 │                  │
```

## 7. Deployment View

### 7.1 Infrastructure

```
┌────────────────────────────────────┐
│        User's Local Machine        │
│                                    │
│  ┌──────────────────────────────┐ │
│  │    Terminal Emulator         │ │
│  │  (bash, zsh, PowerShell)     │ │
│  └───────────┬──────────────────┘ │
│              │                     │
│  ┌───────────▼──────────────────┐ │
│  │    yoo Binary                │ │
│  │  (Go executable)             │ │
│  └───────────┬──────────────────┘ │
│              │                     │
│  ┌───────────▼──────────────────┐ │
│  │  ~/.config/yoo/config.yaml   │ │
│  └──────────────────────────────┘ │
│              │                     │
│  ┌───────────▼──────────────────┐ │
│  │  ~/.local/share/yoo/yoo.db   │ │
│  │  (SQLite Database)           │ │
│  └──────────────────────────────┘ │
└────────────────────────────────────┘
```

### 7.2 File Locations

| OS | Config File | Database File |
|----|-------------|---------------|
| Linux | `~/.config/yoo/config.yaml` | `~/.local/share/yoo/yoo.db` |
| macOS | `~/Library/Application Support/yoo/config.yaml` | `~/Library/Application Support/yoo/yoo.db` |
| Windows | `%APPDATA%\yoo\config.yaml` | `%APPDATA%\yoo\yoo.db` |

## 8. Cross-cutting Concepts

### 8.1 Error Handling

- **Strategy**: Explicit error returns (Go idiomatic)
- **User Errors**: Clear messages via CLI/TUI
- **System Errors**: Logged with context; graceful degradation

### 8.2 Logging

- **Level**: INFO for normal operations, DEBUG for development
- **Output**: STDERR (doesn't interfere with TUI)
- **Format**: Structured logging (future: integrate with `log/slog`)

### 8.3 Configuration

**Hierarchy** (highest precedence first):
1. Command-line flags
2. Environment variables (`YOO_*`)
3. Config file (`config.yaml`)
4. Defaults

**Key Settings:**
- Database path
- Date format
- Default view options

### 8.4 Date Handling

- **Storage**: ISO 8601 format (`YYYY-MM-DD`)
- **Display**: User-configurable via config
- **Default**: Current date in user's local timezone

### 8.5 Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Enter` | Select/Edit |
| `n` | New note |
| `d` | Delete note |
| `Space` | Toggle complete |
| `q`/`Esc` | Quit/Back |
| `?` | Help |

## 9. Architecture Decisions

### ADR-001: Use Pure Go SQLite Driver

**Context**: Need SQLite but want to avoid CGo for easier cross-compilation.

**Decision**: Use `modernc.org/sqlite` instead of `github.com/mattn/go-sqlite3`.

**Consequences**:
- ✅ No CGo dependency
- ✅ Easier cross-compilation
- ✅ Simpler build process
- ⚠️ Slightly slower than CGo version (acceptable for this use case)

### ADR-002: Date-Centric Organization

**Context**: Notes/tasks need to be organized for schedule viewing.

**Decision**: Primary organization is by date; notes always have a `scheduled_date`.

**Consequences**:
- ✅ Natural fit for schedule view
- ✅ Simple querying by date range
- ⚠️ Recurring tasks need special handling (future feature)

### ADR-003: TUI as Primary Interface

**Context**: Need intuitive interface for viewing and managing schedule.

**Decision**: Use Bubble Tea for full-screen TUI as primary interface.

**Consequences**:
- ✅ Rich, interactive interface
- ✅ Better UX than plain CLI output
- ⚠️ Steeper learning curve for non-TUI users
- ⚠️ Testing more complex

### ADR-004: Local-First Architecture

**Context**: Data privacy and offline-first requirements.

**Decision**: All data stored locally in SQLite; no cloud sync.

**Consequences**:
- ✅ Complete data privacy
- ✅ Works offline
- ✅ Fast performance
- ⚠️ No multi-device sync (could be future feature)

## 10. Quality Requirements

### 10.1 Performance

| Requirement | Target | Measurement |
|-------------|--------|-------------|
| TUI Response Time | < 100ms | Time from keypress to screen update |
| Database Query | < 50ms | Time to fetch day's notes |
| Startup Time | < 500ms | Launch to TUI ready |

### 10.2 Usability

- Intuitive keyboard shortcuts
- Clear visual hierarchy
- Helpful error messages
- Online help accessible via `?`

### 10.3 Reliability

- Database transactions for data integrity
- Automatic database backup on major operations
- Graceful handling of corrupted database

### 10.4 Maintainability

- Test coverage > 70%
- All public APIs documented
- Clear separation of concerns
- Idiomatic Go code

## 11. Risks and Technical Debt

### 11.1 Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Database corruption | High | Regular backups; database integrity checks |
| Terminal compatibility issues | Medium | Test on multiple terminal emulators |
| Large dataset performance | Low | Implement pagination; optimize queries |

### 11.2 Technical Debt

- **Current**: No automated tests
- **Plan**: Add unit tests for database layer, integration tests for commands
- **Timeline**: Before v1.0 release

## 12. Glossary

| Term | Definition |
|------|------------|
| **Note** | A task, action, reminder, or to-do item in the schedule |
| **Schedule** | Date-organized view of notes |
| **TUI** | Terminal User Interface - full-screen interactive terminal application |
| **yoo** | "Your Own Orchestrator" - the name of this application |
| **Query** | Viewing/retrieving notes from the schedule |
| **Scheduled Date** | The date a note is assigned to appear in the schedule |