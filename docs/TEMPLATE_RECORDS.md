# Template Records (Log Shape) - User Guide

## Overview

Template records provide a powerful way to store **repeating structured data** directly in your yoo database. Instead of managing external CSV files or spreadsheets, you can define a schema in your template and store records as first-class data.

Think of template records as having a **mini-database table** for each templated note. Perfect for:
- Job applications tracking
- Contact lists
- Expense tracking
- Project task lists
- Research paper notes
- Meeting attendance
- Any tabular/repeating data

## Key Concepts

### Record Schema
A **record schema** defines the structure of your repeating data. It's like defining columns in a spreadsheet:

```yaml
record_schema:
  fields:
    - name: "company"
      type: "text"
      required: true
    - name: "date"
      type: "date"
      required: true
    - name: "status"
      type: "enum"
      values: ["Applied", "Interview", "Offer", "Rejected"]
      default: "Applied"
```

### Records vs. Artifacts

| Feature | Records (Database) | Artifacts (Files) |
|---------|-------------------|-------------------|
| **Storage** | SQLite database | File system |
| **Queryable** | Yes, with SQL | No |
| **Structure** | Defined schema | Free-form |
| **Use Case** | Tabular/repeating data | Documents, PDFs, images |
| **Exportable** | Yes (CSV/JSON) | Already files |
| **Example** | Job applications list | Resume PDF |

**You can use both!** A job search note might have:
- **Records**: 10 job applications (company, position, date, status)
- **Artifacts**: Your resume.pdf, cover-letter-template.docx
- **Steps**: Find jobs → Customize → Apply → Document

## Field Types

Templates support the following field types:

| Type | Description | Example |
|------|-------------|---------|
| `text` | Free-form text | "Acme Corporation" |
| `integer` | Whole numbers | 50000 |
| `date` | Date string | "2024-01-15" |
| `enum` | Predefined choices | "Applied" (from list) |
| `url` | Web URL | "https://jobs.acme.com/123" |
| `boolean` | True/false | true |

### Enum Fields
Enum fields are powerful for tracking status and categories:

```yaml
- name: "status"
  type: "enum"
  required: true
  default: "Found"
  values:
    - "Found"
    - "Applied"
    - "Interview"
    - "Offer"
    - "Rejected"
    - "Withdrawn"
```

## CLI Commands

### Add a Record

```bash
# With flags (quick)
yoo records add <note-id> \
  --field company="Acme Corp" \
  --field position="Senior Developer" \
  --field date="2024-01-15" \
  --field status="Applied"

# Interactive mode (guided)
yoo records add <note-id> --interactive
```

### List Records

```bash
# List all records for a note
yoo records list <note-id>

# Filter by status
yoo records list <note-id> --status complete
yoo records list <note-id> --status draft
```

### Edit a Record

```bash
# Update specific fields
yoo records edit <note-id> <record-index> \
  --field status="Interview" \
  --field notes="Phone screen scheduled for Friday"

# Example: Update job application #3
yoo records edit 42 3 --field status="Interview"
```

### Delete a Record

```bash
yoo records delete <note-id> <record-index>

# Example: Delete application #5
yoo records delete 42 5
```

### Export Records

```bash
# Export to CSV (stdout)
yoo records export <note-id> --format csv

# Save to file
yoo records export <note-id> --format csv > applications.csv
yoo records export <note-id> --format csv --output applications.csv

# Export as JSON
yoo records export <note-id> --format json > data.json
```

## Complete Workflow Example

Let's walk through using the built-in "Job Applications" template:

### Step 1: Create a Templated Note

```bash
# First, create a note that uses the Job Applications template
# (Note: Phase 2 of template implementation will add `yoo add --template`)
# For now, you'd create via the database directly or wait for Phase 2

# We'll assume note ID 42 for this example
```

### Step 2: Add Job Applications (Records)

```bash
# Application #1
yoo records add 42 \
  --field company="Acme Corp" \
  --field position="Senior Software Engineer" \
  --field date="2024-01-15" \
  --field status="Applied" \
  --field url="https://jobs.acme.com/senior-eng-123" \
  --field reference="APP-2024-001"

# Application #2
yoo records add 42 \
  --field company="TechStartup Inc" \
  --field position="Lead Developer" \
  --field date="2024-01-16" \
  --field status="Applied" \
  --field url="https://techstartup.com/careers/lead-dev"

# Application #3 (interactive mode)
yoo records add 42 --interactive
> company: GlobalTech
> position: Engineering Manager
> date: 2024-01-17
> status: Interview
> url: https://globaltech.com/job/123
> reference: REF-456
> ...
```

### Step 3: Track Progress

