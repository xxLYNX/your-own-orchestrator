# Records Table View Documentation

## Overview

The Records Table View is an interactive TUI (Text User Interface) component for displaying and managing template records in a table format. It provides a spreadsheet-like interface for viewing, editing, adding, and deleting records with keyboard navigation, filtering, and pagination support.

## Features

### Core Features

- **Table Display**: Shows records in a structured table format with columns matching the record schema
- **Keyboard Navigation**: Full keyboard-driven interface with vim-style bindings
- **Pagination**: Handles large datasets with configurable page size (default: 20 records per page)
- **Status Filtering**: Filter records by status (all, draft, in_progress, complete)
- **Search/Filter**: Full-text search across all record fields
- **CRUD Operations**:
  - Add new records with a form interface
  - Edit existing records
  - Delete records with confirmation dialog
- **Visual Feedback**: Color-coded status indicators and highlighted selections
- **Responsive Design**: Adapts to terminal size

### Status Indicators

Records can have the following statuses:
- `draft` - 📝 New or incomplete records (orange)
- `in_progress` - ▶ Records being worked on (blue)
- `complete` - ✓ Finished records (green)

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down one row |
| `k` / `↑` | Move cursor up one row |
| `h` / `←` / `PgUp` | Previous page |
| `l` / `→` / `PgDown` | Next page |
| `g` / `Home` | Jump to first record |
| `G` / `End` | Jump to last record |

### Actions

| Key | Action |
|-----|--------|
| `a` | Add new record |
| `e` / `Enter` | Edit selected record |
| `d` | Delete selected record (with confirmation) |
| `f` | Cycle through status filters |
| `/` | Open search/filter dialog |
| `c` | Clear all filters |
| `q` / `Esc` | Quit/return to previous view |

### Form Navigation (Add/Edit Mode)

| Key | Action |
|-----|--------|
| `↑` / `Shift+Tab` | Previous field |
| `↓` / `Tab` | Next field |
| `Enter` | Edit field / Submit form |
| `Esc` | Cancel and return to table |

### Delete Confirmation

| Key | Action |
|-----|--------|
| `y` / `Y` | Confirm deletion |
| `n` / `N` / `Esc` | Cancel deletion |

### Search Mode

| Key | Action |
|-----|--------|
| `Enter` | Apply search filter |
| `Esc` | Cancel search |
| `Backspace` | Delete character |
| Any character | Add to search query |

## Usage

### Basic Usage

```go
package main

import (
    "yoo/internal/models"
    "yoo/internal/tui"
)

func main() {
    // Define your record schema
    schema := &models.RecordSchema{
        Fields: []models.RecordField{
            {
                Name:     "date",
                Type:     "date",
                Required: true,
            },
            {
                Name:     "task",
                Type:     "text",
                Required: true,
            },
            {
                Name:     "hours",
                Type:     "integer",
                Required: true,
            },
        },
    }

    // Create or fetch your records
    records := []*models.TemplateRecord{
        // ... your records
    }

    // Launch the table view
    err := tui.ShowRecordsTable(noteID, records, schema)
    if err != nil {
        log.Fatal(err)
    }
}
```

### With Database Integration

```go
func showRecordsForNote(db *database.DB, noteID int64) error {
    // Fetch the note template
    noteTemplate, err := db.GetNoteTemplate(noteID)
    if err != nil {
        return err
    }

    // Get the record schema
    schema := noteTemplate.Template.Definition.RecordSchema
    if schema == nil {
        return fmt.Errorf("note template does not have a record schema")
    }

    // Fetch records
    records, err := db.GetTemplateRecords(noteID)
    if err != nil {
        return err
    }

    // Show the table
    return tui.ShowRecordsTable(noteID, records, schema)
}
```

## Record Schema

### Field Types

The record schema supports the following field types:

| Type | Description | Example |
|------|-------------|---------|
| `text` | Free-form text | "Write documentation" |
| `integer` | Whole numbers | 42 |
| `date` | ISO date format | "2024-01-15" |
| `enum` | Predefined values | "high", "medium", "low" |
| `url` | Web URLs | "https://example.com" |
| `boolean` | True/false values | true, false |

### Field Definition

```go
type RecordField struct {
    Name        string   // Field name (required)
    Type        string   // Field type (required)
    Description string   // Help text (optional)
    Required    bool     // Is field required?
    Default     string   // Default value (optional)
    Values      []string // Valid values for enum type
}
```

### Example Schemas

#### 1. Work Log

