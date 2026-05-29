package models

import "fmt"

// NormalizeShapeTree compiles procedure ordering and normalizes the shape tree.
func NormalizeShapeTree(node *ShapeNode) *ShapeNode {
	if node == nil {
		return nil
	}
	node.normalizeChildTrees()
	compileProcedureDependencies(node)
	return node
}

func (n *ShapeNode) normalizeChildTrees() {
	for i := range n.Steps {
		n.Steps[i] = *NormalizeShapeTree(&n.Steps[i])
	}
	for i := range n.Items {
		n.Items[i] = *NormalizeShapeTree(&n.Items[i])
	}
	for i := range n.ChildList {
		n.ChildList[i] = *NormalizeShapeTree(&n.ChildList[i])
	}
}

func compileProcedureDependencies(node *ShapeNode) {
	var walk func(*ShapeNode)
	walk = func(n *ShapeNode) {
		if n == nil {
			return
		}
		if n.Kind == ShapeProcedure && n.DependencyMode != ProcedureParallel && n.DependencyMode != ProcedureManual {
			applySequentialDeps := func(children *[]ShapeNode) {
				for i := 1; i < len(*children); i++ {
					prev := (*children)[i-1]
					child := &(*children)[i]
					if !child.hasDependencyOn(prev.ID) {
						child.DependsOn = append(child.DependsOn, DependencySpec{
							Target:      TargetRef{ShapeID: prev.ID},
							Requirement: RequirementCompleted,
							Scope:       ScopeSameRepeatContext,
						})
					}
				}
			}
			if len(n.Steps) > 0 {
				applySequentialDeps(&n.Steps)
			} else {
				applySequentialDeps(&n.ChildList)
			}
		}
		for i := range n.Steps {
			walk(&n.Steps[i])
		}
		for i := range n.Items {
			walk(&n.Items[i])
		}
		for i := range n.ChildList {
			walk(&n.ChildList[i])
		}
	}
	walk(node)
}

func (n *ShapeNode) hasDependencyOn(shapeID string) bool {
	for _, dep := range n.DependsOn {
		if dep.Target.ShapeID == shapeID || dep.Target.ShapePath == shapeID {
			return true
		}
	}
	return false
}

// EffectiveRepeatCount returns how many occurrences this shape expands into.
func (n *ShapeNode) EffectiveRepeatCount(inputs map[string]interface{}) int {
	if n == nil {
		return 1
	}
	if n.Repeat != nil {
		return n.Repeat.ResolveCount(inputs)
	}
	return 1
}

// NeedsRepeatPicker reports whether the UI should show occurrence selection for this node.
func NeedsRepeatPicker(node *ShapeNode, stack RepeatStack, inputs map[string]interface{}) bool {
	if node == nil {
		return false
	}
	count := node.EffectiveRepeatCount(inputs)
	if count <= 1 {
		return false
	}
	_, has := stack.FrameForShape(node.ID)
	return !has
}

// SimpleNoteStructure returns the implicit single-action structure for line-item notes.
func SimpleNoteStructure(title string) *ShapeNode {
	return &ShapeNode{
		ID:     "note-action",
		Kind:   ShapeAction,
		Title:  title,
		Repeat: &RepeatSpec{Count: "1"},
	}
}

// ValidateDependencySpecs ensures dependency metadata is well-formed.
func ValidateDependencySpecs(node *ShapeNode) error {
	if node == nil {
		return nil
	}
	for _, dep := range node.DependsOn {
		if dep.Requirement == "" {
			return fmt.Errorf("shape %q dependency missing requirement", node.ID)
		}
		if !dep.Requirement.IsValid() {
			return fmt.Errorf("shape %q dependency has unknown requirement %q", node.ID, dep.Requirement)
		}
		if dep.Target.ShapeID == "" && dep.Target.ShapePath == "" && dep.Target.NoteAlias == "" && dep.Target.NoteID == nil {
			return fmt.Errorf("shape %q dependency missing target", node.ID)
		}
	}
	for i := range node.ChildNodes() {
		child := node.ChildNodes()[i]
		if err := ValidateDependencySpecs(&child); err != nil {
			return err
		}
	}
	return nil
}
