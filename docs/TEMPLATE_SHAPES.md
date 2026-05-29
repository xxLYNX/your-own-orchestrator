# Template Shapes - Architecture Guide

## Executive Summary

Templates in yoo are **fractal composition trees** built from composable **shapes**. A **Note** may optionally use a **Template**; the template's `composition` field defines nested shapes that you drill into in the orchestrator TUI.

### Shape primitives

| Shape | Purpose |
|-------|---------|
| **Procedure** | Ordered stages (may nest checklist, log, artifact, repeat) |
| **Checklist** | Unordered items within a procedure step |
| **Log** | Schema + repeating DB rows (operational data) |
| **Artifact** | External files/URLs; export target for logs |
| **Repeat** | Run a child subtree N times (e.g. `target_count` job applications) |

Legacy flat fields (`steps`, `record_schema`) are still accepted and compiled into a composition tree automatically.

### Example: Job Applications

```
Note
├── Procedure: Preparation (once)
├── Repeat × target_count: Single application
│   ├── Procedure: Application workflow + checklist
│   └── Log: Application record (scoped by repeat index)
└── Procedure: Wrap up → export log → artifacts
```

Navigate with **enter** (drill in), **esc** (up/back), **l** (log panel), **p** (procedure/steps), **f** (artifacts).

---

## Terminology

Templates use a few precise terms — they overlap in casual speech but mean different things in code:

| Term | Meaning |
|------|---------|
| **Shape** | A composable primitive kind (`procedure`, `checklist`, `log`, …). |
| **Node** | One entry in the template structure tree (has `id`, `kind`, `title`). |
| **Occurrence** | A node at a specific **repeat stack** — e.g. application **#4** of 10. The walker expands repeats into occurrences. |
| **Instance** | A persisted runtime row (`ShapeState`) tracking status for a trackable occurrence (checklist, action, procedure). Log and artifact nodes do not get instances. |
| **Item** | A child node inside a checklist — the thing behind each checkbox. The fourth row in a checklist is the fourth **item** (identified by its node `id` in `item_completion`). |
| **Record** | A row in the log table (`template_records`) — operational data, separate from shape instance state. |

**Example:** In a repeated job-application checklist, *“Tailor resume for this role”* is an **item** (structure node). When you are on application #4, you are looking at one **occurrence** of that checklist. Checking the box updates the checklist **instance** for that occurrence. Saving company/position updates a **record** in the log.

---

## Dependencies

Shapes can declare `depends_on` edges. The engine evaluates them at runtime and **hard-blocks** forward progress (checking items, marking complete) until satisfied.

### Supported requirements

| Requirement | Meaning |
|-------------|---------|
| `completed` | Target shape instance is done. |
| `started` | Target has been started or finished. |
| `has_record` | Log row(s) exist — see scopes below. |
| `has_artifact` | At least one artifact is attached to the note. |

`failed` is a terminal **instance status** (like `complete`). If a step failed, a `completed` dependency on it stays unsatisfied — a separate `not_failed` requirement is unnecessary.

#### Why no `approved` requirement?

Approval can be modeled today with a checklist item or action (“Manager signed off”) and a normal `completed` dependency. A dedicated `approved` requirement would only matter for **external attestation** — e.g. a second user, timestamped sign-off, or integration callback stored outside shape state. None of that exists yet; if it does later, either add an approval field on `ShapeState` or reintroduce the requirement deliberately.

### `has_record` scopes

Log data and checklist state are separate. `has_record` ties workflow progress to **evidence in the log**.

| Scope | Use when |
|-------|----------|
| `same_repeat_context` (default) | **This iteration** must have a log row — e.g. application #4 cannot be confirmed until application #4 is logged. |
| `all_occurrences` | **Every iteration** must have a log row — e.g. wrap-up blocked until all 10 applications have a record when `target_count: 10`. |

Example (per-iteration gate):

```yaml
- id: confirm-logged
  kind: action
  title: Confirm application record is complete
  depends_on:
    - target: application-log
      requirement: has_record
      scope: same_repeat_context
```

Example (all iterations before wrap-up):

```yaml
- id: wrapup
  kind: procedure
  depends_on:
    - target: application-log
      requirement: has_record
      scope: all_occurrences
```

Records are stored with the repeat stack of the iteration you were in when you added them (e.g. `[{"shape_id":"applications","index":4}]`), so scoped checks align with how the TUI filters the log panel.

---

## Legacy overview (flat model)

The original three-tab model maps to shapes as follows:

1. **LOG SHAPE** (Records) - Repeating structured data stored in the database
2. **CHECKLIST SHAPE** (Steps) - Sequential workflow steps to complete
3. **ARTIFACT SHAPE** (Files) - Deliverable files and documents

These shapes can be used independently or combined. For example, a job application note might use all three:
- **Log**: Track 10 job applications (company, position, date, status)
- **Checklist**: Find jobs → Customize → Apply → Document
- **Artifacts**: Attach resume.pdf, cover-letter.docx

## The Three Shapes

### 1. Log Shape (Database Records)

**Purpose**: Store repeating, structured data as database records (like rows in a spreadsheet).

**When to Use**:
- Tracking multiple similar items
- Need to query/filter data
- Want status tracking across many items
- Data naturally fits in a table
- Need to export to CSV/spreadsheet

**Examples**:
- Job applications (company, position, date, status)
- Research papers (title, authors, year, relevance)
- Contacts (name, email, company, context)
- Expenses (date, amount, category, vendor)
- Tasks/issues (title, owner, priority, status)

**Database Schema**:
```sql
CREATE TABLE template_records (
    id INTEGER PRIMARY KEY,
    note_template_id INTEGER,
    record_index INTEGER,           -- 1, 2, 3... (for record #1, #2, #3)
    data TEXT NOT NULL,             -- JSON: {"company": "Acme", "status": "Applied"}
    status TEXT DEFAULT 'draft',
    created_at DATETIME,
    updated_at DATETIME
);
```

**Template Definition**:
```yaml
record_schema:
  fields:
    - name: "company"
      type: "text"
      required: true
    - name: "position"
      type: "text"
      required: true
    - name: "status"
      type: "enum"
      values: ["Applied", "Interview", "Offer", "Rejected"]
      default: "Applied"
```

**CLI Usage**:
```bash
# Add a record
yoo records add <note-id> --field company="Acme" --field status="Applied"

# List all records
yoo records list <note-id>

# Update a record
yoo records edit <note-id> 3 --field status="Interview"

# Export to CSV
yoo records export <note-id> --format csv > data.csv
```

**Advantages**:
- ✅ Queryable with SQL
- ✅ Filterable and sortable
- ✅ Type validation
- ✅ Efficient storage
- ✅ Exportable to CSV/JSON
- ✅ No external file management

### 2. Checklist Shape (Workflow Steps)

**Purpose**: Define sequential steps in a workflow with checklist items and completion tracking.

**When to Use**:
- Sequential process/workflow
- Need guided execution
- Want checklist-style tracking
- Process is more important than data
- Steps have sub-tasks

**Examples**:
- Research process (find sources → read → synthesize → write)
- Application workflow (find → customize → submit → document)
- Review process (initial review → testing → approval → deployment)
- Content creation (outline → draft → edit → publish)

**Database Schema**:
```sql
CREATE TABLE note_steps (
    id INTEGER PRIMARY KEY,
    note_template_id INTEGER,
    step_number INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT 0,
    completed_at DATETIME,
    notes TEXT
);
```

**Template Definition**:
```yaml
steps:
  - id: 1
    title: "Find job opportunities"
    description: "Search job boards for relevant positions"
    checklist:
      - "Search LinkedIn for roles"
      - "Check Indeed for positions"
      - "Review job requirements"
    estimated_time: "1-2 hours"
  
  - id: 2
    title: "Customize applications"
    description: "Tailor resume and cover letter"
    estimated_time: "30-60 minutes per application"
    output_required: "customized_materials"
```

**CLI Usage** (Phase 2 - to be implemented):
```bash
# Complete a step
yoo step complete <note-id> 2

# List steps with status
yoo step list <note-id>

# Add notes to a step
yoo step note <note-id> 2 "Used template from last year"
```

**Advantages**:
- ✅ Clear progress tracking
- ✅ Guided execution
- ✅ Estimated time tracking
- ✅ Can link to required outputs
- ✅ Natural for workflows

### 3. Artifact Shape (Files & References)

**Purpose**: Track files, URLs, and other deliverables associated with a note.

**When to Use**:
- Actual files (PDFs, images, documents)
- Unstructured or free-form data
- Need to preserve original format
- Files created by external tools
- URLs and external references

**Examples**:
- Resume PDF, cover letter DOCX
- Photos, screenshots, diagrams
- Generated reports, presentations
- Research paper PDFs
- URLs to job postings, documentation