```go
schema := &models.RecordSchema{
    Fields: []models.RecordField{
        {
            Name:        "date",
            Type:        "date",
            Description: "Date of work",
            Required:    true,
        },
        {
            Name:        "task",
            Type:        "text",
            Description: "Task description",
            Required:    true,
        },
        {
            Name:        "hours",
            Type:        "integer",
            Description: "Hours worked",
            Required:    true,
            Default:     "8",
        },
        {
            Name:        "category",
            Type:        "enum",
            Description: "Task category",
            Required:    true,
            Values:      []string{"dev", "testing", "docs", "meetings"},
            Default:     "dev",
        },
    },
}
```

#### 2. Contact List

```go
schema := &models.RecordSchema{
    Fields: []models.RecordField{
        {Name: "name", Type: "text", Required: true},
        {Name: "email", Type: "text", Required: true},
        {Name: "phone", Type: "text", Required: false},
        {Name: "company", Type: "text", Required: false},
        {Name: "notes", Type: "text", Required: false},
    },
}
```

#### 3. Task Tracker

```go
schema := &models.RecordSchema{
    Fields: []models.RecordField{
        {
            Name:     "task",
            Type:     "text",
            Required: true,
        },
        {
            Name:     "priority",
            Type:     "enum",
            Values:   []string{"low", "medium", "high", "urgent"},
            Required: true,
            Default:  "medium",
        },
        {
            Name:     "due_date",
            Type:     "date",
            Required: false,
        },
        {
            Name:     "completed",
            Type:     "boolean",
            Required: true,
            Default:  "false",
        },
    },
}
```

#### 4. Expense Tracker

```go
schema := &models.RecordSchema{
    Fields: []models.RecordField{
        {Name: "date", Type: "date", Required: true},
        {Name: "description", Type: "text", Required: true},
        {Name: "amount", Type: "integer", Required: true},
        {
            Name:     "category",
            Type:     "enum",
            Values:   []string{"food", "transport", "utilities", "entertainment"},
            Required: true,
        },
        {Name: "receipt_url", Type: "url", Required: false},
    },
}
```

## Architecture

### Model Structure

```go
type RecordsTableModel struct {
    noteID        int64                        // Note/template ID
    records       []*models.TemplateRecord     // All records
    recordSchema  *models.RecordSchema         // Schema definition
    cursor        int                          // Current cursor position
    page          int                          // Current page number
    perPage       int                          // Records per page
    filter        string                       // Search filter text
    statusFilter  string                       // Status filter
    width, height int                          // Terminal dimensions
    viewMode      ViewMode                     // Current view mode
    editingRecord *models.TemplateRecord       // Record being edited
    formData      map[string]interface{}       // Form field data
    formCursor    int                          // Form field cursor
    searchInput   string                       // Search input buffer
    err           error                        // Last error
    quitting      bool                         // Is quitting?
}
```

### View Modes

The table operates in different view modes:

- `ViewModeTable` - Main table view
- `ViewModeAdd` - Add new record form
- `ViewModeEdit` - Edit existing record form
- `ViewModeDelete` - Delete confirmation dialog
- `ViewModeSearch` - Search/filter input

### Data Flow

```
┌─────────────────┐
│  User Input     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Update()       │ ◄── Handle key events
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Model State    │ ◄── Update cursor, filters, etc.
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  View()         │ ◄── Render current state
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Terminal       │
└─────────────────┘
```

## Integration Patterns

### Pattern 1: Standalone View

Launch the table as a standalone TUI application:

```go
func main() {
    records := loadRecords()
    schema := defineSchema()
    
    if err := tui.ShowRecordsTable(1, records, schema); err != nil {
        log.Fatal(err)
    }
}
```

### Pattern 2: As a Sub-View

Integrate into a larger TUI application:

```go
type MainModel struct {
    mode        string
    tableModel  tui.RecordsTableModel
    // ... other fields
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.mode {
        case "records":
            // Delegate to table model
            newModel, cmd := m.tableModel.Update(msg)
            m.tableModel = newModel.(tui.RecordsTableModel)
            return m, cmd
        }
    }
    return m, nil
}

func (m MainModel) View() string {
    switch m.mode {
    case "records":
        return m.tableModel.View()
    }
    return "..."
}
```

### Pattern 3: With Database Operations

Extend the model to persist changes:

```go
type RecordsTableWithDB struct {
    tui.RecordsTableModel
    db *database.DB
}

// Override methods to add database operations
func (m *RecordsTableWithDB) addRecord() {
    record := &models.TemplateRecord{
        NoteTemplateID: m.NoteID,
        Data:           m.FormData,
        Status:         "draft",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    
    if err := m.db.CreateTemplateRecord(record); err != nil {
        m.Err = err
        return
    }
    
    m.Records = append(m.Records, record)
}
```

