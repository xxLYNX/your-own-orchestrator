package models

import (
	"fmt"
	"strings"
)

// ShapeNode is a node in the fractal template structure tree.
type ShapeNode struct {
	ID          string `json:"id" yaml:"id"`
	Kind        string `json:"kind" yaml:"kind"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Procedure: ordered child steps (each may nest further shapes).
	Steps []ShapeNode `json:"steps,omitempty" yaml:"steps,omitempty"`

	// Checklist: unordered items (strings become leaf nodes, or nested shapes).
	Items []ShapeNode `json:"items,omitempty" yaml:"items,omitempty"`

	// ChildList is an alternate unified child list used by structure-first YAML (`children:`).
	ChildList []ShapeNode `json:"children,omitempty" yaml:"children,omitempty"`

	// Execution modifiers.
	Repeat         *RepeatSpec      `json:"repeat,omitempty" yaml:"repeat,omitempty"`
	DependsOn      []DependencySpec `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	DependencyMode ProcedureMode    `json:"dependency_mode,omitempty" yaml:"dependency_mode,omitempty"`

	// Log shape: structured repeating rows.
	RecordSchema *RecordSchema `json:"record_schema,omitempty" yaml:"record_schema,omitempty"`

	// Artifact shape: template-level artifact definition (input/output spec).
	Artifact *TemplateOutput `json:"artifact,omitempty" yaml:"artifact,omitempty"`

	EstimatedTime  string `json:"estimated_time,omitempty" yaml:"estimated_time,omitempty"`
	OutputRequired string `json:"output_required,omitempty" yaml:"output_required,omitempty"`
}

// NavContext tracks where the user is inside a note's structure tree.
type NavContext struct {
	Path        []string    `json:"path"`
	RepeatStack RepeatStack `json:"repeat_stack,omitempty"`
}

// String returns a breadcrumb-style path for display.
func (nc NavContext) String() string {
	if len(nc.Path) == 0 {
		return "score"
	}
	parts := append([]string(nil), nc.Path...)
	if suffix := nc.RepeatStack.String(); suffix != "" {
		parts[len(parts)-1] = fmt.Sprintf("%s %s", parts[len(parts)-1], suffix)
	}
	return strings.Join(parts, " › ")
}

// NormalizeTemplateDefinition validates and normalizes a template definition.
func NormalizeTemplateDefinition(def *TemplateDefinition) error {
	if def == nil {
		return fmt.Errorf("template definition is nil")
	}
	if def.Structure == nil {
		return fmt.Errorf("template must define structure")
	}

	def.Structure = NormalizeShapeTree(def.Structure)
	def.Structure.NormalizeIDs("root")
	if err := def.Structure.Validate(); err != nil {
		return err
	}
	if err := ValidateDependencySpecs(def.Structure); err != nil {
		return err
	}
	return nil
}

// ChildNodes returns the active child list for a node.
func (n *ShapeNode) ChildNodes() []ShapeNode {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case ShapeProcedure:
		if len(n.Steps) > 0 {
			return n.Steps
		}
		return n.ChildList
	case ShapeChecklist, ShapeGroup:
		if len(n.Items) > 0 {
			return n.Items
		}
		return n.ChildList
	default:
		return n.ChildList
	}
}

// DisplayTitle returns the best label for a node.
func (n *ShapeNode) DisplayTitle() string {
	if n.Title != "" {
		return n.Title
	}
	switch n.Kind {
	case ShapeAction:
		return "Action"
	case ShapeProcedure:
		return "Procedure"
	case ShapeChecklist:
		return "Checklist"
	case ShapeLog:
		return "Log"
	case ShapeArtifact:
		return "Artifacts"
	case ShapeGroup:
		return "Group"
	default:
		return n.ID
	}
}

