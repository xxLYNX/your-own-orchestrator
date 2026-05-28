package models

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// FieldByName returns the schema field with the given name, or nil.
func (rs *RecordSchema) FieldByName(name string) *RecordField {
	if rs == nil {
		return nil
	}
	for i := range rs.Fields {
		if rs.Fields[i].Name == name {
			return &rs.Fields[i]
		}
	}
	return nil
}

// FieldNames returns field names in schema order.
func (rs *RecordSchema) FieldNames() []string {
	if rs == nil {
		return nil
	}
	names := make([]string, len(rs.Fields))
	for i, field := range rs.Fields {
		names[i] = field.Name
	}
	return names
}

// ParseStringValue coerces a raw string into the typed value for this field.
func (rf RecordField) ParseStringValue(raw string) (interface{}, error) {
	switch rf.Type {
	case "integer":
		val, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("field '%s' must be an integer", rf.Name)
		}
		return val, nil
	case "boolean":
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("field '%s' must be true/false", rf.Name)
		}
		return val, nil
	default:
		return raw, nil
	}
}

// PromptLabel builds the interactive CLI prompt for a field.
func (rf RecordField) PromptLabel() string {
	prompt := rf.Name
	if rf.Description != "" {
		prompt += fmt.Sprintf(" (%s)", rf.Description)
	}
	if rf.Required {
		prompt += " [required]"
	}
	if rf.Default != "" {
		prompt += fmt.Sprintf(" [default: %s]", rf.Default)
	}
	if rf.Type == "enum" && len(rf.Values) > 0 {
		prompt += fmt.Sprintf(" [%s]", strings.Join(rf.Values, ", "))
	}
	return prompt + ": "
}

// ApplyFieldString sets one field on data using schema-aware type coercion.
func (rs *RecordSchema) ApplyFieldString(data map[string]interface{}, name, raw string) error {
	if field := rs.FieldByName(name); field != nil {
		val, err := field.ParseStringValue(raw)
		if err != nil {
			return err
		}
		data[name] = val
		return nil
	}
	data[name] = raw
	return nil
}

// ApplyFieldFlags parses name=value flags and applies them to data.
func (rs *RecordSchema) ApplyFieldFlags(data map[string]interface{}, flags []string) error {
	for _, fieldStr := range flags {
		name, value, err := ParseFieldFlag(fieldStr)
		if err != nil {
			return err
		}
		if err := rs.ApplyFieldString(data, name, value); err != nil {
			return err
		}
	}
	return nil
}

// ParseFieldFlag parses a single name=value flag.
func ParseFieldFlag(fieldStr string) (name, value string, err error) {
	parts := strings.SplitN(fieldStr, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid field format '%s', expected name=value", fieldStr)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

// PromptForMissingFields interactively fills unset fields in data.
func (rs *RecordSchema) PromptForMissingFields(data map[string]interface{}, reader *bufio.Reader) error {
	for _, field := range rs.Fields {
		if _, exists := data[field.Name]; exists {
			continue
		}

		fmt.Print(field.PromptLabel())
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" && field.Default != "" {
			input = field.Default
		}
		if input == "" && field.Required {
			return fmt.Errorf("required field '%s' cannot be empty", field.Name)
		}
		if input == "" {
			continue
		}

		val, err := field.ParseStringValue(input)
		if err != nil {
			return err
		}
		data[field.Name] = val
	}
	return nil
}

// FilterTemplateRecords returns records matching optional status and search filters.
func FilterTemplateRecords(records []*TemplateRecord, statusFilter, search string) []*TemplateRecord {
	if statusFilter == "all" {
		statusFilter = ""
	}
	search = strings.TrimSpace(search)
	if statusFilter == "" && search == "" {
		return records
	}

	filtered := make([]*TemplateRecord, 0, len(records))
	searchLower := strings.ToLower(search)

	for _, record := range records {
		if statusFilter != "" && record.Status != statusFilter {
			continue
		}
		if search != "" {
			match := false
			for _, value := range record.Data {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), searchLower) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, record)
	}
	return filtered
}
