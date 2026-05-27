# Template System Implementation Roadmap

**Project**: yoo Note Templates & Workflow System  
**Status**: Planning  
**Target Version**: v0.2.0 - v0.5.0  
**Last Updated**: 2026-05-27

---

## Overview

This roadmap outlines the implementation of the template system for yoo, transforming it from a simple note manager into a structured workflow tool while maintaining simplicity for basic use cases.

**Goal**: Enable users to create structured, repeatable workflows with inputs, steps, and outputs while keeping simple notes simple.

---

## Implementation Phases

### Phase 1: Foundation (v0.2.0) - Weeks 1-3

**Goal**: Establish core template infrastructure and data models.

#### Database Schema Migration

**Priority**: Critical  
**Estimated Time**: 1 week

- [ ] Create migration system for schema versioning
- [ ] Add new tables:
  - `templates` - Template definitions
  - `note_templates` - Link notes to templates
  - `note_steps` - Step tracking
  - `artifacts` - Input/output artifacts
  - `note_references` - External references
- [ ] Modify `notes` table:
  - Add `is_templated` column
  - Add `template_progress` column
- [ ] Create indexes for performance
- [ ] Write migration rollback logic
- [ ] Test migration on existing databases

**Files to Create/Modify**:
- `internal/database/migrations.go` (new)
- `internal/database/migrations/001_add_templates.sql` (new)
- `internal/database/db.go` (modify)

#### Template Data Models

**Priority**: Critical  
**Estimated Time**: 1 week

- [ ] Define Go structs for templates
  - `Template`
  - `TemplateDefinition`
  - `TemplateInput`
  - `TemplateStep`
  - `TemplateOutput`
  - `NoteTemplate`
  - `StepInstance`
  - `Artifact`
  - `Reference`
- [ ] Implement JSON/YAML serialization
- [ ] Add validation logic
- [ ] Write unit tests for models

**Files to Create**:
- `internal/models/template.go` (new)
- `internal/models/template_test.go` (new)

#### Template Storage Layer

**Priority**: Critical  
**Estimated Time**: 1 week

- [ ] Implement template CRUD operations
  - `CreateTemplate()`
  - `GetTemplate()`
  - `UpdateTemplate()`
  - `DeleteTemplate()`
  - `ListTemplates()`
- [ ] Implement template-note association
  - `AttachTemplateToNote()`
  - `GetNoteTemplate()`
- [ ] Add step tracking functions
  - `CompleteStep()`
  - `GetStepProgress()`
- [ ] Add artifact management
  - `AddArtifact()`
  - `GetArtifacts()`
- [ ] Write integration tests

**Files to Create**:
- `internal/database/template.go` (new)
- `internal/database/template_test.go` (new)
- `internal/database/artifact.go` (new)
- `internal/database/reference.go` (new)

#### Built-in Templates

**Priority**: High  
**Estimated Time**: 3-5 days

- [ ] Create 3 initial templates:
  - Job Applications
  - Research Task
  - Project Milestone
- [ ] Define template YAML structure
- [ ] Implement template loading from files
- [ ] Add template validation
- [ ] Create template documentation

**Files Already Created**:
- `templates/job-applications.yaml` ✓
- `templates/research-task.yaml` ✓
- `templates/project-milestone.yaml` ✓
- `templates/README.md` ✓

#### Basic CLI Commands

**Priority**: High  
**Estimated Time**: 1 week

- [ ] Implement `yoo templates list`
- [ ] Implement `yoo templates show <name>`
- [ ] Add template flag to `yoo add`
- [ ] Add input flags support
- [ ] Update help documentation

**Files to Create/Modify**:
- `cmd/template.go` (new)
- `cmd/add.go` (modify)

**Deliverables**:
- ✅ Database schema supports templates
- ✅ Template data models defined
- ✅ Basic template CRUD working
- ✅ 3 built-in templates available
- ✅ Can list and view templates via CLI

**Success Criteria**:
- Migration runs successfully on empty and populated databases
- All tests pass
- Can create and retrieve templates from database
- CLI commands respond in < 100ms

---

### Phase 2: Core Functionality (v0.3.0) - Weeks 4-6

**Goal**: Enable full template workflow via CLI.

#### Templated Note Creation

**Priority**: Critical  
**Estimated Time**: 1 week

- [ ] Implement note creation from template
- [ ] Parse and validate input parameters
- [ ] Initialize template instance data
- [ ] Create step instances
- [ ] Link template to note
- [ ] Handle template instantiation errors
- [ ] Add validation for required inputs

**Files to Create/Modify**:
- `internal/template/instantiate.go` (new)
- `cmd/add.go` (modify)

#### Step Management

**Priority**: Critical  
**Estimated Time**: 1 week

