# Quick Start Guide - Steps View

Get started with the interactive steps checklist view in 5 minutes.

## What is Steps View?

An interactive terminal interface for managing workflow steps in templated notes. Track progress, check off completed steps, and add notes—all from your keyboard.

```
Progress: [████████░░░░] 4/7 steps (57%)

✓ 1. Find job opportunities    Completed 2024-01-15
○ 2. Customize applications
✓ 3. Submit applications        Completed 2024-01-18
> ○ 4. Document applications
```

## Quick Start

### 1. Basic Usage

```go
package main

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

    // Load your note (must be templated)
    noteID := int64(1)
    noteTemplate, err := database.GetNoteTemplate(db, noteID)
    if err != nil {
        panic(err)
    }

    // Load template definition
    template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
    if err != nil {
        panic(err)
    }

    // Launch the steps view!
    if err := tui.ShowSteps(db, noteID, noteTemplate.ID, template); err != nil {
        panic(err)
    }
}
```

### 2. Create a New Templated Note

```go
// Get a template
template, _ := database.GetTemplateByID(db, templateID)

// Create a note
note := &database.Note{
    Title:       "My Project Workflow",
    Description: "Track project steps",
    ScheduledAt: time.Now(),
    Status:      "pending",
    IsTemplated: true,
}
database.CreateNote(db, note)

// Setup inputs
inputs := map[string]interface{}{
    "project_name": "My Cool App",
    "deadline":     "2024-02-01",
}

// Create template instance
instance := models.NewTemplateInstance(&template.Definition, inputs)

// Attach to note (creates steps automatically)
noteTemplate, _ := database.AttachTemplateToNote(db, note.ID, template.ID, instance)

// Launch steps view
tui.ShowSteps(db, note.ID, noteTemplate.ID, template)
```

## Essential Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Space` | Toggle step completion |
| `n` | Add/edit note |
| `v` | View full details |
| `q` | Quit |

## Common Tasks

### Check off a step
1. Navigate to step with `↑`/`↓` or `j`/`k`
2. Press `Space` or `Enter`
3. Step is immediately marked complete ✓

### Add a note to a step
1. Navigate to the step
2. Press `n` to enter note mode
3. Type your note
4. Press `Enter` to save (or `Esc` to cancel)

### View step details
1. Navigate to any step
2. Press `v` to see full details
3. Press `v` again or `Esc` to return

### Track progress
- Progress bar at the top shows real-time completion
- Completed steps show green checkmarks
- Percentage updates automatically

## Integration with Schedule View

```go
// In your schedule view handler
func handleNoteSelection(db *sql.DB, note *database.Note) error {
    if !note.IsTemplated {
        fmt.Println("This note doesn't have steps")
        return nil
    }

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
    return tui.ShowSteps(db, note.ID, noteTemplate.ID, template)
}
```

## Programmatic Step Management

### Complete a step
```go
err := database.CompleteStep(db, noteTemplateID, stepNumber)
```

### Add notes
```go
err := database.UpdateStepNotes(db, noteTemplateID, stepNumber, "My notes here")
```

### Check progress
```go
steps, _ := database.ListNoteSteps(db, noteTemplateID)
completed := 0
for _, step := range steps {
    if step.Completed {
        completed++
    }
}
progress := float64(completed) / float64(len(steps))
```

## Example Template

Create a template YAML file:

```yaml
name: "Simple Project Workflow"
version: "1.0.0"
description: "Basic project management workflow"
category: "project"

definition:
  inputs:
    - name: project_name
      type: text
      description: "Name of the project"
      required: true

  steps:
    - id: 1
      title: "Plan project scope"
      description: "Define goals and requirements"
      checklist:
        - "List all features"
        - "Define success criteria"
        - "Set timeline"
      estimated_time: "2-3 hours"

    - id: 2
      title: "Setup development environment"
      description: "Configure tools and dependencies"
      checklist:
        - "Install dependencies"
        - "Configure IDE"
        - "Setup version control"
      estimated_time: "1 hour"

    - id: 3
      title: "Implement features"
      description: "Build the core functionality"
      estimated_time: "1-2 weeks"

    - id: 4
      title: "Test and deploy"
      description: "Verify everything works"
      checklist:
        - "Write tests"
        - "Run QA checks"
        - "Deploy to production"
      estimated_time: "3-5 days"

  metadata:
    tags: ["project", "workflow"]
    estimated_duration: "2-3 weeks"
```

Import it:
```bash
yoo template import simple-project.yaml
```

## Tips & Tricks

### 1. Vim-style Navigation
Use `j` (down) and `k` (up) for quick navigation without leaving home row.

### 2. Quick Toggle
Press `Space` for instant step completion—no confirmation needed.

### 3. Context in Notes
Add dates, links, or specifics in notes for future reference:
```
Applied via company portal. Confirmation #12345.
Contact: jane@company.com
Deadline: Jan 30
```

### 4. View Details First
Press `v` on a new step to see full checklist before starting.

### 5. Progress Tracking
The progress bar updates instantly—watch your completion grow!

## Troubleshooting

### "Failed to load steps"
- Ensure the note is templated (`IsTemplated = true`)
- Verify steps were created when attaching template
- Check database connection

### "No steps found"
- Template must have steps defined
- Use `database.AttachTemplateToNote()` to create steps
- Verify `noteTemplateID` is correct

### Display issues
- Ensure terminal supports UTF-8 and ANSI colors
- Try a different terminal emulator
- Check terminal size (minimum 80x24 recommended)

## Next Steps

- **Full Documentation:** See `STEPS_VIEW_README.md`
- **Visual Demo:** Check `STEPS_VIEW_DEMO.txt`
- **Code Examples:** Browse `steps_view_example.go`
- **Implementation:** Read `IMPLEMENTATION_SUMMARY.md`

## Quick Reference Card

```
╔══════════════════════════════════════════╗
║         STEPS VIEW CHEAT SHEET           ║
╠══════════════════════════════════════════╣
║ Navigation                               ║
║ ↑/k      Move up                         ║
║ ↓/j      Move down                       ║
║                                          ║
║ Actions                                  ║
║ Space    Toggle completion               ║
║ n        Add/edit note                   ║
║ v        View details                    ║
║ q/Esc    Quit/back                       ║
║                                          ║
║ Note Input Mode                          ║
║ Enter    Save note                       ║
║ Esc      Cancel                          ║
║ Ctrl+U   Clear all                       ║
║ Ctrl+W   Delete word                     ║
╚══════════════════════════════════════════╝
```

---

**Ready to go?** Just load a templated note and call `tui.ShowSteps()`. Everything else is intuitive!