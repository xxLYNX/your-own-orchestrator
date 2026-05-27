# Phase 2 Complete: Templated Notes & Workflow Management

## Overview

Phase 2 is now **complete**! You can now create notes from templates and manage all three template shapes:
- ✅ **LOG SHAPE** (Records) - Database-native structured data
- ✅ **CHECKLIST SHAPE** (Steps) - Sequential workflow tracking
- ✅ **ARTIFACT SHAPE** (Files) - File and URL references

## What's New in Phase 2

### 1. Create Templated Notes
```bash
yoo add "Task name" --template "Template Name" --input key=value
```

### 2. Step Management
```bash
yoo step list <note-id>          # List all steps with status
yoo step show <note-id> <step>   # Show step details
yoo step complete <note-id> <step>
yoo step uncomplete <note-id> <step>
yoo step note <note-id> <step> "text"
```

### 3. Artifact Management
```bash
yoo artifact list <note-id>
yoo artifact add <note-id> --type input --name X --file path
yoo artifact add <note-id> --type output --name Y --url https://...
yoo artifact show <note-id> <name>
yoo artifact delete <note-id> <name>
yoo artifact open <note-id> <name>
```

### 4. Record Management (from Phase 1.5)
```bash
yoo records add <note-id> --field name=value
yoo records list <note-id>
yoo records edit <note-id> <index> --field name=newvalue
yoo records delete <note-id> <index>
yoo records export <note-id> --format csv
```

## Complete Workflow Example: Job Applications

### Step 1: Create a Templated Note

```bash
./bin/yoo add "Apply to 10 tech jobs in January 2024" \
  --template "Job Applications" \
  --input target_count=10 \
  --input resume=~/Documents/resume-2024.pdf
```

**Output:**
```
✓ Created templated note: Apply to 10 tech jobs in January 2024
  Template: Job Applications (v1.0.0)
  Note ID: 4
  Steps: 4
  Record schema: 10 fields
  Add records with: yoo records add 4
  Complete steps with: yoo step complete 4 <step-number>
  Outputs required: 2
```

### Step 2: View Your Workflow Steps

```bash
./bin/yoo step list 4
```

**Output:**
```
Steps for: Apply to 10 tech jobs in January 2024

STATUS   STEP   TITLE                    COMPLETED
------   ----   -----                    ---------
○        1      Find job opportunities
○        2      Customize applications
○        3      Submit applications
○        4      Document applications

Progress: 0/4 steps completed (0%)
```

### Step 3: Track Your Job Applications (Records)

```bash
# Application #1
./bin/yoo records add 4 \
  --field company="Acme Corp" \
  --field position="Senior Software Engineer" \
  --field date="2024-01-15" \
  --field status="Applied" \
  --field url="https://jobs.acme.com/senior-eng"

# Application #2
./bin/yoo records add 4 \
  --field company="TechStartup Inc" \
  --field position="Lead Developer" \
  --field date="2024-01-16" \
  --field status="Applied"

# Application #3 (interactive mode)
./bin/yoo records add 4 --interactive
```

### Step 4: View Your Applications

```bash
./bin/yoo records list 4
```

**Output:**
```
Records for: Apply to 10 tech jobs in January 2024

#    company          date        position                  status     Status
---  ---              ---         ---                       ---        ---
1    Acme Corp        2024-01-15  Senior Software Engineer  Applied    draft
2    TechStartup Inc  2024-01-16  Lead Developer            Applied    draft
3    BigTech Co       2024-01-17  Staff Engineer            Interview  draft

Total: 3 record(s)
```

### Step 5: Complete Workflow Steps

```bash
# Mark step 1 as complete
./bin/yoo step complete 4 1
```

**Output:**
```
✓ Step 1 marked as complete: Find job opportunities
  Note: Apply to 10 tech jobs in January 2024

Progress: 1/4 steps completed (25%)
```

```bash
# Check all steps
./bin/yoo step list 4
```

**Output:**
```
STATUS   STEP   TITLE                    COMPLETED
------   ----   -----                    ---------
✓        1      Find job opportunities   2024-01-27 23:05:12
○        2      Customize applications
○        3      Submit applications
○        4      Document applications

Progress: 1/4 steps completed (25%)
```

### Step 6: Add Step Notes

```bash
./bin/yoo step note 4 1 "Found 15 great opportunities on LinkedIn and Indeed"
```

### Step 7: View Step Details

```bash
./bin/yoo step show 4 1
```

**Output:**
```
Step 1: Find job opportunities

Status: ✓ Completed on 2024-01-27 23:05:12
Description: Search job boards and company sites for relevant positions

Checklist:
  - Search LinkedIn for relevant roles
  - Check Indeed for positions
  - Browse Glassdoor listings
  - Check company career pages directly
  - Review job requirements and fit

Estimated Time: 1-2 hours

Notes:
  Found 15 great opportunities on LinkedIn and Indeed
```

