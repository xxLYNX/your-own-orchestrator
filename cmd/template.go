package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	templateCategory string
	templateFormat   string
)

// templatesCmd represents the templates command
var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"template", "tpl"},
	Short:   "Manage note templates",
	Long: `Manage note templates for structured workflows.

Templates define reusable patterns with inputs, procedural steps, and expected outputs.
They help structure complex tasks while keeping simple notes simple.`,
}

// templatesListCmd lists all templates
var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available templates",
	Long: `List all available templates, including built-in and custom templates.

Optionally filter by category using the --category flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// List templates
		templates, err := database.ListTemplates(db.Conn(), templateCategory)
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		if len(templates) == 0 {
			if templateCategory != "" {
				fmt.Printf("No templates found in category '%s'\n", templateCategory)
			} else {
				fmt.Println("No templates available")
				fmt.Println("\nTo import a template, use: yoo template import <file>")
			}
			return nil
		}

		// Display templates
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tCATEGORY\tDESCRIPTION\tTYPE")
		fmt.Fprintln(w, "----\t-------\t--------\t-----------\t----")

		for _, tpl := range templates {
			tplType := "custom"
			if tpl.IsBuiltin {
				tplType = "builtin"
			}

			desc := tpl.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				tpl.Name,
				tpl.Version,
				tpl.Category,
				desc,
				tplType,
			)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d template(s)\n", len(templates))
		fmt.Println("\nUse 'yoo template show <name>' to see template details")

		return nil
	},
}

// templatesShowCmd shows details of a template
var templatesShowCmd = &cobra.Command{
	Use:   "show <template-name>",
	Short: "Show template details",
	Long:  `Display detailed information about a specific template including inputs, steps, and outputs.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get template
		template, err := database.GetTemplateByName(db.Conn(), templateName)
		if err != nil {
			return fmt.Errorf("template '%s' not found: %w", templateName, err)
		}

		// Display template details
		fmt.Printf("Template: %s\n", template.Name)
		fmt.Printf("Version: %s\n", template.Version)
		fmt.Printf("Category: %s\n", template.Category)
		fmt.Printf("Description: %s\n", template.Description)
		if template.IsBuiltin {
			fmt.Println("Type: Built-in")
		} else {
			fmt.Println("Type: Custom")
		}
		fmt.Println()

		// Display metadata
		if len(template.Definition.Metadata.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(template.Definition.Metadata.Tags, ", "))
		}
		if template.Definition.Metadata.EstimatedDuration != "" {
			fmt.Printf("Estimated Duration: %s\n", template.Definition.Metadata.EstimatedDuration)
		}
		if template.Definition.Metadata.Difficulty != "" {
			fmt.Printf("Difficulty: %s\n", template.Definition.Metadata.Difficulty)
		}
		fmt.Println()

		// Display inputs
		fmt.Println("INPUTS:")
		if len(template.Definition.Inputs) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, input := range template.Definition.Inputs {
				required := ""
				if input.Required {
					required = " [required]"
				}
				fmt.Printf("  • %s (%s)%s\n", input.Name, input.Type, required)
				fmt.Printf("    %s\n", input.Description)
				if input.Default != "" {
					fmt.Printf("    Default: %s\n", input.Default)
				}
			}
		}
		fmt.Println()

		// Display steps
		fmt.Println("STEPS:")
		for _, step := range template.Definition.Steps {
			fmt.Printf("  %d. %s\n", step.ID, step.Title)
			if step.Description != "" {
				fmt.Printf("     %s\n", step.Description)
			}
			if len(step.Checklist) > 0 {
				for _, item := range step.Checklist {
					fmt.Printf("     ☐ %s\n", item)
				}
			}
			if step.EstimatedTime != "" {
				fmt.Printf("     Estimated: %s\n", step.EstimatedTime)
			}
		}
		fmt.Println()

		// Display outputs
		fmt.Println("OUTPUTS:")
		if len(template.Definition.Outputs) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, output := range template.Definition.Outputs {
				required := ""
				if output.Required {
					required = " [required]"
				}
				fmt.Printf("  • %s (%s)%s\n", output.Name, output.Type, required)
				fmt.Printf("    %s\n", output.Description)
				if output.Format != "" {
					fmt.Printf("    Format: %s\n", output.Format)
				}
			}
		}
		fmt.Println()

		// Display examples
		if len(template.Definition.Examples) > 0 {
			fmt.Println("EXAMPLES:")
			for _, example := range template.Definition.Examples {
				fmt.Printf("  %s\n", example.Description)
				fmt.Printf("  $ %s\n", example.Command)
				fmt.Println()
			}
		}

		// Display notes
		if template.Definition.Notes != "" {
			fmt.Println("NOTES:")
			fmt.Println(template.Definition.Notes)
			fmt.Println()
		}

		return nil
	},
}

