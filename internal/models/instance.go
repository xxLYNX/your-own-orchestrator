package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RepeatFrame identifies one occurrence of a repeated shape.
type RepeatFrame struct {
	ShapeID string `json:"shape_id"`
	Index   int    `json:"index"`
}

// RepeatStack is ordered from outermost to innermost repeat context.
type RepeatStack []RepeatFrame

// EmptyRepeatStackJSON is the canonical serialized empty stack.
const EmptyRepeatStackJSON = "[]"

// JSON serializes the repeat stack for persistence.
func (s RepeatStack) JSON() string {
	if len(s) == 0 {
		return EmptyRepeatStackJSON
	}
	b, err := json.Marshal(s)
	if err != nil {
		return EmptyRepeatStackJSON
	}
	return string(b)
}

// ParseRepeatStackJSON parses persisted repeat stack JSON.
func ParseRepeatStackJSON(raw string) RepeatStack {
	if raw == "" || raw == EmptyRepeatStackJSON {
		return nil
	}
	var stack RepeatStack
	if err := json.Unmarshal([]byte(raw), &stack); err != nil {
		return nil
	}
	return stack
}

// Equal reports whether two stacks match.
func (s RepeatStack) Equal(other RepeatStack) bool {
	if len(s) != len(other) {
		return false
	}
	for i := range s {
		if s[i].ShapeID != other[i].ShapeID || s[i].Index != other[i].Index {
			return false
		}
	}
	return true
}

// FrameForShape returns the active frame for a shape ID, if any.
func (s RepeatStack) FrameForShape(shapeID string) (RepeatFrame, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i].ShapeID == shapeID {
			return s[i], true
		}
	}
	return RepeatFrame{}, false
}

// WithFrame returns a copy of the stack with an appended frame.
func (s RepeatStack) WithFrame(shapeID string, index int) RepeatStack {
	out := append(RepeatStack{}, s...)
	out = append(out, RepeatFrame{ShapeID: shapeID, Index: index})
	return out
}

// ParentContext returns the stack with the innermost frame removed.
func (s RepeatStack) ParentContext() RepeatStack {
	if len(s) == 0 {
		return nil
	}
	return append(RepeatStack{}, s[:len(s)-1]...)
}

// String renders a breadcrumb suffix for display.
func (s RepeatStack) String() string {
	if len(s) == 0 {
		return ""
	}
	parts := make([]string, len(s))
	for i, frame := range s {
		parts[i] = fmt.Sprintf("#%d", frame.Index)
	}
	return strings.Join(parts, " › ")
}
