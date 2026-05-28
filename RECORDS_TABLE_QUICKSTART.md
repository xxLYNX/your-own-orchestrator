# Records Table Quick Start Guide

## What is it?

An interactive terminal UI for viewing and managing structured records in a spreadsheet-like table. Perfect for logs, task lists, contact databases, and any repeating data.

## Try it now (Demo)

```bash
cd your-own-orchestrator
go build -o demo-records ./cmd/demo-records/
./demo-records work-log
```

**Available demos:** `work-log`, `contacts`, `tasks`, `expenses`

## Basic Usage (3 steps)

### 1. Define your schema

```go
schema := &models.RecordSchema{
    Fields: []models.RecordField{
        {Name: "date", Type: "date", Required: true},
        {Name: "task", Type: "text", Required: true},
        {Name: "hours", Type: "integer", Required: true},
    },
}
```

### 2. Create or load records

```go
records := []*models.TemplateRecord{
    {
        ID:             1,
        NoteTemplateID: noteID,
        RecordIndex:    1,
        Data: map[string]interface{}{
            "date":  "2024-01-15",
            "task":  "Write documentation",
            "hours": 3,
        },
        Status:    "complete",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    },
}
```

### 3. Show the table

```go
import "yoo/internal/tui"

err := tui.ShowRecordsTable(noteID, records, schema)
```

That's it! 🎉

## Essential Keyboard Shortcuts

| Key | What it does |
|-----|--------------|
| `j`/`k` or `↑`/`↓` | Navigate up/down |
| `a` | Add new record |
| `e` or `Enter` | Edit selected record |
| `d` | Delete record (with confirmation) |
| `f` | Filter by status (cycles through all/draft/in_progress/complete) |
| `/` | Search records |
| `q` or `Esc` | Quit |

## Field Types

| Type | Example | Use for |
|------|---------|---------|
| `text` | "Hello world" | Any text |
| `integer` | 42 | Numbers |
| `date` | "2024-01-15" | Dates |
| `enum` | "high" | Fixed choices |
| `url` | "https://..." | Web links |
| `boolean` | true/false | Yes/No values |

## Common Patterns

### Daily Log
```go
{Name: "date", Type: "date", Required: true},
{Name: "entry", Type: "text", Required: true},
{Name: "mood", Type: "enum", Values: []string{"great", "good", "okay", "bad"}},
```

### Task List
```go
{Name: "task", Type: "text", Required: true},
{Name: "priority", Type: "enum", Values: []string{"low", "medium", "high"}},
{Name: "due_date", Type: "date", Required: false},
```

### Contact List
```go
{Name: "name", Type: "text", Required: true},
{Name: "email", Type: "text", Required: true},
{Name: "phone", Type: "text", Required: false},
```

### Expense Tracker
```go
{Name: "date", Type: "date", Required: true},
{Name: "description", Type: "text", Required: true},
{Name: "amount", Type: "integer", Required: true},
{Name: "category", Type: "enum", Values: []string{"food", "transport", "other"}},
```

## Status Workflow

Records have three built-in statuses:

- **draft** 📝 - Just created, incomplete
- **in_progress** ▶️ - Being worked on  
- **complete** ✅ - Finished

Use the status filter (`f` key) to view only records in a specific status.

## Integration with Database

The table currently works with in-memory data. To persist changes, extend the model:

```go
type RecordsTableWithDB struct {
    tui.RecordsTableModel
    db *database.DB
}

func (m *RecordsTableWithDB) addRecord() {
    // Create record in memory
    record := &models.TemplateRecord{...}
    
    // Save to database
    if err := m.db.CreateTemplateRecord(record); err != nil {
        m.Err = err
        return
    }
    
    // Update local list
    m.Records = append(m.Records, record)
}
```

Similar approach for `updateRecord()` and `deleteRecord()`.

## Tips & Tricks

1. **Large datasets?** Table supports 20 records per page. Use search (`/`) to find specific records.

2. **Required fields** are marked with `*` in the form view.

3. **Clear filters** anytime with `c` key to see all records again.

4. **Enum fields** should define valid values:
   ```go
   {Name: "status", Type: "enum", Values: []string{"todo", "doing", "done"}}
   ```

5. **Default values** make adding records faster:
   ```go
   {Name: "hours", Type: "integer", Default: "8"}
   ```

## Next Steps

- **Full docs:** `docs/RECORDS_TABLE.md` (comprehensive guide)
- **Examples:** `internal/tui/records_table_example.go` (6 complete examples)
- **Code:** `internal/tui/records_table.go` (implementation details)
- **Demo source:** `cmd/demo-records/main.go` (see how it works)

## Common Issues

**Q: Schema not showing up?**  
A: Ensure `RecordSchema` is not nil and has at least one field.

**Q: Changes not persisting?**  
A: By default, changes are in-memory only. See "Integration with Database" above.

**Q: Form fields not editable?**  
A: Current version uses simple form. Rich field editors are a future enhancement.

## Get Help

Run the demo and experiment! It's the fastest way to learn:

```bash
./demo-records work-log  # Start here
./demo-records contacts  # Then try this
```

Press `?` in any demo for keyboard shortcuts (note: help key not implemented yet, but shortcuts are shown at bottom).

---

**Ready to use it in your app?** See `docs/RECORDS_TABLE.md` for complete integration guide.