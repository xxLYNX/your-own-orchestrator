package engine

import "yoo/internal/models"

// recordRequirementMet evaluates has_record for a dependency target log shape.
func recordRequirementMet(root *models.ShapeNode, dep models.DependencySpec, stack models.RepeatStack, inputs map[string]interface{}, eval EvalContext) bool {
	targetID := dep.Target.ShapeID
	if targetID == "" {
		targetID = dep.Target.ShapePath
	}

	scope := dep.Scope
	if scope == "" {
		scope = models.ScopeSameRepeatContext
	}

	switch scope {
	case models.ScopeAllOccurrences, models.ScopeAllOccurrencesInSameParent:
		repeatNode, count := enclosingRepeatForTarget(root, targetID, inputs)
		if repeatNode == nil || count <= 1 {
			return eval.HasRecordInStack(stack)
		}
		base := stackWithoutFrame(stack, repeatNode.ID)
		for i := 1; i <= count; i++ {
			iterStack := base.WithFrame(repeatNode.ID, i)
			if !eval.HasRecordInStack(iterStack) {
				return false
			}
		}
		return true
	default:
		return eval.HasRecordInStack(stack)
	}
}

// enclosingRepeatForTarget returns the innermost repeated ancestor on the path to targetID.
func enclosingRepeatForTarget(root *models.ShapeNode, targetID string, inputs map[string]interface{}) (*models.ShapeNode, int) {
	if root == nil || targetID == "" {
		return nil, 1
	}
	path := root.PathTo(targetID)
	if path == nil {
		if root.FindByPath([]string{root.ID, targetID}) != nil {
			path = []string{root.ID, targetID}
		}
	}
	if path == nil {
		return nil, 1
	}

	var (
		repeatNode *models.ShapeNode
		repeatCount int
	)
	node := root
	for i := 1; i < len(path); i++ {
		childID := path[i]
		found := false
		for j := range node.ChildNodes() {
			child := node.ChildNodes()[j]
			if child.ID != childID {
				continue
			}
			count := child.EffectiveRepeatCount(inputs)
			if count > 1 {
				repeatNode = &child
				repeatCount = count
			}
			node = &child
			found = true
			break
		}
		if !found {
			break
		}
	}
	return repeatNode, repeatCount
}

func stackWithoutFrame(stack models.RepeatStack, shapeID string) models.RepeatStack {
	if len(stack) == 0 {
		return nil
	}
	out := make(models.RepeatStack, 0, len(stack))
	for _, frame := range stack {
		if frame.ShapeID == shapeID {
			continue
		}
		out = append(out, frame)
	}
	return out
}