**Database Schema**:
```sql
CREATE TABLE artifacts (
    id INTEGER PRIMARY KEY,
    note_template_id INTEGER,
    artifact_type TEXT NOT NULL,     -- "input" or "output"
    name TEXT NOT NULL,
    type TEXT NOT NULL,              -- "file", "folder", "url"
    value TEXT NOT NULL,             -- File path or URL
    description TEXT,
    required BOOLEAN DEFAULT 0
);
```

**Template Definition**:
```yaml
inputs:
  - name: "resume"
    type: "file"
    description: "Your current resume/CV"
    required: true

outputs:
  - name: "applications_list"
    type: "file"
    description: "Spreadsheet of all applications"
    format: "csv|xlsx"
    required: true
  
  - name: "confirmation_emails"
    type: "folder"
    description: "Folder with confirmation emails"
    required: true
```

**CLI Usage** (Phase 2 - to be implemented):
```bash
# Add an artifact
yoo artifact add <note-id> --type output --name resume --file ~/resume.pdf

# List artifacts
yoo artifact list <note-id>

# Open an artifact
yoo artifact open <note-id> resume
```

**Advantages**:
- ✅ Track files without moving them
- ✅ Support any file type
- ✅ URLs and external references
- ✅ Input/output distinction
- ✅ Required vs. optional

## Combining Shapes

The power of yoo templates comes from combining shapes. Here's how they complement each other:

### Example 1: Job Applications (All Three Shapes)

```yaml
name: "Job Applications"

# Inputs (Artifact Shape)
inputs:
  - name: "resume"
    type: "file"
    required: true

# Log Shape - Track each application
record_schema:
  fields:
    - name: "company"
      type: "text"
      required: true
    - name: "status"
      type: "enum"
      values: ["Found", "Applied", "Interview", "Offer", "Rejected"]

# Checklist Shape - Overall workflow
steps:
  - id: 1
    title: "Find opportunities"
  - id: 2
    title: "Customize materials"
  - id: 3
    title: "Submit applications"
  - id: 4
    title: "Document results"
    output_required: "applications_list"

# Outputs (Artifact Shape)
outputs:
  - name: "applications_list"
    type: "file"
    format: "csv"
    required: true
```

**Usage Flow**:
```bash
# 1. Create note with template (Phase 2)
yoo add "Apply to 10 jobs" --template "Job Applications" --input resume=~/resume.pdf

# 2. Add application records (LOG SHAPE)
yoo records add 42 --field company="Acme" --field status="Applied"
yoo records add 42 --field company="TechCo" --field status="Applied"
# ... add 8 more

# 3. Complete workflow steps (CHECKLIST SHAPE)
yoo step complete 42 1  # Found opportunities
yoo step complete 42 2  # Customized materials
yoo step complete 42 3  # Submitted applications

# 4. Export and attach (ARTIFACT SHAPE)
yoo records export 42 --format csv > applications.csv
yoo artifact add 42 --type output --name applications_list --file ./applications.csv
yoo step complete 42 4  # Document results

# 5. View in TUI (Phase 3)
yoo schedule  # Shows note with progress: 10/10 records, 4/4 steps, 1/1 artifacts
```

### Example 2: Research Only Needs Log + Checklist

```yaml
name: "Literature Review"

record_schema:
  fields:
    - name: "paper_title"
      type: "text"
    - name: "status"
      type: "enum"
      values: ["To Read", "Reading", "Read", "Cited"]

steps:
  - id: 1
    title: "Find papers"
  - id: 2
    title: "Read and annotate"
  - id: 3
    title: "Synthesize findings"

# No artifacts needed - papers tracked as records
```

### Example 3: Simple Task List Only Needs Records

```yaml
name: "Shopping List"

record_schema:
  fields:
    - name: "item"
      type: "text"
    - name: "quantity"
      type: "integer"
    - name: "purchased"
      type: "boolean"
      default: false

# No steps needed - just a list
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         NOTE                                │
│  (id, title, description, scheduled_at, status)            │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ links to
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   NOTE_TEMPLATE                             │
│  (id, note_id, template_id, template_data)                 │
└─────┬───────────────────┬───────────────────┬───────────────┘
      │                   │                   │
      │                   │                   │
      │ has many          │ has many          │ has many
      ▼                   ▼                   ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────────┐
│  RECORDS    │   │   STEPS     │   │   ARTIFACTS     │
│  (LOG)      │   │ (CHECKLIST) │   │   (FILES)       │
├─────────────┤   ├─────────────┤   ├─────────────────┤
│ record_idx  │   │ step_number │   │ artifact_type   │
│ data (JSON) │   │ title       │   │ name            │
│ status      │   │ completed   │   │ type            │
│             │   │ notes       │   │ value (path)    │
└─────────────┘   └─────────────┘   └─────────────────┘
```

