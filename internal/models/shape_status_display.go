package models

// InstanceStatusMarker returns a compact status glyph for CLI/TUI display.
func InstanceStatusMarker(state *ShapeState, blocked bool) string {
	if state == nil {
		return "○"
	}
	if blocked && !state.Completed && state.Status != StatusSkipped && state.Status != StatusFailed {
		return "🔒"
	}
	switch state.Status {
	case StatusComplete:
		if state.Completed {
			return "✓"
		}
		return "✓"
	case StatusSkipped:
		return "⊘"
	case StatusFailed:
		return "✗"
	case StatusInProgress:
		return "◐"
	default:
		if state.Completed {
			return "✓"
		}
		return "○"
	}
}

// InstanceStatusLabel returns a human-readable status name.
func InstanceStatusLabel(state *ShapeState, blocked bool) string {
	if state == nil {
		return "open"
	}
	if blocked && !state.Completed && state.Status != StatusSkipped && state.Status != StatusFailed {
		return "blocked"
	}
	switch state.Status {
	case StatusComplete:
		return "complete"
	case StatusSkipped:
		return "skipped"
	case StatusFailed:
		return "failed"
	case StatusInProgress:
		return "in_progress"
	case StatusNotStarted:
		return "open"
	default:
		if state.Completed {
			return "complete"
		}
		return string(state.Status)
	}
}