### Pattern 4: With Custom Commands

Use Bubble Tea commands for async operations:

```go
type recordSavedMsg struct{ record *models.TemplateRecord }

func saveRecordCmd(db *database.DB, record *models.TemplateRecord) tea.Cmd {
    return func() tea.Msg {
        if err := db.SaveRecord(record); err != nil {
            return err
        }
        return recordSavedMsg{record}
    }
}

// In Update()
case tea.KeyMsg:
    if msg.String() == "s" {
        return m, saveRecordCmd(m.db, m.currentRecord)
    }

case recordSavedMsg:
    // Update UI to show success
```

## Styling

The records table uses styles from `internal/tui/styles.go` for consistency:

- `TitleWithBorderStyle` - Main title
- `TableHeaderStyle` - Column headers
- `TableRowStyle` - Regular rows
- `TableRowSelectedStyle` - Selected row
- `StatusDraftStyle` - Draft status badge
- `StatusInProgressStyle` - In-progress status badge
- `StatusCompletedStyle` - Complete status badge
- `HelpWithBorderStyle` - Help footer

### Customizing Colors

To customize colors, modify the color constants in `styles.go`:

```go
var (
    ColorPrimary   = lipgloss.Color("#7D56F4")  // Purple
    ColorSuccess   = lipgloss.Color("#04B575")  // Green
    ColorWarning   = lipgloss.Color("#F59E0B")  // Orange
    ColorError     = lipgloss.Color("#EF4444")  // Red
)
```

## Performance Considerations

### Memory Usage

- Records are held in memory during the session
- For very large datasets (>10,000 records), consider implementing lazy loading

### Pagination

- Default page size: 20 records
- Adjust `perPage` field for different performance characteristics
- Smaller pages = faster navigation but more page switching
- Larger pages = more scrolling but fewer page switches

### Filtering

- Filtering creates a new slice (copy) of filtered records
- Search performs string matching on all fields
- For complex filtering, consider adding indexed search

## Troubleshooting

### Common Issues

#### 1. Schema Not Displaying

**Problem**: Table shows "No schema defined for records."

**Solution**: Ensure the `RecordSchema` is not nil and has at least one field:

```go
if schema == nil || len(schema.Fields) == 0 {
    return fmt.Errorf("invalid schema")
}
```

#### 2. Navigation Not Working

**Problem**: Cursor doesn't move or moves incorrectly.

**Solution**: Check that records are properly loaded and pagination is calculated:

```go
// Verify records
fmt.Printf("Total records: %d\n", len(records))
fmt.Printf("Filtered records: %d\n", len(m.getFilteredRecords()))
```

#### 3. Forms Not Saving

**Problem**: Add/edit forms don't persist changes.

**Solution**: Current implementation saves in-memory. To persist, add database calls:

```go
// After creating/updating record
if err := db.SaveRecord(record); err != nil {
    return err
}
```

#### 4. Filters Not Working

**Problem**: Status or search filters show no results.

**Solution**: Verify field names match exactly and values are correct:

```go
// Check field names
for name := range record.Data {
    fmt.Printf("Field: %s\n", name)
}
```

## Future Enhancements

Potential improvements for future versions:

1. **Rich Field Editors**
   - Date picker for date fields
   - Dropdown for enum fields
   - Multi-line editor for text fields

2. **Sorting**
   - Click/key to sort by column
   - Multi-column sorting

3. **Export**
   - Export to CSV, JSON, or Markdown
   - Copy selected records

4. **Bulk Operations**
   - Multi-select with checkboxes
   - Bulk status updates
   - Bulk delete

5. **Advanced Filtering**
   - Column-specific filters
   - Date range filtering
   - Numeric comparisons (>, <, =)

6. **Validation**
   - Real-time field validation
   - Required field indicators
   - Format validation (email, URL, etc.)

7. **Undo/Redo**
   - Command history
   - Revert changes

8. **Templates**
   - Save common record templates
   - Quick fill from templates

## Examples

See `internal/tui/records_table_example.go` for complete working examples:

- `ExampleRecordsTable()` - Basic usage
- `ExampleRecordsTableWithDatabase()` - Database integration
- `ExampleRecordSchemaPatterns()` - Common schema patterns

## Related Documentation

- [Template System](./TEMPLATES.md) - Understanding template definitions
- [TUI Styles](./TUI_STYLES.md) - Customizing appearance
- [Database Schema](./DATABASE.md) - Database table structures
- [Keyboard Navigation](./KEYBOARD.md) - Complete keyboard reference

## License

This component is part of the Your Own Orchestrator (yoo) project.