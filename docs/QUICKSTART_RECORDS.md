# Quick Start: Template Records

Get started with database-native records in 5 minutes.

## What Are Records?

Records let you store **repeating structured data** directly in yoo's database instead of external files. Perfect for tracking job applications, contacts, expenses, or any tabular data.

Think of it as having a mini-spreadsheet for each note, but queryable and integrated into your workflow.

## Example: Track 10 Job Applications

### Step 1: Check Built-in Templates

```bash
./bin/yoo templates list
```

You should see "Job Applications" - it already has a record schema defined.

### Step 2: View the Template

```bash
./bin/yoo templates show "Job Applications"
```

Look for the record schema fields: `company`, `position`, `date`, `status`, etc.

### Step 3: Create a Note (Manual for Now)

Phase 2 will add `yoo add --template`, but for now, create a regular note:

```bash
./bin/yoo add "Apply to 10 tech jobs in January"
```

Note the ID (let's say it's `42`).

### Step 4: Add Your First Application

```bash
./bin/yoo records add 42 \
  --field company="Acme Corp" \
  --field position="Senior Developer" \
  --field date="2024-01-15" \
  --field status="Applied" \
  --field url="https://jobs.acme.com/senior-dev"
```

Or use interactive mode:

```bash
./bin/yoo records add 42 --interactive
```

### Step 5: Add More Applications

```bash
./bin/yoo records add 42 \
  --field company="TechStartup Inc" \
  --field position="Lead Engineer" \
  --field date="2024-01-16" \
  --field status="Applied"

./bin/yoo records add 42 \
  --field company="BigTech Co" \
  --field position="Staff Engineer" \
  --field date="2024-01-17" \
  --field status="Interview"
```

Repeat until you have 10 applications.

### Step 6: View Your Applications

```bash
./bin/yoo records list 42
```

You'll see a nice table:

```
Records for: Apply to 10 tech jobs in January

#    company              position              date        status     url
---  ---                  ---                   ---         ---        ---
1    Acme Corp            Senior Developer      2024-01-15  Applied    https://jobs.acme.com/...
2    TechStartup Inc      Lead Engineer         2024-01-16  Applied    
3    BigTech Co           Staff Engineer        2024-01-17  Interview  

Total: 3 record(s)
```

### Step 7: Update Progress

```bash
# Got an interview!
./bin/yoo records edit 42 1 --field status="Interview"

# Got rejected :(
./bin/yoo records edit 42 2 --field status="Rejected"

# Got an offer!
./bin/yoo records edit 42 3 --field status="Offer"
```

### Step 8: Filter and Query

```bash
# Show only interviews
./bin/yoo records list 42 --status Interview

# Show only completed applications
./bin/yoo records list 42 --status complete
```

### Step 9: Export Your Data

```bash
# Export to CSV
./bin/yoo records export 42 --format csv > my-job-search.csv

# Export to JSON
./bin/yoo records export 42 --format json > my-job-search.json

# Open in Excel/Google Sheets
open my-job-search.csv
```

## Common Operations

### Add a record (quick)
```bash
yoo records add <note-id> --field name="value" --field name2="value2"
```

### Add a record (interactive)
```bash
yoo records add <note-id> --interactive
```

### List all records
```bash
yoo records list <note-id>
```

### Update a record
```bash
yoo records edit <note-id> <record-index> --field name="new value"
```

### Delete a record
```bash
yoo records delete <note-id> <record-index>
```

### Export to CSV
```bash
yoo records export <note-id> --format csv > output.csv
```

## Field Types

When adding records, remember the types:

- **text**: `--field company="Acme Corp"`
- **integer**: `--field count=10` (no quotes)
- **date**: `--field date="2024-01-15"`
- **boolean**: `--field active=true` (no quotes)
- **enum**: `--field status="Applied"` (must match template values)

## Tips

### Use Defaults
Many fields have defaults. Check the template:
```bash
yoo templates show "Job Applications"
```

### Interactive Mode for Complex Records
If a template has 10+ fields, use interactive mode:
```bash
yoo records add 42 --interactive
```

### Export Regularly
Keep backups:
```bash
yoo records export 42 --format json > backup-$(date +%Y%m%d).json
```

### Query from SQLite (Advanced)
```bash
sqlite3 ~/.local/share/yoo/yoo.db

SELECT 
  json_extract(data, '$.company') as company,
  json_extract(data, '$.status') as status
FROM template_records
WHERE json_extract(data, '$.status') = 'Interview';
```

## What's Next?

- **Phase 2**: Create templated notes directly with `yoo add --template`
- **Phase 3**: TUI support for visual record editing
- **Read more**: 
  - `docs/TEMPLATE_RECORDS.md` - Complete guide
  - `docs/TEMPLATE_SHAPES.md` - Architecture overview

## Troubleshooting

**"Template does not support records"**
→ Not all templates have `record_schema`. Check with `yoo templates show <name>`

**"Required field missing"**
→ You must provide all required fields. Use `--interactive` to see prompts.

**"Invalid field type"**
→ Numbers don't use quotes: `--field count=10` not `--field count="10"`

## Create Your Own Template

Create `my-template.yaml`:

```yaml
name: "My Template"
version: "1.0.0"
description: "Track my things"
category: "custom"

record_schema:
  fields:
    - name: "title"
      type: "text"
      required: true
    
    - name: "status"
      type: "enum"
      values: ["New", "In Progress", "Done"]
      default: "New"
    
    - name: "notes"
      type: "text"
      required: false

steps:
  - id: 1
    title: "Add items to list"
  
  - id: 2
    title: "Complete items"

metadata:
  tags: ["custom"]
```

Import it:
```bash
yoo templates import my-template.yaml
```

---

**Questions?** Check the full docs or open an issue on GitHub.