### Step 8: Update Application Progress

```bash
# Got an interview!
./bin/yoo records edit 4 1 --field status="Interview"

# Got an offer!
./bin/yoo records edit 4 3 --field status="Offer"
```

### Step 9: Attach Artifacts

```bash
# Add your resume as input
./bin/yoo artifact add 4 \
  --type input \
  --name resume \
  --file ~/Documents/resume-2024.pdf \
  --description "Updated resume for tech positions" \
  --required

# List artifacts
./bin/yoo artifact list 4
```

**Output:**
```
Artifacts for: Apply to 10 tech jobs in January 2024

INPUTS:
NAME      TYPE    VALUE                             REQUIRED   DESCRIPTION
----      ----    -----                             --------   -----------
✓ resume  file    /home/user/Documents/resume...    [required] Updated resume for tech...

OUTPUTS:
(No output artifacts yet)

Summary: 1 input(s), 0 output(s)
```

### Step 10: Export Your Data

```bash
# Export to CSV
./bin/yoo records export 4 --format csv > my-job-search.csv

# Export to JSON
./bin/yoo records export 4 --format json > my-job-search.json
```

**CSV Output:**
```csv
index,status,company,position,date,status,url
1,draft,Acme Corp,Senior Software Engineer,2024-01-15,Interview,https://jobs.acme.com/senior-eng
2,draft,TechStartup Inc,Lead Developer,2024-01-16,Applied,
3,draft,BigTech Co,Staff Engineer,2024-01-17,Offer,
```

### Step 11: Complete More Steps

```bash
# Mark steps as complete as you progress
./bin/yoo step complete 4 2  # Customized applications
./bin/yoo step complete 4 3  # Submitted applications

# Check progress
./bin/yoo step list 4
```

**Output:**
```
STATUS   STEP   TITLE                    COMPLETED
------   ----   -----                    ---------
✓        1      Find job opportunities   2024-01-27 23:05:12
✓        2      Customize applications   2024-01-27 23:12:45
✓        3      Submit applications      2024-01-27 23:18:30
○        4      Document applications

Progress: 3/4 steps completed (75%)
```

### Step 12: Add Output Artifacts

```bash
# Export your tracking spreadsheet
./bin/yoo records export 4 --format csv --output applications-list.csv

# Add it as an output artifact
./bin/yoo artifact add 4 \
  --type output \
  --name applications_list \
  --file ./applications-list.csv \
  --description "Complete list of all applications submitted" \
  --required

# Mark final step complete
./bin/yoo step complete 4 4
```

**Output:**
```
✓ Step 4 marked as complete: Document applications
  Note: Apply to 10 tech jobs in January 2024

Progress: 4/4 steps completed (100%)
```

## All CLI Commands Reference

### Templated Notes

```bash
# Create from template
yoo add "Title" --template "Template Name" --input key=value --input key2=value2

# List templates
yoo templates list

# Show template details
yoo templates show "Template Name"
```

### Steps

```bash
# List all steps
yoo step list <note-id>

# Show step details (with checklist)
yoo step show <note-id> <step-number>

# Complete a step
yoo step complete <note-id> <step-number>

# Uncomplete a step
yoo step uncomplete <note-id> <step-number>

# Add notes to a step
yoo step note <note-id> <step-number> "Your notes here"
```

### Records (Log Shape)

```bash
# Add a record (quick)
yoo records add <note-id> --field name=value --field name2=value2

# Add a record (interactive)
yoo records add <note-id> --interactive

# List all records
yoo records list <note-id>

# Filter by status
yoo records list <note-id> --status complete

# Edit a record
yoo records edit <note-id> <record-index> --field name=newvalue

# Delete a record
yoo records delete <note-id> <record-index>

# Export records
yoo records export <note-id> --format csv > output.csv
yoo records export <note-id> --format json > output.json
yoo records export <note-id> --format csv --output file.csv
```

### Artifacts (Files & URLs)

```bash
# List artifacts
yoo artifact list <note-id>

# Add a file artifact
yoo artifact add <note-id> \
  --type input \
  --name artifact-name \
  --file /path/to/file.pdf \
  --description "Description" \
  --required

# Add a URL artifact
yoo artifact add <note-id> \
  --type output \
  --name report \
  --url https://example.com/report \
  --description "Final report"

# Show artifact details
yoo artifact show <note-id> <artifact-name>

# Delete an artifact
yoo artifact delete <note-id> <artifact-name>

# Open an artifact (file or URL)
yoo artifact open <note-id> <artifact-name>
```

## Template Input Types

When creating templated notes, inputs are automatically type-converted:

