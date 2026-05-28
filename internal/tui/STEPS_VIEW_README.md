# Steps View - Interactive Checklist TUI

A beautiful, interactive terminal user interface (TUI) for managing workflow steps in templated notes. The steps view provides an intuitive checklist experience with progress tracking, step details, and note-taking capabilities.

## Overview

The Steps View (`steps_view.go`) is a Bubble Tea-powered TUI component that displays and manages workflow steps for templated notes. It provides:

- **Visual Progress Tracking** - Real-time progress bar showing completion status
- **Interactive Navigation** - Keyboard-driven step navigation and management
- **Step Details** - View descriptions, checklists, estimated times, and notes
- **Note-Taking** - Add contextual notes to individual steps
- **Persistent State** - All changes are saved to the database in real-time

## Features

### 1. Progress Bar
```
Progress: [████████░░░░░░] 4/7 steps (57%)
```
- Visual representation of overall completion
- Shows completed/total steps and percentage
- Updates in real-time as steps are toggled

### 2. Step List Display
```
✓ 1. Find job opportunities          Completed 2024-01-15
○ 2. Customize applications
✓ 3. Submit applications              Completed 2024-01-18
> ○ 4. Document applications
```
- Checkbox indicators (✓ completed, ○ pending)
- Cursor indicator (>) for selected step
- Completion timestamps for finished steps
- Strikethrough styling for completed steps

### 3. Step Details Panel
```
Selected: 2. Customize applications
Description: Tailor resume and cover letter for each role

Checklist:
  ☐ Review job description keywords
  ☐ Adjust resume to highlight relevant experience
  ☐ Customize cover letter if required

Estimated: 30-60 minutes per application

Notes:
  Remember to emphasize Go experience and distributed systems
```

### 4. Detailed View Mode
Press `v` to toggle full-screen details for the selected step:
- Complete step information
- All checklist items
- Estimated time with visual badge
- Required outputs (if any)
- Notes section with edit prompt
- Completion timestamp

### 5. Note Input Mode
Press `n` to add or edit notes for the current step:
- Multi-line text input
- Pre-populated with existing notes
- Inline editing controls
- Auto-saves on Enter

## Keyboard Shortcuts

| Key | Action | Description |
|-----|--------|-------------|
| `↑` or `k` | Move Up | Navigate to previous step |
| `↓` or `j` | Move Down | Navigate to next step |
| `Space` or `Enter` | Toggle | Mark step as complete/incomplete |
| `n` | Add Note | Add or edit notes for current step |
| `v` | View Details | Toggle detailed view mode |
| `Esc` or `q` | Quit | Return to previous view / exit |

### Note Input Mode Keys

| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Save | Save note and return to list |
| `Esc` | Cancel | Discard changes and return |
| `Backspace` | Delete Char | Delete previous character |
| `Ctrl+U` | Clear Line | Clear entire input |
| `Ctrl+W` | Delete Word | Delete previous word |

## Usage

### Basic Usage

```go
import (
    "database/sql"
    "yoo/internal/database"
    "yoo/internal/tui"
)

func main() {
    // Open database
    db, err := sql.Open("sqlite", "yoo.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Load note and template
    noteID := int64(1)
    noteTemplate, err := database.GetNoteTemplate(db, noteID)
    if err != nil {
        panic(err)
    }

    template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
    if err != nil {
        panic(err)
    }

    // Show interactive steps view
    if err := tui.ShowSteps(db, noteID, noteTemplate.ID, template); err != nil {
        panic(err)
    }
}
```

### Creating a New Templated Note with Steps

```go
// Get template
template, err := database.GetTemplateByID(db, templateID)
if err != nil {
    return err
}

// Create note
note := &database.Note{
    Title:       "Job Application - Software Engineer",
    Description: "Tracking application workflow",
    ScheduledAt: time.Now(),
    Status:      "pending",
    IsTemplated: true,
}
err = database.CreateNote(db, note)
if err != nil {
    return err
}

// Create template instance with inputs
inputs := map[string]interface{}{
    "company_name": "Acme Corp",
    "position":     "Software Engineer",
}
instance := models.NewTemplateInstance(&template.Definition, inputs)

// Attach template to note (creates steps)
noteTemplate, err := database.AttachTemplateToNote(db, note.ID, template.ID, instance)
if err != nil {
    return err
}

// Launch steps view
err = tui.ShowSteps(db, note.ID, noteTemplate.ID, template)
```

### Programmatic Step Management

```go
// Complete a step
err := database.CompleteStep(db, noteTemplateID, stepNumber)

// Uncomplete a step
err := database.UncompleteStep(db, noteTemplateID, stepNumber)

// Add notes to a step
err := database.UpdateStepNotes(db, noteTemplateID, stepNumber, "Important notes here")

// Get all steps
steps, err := database.ListNoteSteps(db, noteTemplateID)

// Calculate progress
completed := 0
for _, step := range steps {
    if step.Completed {
        completed++
    }
}
progress := float64(completed) / float64(len(steps))
```

## Integration with Schedule View

The steps view integrates seamlessly with the schedule view:

```go
// In your schedule view, when user selects a templated note:
if note.IsTemplated {
    // Load template data
    noteTemplate, err := database.GetNoteTemplate(db, note.ID)
    if err != nil {
        return err
    }

    template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
    if err != nil {
        return err
    }

    // Launch steps view
    if err := tui.ShowSteps(db, note.ID, noteTemplate.ID, template); err != nil {
        return err
    }

    // After steps view closes, recalculate progress
    steps, _ := database.ListNoteSteps(db, noteTemplate.ID)
    completed := 0
    for _, step := range steps {
        if step.Completed {
            completed++
        }
    }
    
    // Update note with new progress
    note.TemplateProgress = float64(completed) / float64(len(steps))
    database.UpdateNote(db, note)
}
```

## Model Structure

### StepsViewModel

```go
type StepsViewModel struct {
    db             *sql.DB               // Database connection
    noteID         int64                 // Parent note ID
    noteTemplateID int64                 // Note template ID
    steps          []*models.StepInstance // List of steps
    template       *models.Template      // Template definition
    cursor         int                   // Current selection
    showingDetails bool                  // Detailed view mode
    addingNote     bool                  // Note input mode
    noteInput      string                // Note text buffer
    width          int                   // Terminal width
    height         int                   // Terminal height
    err            error                 // Error state
    quitting       bool                  // Quit flag
}
```

### StepInstance (from models)

```go
type StepInstance struct {
    ID          int64      // Database ID
    StepNumber  int        // Step number (1-based)
    Title       string     // Step title
    Description string     // Step description
    Completed   bool       // Completion status
    CompletedAt *time.Time // Completion timestamp
    Notes       string     // User notes
    CreatedAt   time.Time  // Creation timestamp
}
```

## Database Schema

The steps view requires the following tables:

### note_steps

```sql
CREATE TABLE note_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    step_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT 0,
    completed_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);
```

### Required Database Functions

- `database.ListNoteSteps(db, noteTemplateID)` - Get all steps
- `database.CompleteStep(db, noteTemplateID, stepNumber)` - Mark complete
- `database.UncompleteStep(db, noteTemplateID, stepNumber)` - Mark incomplete
- `database.UpdateStepNotes(db, noteTemplateID, stepNumber, notes)` - Save notes
- `database.GetNoteTemplate(db, noteID)` - Get note template
- `database.GetTemplateByID(db, templateID)` - Get template definition

## Styling

The steps view uses the centralized styling system from `styles.go`:

- **Colors**: Primary (#7D56F4), Success (#04B575), Muted (#666666)
- **Components**: Progress bars, checkboxes, borders, status badges
- **Helper Functions**: `Checkbox()`, `ProgressBar()`, `StatusBadge()`, `KeyBindings()`

### Custom Styling

To customize the appearance:

```go
// Modify styles.go or override in your view
var customStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#custom")).
    Background(lipgloss.Color("#color")).
    Bold(true)
```

## Error Handling

The steps view handles errors gracefully:

```go
// Database errors
if err != nil {
    m.err = err
    return m, nil
}

// Error display
if m.err != nil {
    return ErrorMessageStyle.Render("Error: " + m.err.Error()) + "\n\n" +
        HelpStyle.Render("Press q to quit")
}
```

## Performance Considerations

- **Database Queries**: Steps are loaded once on initialization
- **Real-time Updates**: Changes are persisted immediately to database
- **Memory Usage**: Minimal - only current template and steps in memory
- **Render Optimization**: Only re-renders on state changes

## Advanced Features

### Checklist Items

Steps can include sub-checklist items defined in the template:

```yaml
steps:
  - id: 1
    title: "Customize applications"
    description: "Tailor resume and cover letter"
    checklist:
      - "Review job description keywords"
      - "Adjust resume highlights"
      - "Customize cover letter"
    estimated_time: "30-60 minutes"
```

These are displayed in the step details but are not individually trackable (template-level only).

### Estimated Time Display

Shows time estimates with visual badges:
```
Estimated: 🕐 30-60 minutes per application
```

### Output Requirements

Steps can require specific outputs to be provided:

```yaml
steps:
  - id: 3
    title: "Submit applications"
    output_required: "application_tracking_sheet"
```

Displayed as:
```
Output Required: → application_tracking_sheet
```

## Testing

Example test scenarios:

```go
func TestStepsViewNavigation(t *testing.T) {
    // Test cursor movement
    // Test step completion toggling
    // Test note input
}

func TestStepsViewProgress(t *testing.T) {
    // Test progress calculation
    // Test completion state persistence
}
```

## Troubleshooting

### Steps Not Loading

```
Error: failed to load steps
```

**Solution**: Ensure the note_template_id is valid and steps were created when attaching the template.

### Completion Not Persisting

**Solution**: Check database connection and transaction handling. Verify the `CompleteStep` function is being called.

### Notes Not Saving

**Solution**: Verify `UpdateStepNotes` is called correctly and the step_number exists.

### Display Issues

**Solution**: Ensure terminal supports UTF-8 and ANSI colors. Use `tea.WithAltScreen()` for full-screen apps.

## Contributing

When extending the steps view:

1. Follow the existing Bubble Tea pattern (Init/Update/View)
2. Use styles from `styles.go` for consistency
3. Add keyboard shortcuts to the help text
4. Update this README with new features
5. Test with various terminal sizes

## License

Part of the Your Own Orchestrator (YOO) project.

## See Also

- `schedule.go` - Schedule view component
- `styles.go` - Centralized styling system
- `models/template.go` - Template data structures
- `database/steps.go` - Step database operations