package models

// GroupArtifacts splits artifacts into inputs and outputs.
func GroupArtifacts(artifacts []*Artifact) (inputs, outputs []*Artifact) {
	for _, art := range artifacts {
		if art.ArtifactType == "input" {
			inputs = append(inputs, art)
		} else {
			outputs = append(outputs, art)
		}
	}
	return inputs, outputs
}

// ArtifactProvidedIcon returns a status indicator for whether a value is set.
func ArtifactProvidedIcon(value string) string {
	if value != "" {
		return "✓"
	}
	return " "
}