// NormalizeIDs assigns IDs to nodes missing them.
func (n *ShapeNode) NormalizeIDs(prefix string) {
	if n.ID == "" {
		n.ID = prefix
	}
	normalizeSlice := func(children *[]ShapeNode, prefixFn func(int) string) {
		for i := range *children {
			(*children)[i].NormalizeIDs(prefixFn(i))
		}
	}
	switch n.Kind {
	case ShapeProcedure:
		if len(n.Steps) > 0 {
			normalizeSlice(&n.Steps, func(i int) string { return fmt.Sprintf("%s.s%d", n.ID, i+1) })
		} else {
			normalizeSlice(&n.ChildList, func(i int) string { return fmt.Sprintf("%s.s%d", n.ID, i+1) })
		}
	case ShapeChecklist, ShapeGroup:
		if len(n.Items) > 0 {
			normalizeSlice(&n.Items, func(i int) string { return fmt.Sprintf("%s.i%d", n.ID, i+1) })
		} else {
			normalizeSlice(&n.ChildList, func(i int) string { return fmt.Sprintf("%s.i%d", n.ID, i+1) })
		}
	default:
		normalizeSlice(&n.ChildList, func(i int) string { return fmt.Sprintf("%s.c%d", n.ID, i+1) })
	}
}

// Validate validates a structure subtree.
func (n *ShapeNode) Validate() error {
	if n == nil {
		return fmt.Errorf("nil shape node")
	}
	if n.Kind == "" {
		return fmt.Errorf("shape node %q missing kind", n.ID)
	}

	switch n.Kind {
	case ShapeAction:
		// leaf action — always valid
	case ShapeProcedure:
		if len(n.ChildNodes()) == 0 {
			return fmt.Errorf("procedure %q must have steps or children", n.ID)
		}
		for i := range n.ChildNodes() {
			if err := n.ChildNodes()[i].Validate(); err != nil {
				return err
			}
		}
	case ShapeChecklist, ShapeGroup:
		if len(n.ChildNodes()) == 0 && n.Title == "" {
			return fmt.Errorf("checklist %q must have items or a title", n.ID)
		}
		for i := range n.ChildNodes() {
			if err := n.ChildNodes()[i].Validate(); err != nil {
				return err
			}
		}
	case ShapeLog:
		if n.RecordSchema == nil || len(n.RecordSchema.Fields) == 0 {
			return fmt.Errorf("log %q must define record_schema", n.ID)
		}
	case ShapeArtifact:
		// Artifact container nodes may have empty items (filled at runtime).
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
	for i := range n.ChildNodes() {
		child := n.ChildNodes()[i]
		if found := child.FindByPath(rest); found != nil {
			return found
		}
	}
	return nil
}

// ChecklistForScope returns the checklist to edit at this node (direct child first, else nested).
func (n *ShapeNode) ChecklistForScope() *ShapeNode {
	if n == nil {
		return nil
	}
	for i := range n.Steps {
		if n.Steps[i].Kind == ShapeChecklist {
			return &n.Steps[i]
		}
	}
	for i := range n.ChildNodes() {
		child := n.ChildNodes()[i]
		if child.Kind == ShapeChecklist {
			return &child
		}
	}
	return nil
}

// NavigableChildren returns tree rows, omitting checklists shown inline at this level.
func (n *ShapeNode) NavigableChildren() []ShapeNode {
	children := n.ChildNodes()
	if n.ChecklistForScope() != nil && n.Kind == ShapeProcedure {
		filtered := make([]ShapeNode, 0, len(children))
		for i := range children {
			if children[i].Kind == ShapeChecklist {
				continue
			}
			filtered = append(filtered, children[i])
		}
		return filtered
	}
	return children
}

// FindFirstLogNode returns the first log shape node in the tree.
func (n *ShapeNode) FindFirstLogNode() *ShapeNode {
	if n == nil {
		return nil
	}
	if n.Kind == ShapeLog && n.RecordSchema != nil {
		return n
	}
	for i := range n.ChildNodes() {
		child := n.ChildNodes()[i]
		if found := child.FindFirstLogNode(); found != nil {
			return found
		}
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
		for i := range node.ChildNodes() {
			child := node.ChildNodes()[i]
			if found := walk(&child, current); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(n, nil)
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
		for i := range node.ChildNodes() {
			walk(&node.ChildNodes()[i])
		}
	}
	walk(n)
	return kinds
}
