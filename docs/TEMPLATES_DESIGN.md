# Note Templates and Workflow System - Design Document

**Status**: Proposal  
**Version**: 0.1.0  
**Date**: 2026-05-27  
**Author**: yoo development team

---

## Executive Summary

This document proposes a **Template System** for yoo that allows notes to evolve from simple reminders into structured workflows with inputs, procedural steps, and outputs, while maintaining simplicity for basic use cases.

### Problem Statement

Current limitation: Notes are flat entities with only title, description, and metadata. Many real-world tasks require:
- **Inputs**: Prerequisites or resources needed
- **Steps**: Procedural actions to complete
- **Outputs**: Deliverables or artifacts produced
- **External References**: Links to files, URLs, or other systems

**Example**: "Apply to 10 jobs"
- Input: Job boards, resume, cover letter template
- Steps: 1) Find jobs, 2) Apply, 3) Document in spreadsheet
- Output: Applications list for accountability partner

---

## Design Principles

1. **Simple by Default**: Basic notes remain simple text
2. **Complexity on Demand**: Templates available when needed
3. **Progressive Enhancement**: Notes can be upgraded to use templates
4. **Template Reusability**: Common patterns captured as reusable templates
5. **External Integration**: Easy linking to files, URLs, other tools

---

## Core Concepts

### 1. Template Types

```
┌─────────────────────────────────────────────────┐
│            Note Hierarchy                       │
├─────────────────────────────────────────────────┤
│                                                 │
│  Simple Note                                    │
│  └─ Title + Description (current)              │
│                                                 │
│  Templated Note                                 │
│  ├─ Template Reference                          │
│  ├─ Input Artifacts                             │
│  ├─ Steps/Checklist                             │
│  ├─ Output Artifacts                            │
│  └─ External References                         │
│                                                 │
│  Workflow Note (Future)                         │
│  └─ Multi-stage process with state machine     │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 2. Template Structure

```yaml
template:
  id: "job-applications"
  name: "Job Applications"
  version: "1.0"
  description: "Track multiple job applications with documentation"
  
  inputs:
    - name: "target_count"
      type: "integer"
      description: "Number of jobs to apply to"
      required: true
    - name: "resume"
      type: "file"
      description: "Resume file path"
      required: true
    - name: "cover_letter_template"
      type: "file"
      description: "Cover letter template"
      required: false
      
  steps:
    - id: 1
      title: "Find job opportunities"
      description: "Search job boards for relevant positions"
      checklist:
        - "Check LinkedIn"
        - "Check Indeed"
        - "Check company career pages"
    - id: 2
      title: "Customize applications"
      description: "Tailor resume and cover letter for each role"
    - id: 3
      title: "Submit applications"
      description: "Apply to positions and track submissions"
    - id: 4
      title: "Document applications"
      description: "Record all applications in tracking spreadsheet"
      output_required: "applications_list"
      
  outputs:
    - name: "applications_list"
      type: "file"
      description: "Spreadsheet of all applications"
      format: "csv|xlsx|google-sheets"
      required: true
    - name: "confirmation_emails"
      type: "folder"
      description: "Folder with application confirmations"
      required: false
      
  metadata:
    category: "career"
    tags: ["job-search", "career-development"]
    estimated_duration: "2-3 hours"
```

---

## Database Schema Changes

### New Tables

```sql
-- Template definitions
CREATE TABLE templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT,
    category TEXT,
    definition TEXT NOT NULL,  -- JSON/YAML template structure
    is_builtin BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Template usage tracking
CREATE TABLE note_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL,
    template_data TEXT,  -- JSON: instantiated template with values
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES templates(id)
);

-- Step tracking for templated notes
CREATE TABLE note_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    step_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT 0,
    completed_at DATETIME,
    notes TEXT,  -- User notes on this step
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);

-- Artifacts (inputs/outputs)
CREATE TABLE artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    artifact_type TEXT NOT NULL,  -- 'input' or 'output'
    name TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'file', 'url', 'text', 'folder'
    value TEXT NOT NULL,  -- Path, URL, or content
    description TEXT,
    required BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);

-- External references
CREATE TABLE note_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    ref_type TEXT NOT NULL,  -- 'file', 'url', 'command', 'note'
    ref_value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE INDEX idx_note_templates_note ON note_templates(note_id);