```bash
# View all applications
yoo records list 42

# Output:
# Records for: Apply to 10 tech jobs
#
# #    company              position                date        status     url                           ...
# ---  ---                  ---                     ---         ---        ---                           ...
# 1    Acme Corp            Senior Software Eng...  2024-01-15  Applied    https://jobs.acme.com/sen...  ...
# 2    TechStartup Inc      Lead Developer          2024-01-16  Applied    https://techstartup.com/c...  ...
# 3    GlobalTech           Engineering Manager     2024-01-17  Interview  https://globaltech.com/jo...  ...
#
# Total: 3 record(s)

# Filter for interviews only
yoo records list 42 --status Interview
```

### Step 4: Update as You Progress

```bash
# Got an interview for application #1!
yoo records edit 42 1 \
  --field status="Interview" \
  --field follow_up_date="2024-01-22" \
  --field notes="Technical phone screen scheduled"

# Got rejected from application #2
yoo records edit 42 2 --field status="Rejected"

# Got an offer from #3!
yoo records edit 42 3 \
  --field status="Offer" \
  --field salary_range="$120k-$140k" \
  --field notes="Verbal offer received, awaiting written offer"
```

### Step 5: Export Your Data

```bash
# Export to CSV for sharing or backup
yoo records export 42 --format csv > my-job-search-2024.csv

# Open in Excel/Google Sheets
# Share with accountability partner
# Analyze your application success rate
```

## Creating Templates with Record Schemas

### Simple Example: Contact List

```yaml
name: "Contact List"
version: "1.0.0"
description: "Maintain a list of professional contacts"
category: "networking"

inputs:
  - name: "context"
    type: "text"
    description: "Context for this contact list (e.g., conference, project)"
    required: true

record_schema:
  fields:
    - name: "name"
      type: "text"
      required: true
    
    - name: "email"
      type: "text"
      required: true
    
    - name: "company"
      type: "text"
      required: false
    
    - name: "role"
      type: "text"
      required: false
    
    - name: "met_date"
      type: "date"
      required: true
    
    - name: "context"
      type: "text"
      required: false
    
    - name: "follow_up"
      type: "boolean"
      default: false

steps:
  - id: 1
    title: "Collect contact information"
    description: "Gather names and emails from event/project"
  
  - id: 2
    title: "Send follow-ups"
    description: "Reach out to contacts marked for follow-up"

metadata:
  tags: ["networking", "contacts"]
  estimated_duration: "1-2 hours"
```

### Advanced Example: Research Paper Tracker

```yaml
name: "Research Papers"
version: "1.0.0"
description: "Track papers to read for literature review"
category: "research"

inputs:
  - name: "topic"
    type: "text"
    description: "Research topic"
    required: true
  
  - name: "target_count"
    type: "integer"
    description: "Number of papers to review"
    required: true

record_schema:
  fields:
    - name: "title"
      type: "text"
      required: true
    
    - name: "authors"
      type: "text"
      required: true
    
    - name: "year"
      type: "integer"
      required: true
    
    - name: "venue"
      type: "text"
      description: "Conference or journal name"
      required: false
    
    - name: "url"
      type: "url"
      required: false
    
    - name: "status"
      type: "enum"
      required: true
      default: "To Read"
      values:
        - "To Read"
        - "Reading"
        - "Read"
        - "Cited"
        - "Not Relevant"
    
    - name: "relevance"
      type: "enum"
      required: false
      values:
        - "High"
        - "Medium"
        - "Low"
    
    - name: "notes"
      type: "text"
      required: false

steps:
  - id: 1
    title: "Find papers"
    description: "Search databases and add papers to list"
  
  - id: 2
    title: "Read and annotate"
    description: "Read each paper and take notes"
  
  - id: 3
    title: "Synthesize findings"
    description: "Write summary of key insights"

outputs:
  - name: "literature_review"
    type: "file"
    description: "Written literature review document"
    format: "markdown|docx|pdf"
    required: true

metadata:
  tags: ["research", "academic", "papers"]
  estimated_duration: "2-4 weeks"
```

## Hybrid Templates: Records + Steps + Artifacts

The most powerful templates combine all three shapes:

```yaml
name: "Product Launch"
version: "1.0.0"
description: "Launch a new product with feature tracking"

# Traditional inputs
inputs:
  - name: "product_name"
    type: "text"
    required: true
  - name: "launch_date"
    type: "date"
    required: true

# LOG SHAPE: Track each feature/task
record_schema:
  fields:
    - name: "feature"
      type: "text"
      required: true
    - name: "owner"
      type: "text"
      required: true
    - name: "status"
      type: "enum"
      values: ["Not Started", "In Progress", "Done", "Blocked"]
      default: "Not Started"
    - name: "priority"
      type: "enum"
      values: ["P0", "P1", "P2", "P3"]

# CHECKLIST SHAPE: Overall workflow steps
steps:
  - id: 1
    title: "Define features"
    description: "Create feature list (add as records)"
  
  - id: 2
    title: "Build features"
    description: "Implement all features, tracking status in records"
  
  - id: 3
    title: "Create marketing materials"
    description: "Write copy, design graphics"
    output_required: "marketing_assets"
  
  - id: 4
    title: "Launch"
    description: "Deploy to production and announce"

# ARTIFACT SHAPE: Deliverable files
outputs:
  - name: "marketing_assets"
    type: "folder"
    required: true
  
  - name: "launch_announcement"
    type: "file"
    format: "markdown"
    required: true

metadata:
  tags: ["product", "launch", "project-management"]
  estimated_duration: "4-8 weeks"
```

