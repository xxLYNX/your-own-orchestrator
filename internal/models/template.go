package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Template represents a reusable note structure
type Template struct {
	ID          int64              `json:"id" db:"id"`
	Name        string             `json:"name" db:"name"`
	Version     string             `json:"version" db:"version"`
	Description string             `json:"description" db:"description"`
	Category    string             `json:"category" db:"category"`
	Definition  TemplateDefinition `json:"definition" db:"definition"` // Stored as JSON
	IsBuiltin   bool               `json:"is_builtin" db:"is_builtin"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// TemplateDefinition is the complete template structure
type TemplateDefinition struct {
	Inputs       []TemplateInput   `json:"inputs" yaml:"inputs"`
	Composition  *ShapeNode        `json:"composition,omitempty" yaml:"composition,omitempty"`
	Steps        []TemplateStep    `json:"steps" yaml:"steps"`
	Outputs      []TemplateOutput  `json:"outputs" yaml:"outputs"`
	RecordSchema *RecordSchema     `json:"record_schema,omitempty" yaml:"record_schema,omitempty"`
	Metadata     TemplateMetadata  `json:"metadata" yaml:"metadata"`
	Examples     []TemplateExample `json:"examples,omitempty" yaml:"examples,omitempty"`
	Notes        string            `json:"notes,omitempty" yaml:"notes,omitempty"`
}

// GetComposition returns the normalized composition tree for this template.
func (td *TemplateDefinition) GetComposition() (*ShapeNode, error) {
	if err := NormalizeTemplateDefinition(td); err != nil {
		return nil, err
	}
	return td.Composition, nil
}

// RecordSchema defines the structure for repeating log-style records
type RecordSchema struct {
	Fields []RecordField `json:"fields" yaml:"fields"`
}

// RecordField defines a single field in a record
type RecordField struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"` // text, integer, date, enum, url, boolean
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool     `json:"required" yaml:"required"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Values      []string `json:"values,omitempty" yaml:"values,omitempty"` // For enum types
}

// TemplateInput defines required or optional inputs
type TemplateInput struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"` // file, url, text, integer, boolean, date
	Description string   `json:"description" yaml:"description"`
	Required    bool     `json:"required" yaml:"required"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []string `json:"options,omitempty" yaml:"options,omitempty"` // For enum-like inputs
}

// TemplateStep defines a procedural step in the workflow
type TemplateStep struct {
	ID             int      `json:"id" yaml:"id"`
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description" yaml:"description"`
	Checklist      []string `json:"checklist,omitempty" yaml:"checklist,omitempty"`
	EstimatedTime  string   `json:"estimated_time,omitempty" yaml:"estimated_time,omitempty"`
	OutputRequired string   `json:"output_required,omitempty" yaml:"output_required,omitempty"` // Name of output that must be provided
}

// TemplateOutput defines expected outputs/deliverables
type TemplateOutput struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"` // file, folder, text, url
	Description string   `json:"description" yaml:"description"`
	Format      string   `json:"format,omitempty" yaml:"format,omitempty"` // csv, xlsx, pdf, markdown, etc.
	Required    bool     `json:"required" yaml:"required"`
	Fields      []string `json:"fields,omitempty" yaml:"fields,omitempty"` // Expected fields in output (e.g., for CSV)
}

// TemplateMetadata contains template metadata
type TemplateMetadata struct {
	Tags              []string `json:"tags" yaml:"tags"`
	EstimatedDuration string   `json:"estimated_duration" yaml:"estimated_duration"`
	Difficulty        string   `json:"difficulty,omitempty" yaml:"difficulty,omitempty"` // easy, medium, hard
}

// TemplateExample provides usage examples
type TemplateExample struct {
	Description string `json:"description" yaml:"description"`
	Command     string `json:"command" yaml:"command"`
}

// NoteTemplate links a note to a template with instance data
type NoteTemplate struct {
	ID           int64            `json:"id" db:"id"`
	NoteID       int64            `json:"note_id" db:"note_id"`
	TemplateID   int64            `json:"template_id" db:"template_id"`
	TemplateData TemplateInstance `json:"template_data" db:"template_data"` // Stored as JSON
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
}

// TemplateInstance is the instantiated template with user-provided values
type TemplateInstance struct {
	Inputs  map[string]interface{} `json:"inputs"`
	Steps   []StepInstance         `json:"steps"`
	Outputs map[string]*Artifact   `json:"outputs"` // Key is output name
}

