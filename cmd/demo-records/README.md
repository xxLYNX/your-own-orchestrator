# Records Table Demo

This is a standalone demo program for testing the interactive Records Table TUI component.

## Overview

The demo provides several pre-configured scenarios with sample data to showcase the Records Table features:

- **work-log** - Work time tracking with tasks, hours, and categories
- **contacts** - Contact list with names, emails, and company info
- **tasks** - Task tracker with priorities and due dates
- **expenses** - Expense tracker with amounts and categories

## Building

```bash
cd your-own-orchestrator
go build -o demo-records ./cmd/demo-records/
```

## Running

### Default scenario (work-log):
```bash
./demo-records
```

### Specific scenario:
```bash
./demo-records work-log
./demo-records contacts
./demo-records tasks
./demo-records expenses
```

## Features Demonstrated

### Navigation
- ✅ Cursor navigation with arrow keys and vim bindings (j/k)
- ✅ Page navigation (h/l, PgUp/PgDown)
- ✅ Jump to first/last record (g/G, Home/End)

### Filtering & Search
- ✅ Status filtering (press `f` to cycle)
- ✅ Full-text search (press `/`)
- ✅ Clear filters (press `c`)

### CRUD Operations
- ✅ Add new records (press `a`)
- ✅ Edit existing records (press `e` or Enter)
- ✅ Delete records (press `d` with confirmation)

### Display
- ✅ Table layout with columns
- ✅ Status indicators with colors
- ✅ Pagination (20 records per page)
- ✅ Highlighted selection
- ✅ Filter status display

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` / `PgUp` | Previous page |
| `l` / `→` / `PgDown` | Next page |
| `g` / `Home` | First record |
| `G` / `End` | Last record |
| `a` | Add new record |
| `e` / `Enter` | Edit record |
| `d` | Delete record |
| `f` | Filter by status |
| `/` | Search |
| `c` | Clear filters |
| `q` / `Esc` | Quit |

### In Add/Edit Form
| Key | Action |
|-----|--------|
| `↑` / `Shift+Tab` | Previous field |
| `↓` / `Tab` | Next field |
| `Enter` | Edit/Submit |
| `Esc` | Cancel |

## Scenarios

### Work Log
Tracks daily work activities with:
- Date
- Task description
- Hours worked
- Category (development, testing, documentation, meetings, other)
- Notes

Sample: 6 records with various statuses

### Contacts
Manages contact information:
- Name
- Email
- Phone (optional)
- Company (optional)
- Notes (optional)

Sample: 4 records

### Tasks
Task tracking system:
- Task description
- Priority (low, medium, high, urgent)
- Due date (optional)
- Assigned to (optional)
- Completed (boolean)

Sample: 3 records

### Expenses
Expense tracking:
- Date
- Description
- Amount
- Category (food, transport, utilities, entertainment, other)
- Receipt URL (optional)

Sample: 4 records

## Notes

- Changes are **not** persisted - this is a demo with in-memory data only
- The demo exits when you press `q` or `Esc` in the main table view
- Form editing is simplified - in a full implementation, each field would have a custom editor
- Database integration example is shown in `internal/tui/records_table_example.go`

## See Also

- [Records Table Documentation](../../docs/RECORDS_TABLE.md) - Complete documentation
- [Integration Example](../../internal/tui/records_table_example.go) - Code examples
- [TUI Styles](../../internal/tui/styles.go) - Style customization

## License

Part of the Your Own Orchestrator (yoo) project.