Usage:
```bash
# Add feature records
yoo records add 50 --field feature="User authentication" --field owner="Alice" --field priority="P0"
yoo records add 50 --field feature="Dashboard redesign" --field owner="Bob" --field priority="P1"

# View feature status
yoo records list 50

# Update as work progresses
yoo records edit 50 1 --field status="Done"

# Complete workflow steps
yoo step complete 50 1
yoo step complete 50 2

# Attach artifacts
yoo artifact add 50 --type output --name marketing_assets --file ./assets/

# Export feature list for stakeholders
yoo records export 50 --format csv > launch-features.csv
```

## Best Practices

### When to Use Records
✅ **Use records when:**
- Data is structured and repeating
- You want to query/filter the data
- You need to track status across many items
- Data naturally fits in a table
- You want to export to CSV/spreadsheet

Examples: job applications, contacts, expenses, task lists, paper citations

### When to Use Artifacts
✅ **Use artifacts when:**
- You have actual files (PDFs, images, documents)
- Data is unstructured or free-form
- You need to keep the original format
- Files are created by external tools

Examples: resume, photos, reports, generated documents

### When to Use Steps
✅ **Use steps when:**
- You have a sequential workflow
- You need checklist-style tracking
- Process is more important than data
- You want guided execution

Examples: research process, application workflow, review process

### Combining All Three
The most powerful templates use **all three shapes** together:
- **Records** = What you're tracking (the data)
- **Steps** = How you do it (the process)
- **Artifacts** = What you produce (the deliverables)

## Database Queries (Advanced)

Since records are stored in SQLite, you can write custom queries:

```bash
# Connect to your database
sqlite3 ~/.local/share/yoo/yoo.db

# Find all "Interview" status applications across ALL notes
SELECT 
    n.title,
    tr.record_index,
    json_extract(tr.data, '$.company') as company,
    json_extract(tr.data, '$.status') as status
FROM template_records tr
JOIN note_templates nt ON tr.note_template_id = nt.id
JOIN notes n ON nt.note_id = n.id
WHERE json_extract(tr.data, '$.status') = 'Interview';

# Count applications by status
SELECT 
    json_extract(tr.data, '$.status') as status,
    COUNT(*) as count
FROM template_records tr
GROUP BY status;

# Find recently updated records
SELECT * FROM template_records 
WHERE updated_at > datetime('now', '-7 days')
ORDER BY updated_at DESC;
```

## Tips & Tricks

### 1. Use Enum Fields for Consistency
Define status values upfront to avoid typos:
```yaml
- name: "status"
  type: "enum"
  values: ["New", "In Progress", "Done"]  # Not "new", "in progress", "done"
```

### 2. Set Smart Defaults
```yaml
- name: "date"
  type: "date"
  required: true
  default: "today"  # Or specific date

- name: "priority"
  type: "enum"
  default: "Medium"
  values: ["Low", "Medium", "High"]
```

### 3. Keep Field Names Simple
Use `snake_case` without spaces:
- ✅ `follow_up_date`
- ❌ `Follow-Up Date`

### 4. Export Regularly
```bash
# Daily backup
yoo records export 42 --format json > backup-$(date +%Y%m%d).json

# Weekly CSV for review
yoo records export 42 --format csv > weekly-review.csv
```

### 5. Interactive Mode for Complex Records
For records with many fields, use interactive mode:
```bash
yoo records add 42 --interactive
# Much easier than typing 10 --field flags!
```

## Troubleshooting

### "Template does not support records"
The template doesn't have a `record_schema` section. Not all templates support records.

### "Required field 'X' is missing"
You must provide all required fields. Check template definition:
```bash
yoo templates show "Template Name"
```

### "Field 'X' must be an integer"
Type mismatch. Use quotes for text, not for numbers:
```bash
# ❌ Wrong
--field year="2024"

# ✅ Correct
--field year=2024
```

### "Unknown field 'X'"
Field name doesn't exist in schema. Check spelling and available fields.

## Next Steps

- **Phase 2**: Create templated notes with `yoo add --template "Job Applications"`
- **Phase 3**: TUI support for viewing/editing records in a visual table
- **Phase 4**: Template variables, conditional fields, auto-calculations

---

**Questions or feedback?** Open an issue on GitHub or check the main documentation.