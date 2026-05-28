package models

import "strconv"

// Shape kind constants — composable, nestable primitives.
const (
	ShapeAction    = "action"
	ShapeProcedure = "procedure"
	ShapeChecklist = "checklist"
	ShapeLog       = "log"
	ShapeArtifact  = "artifact"
	ShapeGroup     = "group"
)

// DependencyRequirement describes what must hold for a dependency edge.
type DependencyRequirement string

const (
	RequirementCompleted   DependencyRequirement = "completed"
	RequirementStarted     DependencyRequirement = "started"
	RequirementHasRecord   DependencyRequirement = "has_record"
	RequirementHasArtifact DependencyRequirement = "has_artifact"
	RequirementApproved    DependencyRequirement = "approved"
	RequirementNotFailed   DependencyRequirement = "not_failed"
)

// DependencyScope disambiguates dependency matching under repeats.
type DependencyScope string

const (
	ScopeAnyOccurrence              DependencyScope = "any_occurrence"
	ScopeAllOccurrences             DependencyScope = "all_occurrences"
	ScopeSameRepeatContext          DependencyScope = "same_repeat_context"
	ScopeAllOccurrencesInSameParent DependencyScope = "all_occurrences_in_same_parent"
	ScopeSpecificRepeatContext      DependencyScope = "specific_repeat_context"
	ScopePreviousOccurrence         DependencyScope = "previous_occurrence"
	ScopeLatestOccurrence           DependencyScope = "latest_occurrence"
	ScopeWholeNote                  DependencyScope = "whole_note"
	ScopeWholeStructure             DependencyScope = "whole_structure"
)

// RepeatSpec controls how many concrete occurrences a shape expands into.
type RepeatSpec struct {
	Count string `json:"count,omitempty" yaml:"count,omitempty"`
}

// ResolveCount returns the resolved repeat count (minimum 1).
func (r *RepeatSpec) ResolveCount(inputs map[string]interface{}) int {
	if r == nil || r.Count == "" {
		return 1
	}
	if v, ok := inputs[r.Count]; ok {
		switch n := v.(type) {
		case int:
			if n > 0 {
				return n
			}
		case int64:
			if n > 0 {
				return int(n)
			}
		case float64:
			if n > 0 {
				return int(n)
			}
		case string:
			if i, err := strconv.Atoi(n); err == nil && i > 0 {
				return i
			}
		}
	}
	if i, err := strconv.Atoi(r.Count); err == nil && i > 0 {
		return i
	}
	return 1
}

// TargetRef identifies the dependency target.
type TargetRef struct {
	NoteID    *int64 `json:"note_id,omitempty" yaml:"note_id,omitempty"`
	NoteAlias string `json:"note,omitempty" yaml:"note,omitempty"`
	ShapePath string `json:"shape_path,omitempty" yaml:"shape_path,omitempty"`
	ShapeID   string `json:"target,omitempty" yaml:"target,omitempty"`
}

// DependencySpec blocks a shape until its target satisfies requirement/scope.
type DependencySpec struct {
	Target      TargetRef             `json:"target" yaml:"target"`
	Requirement DependencyRequirement `json:"requirement" yaml:"requirement"`
	Scope       DependencyScope       `json:"scope,omitempty" yaml:"scope,omitempty"`
}

// ProcedureMode controls auto-generated sibling dependencies for procedures.
type ProcedureMode string

const (
	ProcedureSequential ProcedureMode = "sequential"
	ProcedureParallel   ProcedureMode = "parallel"
	ProcedureManual     ProcedureMode = "manual"
)