- [ ] Implement `yoo step complete <note-id> <step-num>`
- [ ] Implement `yoo step list <note-id>`
- [ ] Implement `yoo step show <note-id> <step-num>`
- [ ] Add step notes/comments
- [ ] Calculate progress percentage
- [ ] Update note status based on step completion

**Files to Create**:
- `cmd/step.go` (new)
- `internal/template/steps.go` (new)

#### Artifact Management

**Priority**: High  
**Estimated Time**: 1 week

- [ ] Implement `yoo artifact add`
- [ ] Implement `yoo artifact list <note-id>`
- [ ] Implement `yoo artifact show <artifact-id>`
- [ ] Support different artifact types (file, url, text)
- [ ] Validate artifact existence (files)
- [ ] Track required vs optional artifacts

**Files to Create**:
- `cmd/artifact.go` (new)
- `internal/template/artifacts.go` (new)

#### Reference Management

**Priority**: Medium  
**Estimated Time**: 3-5 days

- [ ] Implement `yoo ref add <note-id>`
- [ ] Implement `yoo ref list <note-id>`
- [ ] Implement `yoo ref remove <ref-id>`
- [ ] Support file, url, command reference types
- [ ] Add reference descriptions
- [ ] Implement reference opening (xdg-open)

**Files to Create**:
- `cmd/ref.go` (new)

#### Progress Tracking

**Priority**: High  
**Estimated Time**: 3-5 days

- [ ] Implement `yoo progress <note-id>`
- [ ] Calculate overall completion percentage
- [ ] Show steps completed/total
- [ ] Show artifacts provided/required
- [ ] Identify blocking items
- [ ] Display estimated time remaining

**Files to Create/Modify**:
- `cmd/progress.go` (new)
- `internal/template/progress.go` (new)

#### Template Management

**Priority**: Medium  
**Estimated Time**: 1 week

- [ ] Implement `yoo template create`
- [ ] Implement `yoo template edit <name>`
- [ ] Implement `yoo template import <file>`
- [ ] Implement `yoo template export <name>`
- [ ] Implement `yoo template delete <name>`
- [ ] Add template validation
- [ ] Support custom template directories

**Files to Create/Modify**:
- `cmd/template.go` (modify)
- `internal/template/manager.go` (new)

**Deliverables**:
- ✅ Can create notes from templates via CLI
- ✅ Full step management working
- ✅ Artifact tracking implemented
- ✅ Reference system working
- ✅ Progress calculation accurate
- ✅ Template CRUD complete

**Success Criteria**:
- Can complete full workflow (create note, track steps, add artifacts)
- Progress accurately reflects completion state
- All artifact types supported
- Custom templates can be imported

---

### Phase 3: TUI Integration (v0.4.0) - Weeks 7-9

**Goal**: Rich interactive experience in TUI for templated notes.

#### Template Selection View

**Priority**: High  
**Estimated Time**: 1 week

- [ ] Design template selection screen
- [ ] Implement template browser
- [ ] Show template details on select
- [ ] Support filtering by category
- [ ] Add search functionality
- [ ] Handle "no template" option

**Files to Create**:
- `internal/tui/template_selector.go` (new)

#### Templated Note View

**Priority**: Critical  
**Estimated Time**: 2 weeks

- [ ] Design templated note layout
  - Input section
  - Steps section with progress
  - Outputs section
  - References section
- [ ] Implement interactive step completion
- [ ] Show completion checkboxes
- [ ] Display progress bar
- [ ] Highlight required vs optional items
- [ ] Color-code by status (pending, completed)
- [ ] Support keyboard navigation

**Files to Create/Modify**:
- `internal/tui/templated_note.go` (new)
- `internal/tui/schedule.go` (modify)

#### Artifact Management in TUI

**Priority**: High  
**Estimated Time**: 1 week

- [ ] Design artifact input form
- [ ] File picker integration
- [ ] URL input with validation
- [ ] Text input for descriptions
- [ ] Show artifact status indicators
- [ ] Support artifact viewing/opening

**Files to Create**:
- `internal/tui/artifact_form.go` (new)

#### Reference Management in TUI

**Priority**: Medium  
**Estimated Time**: 3-5 days

- [ ] Add reference list view
- [ ] Implement reference addition form
- [ ] Support opening references
- [ ] Show reference types with icons
- [ ] Allow reference editing/deletion

**Files to Create**:
- `internal/tui/reference_list.go` (new)

#### Progress Visualization

**Priority**: Medium  
**Estimated Time**: 3-5 days

- [ ] Design progress dashboard
- [ ] Show overall completion percentage
- [ ] Display step-by-step progress
- [ ] Show artifact completion status
- [ ] Add time estimates
- [ ] Create progress chart/graph

**Files to Create**:
- `internal/tui/progress_view.go` (new)

**Deliverables**:
- ✅ Template selection in TUI
- ✅ Rich templated note view
- ✅ Interactive step management
- ✅ Artifact and reference management
- ✅ Visual progress tracking

