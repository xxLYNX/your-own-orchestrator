package models

// OccurrenceVisit is invoked for each concrete shape occurrence while walking a tree.
type OccurrenceVisit func(node *ShapeNode, path []string, stack RepeatStack)

// WalkOptions configures tree traversal behavior.
type WalkOptions struct {
	// SkipChildKinds prevents descending into children of these kinds.
	SkipChildKinds map[string]bool
}

// WalkOccurrences traverses the structure tree, expanding repeat modifiers into
// concrete occurrences. path is the ID path to the parent of node (excluding node.ID).
func WalkOccurrences(node *ShapeNode, path []string, stack RepeatStack, inputs map[string]interface{}, visit OccurrenceVisit) {
	WalkOccurrencesWithOptions(node, path, stack, inputs, visit, nil)
}

// WalkOccurrencesWithOptions is like WalkOccurrences but honors traversal options.
func WalkOccurrencesWithOptions(node *ShapeNode, path []string, stack RepeatStack, inputs map[string]interface{}, visit OccurrenceVisit, opts *WalkOptions) {
	if node == nil {
		return
	}
	currentPath := append(append([]string{}, path...), node.ID)
	count := node.EffectiveRepeatCount(inputs)
	for i := 1; i <= count; i++ {
		iterStack := stack
		if count > 1 {
			iterStack = stack.WithFrame(node.ID, i)
		}
		visit(node, currentPath, iterStack)
		for _, child := range node.ChildNodes() {
			if opts != nil && opts.SkipChildKinds[child.Kind] {
				continue
			}
			WalkOccurrencesWithOptions(&child, currentPath, iterStack, inputs, visit, opts)
		}
	}
}
