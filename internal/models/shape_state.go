package models

import "strings"

// InstanceStatus is the runtime lifecycle for a shape occurrence.
type InstanceStatus string

const (
	StatusNotStarted InstanceStatus = "not_started"
	StatusInProgress InstanceStatus = "in_progress"
	StatusComplete   InstanceStatus = "complete"
)

// ShapeStateData holds per-shape runtime payload stored as JSON.
type ShapeStateData struct {
	ItemCompletion map[string]bool `json:"item_completion,omitempty"`
}

// ShapeState tracks runtime completion for a node in the structure tree.
type ShapeState struct {
	ID             int64          `json:"id" db:"id"`
	NoteTemplateID int64          `json:"note_template_id" db:"note_template_id"`
	ShapePath      string         `json:"shape_path" db:"shape_path"`
	RepeatStack    RepeatStack    `json:"repeat_stack" db:"repeat_stack_json"`
	Kind           string         `json:"kind" db:"kind"`
	Title          string         `json:"title" db:"title"`
	Status         InstanceStatus `json:"status" db:"status"`
	Completed      bool           `json:"completed" db:"completed"`
	CompletedAt    *string        `json:"completed_at,omitempty" db:"completed_at"`
	Notes          string         `json:"notes,omitempty" db:"notes"`
	Data           ShapeStateData `json:"data"`
	CreatedAt      string         `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      string         `json:"updated_at,omitempty" db:"updated_at"`
}

// ShapeStateInit describes a row to create when a note is instantiated.
type ShapeStateInit struct {
	Path        string
	RepeatStack RepeatStack
	Kind        string
	Title       string
	ItemIDs     []string
}

// ShapePath joins navigation IDs into a stable storage key.
func ShapePath(ids []string) string {
	return strings.Join(ids, ".")
}

// ChecklistItemView is a display row for scoped checklist editing.
type ChecklistItemView struct {
	ItemID    string
	Title     string
	Completed bool
}

// ChecklistItemsFromState builds display rows from a checklist node and its persisted state.
func ChecklistItemsFromState(node *ShapeNode, state *ShapeState) []ChecklistItemView {
	if node == nil || state == nil {
		return nil
	}
	var items []ChecklistItemView
	for _, item := range node.ChildNodes() {
		completed := false
		if state.Data.ItemCompletion != nil {
			completed = state.Data.ItemCompletion[item.ID]
		}
		items = append(items, ChecklistItemView{
			ItemID:    item.ID,
			Title:     item.DisplayTitle(),
			Completed: completed,
		})
	}
	return items
}

// AllChecklistItemsComplete reports whether every item in state is checked off.
func AllChecklistItemsComplete(state *ShapeState) bool {
	if state == nil || len(state.Data.ItemCompletion) == 0 {
		return false
	}
	for _, done := range state.Data.ItemCompletion {
		if !done {
			return false
		}
	}
	return true
}

// shapeStateWalkOpts skips runtime state rows for log and artifact nodes.
var shapeStateWalkOpts = &WalkOptions{
	SkipChildKinds: map[string]bool{
		ShapeLog:      true,
		ShapeArtifact: true,
	},
}

// CollectShapeStateInits walks a structure tree and returns runtime state slots.
func (n *ShapeNode) CollectShapeStateInits(path []string, stack RepeatStack, inputs map[string]interface{}) []ShapeStateInit {
	var out []ShapeStateInit
	WalkOccurrencesWithOptions(n, path, stack, inputs, func(node *ShapeNode, currentPath []string, iterStack RepeatStack) {
		node.appendOccurrenceStateInit(currentPath, iterStack, &out)
	}, shapeStateWalkOpts)
	return out
}

func (n *ShapeNode) appendOccurrenceStateInit(path []string, stack RepeatStack, out *[]ShapeStateInit) {
	pathStr := ShapePath(path)

	switch n.Kind {
	case ShapeChecklist:
		var itemIDs []string
		for _, item := range n.ChildNodes() {
			itemIDs = append(itemIDs, item.ID)
		}
		*out = append(*out, ShapeStateInit{
			Path:        pathStr,
			RepeatStack: stack,
			Kind:        ShapeChecklist,
			Title:       n.DisplayTitle(),
			ItemIDs:     itemIDs,
		})
	case ShapeAction:
		*out = append(*out, ShapeStateInit{
			Path:        pathStr,
			RepeatStack: stack,
			Kind:        ShapeAction,
			Title:       n.DisplayTitle(),
		})
	case ShapeProcedure:
		if n.ID != "root" && n.Title != "" {
			*out = append(*out, ShapeStateInit{
				Path:        pathStr,
				RepeatStack: stack,
				Kind:        ShapeProcedure,
				Title:       n.DisplayTitle(),
			})
		}
	}
}

// FindStateByStack returns the first state row matching a repeat stack.
func FindStateByStack(states []*ShapeState, stack RepeatStack) *ShapeState {
	for _, state := range states {
		if state.RepeatStack.Equal(stack) {
			return state
		}
	}
	return nil
}

// NewShapeStateData builds default runtime payload for a shape state init.
func NewShapeStateData(init ShapeStateInit) ShapeStateData {
	data := ShapeStateData{}
	if init.Kind == ShapeChecklist {
		data.ItemCompletion = make(map[string]bool, len(init.ItemIDs))
		for _, id := range init.ItemIDs {
			data.ItemCompletion[id] = false
		}
	}
	return data
}