**Success Criteria**:
- TUI feels responsive (< 50ms updates)
- Navigation is intuitive
- All keyboard shortcuts documented
- Progress indicators are clear
- Works on different terminal sizes

---

### Phase 4: Advanced Features (v0.5.0) - Weeks 10-14

**Goal**: Power user features and extensibility.

#### Interactive Template Creation

**Priority**: High  
**Estimated Time**: 2 weeks

- [ ] Design template creation wizard
- [ ] Interactive step builder
- [ ] Input definition interface
- [ ] Output specification interface
- [ ] Template validation preview
- [ ] Save to custom location
- [ ] Generate template from existing note

**Files to Create**:
- `internal/tui/template_wizard.go` (new)
- `cmd/template_wizard.go` (new)

#### Template Variables

**Priority**: Medium  
**Estimated Time**: 1 week

- [ ] Define variable syntax `{variable_name}`
- [ ] Implement variable substitution in steps
- [ ] Support variables in descriptions
- [ ] Add variable validation
- [ ] Document variable usage

**Files to Create/Modify**:
- `internal/template/variables.go` (new)
- `internal/template/instantiate.go` (modify)

#### Conditional Steps

**Priority**: Medium  
**Estimated Time**: 1 week

- [ ] Define condition syntax
- [ ] Implement condition evaluation
- [ ] Skip steps based on conditions
- [ ] Update progress calculation
- [ ] Add condition examples

**Files to Create**:
- `internal/template/conditions.go` (new)

#### External Tool Hooks

**Priority**: Medium  
**Estimated Time**: 1 week

- [ ] Define hook syntax
- [ ] Implement pre/post step hooks
- [ ] Support shell command execution
- [ ] Add hook timeout handling
- [ ] Capture hook output
- [ ] Document security considerations

**Files to Create**:
- `internal/template/hooks.go` (new)

#### Template Analytics

**Priority**: Low  
**Estimated Time**: 1 week

- [ ] Track template usage statistics
- [ ] Calculate average completion time
- [ ] Identify popular templates
- [ ] Show template success rate
- [ ] Generate usage reports

**Files to Create**:
- `internal/analytics/template_stats.go` (new)
- `cmd/stats.go` (new)

#### Template Sharing/Marketplace

**Priority**: Low  
**Estimated Time**: 2 weeks

- [ ] Design template export format
- [ ] Implement template packaging
- [ ] Create template metadata standard
- [ ] Add template verification
- [ ] Support template URL imports
- [ ] Create template gallery (documentation)

**Files to Create**:
- `internal/template/share.go` (new)
- `cmd/template_share.go` (new)

**Deliverables**:
- ✅ Template creation wizard
- ✅ Variable substitution working
- ✅ Conditional steps supported
- ✅ External hooks functional
- ✅ Template sharing enabled

**Success Criteria**:
- Can create complex templates without editing YAML
- Variables work in all contexts
- Hooks execute safely with proper sandboxing
- Template sharing is secure
- Documentation is comprehensive

---

## Technical Considerations

### Database Performance

- **Indexing Strategy**:
  - Index on `note_templates.note_id`
  - Index on `artifacts.note_template_id`
  - Index on `note_steps.note_template_id`
  - Composite index on `(note_id, step_number)`

- **Query Optimization**:
  - Use prepared statements
  - Batch operations where possible
  - Lazy load template details
  - Cache frequently accessed templates

### Data Integrity

- **Foreign Keys**: Enforce relationships with CASCADE delete
- **Transactions**: Use for multi-table operations
- **Validation**: Validate at multiple layers (input, business logic, database)
- **Backups**: Automated backup before migrations

### Security

- **Input Sanitization**: Validate all user inputs
- **File Path Validation**: Prevent directory traversal
- **Command Execution**: Whitelist or sandbox hook commands
- **SQL Injection**: Use parameterized queries exclusively

### Testing Strategy

- **Unit Tests**: 80%+ coverage for business logic
- **Integration Tests**: Test full workflows end-to-end
- **Database Tests**: Test migrations and CRUD operations
- **TUI Tests**: Test keyboard interactions and rendering
- **Performance Tests**: Benchmark database operations

---

## Dependencies

### New Go Packages Needed

```go
// YAML parsing (already have via Viper)
gopkg.in/yaml.v3

// Template variable substitution
text/template (stdlib)

// Condition evaluation
github.com/Knetic/govaluate  // or similar

// File watching (for template hot reload)
github.com/fsnotify/fsnotify  // already have via Viper
```

### External Tools

- None required (all functionality in Go)

---

## Migration Strategy

### Backward Compatibility

1. **Existing Notes**: Remain as simple notes (not templated)
2. **Database Schema**: Add columns with defaults
3. **CLI Commands**: All existing commands work unchanged
4. **Config Files**: No breaking changes