| Template Type | CLI Example | Conversion |
|---------------|-------------|------------|
| `text` | `--input name="John"` | String |
| `integer` | `--input count=10` | Integer (no quotes!) |
| `boolean` | `--input active=true` | Boolean |
| `file` | `--input resume=~/file.pdf` | String (path) |
| `url` | `--input link=https://...` | String |
| `date` | `--input date=2024-01-15` | String |

**Important:** For integer and boolean inputs, don't use quotes!
- ✅ `--input count=10`
- ❌ `--input count="10"` (will fail validation)

## Built-in Templates

### Job Applications
Track multiple job applications with structured workflow.

**Inputs:** `target_count` (integer), `resume` (file)

**Record Schema:** company, position, date, status, url, reference, contact_person, follow_up_date, salary_range, notes

**Steps:** Find opportunities → Customize → Submit → Document

**Outputs:** applications_list (CSV), confirmation_emails (folder)

### Research Task
Conduct comprehensive research with source tracking.

**Inputs:** `topic` (text), `deadline` (date)

**Steps:** Find sources → Read and annotate → Synthesize → Write report

**Outputs:** research_notes, final_report

### Project Milestone
Complete a project milestone with planning and documentation.

**Inputs:** `milestone_name` (text), `deadline` (date)

**Steps:** Plan → Implement → Test → Document → Review

**Outputs:** implementation, documentation, test_results

## Tips & Best Practices

### 1. Use Interactive Mode for Complex Records
```bash
# Instead of typing 10 --field flags:
yoo records add 42 --interactive
```

### 2. Export Records Regularly
```bash
# Daily backup
yoo records export 42 --format json > backup-$(date +%Y%m%d).json
```

### 3. Add Step Notes as You Go
```bash
yoo step note 4 2 "Used template from last year, updated skills section"
```

### 4. Use Absolute Paths for Artifacts
```bash
# ✅ Good
yoo artifact add 4 --type input --name resume --file ~/Documents/resume.pdf

# ❌ Avoid
yoo artifact add 4 --type input --name resume --file ../resume.pdf
```

### 5. Filter Records by Status
```bash
# Show only interviews
yoo records list 42 --status Interview

# Show completed applications
yoo records list 42 --status complete
```

### 6. Open Artifacts Quickly
```bash
# Opens in default app (PDF viewer, browser, etc.)
yoo artifact open 4 resume
```

## Common Workflows

### Simple Task List (Records Only)

```bash
# Create note (Phase 3 will add templates without steps)
yoo add "Shopping list" --template "Simple List" --input category=groceries

# Add items
yoo records add 5 --field item="Milk" --field purchased=false
yoo records add 5 --field item="Bread" --field purchased=false
yoo records add 5 --field item="Eggs" --field purchased=false

# Mark as purchased
yoo records edit 5 1 --field purchased=true
```

### Research Paper Tracking

```bash
# Create note
yoo add "Literature review on ML" --template "Research Task" --input topic="Machine Learning"

# Add papers as records
yoo records add 6 \
  --field title="Attention Is All You Need" \
  --field authors="Vaswani et al." \
  --field year=2017 \
  --field status="To Read"

# Complete steps as you progress
yoo step complete 6 1  # Found sources
yoo step complete 6 2  # Read and annotated
```

### Content Creation Pipeline

```bash
# Create note
yoo add "Write blog post on templates" --template "Content Creation"

# Track progress through steps
yoo step list 7
yoo step complete 7 1  # Outline
yoo step complete 7 2  # Draft

# Attach drafts as artifacts
yoo artifact add 7 --type output --name draft --file ~/blog-draft.md
```

## Troubleshooting

### "Template not found"
Make sure the template exists:
```bash
yoo templates list
yoo templates show "Template Name"
```

### "Input validation failed"
Check input types (integers don't use quotes):
```bash
# ❌ Wrong
--input count="10"

# ✅ Correct
--input count=10
```

### "Required field missing"
Use `--interactive` mode to see all required fields:
```bash
yoo records add 42 --interactive
```

### "Note is not templated"
You can only use step/artifact commands on templated notes. Check:
```bash
sqlite3 ~/.local/share/yoo/yoo.db "SELECT id, title, is_templated FROM notes"
```

## What's Next?

### Phase 3 (TUI Integration) - Coming Soon
- Visual table editor for records
- Interactive step completion
- Artifact browser
- Progress visualization

### Phase 4 (Advanced Features) - Future
- Template variables
- Conditional fields
- Auto-calculated fields
- Template marketplace

## Summary

Phase 2 provides a complete CLI workflow for orchestrating structured work:

1. **Create** templated notes with inputs
2. **Track** repeating data with database records
3. **Progress** through workflow steps
4. **Attach** files and outputs as artifacts
5. **Export** your data anytime

All three template shapes work together to give you a powerful orchestration tool!

---

**Questions?** Check `docs/TEMPLATE_SHAPES.md` for architecture details or `docs/QUICKSTART_RECORDS.md` for a quick tutorial.