## Data Flow

### Creating a Templated Note

1. **User creates note from template**
   ```bash
   yoo add "My Project" --template "Project Template" --input target=10
   ```

2. **System creates records**:
   - Insert into `notes` table
   - Insert into `note_templates` table (links note to template)
   - Initialize steps from template definition
   - Validate required inputs

3. **User populates log shape**:
   ```bash
   yoo records add <note-id> --field name="Task 1" --field status="New"
   ```
   - Insert into `template_records` with JSON data
   - Auto-increment `record_index`

4. **User progresses through checklist**:
   ```bash
   yoo step complete <note-id> 1
   ```
   - Update `note_steps.completed = true`
   - Set `completed_at` timestamp
   - Update overall progress

5. **User attaches artifacts**:
   ```bash
   yoo artifact add <note-id> --name output --file ./result.pdf
   ```
   - Insert into `artifacts` table
   - Store file path or URL
   - Mark as input or output

6. **System calculates progress**:
   - Steps: `completed_steps / total_steps`
   - Records: count by status
   - Artifacts: `provided_required / total_required`
   - Overall: weighted combination

## Design Principles

### 1. Optional Everything
Templates can use any combination of shapes:
- Just records (shopping list)
- Just steps (process workflow)
- Just artifacts (file collection)
- Any combination

### 2. Database-First for Structure
Records (log shape) live in SQLite:
- ✅ Queryable
- ✅ Filterable
- ✅ Type-safe
- ✅ Exportable

### 3. File System for Blobs
Artifacts (file shape) stay on disk:
- ✅ Any file type
- ✅ No size limits
- ✅ Use native tools
- ✅ Just track paths

### 4. Process Tracking Built-in
Steps (checklist shape) track workflows:
- ✅ Sequential execution
- ✅ Progress visualization
- ✅ Time estimates
- ✅ Dependencies

### 5. Composable and Extensible
```yaml
# Start simple
record_schema:
  fields:
    - name: "task"
      type: "text"

# Add workflow
steps:
  - id: 1
    title: "Do the tasks"

# Add deliverables
outputs:
  - name: "results"
    type: "file"
```

## Performance Considerations

### Records (Log Shape)
- **Storage**: JSON in TEXT column (efficient for small-medium data)
- **Queries**: Use SQLite JSON functions (`json_extract()`)
- **Indexes**: On `note_template_id` and `status`
- **Limits**: Practical limit ~1000 records per note (more is fine, but export gets slow)

### Steps (Checklist Shape)
- **Storage**: Individual rows per step
- **Queries**: Simple WHERE on `note_template_id`
- **Limits**: Typically <20 steps per template

### Artifacts (File Shape)
- **Storage**: Only metadata in DB, files on disk
- **Queries**: Fast (small metadata only)
- **Limits**: No practical limit (files stay external)

## Future Enhancements

### Phase 4 Ideas

**Template Variables**:
```yaml
record_schema:
  fields:
    - name: "application_${index}"
      type: "text"
```

**Conditional Fields**:
```yaml
record_schema:
  fields:
    - name: "status"
      type: "enum"
      values: ["Pending", "Approved", "Rejected"]
    
    - name: "rejection_reason"
      type: "text"
      required: true
      condition: "status == 'Rejected'"
```

**Calculated Fields**:
```yaml
record_schema:
  fields:
    - name: "hours"
      type: "integer"
    - name: "rate"
      type: "integer"
    - name: "total"
      type: "integer"
      calculated: "hours * rate"
```

**Cross-Record Aggregations**:
```yaml
outputs:
  - name: "summary"
    type: "text"
    calculated: "COUNT(records WHERE status='Complete')"
```

## Conclusion

The three-shape architecture provides:

1. **Flexibility**: Use only what you need
2. **Power**: Combine shapes for complex workflows
3. **Simplicity**: Each shape has a clear purpose
4. **Efficiency**: Right tool for each type of data
5. **Queryability**: Database-native structured data
6. **Practicality**: Files stay files, data stays data

This architecture makes yoo a true orchestrator: it can track both the **data** (records), the **process** (steps), and the **deliverables** (artifacts) for any workflow.

---

**See Also**:
- `TEMPLATE_RECORDS.md` - Complete guide to using records
- `TEMPLATES_DESIGN.md` - Original template system design
- `TEMPLATE_ROADMAP.md` - Implementation phases