// templatesImportCmd imports a template from a file
var templatesImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a template from a YAML file",
	Long: `Import a template definition from a YAML file.

The template will be added to your yoo installation and can be used to create notes.

Example:
  yoo template import my-template.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		// Parse YAML
		var templateDef struct {
			Name         string                   `yaml:"name"`
			Version      string                   `yaml:"version"`
			Description  string                   `yaml:"description"`
			Category     string                   `yaml:"category"`
			Inputs       []models.TemplateInput   `yaml:"inputs"`
			Steps        []models.TemplateStep    `yaml:"steps"`
			Outputs      []models.TemplateOutput  `yaml:"outputs"`
			RecordSchema *models.RecordSchema     `yaml:"record_schema"`
			Metadata     models.TemplateMetadata  `yaml:"metadata"`
			Examples     []models.TemplateExample `yaml:"examples"`
			Notes        string                   `yaml:"notes"`
		}

		if err := yaml.Unmarshal(data, &templateDef); err != nil {
			return fmt.Errorf("failed to parse template YAML: %w", err)
		}

		// Validate required fields
		if templateDef.Name == "" {
			return fmt.Errorf("template must have a name")
		}
		if templateDef.Version == "" {
			templateDef.Version = "1.0.0"
		}

		// Create template
		template := &models.Template{
			Name:        templateDef.Name,
			Version:     templateDef.Version,
			Description: templateDef.Description,
			Category:    templateDef.Category,
			IsBuiltin:   false,
			Definition: models.TemplateDefinition{
				Inputs:       templateDef.Inputs,
				Steps:        templateDef.Steps,
				Outputs:      templateDef.Outputs,
				RecordSchema: templateDef.RecordSchema,
				Metadata:     templateDef.Metadata,
				Examples:     templateDef.Examples,
				Notes:        templateDef.Notes,
			},
		}

		// Validate template
		if err := template.Definition.Validate(); err != nil {
			return fmt.Errorf("template validation failed: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Check if template already exists
		existing, _ := database.GetTemplateByName(db.Conn(), template.Name)
		if existing != nil {
			return fmt.Errorf("template '%s' already exists. Delete it first or choose a different name", template.Name)
		}

		// Create template
		if err := database.CreateTemplate(db.Conn(), template); err != nil {
			return fmt.Errorf("failed to import template: %w", err)
		}

		fmt.Printf("✓ Template '%s' imported successfully\n", template.Name)
		fmt.Printf("  Version: %s\n", template.Version)
		fmt.Printf("  Category: %s\n", template.Category)
		fmt.Printf("  Steps: %d\n", len(template.Definition.Steps))
		fmt.Println("\nUse it with: yoo add \"task\" --template", template.Name)

		return nil
	},
}

// templatesExportCmd exports a template to a file
var templatesExportCmd = &cobra.Command{
	Use:   "export <template-name>",
	Short: "Export a template to a YAML file",
	Long: `Export a template definition to a YAML file for sharing or backup.

The output will be written to stdout unless redirected to a file.

Example:
  yoo template export job-applications > my-template.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get template
		template, err := database.GetTemplateByName(db.Conn(), templateName)
		if err != nil {
			return fmt.Errorf("template '%s' not found: %w", templateName, err)
		}

		// Create export structure
		export := struct {
			Name        string                   `yaml:"name"`
			Version     string                   `yaml:"version"`
			Description string                   `yaml:"description"`
			Category    string                   `yaml:"category"`
			Inputs      []models.TemplateInput   `yaml:"inputs"`
			Steps       []models.TemplateStep    `yaml:"steps"`
			Outputs     []models.TemplateOutput  `yaml:"outputs"`
			Metadata    models.TemplateMetadata  `yaml:"metadata"`
			Examples    []models.TemplateExample `yaml:"examples,omitempty"`
			Notes       string                   `yaml:"notes,omitempty"`
		}{
			Name:        template.Name,
			Version:     template.Version,
			Description: template.Description,
			Category:    template.Category,
			Inputs:      template.Definition.Inputs,
			Steps:       template.Definition.Steps,
			Outputs:     template.Definition.Outputs,
			Metadata:    template.Definition.Metadata,
			Examples:    template.Definition.Examples,
			Notes:       template.Definition.Notes,
		}

		// Marshal to YAML or JSON based on format flag
		var output []byte
		if templateFormat == "json" {
			output, err = json.MarshalIndent(export, "", "  ")
		} else {
			output, err = yaml.Marshal(export)
		}
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}

		fmt.Print(string(output))

		return nil
	},
}