CREATE INDEX idx_artifacts_note_template ON artifacts(note_template_id);
CREATE INDEX idx_note_references_note ON note_references(note_id);
```

### Modified Notes Table

```sql
-- Add template support to existing notes table
ALTER TABLE notes ADD COLUMN is_templated BOOLEAN DEFAULT 0;
ALTER TABLE notes ADD COLUMN template_progress REAL DEFAULT 0.0;  -- 0.0 to 1.0
```

---

## CLI Commands

### Creating Templated Notes

```bash
# List available templates
yoo templates list

# Show template details
yoo templates show job-applications

# Create note from template
yoo add "Apply to 10 jobs" --template job-applications

# Interactive template setup
yoo add --template job-applications --interactive

# Create note from template with inputs
yoo add "Apply to 10 jobs" \
  --template job-applications \
  --input target_count=10 \
  --input resume=~/Documents/resume.pdf \
  --input cover_letter=~/Documents/cover-letter.txt
```

### Working with Templated Notes

```bash
# View note with template structure
yoo show <note-id>

# Complete a step
yoo step complete <note-id> <step-number>
yoo step complete 42 2  # Complete step 2 of note 42

# Add artifact (input/output)
yoo artifact add <note-id> --type output \
  --name applications_list \
  --file ~/Documents/job-apps.xlsx

# Add reference to external resource
yoo ref add <note-id> --url https://linkedin.com/jobs
yoo ref add <note-id> --file ~/Projects/job-search/notes.txt
yoo ref add <note-id> --command "open spreadsheet.xlsx"

# View progress
yoo progress <note-id>
```

### Template Management

```bash
# Create custom template
yoo template create job-applications.yaml

# Edit template
yoo template edit job-applications

# Export template
yoo template export job-applications > my-template.yaml

# Import template
yoo template import my-template.yaml

# Delete template
yoo template delete job-applications
```

---

## Built-in Templates

### 1. Job Applications Template

**Use Case**: Apply to multiple positions with documentation

```yaml
name: "Job Applications"
inputs: [resume, cover_letter, target_count]
steps: [find, customize, apply, document]
outputs: [applications_list, confirmation_emails]
```

### 2. Research Task Template

**Use Case**: Conduct research with sources and findings

```yaml
name: "Research Task"
inputs: [topic, deadline, depth_level]
steps: [literature_review, data_collection, analysis, write_up]
outputs: [research_notes, bibliography, summary_document]
```

### 3. Project Milestone Template

**Use Case**: Complete project milestone with deliverables

```yaml
name: "Project Milestone"
inputs: [project_spec, resources, deadline]
steps: [planning, implementation, testing, documentation]
outputs: [deliverable, test_results, docs]
```

### 4. Content Creation Template

**Use Case**: Create content with editing and publishing

```yaml
name: "Content Creation"
inputs: [topic, target_audience, platform]
steps: [outline, draft, edit, design, publish]
outputs: [final_content, analytics_link]
```

### 5. Learning Module Template

**Use Case**: Learn new skill with practice and proof

```yaml
name: "Learning Module"
inputs: [skill, resources, proficiency_goal]
steps: [study, practice, build_project, get_feedback]
outputs: [project_demo, certificate, portfolio_entry]
```

### 6. Meeting Preparation Template

**Use Case**: Prepare for important meeting

```yaml
name: "Meeting Prep"
inputs: [agenda, attendees, objectives]
steps: [review_materials, prepare_talking_points, create_slides]
outputs: [presentation, meeting_notes, action_items]
```

---

## TUI Integration

### Template Selection View

```
┌─────────────────────────────────────────────────────────────┐
│ Add Note - Select Template                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ > Simple Note (no template)                                 │
│                                                             │
│   Job Applications                                          │
│   └─ Track multiple job applications with documentation    │
│                                                             │
│   Research Task                                             │
│   └─ Conduct research with sources and findings            │
│                                                             │
│   Project Milestone                                         │
│   └─ Complete project milestone with deliverables          │
│                                                             │
│   [Custom Templates...]                                     │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ ↑/↓: Navigate • Enter: Select • Esc: Cancel                │
└─────────────────────────────────────────────────────────────┘
```

### Templated Note View

```
┌─────────────────────────────────────────────────────────────┐
│ 📋 Apply to 10 jobs                         [Progress: 75%] │
├─────────────────────────────────────────────────────────────┤
│ Template: Job Applications v1.0                             │
│ Due: 2024-06-15 | Priority: High                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 📥 INPUTS                                                   │
│   ✓ Resume: ~/Documents/resume.pdf                         │
│   ✓ Target: 10 applications                                │
│   ✗ Cover Letter: [Not provided]                           │
│                                                             │
│ ✅ STEPS                                                    │
│   ☑ 1. Find job opportunities          [Completed]         │
│   ☑ 2. Customize applications          [Completed]         │
│   ☑ 3. Submit applications             [Completed]         │
│   ☐ 4. Document applications           [In Progress]       │
│                                                             │
│ 📤 OUTPUTS                                                  │
│   ⏳ Applications list: [Pending]                           │
│   ✓ Confirmations: ~/Jobs/confirmations/                   │
│                                                             │
│ 🔗 REFERENCES                                               │
│   • https://linkedin.com/jobs                              │
│   • file://~/Projects/job-search/companies.txt            │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ Enter: Edit • S: Complete step • A: Add artifact • Q: Back │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Foundation (v0.2.0)
- [ ] Database schema changes
- [ ] Template data model in Go
- [ ] Basic template CRUD operations
- [ ] CLI: `yoo templates list/show`
- [ ] Add 2-3 built-in templates

