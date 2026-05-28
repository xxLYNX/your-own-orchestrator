package models

import "strings"

// ShapeStateData holds per-shape runtime payload stored as JSON.
type ShapeStateData struct {
	ItemCompletion map[string]bool `json:"item_completion,omitempty"`
}

// ShapeState tracks runtime completion for a node in the composition tree.
type ShapeState struct {
	ID             int64          `json:"id" db:"id"`
	NoteTemplateID int64          `json:"note_template_id" db:"note_template_id"`
	ShapePath      string         `json:"shape_path" db:"shape_path"`
	RepeatIndex    int            `json:"repeat_index" db:"repeat_index"`
	Kind           string         `json:"kind" db:"kind"`
	Title          string         `json:"title" db:"title"`
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
	RepeatIndex int
	Kind        string
	Title       string
	ItemIDs     []string
}

// ShapePath joins navigation IDs into a stable storage key.
func ShapePath(ids []string) string {
	return strings.Join(ids, ".")
}

// FindFirstChecklist returns the first checklist node in this subtree.
func (n *ShapeNode) FindFirstChecklist() *ShapeNode {
	if n == nil {
		return nil
	}
	if n.Kind == ShapeChecklist {
		return n
	}
	for i := range n.Steps {
		if found := n.Steps[i].FindFirstChecklist(); found != nil {
			return found
		}
	}
	if n.RepeatBody != nil {
		return n.RepeatBody.FindFirstChecklist()
	}
	return nil
}

// CollectShapeStateInits walks a composition tree and returns runtime state slots.
func (n *ShapeNode) CollectShapeStateInits(path []string, repeatIndex int, inputs map[string]interface{}) []ShapeStateInit {
	var out []ShapeStateInit
	n.collectShapeStateInits(path, repeatIndex, inputs, &out)
	return out
}

func (n *ShapeNode) collectShapeStateInits(path []string, repeatIndex int, inputs map[string]interface{}, out *[]ShapeStateInit) {
	if n == nil {
		return
	}

	currentPath := append(append([]string{}, path...), n.ID)
	pathStr := ShapePath(currentPath)

	switch n.Kind {
	case ShapeRepeat:
		count := n.ResolveRepeatCount(inputs)
		if n.RepeatBody != nil {
			for i := 1; i <= count; i++ {
				n.RepeatBody.collectShapeStateInits(currentPath, i, inputs, out)
			}
		}
		return

	case ShapeChecklist:
		var itemIDs []string
		for _, item := range n.Items {
			itemIDs = append(itemIDs, item.ID)
		}
		*out = append(*out, ShapeStateInit{
			Path:        pathStr,
			RepeatIndex: repeatIndex,
			Kind:        ShapeChecklist,
			Title:       n.DisplayTitle(),
			ItemIDs:     itemIDs,
		})
		return

	case ShapeProcedure:
		if n.ID != "root" && n.Title != "" {
			*out = append(*out, ShapeStateInit{
				Path:        pathStr,
				RepeatIndex: repeatIndex,
				Kind:        ShapeProcedure,
				Title:       n.DisplayTitle(),
			})
		}
		for i := range n.Steps {
			switch n.Steps[i].Kind {
			case ShapeLog, ShapeArtifact:
				continue
			}
			n.Steps[i].collectShapeStateInits(currentPath, repeatIndex, inputs, out)
		}
		if n.RepeatBody != nil {
			n.RepeatBody.collectShapeStateInits(currentPath, repeatIndex, inputs, out)
		}
	}
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
	for _, item := range node.Items {
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