// templatesDeleteCmd deletes a template
var templatesDeleteCmd = &cobra.Command{
	Use:   "delete <template-name>",
	Short: "Delete a template",
	Long: `Delete a custom template from your yoo installation.

Built-in templates cannot be deleted.

Warning: This action cannot be undone.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get template
		template, err := database.GetTemplateByName(db.Conn(), templateName)
		if err != nil {
			return fmt.Errorf("template '%s' not found: %w", templateName, err)
		}

		// Check if built-in
		if template.IsBuiltin {
			return fmt.Errorf("cannot delete built-in template '%s'", templateName)
		}

		// Delete template
		if err := database.DeleteTemplate(db.Conn(), template.ID); err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}

		fmt.Printf("✓ Template '%s' deleted successfully\n", templateName)

		return nil
	},
}

// loadBuiltinTemplates loads built-in templates from the templates directory
var loadBuiltinTemplatesCmd = &cobra.Command{
	Use:    "load-builtins",
	Short:  "Load built-in templates from templates directory",
	Hidden: true, // Hidden from help, used internally
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Find templates directory
		templatesDir := "templates"
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			return fmt.Errorf("templates directory not found")
		}

		// Read all YAML files in templates directory
		files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
		if err != nil {
			return fmt.Errorf("failed to read templates directory: %w", err)
		}

		loaded := 0
		for _, file := range files {
			// Skip README
			if strings.HasSuffix(file, "README.yaml") {
				continue
			}

			data, err := os.ReadFile(file)
			if err != nil {
				fmt.Printf("⚠ Failed to read %s: %v\n", file, err)
				continue
			}

			// Parse template
			var templateDef struct {
				Name         string                   `yaml:"name"`
				Version      string                   `yaml:"version"`
				Description  string                   `yaml:"description"`
				Category     string                   `yaml:"category"`
				Inputs       []models.TemplateInput   `yaml:"inputs"`
				Steps        []models.TemplateStep    `yaml:"steps"`
				Outputs      []models.TemplateOutput  `yaml:"outputs"`
				RecordSchema *models.RecordSchema     `yaml:"record_schema"`
				Metadata     models.TemplateMetadata  `yaml:"metadata"`
				Examples     []models.TemplateExample `yaml:"examples"`
				Notes        string                   `yaml:"notes"`
			}

			if err := yaml.Unmarshal(data, &templateDef); err != nil {
				fmt.Printf("⚠ Failed to parse %s: %v\n", file, err)
				continue
			}

			template := &models.Template{
				Name:        templateDef.Name,
				Version:     templateDef.Version,
				Description: templateDef.Description,
				Category:    templateDef.Category,
				IsBuiltin:   true,
				Definition: models.TemplateDefinition{
					Inputs:       templateDef.Inputs,
					Steps:        templateDef.Steps,
					Outputs:      templateDef.Outputs,
					RecordSchema: templateDef.RecordSchema,
					Metadata:     templateDef.Metadata,
					Examples:     templateDef.Examples,
					Notes:        templateDef.Notes,
				},
			}

			// Check if already exists
			existing, _ := database.GetTemplateByName(db.Conn(), template.Name)
			if existing != nil {
				fmt.Printf("⊘ Template '%s' already exists, skipping\n", template.Name)
				continue
			}

			// Create template
			if err := database.CreateTemplate(db.Conn(), template); err != nil {
				fmt.Printf("⚠ Failed to load %s: %v\n", template.Name, err)
				continue
			}

			fmt.Printf("✓ Loaded built-in template: %s\n", template.Name)
			loaded++
		}

		fmt.Printf("\nLoaded %d built-in template(s)\n", loaded)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)

	// Add subcommands
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesShowCmd)
	templatesCmd.AddCommand(templatesImportCmd)
	templatesCmd.AddCommand(templatesExportCmd)
	templatesCmd.AddCommand(templatesDeleteCmd)
	templatesCmd.AddCommand(loadBuiltinTemplatesCmd)

	// Flags for list command
	templatesListCmd.Flags().StringVarP(&templateCategory, "category", "c", "", "Filter by category")

	// Flags for export command
	templatesExportCmd.Flags().StringVarP(&templateFormat, "format", "f", "yaml", "Output format (yaml or json)")
}
