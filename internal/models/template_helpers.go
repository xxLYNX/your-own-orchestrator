package models

// TypeLabel returns a short type label for list views.
func (t *Template) TypeLabel() string {
	if t.IsBuiltin {
		return "builtin"
	}
	return "custom"
}

// TypeLabelDisplay returns a human-readable type label for detail views.
func (t *Template) TypeLabelDisplay() string {
	if t.IsBuiltin {
		return "Built-in"
	}
	return "Custom"
}
