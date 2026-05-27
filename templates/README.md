# Note Templates

This directory contains reusable templates for common note patterns. Templates help you structure complex tasks with inputs, procedural steps, and expected outputs.

## What Are Templates?

Templates transform simple notes into structured workflows. They're perfect for:

- **Repetitive tasks** with consistent steps (e.g., "Apply to 10 jobs")
- **Complex projects** requiring inputs and outputs (e.g., "Research project")
- **Accountability** when you need to document and share results
- **Procedural work** with defined steps to completion

## Available Templates

### 1. Job Applications (`job-applications.yaml`)

**Use Case**: Apply to multiple positions with documentation

**Inputs**:
- Target count (how many applications)
- Resume file path
- Cover letter template (optional)

**Steps**:
1. Find job opportunities
2. Customize applications
3. Submit applications
4. Document applications

**Outputs**:
- Applications list (spreadsheet for tracking)
- Confirmation emails folder

**Example**:
```bash
yoo add "Apply to 10 tech jobs" \
  --template job-applications \
  --input target_count=10 \
  --input resume=~/Documents/resume.pdf
```

### 2. Research Task (Coming Soon)

Conduct research with source tracking and deliverables.

### 3. Project Milestone (Coming Soon)

Complete project milestones with planning and documentation.

## Using Templates

### List Available Templates

```bash
yoo templates list
```

### View Template Details

```bash
yoo templates show job-applications
```

### Create Note from Template

```bash
# Basic usage
yoo add "My task" --template job-applications

# With inputs
yoo add "Apply to 10 jobs" \
  --template job-applications \
  --input target_count=10 \
  --input resume=~/resume.pdf

# Interactive mode (prompts for inputs)
yoo add --template job-applications --interactive
```

### Working with Templated Notes

```bash
# View note with full template structure
yoo show <note-id>

# Complete a step
yoo step complete <note-id> <step-number>

# Add an output artifact
yoo artifact add <note-id> \
  --type output \
  --name applications_list \
  --file ~/job-applications.xlsx

# Add external reference
yoo ref add <note-id> --url https://linkedin.com/jobs
yoo ref add <note-id> --file ~/notes.txt

# Check progress
yoo progress <note-id>
```

## Creating Custom Templates

### Template Structure

Templates are YAML files with the following structure:

```yaml
name: "Template Name"
version: "1.0.0"
description: "What this template helps you accomplish"
category: "career|research|project|personal"

inputs:
  - name: "input_name"
    type: "file|text|integer|url"
    description: "What this input is for"
    required: true|false
    default: "optional default value"

steps:
  - id: 1
    title: "Step title"
    description: "What to do in this step"
    checklist:
      - "Specific action 1"
      - "Specific action 2"
    estimated_time: "30 minutes"
    output_required: "optional_output_name"

outputs:
  - name: "output_name"
    type: "file|folder|text|url"
    description: "What this output represents"
    format: "csv|xlsx|pdf|markdown"
    required: true|false

metadata:
  tags: ["tag1", "tag2"]
  estimated_duration: "2-4 hours"
  difficulty: "easy|medium|hard"
```

### Example: Weekly Review Template

```yaml
name: "Weekly Review"
version: "1.0.0"
description: "Weekly reflection and planning session"
category: "personal"

inputs:
  - name: "week_number"
    type: "integer"
    description: "Week number of the year"
    required: true

steps:
  - id: 1
    title: "Review last week"
    description: "Go through completed tasks and outcomes"
    checklist:
      - "Check completed notes"
      - "Review calendar events"
      - "Note achievements"
  
  - id: 2
    title: "Identify wins and challenges"
    description: "What went well? What was difficult?"
  
  - id: 3
    title: "Plan next week"
    description: "Set priorities and goals"
    checklist:
      - "List top 3 priorities"
      - "Block calendar time"
      - "Set specific outcomes"

outputs:
  - name: "weekly_summary"
    type: "file"
    description: "Summary of review and plans"
    format: "markdown"
    required: true

metadata:
  tags: ["reflection", "planning"]
  estimated_duration: "30-60 minutes"
  difficulty: "easy"
```

