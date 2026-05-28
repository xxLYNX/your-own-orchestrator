package models

import (
	"fmt"
	"strconv"
	"strings"
)

// Shape kind constants — composable, nestable primitives.
const (
	ShapeProcedure = "procedure"
	ShapeChecklist = "checklist"
	ShapeLog       = "log"
	ShapeArtifact  = "artifact"
	ShapeRepeat    = "repeat"
)

// ShapeNode is a node in the fractal template composition tree.
type ShapeNode struct {
	ID          string `json:"id" yaml:"id"`
	Kind        string `json:"kind" yaml:"kind"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Procedure: ordered child steps (each may nest further shapes).
	Steps []ShapeNode `json:"steps,omitempty" yaml:"steps,omitempty"`

	// Checklist: unordered items (strings become leaf nodes, or nested shapes).
	Items []ShapeNode `json:"items,omitempty" yaml:"items,omitempty"`

	// Log shape: structured repeating rows.
	RecordSchema *RecordSchema `json:"record_schema,omitempty" yaml:"record_schema,omitempty"`

	// Artifact shape: template-level artifact definition (input/output spec).
	Artifact *TemplateOutput `json:"artifact,omitempty" yaml:"artifact,omitempty"`

	// Repeat: run child shape N times (count from template input name or literal).
	RepeatCount string     `json:"repeat_count,omitempty" yaml:"repeat_count,omitempty"`
	RepeatBody  *ShapeNode `json:"repeat_body,omitempty" yaml:"repeat_body,omitempty"`

	// Legacy inline checklist strings on procedure steps (compiled to checklist children).
	Checklist []string `json:"checklist,omitempty" yaml:"checklist,omitempty"`

	EstimatedTime  string `json:"estimated_time,omitempty" yaml:"estimated_time,omitempty"`
	OutputRequired string `json:"output_required,omitempty" yaml:"output_required,omitempty"`
}

// NavContext tracks where the user is inside a note's composition tree.
type NavContext struct {
	Path        []string `json:"path"`
	RepeatIndex int      `json:"repeat_index,omitempty"`
}

// String returns a breadcrumb-style path for display.
func (nc NavContext) String() string {
	if len(nc.Path) == 0 {
		return "score"
	}
	parts := append([]string(nil), nc.Path...)
	if nc.RepeatIndex > 0 {
		parts[len(parts)-1] = fmt.Sprintf("%s #%d", parts[len(parts)-1], nc.RepeatIndex)
	}
	return strings.Join(parts, " › ")
}

// NormalizeTemplateDefinition ensures Composition is populated and legacy fields stay in sync.
func NormalizeTemplateDefinition(def *TemplateDefinition) error {
	if def == nil {
		return fmt.Errorf("template definition is nil")
	}

	if def.Composition == nil {
		def.Composition = LegacyCompileComposition(def)
	}

	if def.Composition == nil {
		return fmt.Errorf("template has no composition")
	}

	def.Composition.NormalizeIDs("root")
	if err := def.Composition.Validate(); err != nil {
		return err
	}

	// Keep legacy accessors working for CLI code paths.
	syncLegacyFields(def)
	return nil
}

func syncLegacyFields(def *TemplateDefinition) {
	def.Steps = nil
	def.RecordSchema = nil

	var walk func(node *ShapeNode)
	walk = func(node *ShapeNode) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ShapeProcedure:
			if node.ID != "root" && node.Title != "" {
				def.Steps = append(def.Steps, TemplateStep{
					ID:             len(def.Steps) + 1,
					Title:          node.DisplayTitle(),
					Description:    node.Description,
					Checklist:      node.inlineChecklistTitles(),
					EstimatedTime:  node.EstimatedTime,
					OutputRequired: node.OutputRequired,
				})
			}
			for i := range node.Steps {
				walk(&node.Steps[i])
			}
		case ShapeLog:
			if node.RecordSchema != nil && def.RecordSchema == nil {
				def.RecordSchema = node.RecordSchema
			}
		case ShapeChecklist:
			for i := range node.Items {
				walk(&node.Items[i])
			}
		case ShapeRepeat:
			if node.RepeatBody != nil {
				walk(node.RepeatBody)
			}
		}
	}

	walk(def.Composition)
}

// LegacyCompileComposition builds a composition tree from legacy flat template fields.
func LegacyCompileComposition(def *TemplateDefinition) *ShapeNode {
	root := &ShapeNode{
		ID:    "root",
		Kind:  ShapeProcedure,
		Title: "Workflow",
	}

	for _, step := range def.Steps {
		stepNode := ShapeNode{
			ID:             fmt.Sprintf("step-%d", step.ID),
			Kind:           ShapeProcedure,
			Title:          step.Title,
			Description:    step.Description,
			EstimatedTime:  step.EstimatedTime,
			OutputRequired: step.OutputRequired,
		}
		if len(step.Checklist) > 0 {
			checklist := ShapeNode{
				ID:    fmt.Sprintf("step-%d-checklist", step.ID),
				Kind:  ShapeChecklist,
				Title: "Checklist",
			}
			for i, item := range step.Checklist {
				checklist.Items = append(checklist.Items, ShapeNode{
					ID:    fmt.Sprintf("step-%d-item-%d", step.ID, i+1),
					Kind:  ShapeChecklist,
					Title: item,
				})
			}
			stepNode.Steps = append(stepNode.Steps, checklist)
		}
		root.Steps = append(root.Steps, stepNode)
	}

	if def.RecordSchema != nil {
		root.Steps = append(root.Steps, ShapeNode{
			ID:           "records",
			Kind:         ShapeLog,
			Title:        "Records",
			RecordSchema: def.RecordSchema,
		})
	}

	if len(def.Inputs) > 0 || len(def.Outputs) > 0 {
		artifactRoot := ShapeNode{
			ID:    "artifacts",
			Kind:  ShapeArtifact,
			Title: "Artifacts",
		}
		for _, input := range def.Inputs {
			artifactRoot.Items = append(artifactRoot.Items, ShapeNode{
				ID:    "input-" + input.Name,
				Kind:  ShapeArtifact,
				Title: input.Name,
				Description: input.Description,
			})
		}
		for _, output := range def.Outputs {
			artifactRoot.Items = append(artifactRoot.Items, ShapeNode{
				ID:    "output-" + output.Name,
				Kind:  ShapeArtifact,
				Title: output.Name,
				Artifact: &output,
				Description: output.Description,
			})
		}
		root.Steps = append(root.Steps, artifactRoot)
	}

	if len(root.Steps) == 0 && def.RecordSchema == nil && len(def.Inputs) == 0 && len(def.Outputs) == 0 {
		return nil
	}

	return root
}

// DisplayTitle returns the best label for a node.
func (n *ShapeNode) DisplayTitle() string {
	if n.Title != "" {
		return n.Title
	}
	switch n.Kind {
	case ShapeProcedure:
		return "Procedure"
	case ShapeChecklist:
		return "Checklist"
	case ShapeLog:
		return "Log"
	case ShapeArtifact:
		return "Artifacts"
	case ShapeRepeat:
		return "Repeat"
	default:
		return n.ID
	}
}

func (n *ShapeNode) inlineChecklistTitles() []string {
	var titles []string
	for _, item := range n.Steps {
		if item.Kind == ShapeChecklist {
			for _, sub := range item.Items {
				if sub.Title != "" {
					titles = append(titles, sub.Title)
				}
			}
		}
	}
	if len(titles) == 0 {
		titles = append(titles, n.Checklist...)
	}
	return titles
}

// NormalizeIDs assigns IDs to nodes missing them.
func (n *ShapeNode) NormalizeIDs(prefix string) {
	if n.ID == "" {
		n.ID = prefix
	}
	switch n.Kind {
	case ShapeProcedure:
		for i := range n.Steps {
			n.Steps[i].NormalizeIDs(fmt.Sprintf("%s.s%d", n.ID, i+1))
		}
	case ShapeChecklist:
		for i := range n.Items {
			n.Items[i].NormalizeIDs(fmt.Sprintf("%s.i%d", n.ID, i+1))
		}
	case ShapeRepeat:
		if n.RepeatBody != nil {
			n.RepeatBody.NormalizeIDs(n.ID + ".body")
		}
	}
}

// Validate validates a composition subtree.
func (n *ShapeNode) Validate() error {
	if n == nil {
		return fmt.Errorf("nil shape node")
	}
	if n.Kind == "" {
		return fmt.Errorf("shape node %q missing kind", n.ID)
	}

	switch n.Kind {
	case ShapeProcedure:
		if len(n.Steps) == 0 && len(n.Checklist) == 0 {
			return fmt.Errorf("procedure %q must have steps or checklist items", n.ID)
		}
		for i := range n.Steps {
			if err := n.Steps[i].Validate(); err != nil {
				return err
			}
		}
	case ShapeChecklist:
		if len(n.Items) == 0 && n.Title == "" {
			return fmt.Errorf("checklist %q must have items or a title", n.ID)
		}
		for i := range n.Items {
			if err := n.Items[i].Validate(); err != nil {
				return err
			}
		}
	case ShapeLog:
		if n.RecordSchema == nil || len(n.RecordSchema.Fields) == 0 {
			return fmt.Errorf("log %q must define record_schema", n.ID)
		}
	case ShapeArtifact:
		// Artifact container nodes may have empty items (filled at runtime).
	case ShapeRepeat:
		if n.RepeatBody == nil {
			return fmt.Errorf("repeat %q must define repeat_body", n.ID)
		}
		if n.RepeatCount == "" {
			return fmt.Errorf("repeat %q must define repeat_count", n.ID)
		}
		if err := n.RepeatBody.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown shape kind %q on node %q", n.Kind, n.ID)
	}

	return nil
}

// FindByPath resolves a node from a path of IDs starting at this node.
func (n *ShapeNode) FindByPath(path []string) *ShapeNode {
	if len(path) == 0 {
		return n
	}
	if path[0] != n.ID {
		return nil
	}
	if len(path) == 1 {
		return n
	}
	rest := path[1:]
	for i := range n.Steps {
		if found := n.Steps[i].FindByPath(rest); found != nil {
			return found
		}
	}
	for i := range n.Items {
		if found := n.Items[i].FindByPath(rest); found != nil {
			return found
		}
	}
	if n.RepeatBody != nil {
		if found := n.RepeatBody.FindByPath(rest); found != nil {
			return found
		}
	}
	return nil
}

// Children returns navigable children for a node.
func (n *ShapeNode) Children() []ShapeNode {
	switch n.Kind {
	case ShapeProcedure:
		return n.Steps
	case ShapeChecklist:
		return n.Items
	case ShapeRepeat:
		if n.RepeatBody != nil {
			return []ShapeNode{*n.RepeatBody}
		}
	}
	return nil
}

// FindFirstLogNode returns the first log shape node in the tree.
func (n *ShapeNode) FindFirstLogNode() *ShapeNode {
	if n == nil {
		return nil
	}
	if n.Kind == ShapeLog && n.RecordSchema != nil {
		return n
	}
	for i := range n.Steps {
		if found := n.Steps[i].FindFirstLogNode(); found != nil {
			return found
		}
	}
	for i := range n.Items {
		if found := n.Items[i].FindFirstLogNode(); found != nil {
			return found
		}
	}
	if n.RepeatBody != nil {
		return n.RepeatBody.FindFirstLogNode()
	}
	return nil
}
// FindLogNodeInSubtree returns the first log node under this node (including self).
func (n *ShapeNode) FindLogNodeInSubtree() *ShapeNode {
	if n == nil {
		return nil
	}
	if n.Kind == ShapeLog && n.RecordSchema != nil {
		return n
	}
	for i := range n.Steps {
		if found := n.Steps[i].FindLogNodeInSubtree(); found != nil {
			return found
		}
	}
	for i := range n.Items {
		if found := n.Items[i].FindLogNodeInSubtree(); found != nil {
			return found
		}
	}
	if n.RepeatBody != nil {
		return n.RepeatBody.FindLogNodeInSubtree()
	}
	return nil
}

// PathTo returns the ID path from this node to targetID, or nil if not found.
func (n *ShapeNode) PathTo(targetID string) []string {
	if n == nil {
		return nil
	}
	var walk func(node *ShapeNode, path []string) []string
	walk = func(node *ShapeNode, path []string) []string {
		current := append(append([]string{}, path...), node.ID)
		if node.ID == targetID {
			return current
		}
		for i := range node.Steps {
			if found := walk(&node.Steps[i], current); found != nil {
				return found
			}
		}
		for i := range node.Items {
			if found := walk(&node.Items[i], current); found != nil {
				return found
			}
		}
		if node.RepeatBody != nil {
			if found := walk(node.RepeatBody, current); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(n, nil)
}

func (n *ShapeNode) ResolveRepeatCount(inputs map[string]interface{}) int {
	if n == nil || n.RepeatCount == "" {
		return 0
	}
	if v, ok := inputs[n.RepeatCount]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	if i, err := strconv.Atoi(n.RepeatCount); err == nil {
		return i
	}
	return 0
}

// ActiveShapeKinds walks the tree and returns unique shape kinds present.
func (n *ShapeNode) ActiveShapeKinds() []string {
	seen := map[string]bool{}
	var kinds []string
	var walk func(*ShapeNode)
	walk = func(node *ShapeNode) {
		if node == nil {
			return
		}
		if !seen[node.Kind] {
			seen[node.Kind] = true
			kinds = append(kinds, node.Kind)
		}
		for i := range node.Steps {
			walk(&node.Steps[i])
		}
		for i := range node.Items {
			walk(&node.Items[i])
		}
		if node.RepeatBody != nil {
			walk(node.RepeatBody)
		}
	}
	walk(n)
	return kinds
}

// FlattenProcedureSteps returns all procedure step nodes in DFS order for DB initialization.
func (n *ShapeNode) FlattenProcedureSteps() []ShapeNode {
	var steps []ShapeNode
	var walk func(*ShapeNode)
	walk = func(node *ShapeNode) {
		if node == nil {
			return
		}
		if node.Kind == ShapeProcedure && node.ID != "root" && node.Title != "" && len(node.Steps) > 0 {
			// Inner procedures with substeps — treat as navigable, also flatten substeps
		}
		if node.Kind == ShapeProcedure && node.ID != "root" && node.Title != "" {
			hasOnlyChecklist := len(node.Steps) == 1 && node.Steps[0].Kind == ShapeChecklist
			if !hasOnlyChecklist || len(node.Checklist) > 0 || node.Description != "" {
				steps = append(steps, *node)
			}
		}
		for i := range node.Steps {
			if node.Steps[i].Kind != ShapeLog && node.Steps[i].Kind != ShapeArtifact {
				walk(&node.Steps[i])
			}
		}
		if node.RepeatBody != nil {
			walk(node.RepeatBody)
		}
	}
	walk(n)
	return steps
}
