package models

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// templateYAML is the on-disk template document shape.
type templateYAML struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Category     string            `yaml:"category"`
	Structure    *ShapeNode        `yaml:"structure"`
	Inputs       []TemplateInput   `yaml:"inputs"`
	Outputs      []TemplateOutput  `yaml:"outputs"`
	RecordSchema *RecordSchema     `yaml:"record_schema"`
	Metadata     TemplateMetadata  `yaml:"metadata"`
	Examples     []TemplateExample `yaml:"examples"`
	Notes        string            `yaml:"notes"`
}

// ParseTemplateYAML parses a template definition from YAML bytes.
func ParseTemplateYAML(data []byte, isBuiltin bool) (*Template, error) {
	var doc templateYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}
	if doc.Name == "" {
		return nil, fmt.Errorf("template must have a name")
	}
	if doc.Version == "" {
		doc.Version = "1.0.0"
	}

	template := &Template{
		Name:        doc.Name,
		Version:     doc.Version,
		Description: doc.Description,
		Category:    doc.Category,
		IsBuiltin:   isBuiltin,
		Definition: TemplateDefinition{
			Structure:    doc.Structure,
			Inputs:       doc.Inputs,
			Outputs:      doc.Outputs,
			RecordSchema: doc.RecordSchema,
			Metadata:     doc.Metadata,
			Examples:     doc.Examples,
			Notes:        doc.Notes,
		},
	}
	if err := template.Definition.Validate(); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}
	return template, nil
}

// MarshalTemplateYAML serializes a template to YAML for export.
func MarshalTemplateYAML(template *Template) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}
	structure, err := template.Definition.GetStructure()
	if err != nil {
		return nil, err
	}
	doc := templateYAML{
		Name:         template.Name,
		Version:      template.Version,
		Description:  template.Description,
		Category:     template.Category,
		Structure:    structure,
		Inputs:       template.Definition.Inputs,
		Outputs:      template.Definition.Outputs,
		RecordSchema: template.Definition.RecordSchema,
		Metadata:     template.Definition.Metadata,
		Examples:     template.Definition.Examples,
		Notes:        template.Definition.Notes,
	}
	return yaml.Marshal(doc)
}

// MarshalTemplateJSON serializes a template to JSON for export.
func MarshalTemplateJSON(template *Template) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}
	structure, err := template.Definition.GetStructure()
	if err != nil {
		return nil, err
	}
	doc := templateYAML{
		Name:         template.Name,
		Version:      template.Version,
		Description:  template.Description,
		Category:     template.Category,
		Structure:    structure,
		Inputs:       template.Definition.Inputs,
		Outputs:      template.Definition.Outputs,
		RecordSchema: template.Definition.RecordSchema,
		Metadata:     template.Definition.Metadata,
		Examples:     template.Definition.Examples,
		Notes:        template.Definition.Notes,
	}
	return json.MarshalIndent(doc, "", "  ")
}
