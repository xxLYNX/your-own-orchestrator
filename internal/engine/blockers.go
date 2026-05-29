package engine

import (
	"fmt"
	"strings"

	"yoo/internal/models"
)

// ErrBlocked indicates forward progress was rejected because dependencies are unsatisfied.
type ErrBlocked struct {
	ShapePath   string
	RepeatStack models.RepeatStack
	Blockers    []Blocker
}

func (e *ErrBlocked) Error() string {
	if len(e.Blockers) == 0 {
		return fmt.Sprintf("blocked: %s", e.ShapePath)
	}
	parts := make([]string, 0, len(e.Blockers))
	for _, b := range e.Blockers {
		target := b.Dependency.Target.ShapeID
		if target == "" {
			target = b.Dependency.Target.ShapePath
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", target, b.Dependency.Requirement))
	}
	return fmt.Sprintf("blocked: waiting for %s", strings.Join(parts, ", "))
}

// BlockersForInstance returns dependency blockers affecting one shape occurrence.
func BlockersForInstance(blockers []Blocker, shapePath string, stack models.RepeatStack) []Blocker {
	var out []Blocker
	for _, b := range blockers {
		if b.ShapePath == shapePath && b.RepeatStack.Equal(stack) {
			out = append(out, b)
		}
	}
	return out
}

// IsInstanceBlocked reports whether an instance has unsatisfied dependencies.
func IsInstanceBlocked(blockers []Blocker, shapePath string, stack models.RepeatStack) bool {
	return len(BlockersForInstance(blockers, shapePath, stack)) > 0
}