### Phase 2: Core Functionality (v0.3.0)
- [ ] Create templated notes via CLI
- [ ] Step tracking and completion
- [ ] Basic artifact management
- [ ] CLI: All template commands
- [ ] Template import/export

### Phase 3: TUI Integration (v0.4.0)
- [ ] Template selection in TUI
- [ ] Templated note view with progress
- [ ] Interactive step completion
- [ ] Artifact management in TUI

### Phase 4: Advanced Features (v0.5.0)
- [ ] Custom template creation wizard
- [ ] Template marketplace/sharing
- [ ] Template variables and expressions
- [ ] Conditional steps
- [ ] External tool integration (hooks)

---

## API Design

### Go Structs

```go
// Template represents a reusable note structure
type Template struct {
    ID          int64
    Name        string
    Version     string
    Description string
    Category    string
    Definition  TemplateDefinition
    IsBuiltin   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// TemplateDefinition is the template structure
type TemplateDefinition struct {
    Inputs   []TemplateInput   `yaml:"inputs"`
    Steps    []TemplateStep    `yaml:"steps"`
    Outputs  []TemplateOutput  `yaml:"outputs"`
    Metadata TemplateMetadata  `yaml:"metadata"`
}

// TemplateInput defines required inputs
type TemplateInput struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"` // file, url, text, integer, etc.
    Description string `yaml:"description"`
    Required    bool   `yaml:"required"`
    Default     string `yaml:"default,omitempty"`
}

// TemplateStep defines a procedural step
type TemplateStep struct {
    ID              int      `yaml:"id"`
    Title           string   `yaml:"title"`
    Description     string   `yaml:"description"`
    Checklist       []string `yaml:"checklist,omitempty"`
    OutputRequired  string   `yaml:"output_required,omitempty"`
}

// TemplateOutput defines expected outputs
type TemplateOutput struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"`
    Description string `yaml:"description"`
    Format      string `yaml:"format,omitempty"`
    Required    bool   `yaml:"required"`
}

// NoteTemplate represents a note using a template
type NoteTemplate struct {
    ID           int64
    NoteID       int64
    TemplateID   int64
    TemplateData TemplateInstance
    CreatedAt    time.Time
}

// TemplateInstance is the instantiated template with values
type TemplateInstance struct {
    Inputs   map[string]interface{}
    Steps    []StepInstance
    Outputs  map[string]Artifact
}

// StepInstance tracks step completion
type StepInstance struct {
    StepNumber  int
    Title       string
    Description string
    Completed   bool
    CompletedAt *time.Time
    Notes       string
}

// Artifact represents an input or output
type Artifact struct {
    Name        string
    Type        string // file, url, text, folder
    Value       string
    Description string
    CreatedAt   time.Time
}

// Reference links to external resources
type Reference struct {
    ID          int64
    NoteID      int64
    Type        string // file, url, command, note
    Value       string
    Description string
    CreatedAt   time.Time
}
```

---

## Examples

### Example 1: Simple Job Application Flow

```bash
# Create the note
yoo add "Apply to 10 tech jobs" --template job-applications \
  --input target_count=10 \
  --input resume=~/resume.pdf

