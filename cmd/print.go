package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"yoo/internal/models"
	"yoo/internal/strutil"
)

func printArtifactGroup(label string, artifacts []*models.Artifact) {
	if len(artifacts) == 0 {
		return
	}
	fmt.Println(label)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, art := range artifacts {
		required := ""
		if art.Required {
			required = " [required]"
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\t%s%s\n",
			models.ArtifactProvidedIcon(art.Value),
			art.Name,
			art.Type,
			art.Description,
			required,
		)
	}
	_ = w.Flush()
	fmt.Println()
}

func printRecordsTable(records []*models.TemplateRecord, schema *models.RecordSchema) {
	if len(records) == 0 {
		return
	}

	fieldNames := schema.FieldNames()
	if len(fieldNames) == 0 {
		for name := range records[0].Data {
			fieldNames = append(fieldNames, name)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(w, "#\t")
	_, _ = fmt.Fprint(w, joinTabs(fieldNames))
	_, _ = fmt.Fprint(w, "\tStatus\n")

	_, _ = fmt.Fprint(w, "---\t")
	for range fieldNames {
		_, _ = fmt.Fprint(w, "---\t")
	}
	_, _ = fmt.Fprint(w, "---\n")

	for _, rec := range records {
		_, _ = fmt.Fprintf(w, "%d\t", rec.RecordIndex)
		for _, fieldName := range fieldNames {
			_, _ = fmt.Fprintf(w, "%s\t", strutil.Truncate(fmt.Sprintf("%v", rec.Data[fieldName]), 30))
		}
		_, _ = fmt.Fprintf(w, "%s\n", rec.Status)
	}
	_ = w.Flush()
}

func joinTabs(values []string) string {
	out := ""
	for i, v := range values {
		if i > 0 {
			out += "\t"
		}
		out += v
	}
	return out
}

func printTemplateInputs(inputs []models.TemplateInput) {
	fmt.Println("INPUTS:")
	if len(inputs) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, input := range inputs {
		printTemplateIOItem(input.Name, input.Type, input.Description, input.Required, input.Default, "")
	}
}

func printTemplateOutputs(outputs []models.TemplateOutput) {
	fmt.Println("OUTPUTS:")
	if len(outputs) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, output := range outputs {
		printTemplateIOItem(output.Name, output.Type, output.Description, output.Required, "", output.Format)
	}
}

func printTemplateIOItem(name, typ, description string, required bool, defaultVal, format string) {
	requiredSuffix := ""
	if required {
		requiredSuffix = " [required]"
	}
	fmt.Printf("  • %s (%s)%s\n", name, typ, requiredSuffix)
	fmt.Printf("    %s\n", description)
	if defaultVal != "" {
		fmt.Printf("    Default: %s\n", defaultVal)
	}
	if format != "" {
		fmt.Printf("    Format: %s\n", format)
	}
}
