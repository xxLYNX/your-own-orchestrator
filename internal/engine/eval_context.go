package engine

import "yoo/internal/models"

// EvalContext carries non-shape-state facts used to evaluate dependency requirements.
type EvalContext struct {
	RecordCountTotal   int
	RecordCountByStack map[string]int
	ArtifactCount      int
}

// HasRecordInStack reports whether at least one log record exists in a repeat scope.
func (e EvalContext) HasRecordInStack(stack models.RepeatStack) bool {
	if len(e.RecordCountByStack) > 0 {
		return e.RecordCountByStack[stack.JSON()] > 0
	}
	return e.RecordCountTotal > 0
}