### User Migration Path

1. Users can continue using yoo as before (simple notes)
2. New `--template` flag is opt-in
3. TUI prompts for template if appropriate
4. Documentation highlights template benefits
5. Built-in templates provide instant value

### Rollback Plan

1. Backup database before migration
2. Implement migration rollback SQL
3. Test rollback on development database
4. Document rollback procedure

---

## Documentation Requirements

### User Documentation

- [ ] Template system overview
- [ ] CLI command reference for templates
- [ ] TUI navigation guide
- [ ] Template YAML syntax guide
- [ ] Built-in template documentation
- [ ] Custom template creation guide
- [ ] Real-world examples and workflows
- [ ] Video tutorials (optional)

### Developer Documentation

- [ ] Database schema documentation
- [ ] API documentation for template system
- [ ] Contributing templates guide
- [ ] Architecture decision records (ADRs)
- [ ] Testing guide

### Files to Create/Update

- `docs/TEMPLATES_DESIGN.md` ✓ (already created)
- `docs/TEMPLATE_ROADMAP.md` ✓ (this file)
- `docs/TEMPLATE_CLI_REFERENCE.md` (new)
- `docs/TEMPLATE_YAML_SPEC.md` (new)
- `templates/README.md` ✓ (already created)
- `CONTRIBUTING.md` (update)
- `README.md` (update)

---

## Success Metrics

### Adoption Metrics

- % of users who create at least one templated note
- % of notes that use templates (target: 20-30%)
- Number of custom templates created
- Template completion rate vs simple notes

### Performance Metrics

- Template instantiation time < 50ms
- Step completion operation < 10ms
- Progress calculation < 5ms
- TUI render time < 100ms

### Quality Metrics

- Bug reports per 1000 template operations
- Test coverage > 80%
- User satisfaction score (via feedback)

---

## Risk Assessment

### High Risk

1. **Database Migration Failure**
   - Mitigation: Thorough testing, backup strategy, rollback plan
   
2. **Performance Degradation**
   - Mitigation: Benchmarking, indexing strategy, query optimization

3. **User Confusion**
   - Mitigation: Clear documentation, gradual rollout, help tooltips

### Medium Risk

1. **Template Complexity**
   - Mitigation: Start simple, progressive enhancement
   
2. **Security Issues with Hooks**
   - Mitigation: Sandboxing, whitelisting, clear warnings

### Low Risk

1. **Template Format Changes**
   - Mitigation: Version templates, support older formats

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1 | 3 weeks | Database, models, basic CLI |
| Phase 2 | 3 weeks | Full CLI workflow |
| Phase 3 | 3 weeks | TUI integration |
| Phase 4 | 5 weeks | Advanced features |
| **Total** | **14 weeks** | **Complete template system** |

### Milestones

- **Week 3**: v0.2.0 Alpha - Foundation complete
- **Week 6**: v0.3.0 Beta - CLI fully functional
- **Week 9**: v0.4.0 RC - TUI integration complete
- **Week 14**: v0.5.0 Release - Full feature set

---

## Next Steps

### Immediate Actions (This Week)

1. Review and approve this roadmap
2. Create GitHub issues for Phase 1 tasks
3. Set up development branch: `feature/templates`
4. Write database migration SQL
5. Begin implementing template data models

### Before Starting Phase 1

- [ ] Get stakeholder approval on design
- [ ] Create detailed task breakdown for Phase 1
- [ ] Set up testing infrastructure
- [ ] Prepare development environment
- [ ] Schedule regular progress reviews

---

## Open Questions for Discussion

1. Should we support template versioning from day one?
2. How should we handle template updates for existing notes?
3. Should templates support inheritance/composition?
4. What's the security model for external hooks?
5. Should we create a template gallery website?
6. Integration with external services (GitHub, Jira)?
7. Should artifacts support cloud storage providers?
8. Mobile/web interface in the future?

---

## Appendix: Command Reference (Proposed)

### Template Commands
```bash
yoo templates list                    # List all templates
yoo templates show <name>             # Show template details
yoo template create <file>            # Create new template
yoo template edit <name>              # Edit template
yoo template import <file>            # Import template
yoo template export <name>            # Export template
yoo template delete <name>            # Delete template
```

### Templated Note Commands
```bash
yoo add "task" --template <name>      # Create from template
yoo show <note-id>                    # Show full note details
yoo step complete <id> <step>         # Complete step
yoo step list <id>                    # List steps
yoo artifact add <id> <args>          # Add artifact
yoo artifact list <id>                # List artifacts
yoo ref add <id> <args>               # Add reference
yoo ref list <id>                     # List references
yoo progress <id>                     # Show progress
```

---

**Status**: Ready for Phase 1 implementation  
**Review Date**: 2026-06-03  
**Next Review**: After Phase 1 completion