### Import Custom Template

```bash
# Import from file
yoo template import my-template.yaml

# Edit existing template
yoo template edit weekly-review

# Export template for sharing
yoo template export weekly-review > shared-template.yaml
```

## Template Design Guidelines

### When to Create a Template

✅ **Good candidates for templates**:
- Tasks you do repeatedly with consistent steps
- Complex workflows requiring documentation
- Tasks with clear inputs, steps, and outputs
- Work that needs to be shared/reviewed by others

❌ **Don't need templates**:
- Simple one-off reminders
- Tasks with unpredictable structure
- Very quick actions (< 5 minutes)

### Template Design Tips

1. **Keep Steps Actionable**: Each step should be clear and specific
2. **Use Checklists**: Break complex steps into sub-actions
3. **Define Required vs Optional**: Mark what's essential vs nice-to-have
4. **Add Examples**: Include usage examples in the template
5. **Estimate Time**: Help users plan by noting time requirements
6. **Think About Outputs**: What artifact proves completion?

### Input Types

- `file`: Path to a file on disk
- `folder`: Path to a directory
- `text`: Simple text value
- `integer`: Numeric value
- `url`: Web link
- `boolean`: Yes/no flag

### Output Types

- `file`: Single file deliverable
- `folder`: Collection of files
- `text`: Text content or summary
- `url`: Link to online resource

## Real-World Examples

### Example 1: Job Applications

```bash
# Day 1: Create the note
yoo add "Apply to 10 senior engineer roles" \
  --template job-applications \
  --input target_count=10 \
  --input resume=~/Documents/resume-senior.pdf

# Day 1-3: Complete steps as you work
yoo step complete 42 1  # Found 15 good opportunities
yoo step complete 42 2  # Customized applications
yoo step complete 42 3  # Submitted all 10

# Day 3: Add proof of completion
yoo artifact add 42 \
  --type output \
  --name applications_list \
  --file ~/JobSearch/applications-march-2024.xlsx

yoo artifact add 42 \
  --type output \
  --name confirmation_emails \
  --folder ~/JobSearch/confirmations/

# Check progress
yoo progress 42
# Output: ✓ 100% complete (4/4 steps, 2/2 outputs)
```

### Example 2: Research Project

```bash
# Create research note
yoo add "Research GraphQL performance" \
  --template research-task \
  --input topic="GraphQL query optimization" \
  --input deadline="2024-06-15"

# Add references as you find them
yoo ref add 43 --url https://graphql.org/learn/queries/
yoo ref add 43 --file ~/Papers/graphql-optimization.pdf
yoo ref add 43 --url https://github.com/graphql/dataloader

# Complete research steps
yoo step complete 43 1  # Literature review
yoo step complete 43 2  # Data collection
yoo step complete 43 3  # Analysis

# Add final deliverable
yoo artifact add 43 \
  --type output \
  --name research_summary \
  --file ~/Research/graphql-performance-summary.md
```

## Template Lifecycle

1. **Create**: Design template with clear structure
2. **Import**: Add to yoo template library
3. **Use**: Create notes from template
4. **Refine**: Update based on usage experience
5. **Share**: Export and share with others

## Future Features

- **Template Marketplace**: Share templates with the community
- **Template Variables**: Use `{input_name}` in step descriptions
- **Conditional Steps**: Skip steps based on inputs
- **Template Inheritance**: Extend existing templates
- **External Tool Hooks**: Trigger commands on completion

## Contributing Templates

Have a great template? Share it with the community!

1. Create your template following the structure above
2. Test it thoroughly
3. Add clear documentation and examples
4. Submit via GitHub pull request to `templates/`

Good template contributions should:
- Solve a common, real-world problem
- Be well-documented with examples
- Include reasonable defaults
- Have clear completion criteria

## Support

- **Documentation**: See `docs/TEMPLATES_DESIGN.md` for detailed design
- **Issues**: Report template bugs or suggestions on GitHub
- **Questions**: Open a discussion on the repository

---

**Note**: Template system is currently in design phase. This directory shows the intended structure and usage patterns. Implementation coming in v0.2.0+.