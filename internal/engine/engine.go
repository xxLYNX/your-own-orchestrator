package engine

import (
	"yoo/internal/models"
)

// Progress tracks aggregate completion units.
type Progress struct {
	Total int
	Done  int
}

// Fraction returns completion ratio in [0,1].
func (p Progress) Fraction() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Done) / float64(p.Total)
}

// EvaluateProgress counts atomic completion units from persisted shape state.
func EvaluateProgress(root *models.ShapeNode, states []*models.ShapeState, inputs map[string]interface{}, recordCount int) Progress {
	var progress Progress

	for _, state := range states {
		switch state.Kind {
		case models.ShapeChecklist:
			if len(state.Data.ItemCompletion) == 0 {
				continue
			}
			for _, done := range state.Data.ItemCompletion {
				progress.Total++
				if done {
					progress.Done++
				}
			}
		case models.ShapeAction, models.ShapeProcedure:
			progress.Total++
			if state.Completed {
				progress.Done++
			}
		}
	}

	if root != nil {
		logTarget := 0
		var countLog func(*models.ShapeNode)
		countLog = func(node *models.ShapeNode) {
			if node == nil {
				return
			}
			if node.Kind == models.ShapeLog {
				target := node.EffectiveRepeatCount(inputs)
				if target > logTarget {
					logTarget = target
				}
			}
			for _, child := range node.ChildNodes() {
				countLog(&child)
			}
		}
		countLog(root)
		if logTarget > 0 {
			if recordCount > logTarget {
				recordCount = logTarget
			}
			progress.Total += logTarget
			progress.Done += recordCount
		}
	}

	return progress
}

// IsStructureComplete reports whether all required instances under root are complete.
func IsStructureComplete(root *models.ShapeNode, states []*models.ShapeState, inputs map[string]interface{}, recordCount int) bool {
	p := EvaluateProgress(root, states, inputs, recordCount)
	return p.Total > 0 && p.Done >= p.Total
}

// Blocker describes a dependency that prevents readiness.
type Blocker struct {
	ShapeID     string
	ShapePath   string
	RepeatStack models.RepeatStack
	Dependency  models.DependencySpec
}

// FindBlockers returns dependencies not yet satisfied for instances under root.
func FindBlockers(root *models.ShapeNode, states []*models.ShapeState, inputs map[string]interface{}, eval EvalContext) []Blocker {
	if root == nil {
		return nil
	}
	index := indexStatesByPath(states)
	var blockers []Blocker

	models.WalkOccurrences(root, nil, nil, inputs, func(node *models.ShapeNode, currentPath []string, iterStack models.RepeatStack) {
		for _, dep := range node.DependsOn {
			if !dependencySatisfied(root, index, dep, iterStack, inputs, eval) {
				blockers = append(blockers, Blocker{
					ShapeID:     node.ID,
					ShapePath:   models.ShapePath(currentPath),
					RepeatStack: iterStack,
					Dependency:  dep,
				})
			}
		}
	})
	return blockers
}

func indexStatesByPath(states []*models.ShapeState) map[string][]*models.ShapeState {
	index := map[string][]*models.ShapeState{}
	for _, state := range states {
		index[state.ShapePath] = append(index[state.ShapePath], state)
	}
	return index
}

func dependencySatisfied(root *models.ShapeNode, index map[string][]*models.ShapeState, dep models.DependencySpec, stack models.RepeatStack, inputs map[string]interface{}, eval EvalContext) bool {
	targetID := dep.Target.ShapeID
	if targetID == "" && dep.Target.ShapePath != "" {
		targetID = dep.Target.ShapePath
	}
	if targetID == "" {
		return true
	}

	switch dep.Requirement {
	case models.RequirementHasRecord:
		return recordRequirementMet(root, dep, stack, inputs, eval)
	case models.RequirementHasArtifact:
		return eval.ArtifactCount > 0
	}

	target := root.FindByPath(root.PathTo(targetID))
	if target == nil {
		target = root.FindByPath([]string{root.ID, targetID})
	}
	if target == nil {
		return false
	}
	targetPath := root.PathTo(target.ID)
	if targetPath == nil {
		return false
	}
	pathStr := models.ShapePath(targetPath)

	scope := dep.Scope
	if scope == "" {
		scope = models.ScopeSameRepeatContext
	}

	candidates := index[pathStr]
	switch scope {
	case models.ScopeAllOccurrences, models.ScopeAllOccurrencesInSameParent:
		count := target.EffectiveRepeatCount(inputs)
		if count <= 1 {
			return instanceRequirementMet(dep.Requirement, models.FindStateByStack(candidates, stack.ParentContext()), eval, stack)
		}
		for i := 1; i <= count; i++ {
			ctx := stack.ParentContext().WithFrame(target.ID, i)
			if !instanceRequirementMet(dep.Requirement, models.FindStateByStack(candidates, ctx), eval, stack) {
				return false
			}
		}
		return true
	case models.ScopeAnyOccurrence:
		for _, state := range candidates {
			if instanceRequirementMet(dep.Requirement, state, eval, stack) {
				return true
			}
		}
		return false
	default:
		return instanceRequirementMet(dep.Requirement, models.FindStateByStack(candidates, stack), eval, stack)
	}
}

func instanceRequirementMet(req models.DependencyRequirement, state *models.ShapeState, eval EvalContext, stack models.RepeatStack) bool {
	if state == nil {
		switch req {
		case models.RequirementHasRecord:
			return eval.HasRecordInStack(stack)
		case models.RequirementHasArtifact:
			return eval.ArtifactCount > 0
		default:
			return false
		}
	}
	switch req {
	case models.RequirementCompleted, "":
		return state.Completed || state.Status == models.StatusComplete || state.Status == models.StatusSkipped
	case models.RequirementStarted:
		return state.Status == models.StatusInProgress || state.Completed || state.Status == models.StatusComplete
	case models.RequirementHasRecord:
		return eval.HasRecordInStack(stack)
	case models.RequirementHasArtifact:
		return eval.ArtifactCount > 0
	default:
		return state.Completed
	}
}