// StepInstance tracks the completion of a specific step
type StepInstance struct {
	ID          int64      `json:"id" db:"id"`
	StepNumber  int        `json:"step_number" db:"step_number"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Completed   bool       `json:"completed" db:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Notes       string     `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Artifact represents an input or output artifact
type Artifact struct {
	ID             int64     `json:"id" db:"id"`
	NoteTemplateID int64     `json:"note_template_id" db:"note_template_id"`
	ArtifactType   string    `json:"artifact_type" db:"artifact_type"` // "input" or "output"
	Name           string    `json:"name" db:"name"`
	Type           string    `json:"type" db:"type"` // file, url, text, folder
	Value          string    `json:"value" db:"value"`
	Description    string    `json:"description" db:"description"`
	Required       bool      `json:"required" db:"required"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Reference links to external resources
type Reference struct {
	ID          int64     `json:"id" db:"id"`
	NoteID      int64     `json:"note_id" db:"note_id"`
	Type        string    `json:"type" db:"ref_type"` // file, url, command, note
	Value       string    `json:"value" db:"ref_value"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TemplateRecord represents a single record in a log-style template
type TemplateRecord struct {
	ID             int64                  `json:"id" db:"id"`
	NoteTemplateID int64                  `json:"note_template_id" db:"note_template_id"`
	RecordIndex    int                    `json:"record_index" db:"record_index"`
	RepeatIndex    int                    `json:"repeat_index" db:"repeat_index"` // 0 = global; 1..N = repeat iteration
	Data           map[string]interface{} `json:"data" db:"data"`                 // Stored as JSON
	Status         string                 `json:"status" db:"status"` // draft, in_progress, complete
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// Template shape identifiers — see shape.go for composable primitives.
const (
	ShapeLogLegacy       = "log"
	ShapeChecklistLegacy = "checklist" // legacy alias; prefer ShapeProcedure in docs
	ShapeArtifactLegacy  = "artifact"
)

// HasLogShape reports whether the template uses the log shape.
func (td *TemplateDefinition) HasLogShape() bool {
	comp, err := td.GetComposition()
	if err != nil || comp == nil {
		return td.RecordSchema != nil && len(td.RecordSchema.Fields) > 0
	}
	return containsShapeKind(comp, ShapeLog)
}

// HasProcedureShape reports whether the template uses ordered procedures.
func (td *TemplateDefinition) HasProcedureShape() bool {
	comp, err := td.GetComposition()
	if err != nil || comp == nil {
		return len(td.Steps) > 0
	}
	return containsShapeKind(comp, ShapeProcedure)
}

// HasChecklistShape is kept for compatibility — means procedure or checklist nodes exist.
func (td *TemplateDefinition) HasChecklistShape() bool {
	return td.HasProcedureShape() || containsShapeKindFromDef(td, ShapeChecklist)
}

// HasArtifactShape reports whether the template uses artifact inputs/outputs.
func (td *TemplateDefinition) HasArtifactShape() bool {
	if len(td.Inputs) > 0 || len(td.Outputs) > 0 {
		return true
	}
	return containsShapeKindFromDef(td, ShapeArtifact)
}

func containsShapeKindFromDef(td *TemplateDefinition, kind string) bool {
	comp, err := td.GetComposition()
	if err != nil || comp == nil {
		return false
	}
	return containsShapeKind(comp, kind)
}

func containsShapeKind(node *ShapeNode, kind string) bool {
	if node == nil {
		return false
	}
	if node.Kind == kind {
		return true
	}
	for i := range node.Steps {
		if containsShapeKind(&node.Steps[i], kind) {
			return true
		}
	}
	for i := range node.Items {
		if containsShapeKind(&node.Items[i], kind) {
			return true
		}
	}
	if node.RepeatBody != nil && containsShapeKind(node.RepeatBody, kind) {
		return true
	}
	return false
}

// ActiveShapes returns human-readable shape kinds in this template.
func (td *TemplateDefinition) ActiveShapes() []string {
	comp, err := td.GetComposition()
	if err != nil || comp == nil {
		var shapes []string
		if td.HasLogShape() {
			shapes = append(shapes, ShapeLog)
		}
		if len(td.Steps) > 0 {
			shapes = append(shapes, ShapeProcedure)
		}
		if td.HasArtifactShape() {
			shapes = append(shapes, ShapeArtifact)
		}
		return shapes
	}
	return comp.ActiveShapeKinds()
}

// Validate checks if the template definition is valid
func (td *TemplateDefinition) Validate() error {
	if err := NormalizeTemplateDefinition(td); err != nil {
		return err
	}

	if td.Composition == nil {
		return fmt.Errorf("template must define a composition tree or legacy shape fields")
	}

	// Validate step IDs are sequential
	for i, step := range td.Steps {
		if step.ID != i+1 {
			return fmt.Errorf("step IDs must be sequential starting from 1, got %d at position %d", step.ID, i)
		}
		if step.Title == "" {
			return fmt.Errorf("step %d must have a title", step.ID)
		}
	}

	// Validate input types
	validInputTypes := map[string]bool{
		"file": true, "url": true, "text": true, "integer": true,
		"boolean": true, "date": true, "folder": true,
	}
	for _, input := range td.Inputs {
		if input.Name == "" {
			return fmt.Errorf("input must have a name")
		}
		if !validInputTypes[input.Type] {
			return fmt.Errorf("invalid input type '%s' for input '%s'", input.Type, input.Name)
		}
	}

	// Validate output types
	validOutputTypes := map[string]bool{
		"file": true, "folder": true, "text": true, "url": true,
	}
	for _, output := range td.Outputs {
		if output.Name == "" {
			return fmt.Errorf("output must have a name")
		}
		if !validOutputTypes[output.Type] {
			return fmt.Errorf("invalid output type '%s' for output '%s'", output.Type, output.Name)
		}
	}

	// Validate output references in steps
	outputMap := make(map[string]bool)
	for _, output := range td.Outputs {
		outputMap[output.Name] = true
	}
	for _, step := range td.Steps {
		if step.OutputRequired != "" && !outputMap[step.OutputRequired] {
			return fmt.Errorf("step %d references non-existent output '%s'", step.ID, step.OutputRequired)
		}
	}

	return nil
}

// ValidateRecord checks if a record matches the schema requirements
func (rs *RecordSchema) ValidateRecord(data map[string]interface{}) error {
	if rs == nil {
		return fmt.Errorf("no record schema defined")
	}

	// Check all required fields are present
	for _, field := range rs.Fields {
		if field.Required {
			if _, exists := data[field.Name]; !exists {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	// Validate field types
	validTypes := map[string]bool{
		"text": true, "integer": true, "date": true, "enum": true,
		"url": true, "boolean": true,
	}
	for _, field := range rs.Fields {
		if !validTypes[field.Type] {
			return fmt.Errorf("invalid field type '%s' for field '%s'", field.Type, field.Name)
		}

		// For enum types, validate values list
		if field.Type == "enum" && len(field.Values) == 0 {
			return fmt.Errorf("enum field '%s' must have values defined", field.Name)
		}
	}

	// Validate data values against field definitions
	for name, value := range data {
		// Find field definition
		var fieldDef *RecordField
		for i := range rs.Fields {
			if rs.Fields[i].Name == name {
				fieldDef = &rs.Fields[i]
				break
			}
		}

		if fieldDef == nil {
			return fmt.Errorf("unknown field '%s'", name)
		}

		// Basic type validation
		switch fieldDef.Type {
		case "integer":
			switch value.(type) {
			case int, int64, float64:
				// OK
			default:
				return fmt.Errorf("field '%s' must be an integer", name)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("field '%s' must be a boolean", name)
			}
		case "enum":
			strVal, ok := value.(string)
			if !ok {
				return fmt.Errorf("field '%s' must be a string", name)
			}
			validValue := false
			for _, allowed := range fieldDef.Values {
				if strVal == allowed {
					validValue = true
					break
				}
			}
			if !validValue {
				return fmt.Errorf("field '%s' has invalid value '%s'", name, strVal)
			}
		}
	}

	return nil
}

// GetRequiredInputs returns all required inputs
func (td *TemplateDefinition) GetRequiredInputs() []TemplateInput {
	var required []TemplateInput
	for _, input := range td.Inputs {
		if input.Required {
			required = append(required, input)
		}
	}
	return required
}

// GetRequiredOutputs returns all required outputs
func (td *TemplateDefinition) GetRequiredOutputs() []TemplateOutput {
	var required []TemplateOutput
	for _, output := range td.Outputs {
		if output.Required {
			required = append(required, output)
		}
	}
	return required
}

// CalculateProgress calculates the completion percentage of a template instance
func (ti *TemplateInstance) CalculateProgress() float64 {
	return ti.CalculateProgressWithRecords(0, nil)
}

// CalculateProgressWithRecords factors in log rows and repeat targets when available.
func (ti *TemplateInstance) CalculateProgressWithRecords(recordCount int, def *TemplateDefinition) float64 {
	stepTotal := len(ti.Steps)
	stepDone := ti.GetCompletedSteps()

	if def == nil || stepTotal == 0 {
		if stepTotal == 0 {
			return 0
		}
		return float64(stepDone) / float64(stepTotal)
	}

	comp, err := def.GetComposition()
	if err != nil || comp == nil {
		return float64(stepDone) / float64(stepTotal)
	}

	var repeatNode *ShapeNode
	var findRepeat func(*ShapeNode)
	findRepeat = func(n *ShapeNode) {
		if n == nil || repeatNode != nil {
			return
		}
		if n.Kind == ShapeRepeat {
			repeatNode = n
			return
		}
		for i := range n.Steps {
			findRepeat(&n.Steps[i])
		}
		if n.RepeatBody != nil {
			findRepeat(n.RepeatBody)
		}
	}
	findRepeat(comp)

	if repeatNode == nil {
		return float64(stepDone) / float64(stepTotal)
	}

	target := repeatNode.ResolveRepeatCount(ti.Inputs)
	if target <= 0 {
		return float64(stepDone) / float64(stepTotal)
	}

	if recordCount > target {
		recordCount = target
	}

	totalUnits := stepTotal + target
	doneUnits := stepDone + recordCount
	return float64(doneUnits) / float64(totalUnits)
}

// GetCompletedSteps returns the number of completed steps
func (ti *TemplateInstance) GetCompletedSteps() int {
	count := 0
	for _, step := range ti.Steps {
		if step.Completed {
			count++
		}
	}
	return count
}

// IsComplete checks if all required steps and outputs are complete
func (ti *TemplateInstance) IsComplete(definition *TemplateDefinition) bool {
	// Check all steps are completed
	for _, step := range ti.Steps {
		if !step.Completed {
			return false
		}
	}

	// Check all required outputs are provided
	for _, output := range definition.GetRequiredOutputs() {
		artifact, exists := ti.Outputs[output.Name]
		if !exists || artifact == nil {
			return false
		}
	}

	return true
}

// MarshalJSON custom JSON marshaling for TemplateDefinition
func (td *TemplateDefinition) MarshalJSON() ([]byte, error) {
	type Alias TemplateDefinition
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(td),
	})
}

// UnmarshalJSON custom JSON unmarshaling for TemplateDefinition
func (td *TemplateDefinition) UnmarshalJSON(data []byte) error {
	type Alias TemplateDefinition
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(td),
	}
	return json.Unmarshal(data, aux)
}

// NewTemplateInstance creates a new template instance from a definition
func NewTemplateInstance(definition *TemplateDefinition, inputs map[string]interface{}) *TemplateInstance {
	_ = NormalizeTemplateDefinition(definition)

	// Initialize steps from normalized legacy sync (top-level procedures)
	steps := make([]StepInstance, len(definition.Steps))
	for i, defStep := range definition.Steps {
		steps[i] = StepInstance{
			StepNumber:  defStep.ID,
			Title:       defStep.Title,
			Description: defStep.Description,
			Completed:   false,
			CreatedAt:   time.Now(),
		}
	}

	// Initialize outputs map
	outputs := make(map[string]*Artifact)
	for _, defOutput := range definition.Outputs {
		outputs[defOutput.Name] = nil // Will be filled when user provides output
	}

	return &TemplateInstance{
		Inputs:  inputs,
		Steps:   steps,
		Outputs: outputs,
	}
}

// ValidateInputs checks if provided inputs match template requirements
func (td *TemplateDefinition) ValidateInputs(inputs map[string]interface{}) error {
	// Check all required inputs are provided
	for _, input := range td.Inputs {
		if input.Required {
			if _, exists := inputs[input.Name]; !exists {
				return fmt.Errorf("required input '%s' is missing", input.Name)
			}
		}
	}

	// Check input types (basic validation)
	for name, value := range inputs {
		// Find the input definition
		var inputDef *TemplateInput
		for i := range td.Inputs {
			if td.Inputs[i].Name == name {
				inputDef = &td.Inputs[i]
				break
			}
		}

		if inputDef == nil {
			return fmt.Errorf("unknown input '%s'", name)
		}

		// Basic type validation
		switch inputDef.Type {
		case "integer":
			switch value.(type) {
			case int, int64, float64:
				// OK
			default:
				return fmt.Errorf("input '%s' must be an integer", name)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("input '%s' must be a boolean", name)
			}
		}
	}

	return nil
}