# Complete steps as you work
yoo step complete 42 1  # Found jobs
yoo step complete 42 2  # Customized applications
yoo step complete 42 3  # Submitted applications

# Add the output artifact
yoo artifact add 42 --type output \
  --name applications_list \
  --file ~/job-applications.xlsx

# View progress
yoo progress 42
# Output: 100% complete (4/4 steps, 1/1 outputs)

# Mark note as complete
yoo complete 42
```

### Example 2: Research with External References

```bash
# Create research note
yoo add "Research GraphQL vs REST" --template research-task \
  --input topic="GraphQL vs REST APIs" \
  --input depth="comprehensive"

# Add references as you find them
yoo ref add 43 --url https://graphql.org/
yoo ref add 43 --url https://restfulapi.net/
yoo ref add 43 --file ~/Papers/graphql-spec.pdf

# Complete steps
yoo step complete 43 1  # Literature review done

# Add output
yoo artifact add 43 --type output \
  --name research_notes \
  --file ~/Research/graphql-vs-rest.md
```

### Example 3: Custom Template

```yaml
# ~/.config/yoo/templates/weekly-review.yaml
name: "Weekly Review"
version: "1.0"
description: "Weekly reflection and planning session"

inputs:
  - name: "week_number"
    type: "integer"
    required: true
  - name: "last_week_notes"
    type: "file"
    required: false

steps:
  - id: 1
    title: "Review last week"
    description: "Go through completed tasks and outcomes"
  - id: 2
    title: "Identify wins and challenges"
    description: "What went well? What was difficult?"
  - id: 3
    title: "Plan next week"
    description: "Set priorities and goals"
  - id: 4
    title: "Update tracking systems"
    description: "Update project boards, calendars, etc."

outputs:
  - name: "weekly_summary"
    type: "file"
    description: "Summary of review and plans"
    format: "markdown"
    required: true
  - name: "next_week_goals"
    type: "text"
    description: "Top 3 goals for next week"
    required: true
```

```bash
# Import and use
yoo template import ~/.config/yoo/templates/weekly-review.yaml
yoo add "Week 22 Review" --template weekly-review --input week_number=22
```

---

## Future Enhancements

### Template Marketplace
- Share templates with community
- Rate and review templates
- Install templates from repository

### Template Variables
```yaml
steps:
  - title: "Apply to {target_count} jobs"
  - title: "Document applications in {output_format}"
```

### Conditional Logic
```yaml
steps:
  - id: 3
    title: "Write cover letters"
    condition: "inputs.include_cover_letter == true"
```

### External Tool Integration
```yaml
outputs:
  - name: "applications_list"
    post_action:
      command: "open {value}"
      on_complete: true
```

### Templates from Notes
```bash
# Convert existing note to template
yoo template from-note <note-id> --name "My Custom Template"
```

---

## Migration Strategy

### Backward Compatibility
- Existing simple notes remain unchanged
- Templates are opt-in feature
- No breaking changes to current API
- Notes can be "upgraded" to use templates later

### User Onboarding
- Help text shows template feature
- `yoo add --help` shows template examples
- TUI prompts for template when appropriate
- Documentation with real-world examples

---

## Success Metrics

- **Adoption**: % of users creating templated notes
- **Template Usage**: Which templates are most popular
- **Completion Rate**: Do templated notes have higher completion?
- **Custom Templates**: How many users create custom templates
- **Feedback**: User satisfaction with template feature

---

## Open Questions

1. Should templates support nesting (template within template)?
2. How to handle template version upgrades for existing notes?
3. Should we support template inheritance?
4. Integration with external systems (Jira, GitHub, etc.)?
5. Should artifacts support cloud storage (S3, Drive, etc.)?

---

## Conclusion

This template system transforms yoo from a simple task manager into a **procedural workflow tool** while maintaining its core simplicity. Users can choose the level of structure they need:

- **Level 0**: Simple text notes (current)
- **Level 1**: Notes with references
- **Level 2**: Templated notes with steps
- **Level 3**: Full workflows with inputs/outputs

This progressive enhancement approach ensures yoo remains intuitive for casual use while supporting power users with complex workflows.

---

**Next Step**: Gather feedback on this design and prioritize Phase 1 implementation.