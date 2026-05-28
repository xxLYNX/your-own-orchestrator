package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
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
		return withDB(func(db *database.DB) error {
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
		})
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
		return withDB(func(db *database.DB) error {
			template, err := database.GetTemplateByName(db.Conn(), templateName)
			if err != nil {
				return fmt.Errorf("template '%s' not found: %w", templateName, err)
			}
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

		// Display structure
		fmt.Println("STRUCTURE:")
		structure, err := template.Definition.GetStructure()
		if err != nil {
			return fmt.Errorf("failed to load structure: %w", err)
		}
		printStructureTree(structure, "  ")
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
		})
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
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		template, err := models.ParseTemplateYAML(data, false)
		if err != nil {
			return err
		}

		return withDB(func(db *database.DB) error {
			existing, _ := database.GetTemplateByName(db.Conn(), template.Name)
			if existing != nil {
				return fmt.Errorf("template '%s' already exists. Delete it first or choose a different name", template.Name)
			}

			if err := database.CreateTemplate(db.Conn(), template); err != nil {
				return fmt.Errorf("failed to import template: %w", err)
			}

			fmt.Printf("✓ Template '%s' imported successfully\n", template.Name)
			fmt.Printf("  Version: %s\n", template.Version)
			fmt.Printf("  Category: %s\n", template.Category)
			if structure, err := template.Definition.GetStructure(); err == nil && structure != nil {
				fmt.Printf("  Structure: %s (%d nodes)\n", structure.DisplayTitle(), countStructureNodes(structure))
			}
			fmt.Println("\nUse it with: yoo add \"task\" --template", template.Name)
			return nil
		})
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
		return withDB(func(db *database.DB) error {
			template, err := database.GetTemplateByName(db.Conn(), templateName)
			if err != nil {
				return fmt.Errorf("template '%s' not found: %w", templateName, err)
			}

			var output []byte
			if templateFormat == "json" {
				output, err = models.MarshalTemplateJSON(template)
			} else {
				output, err = models.MarshalTemplateYAML(template)
			}
			if err != nil {
				return fmt.Errorf("failed to marshal template: %w", err)
			}

			fmt.Print(string(output))
			return nil
		})
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
		return withDB(func(db *database.DB) error {
			template, err := database.GetTemplateByName(db.Conn(), templateName)
			if err != nil {
				return fmt.Errorf("template '%s' not found: %w", templateName, err)
			}
			if template.IsBuiltin {
				return fmt.Errorf("cannot delete built-in template '%s'", templateName)
			}
			if err := database.DeleteTemplate(db.Conn(), template.ID); err != nil {
				return fmt.Errorf("failed to delete template: %w", err)
			}
			fmt.Printf("✓ Template '%s' deleted successfully\n", templateName)
			return nil
		})
	},
}

// loadBuiltinTemplates loads built-in templates from the templates directory
var loadBuiltinTemplatesCmd = &cobra.Command{
	Use:    "load-builtins",
	Short:  "Load built-in templates from templates directory",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		templatesDir := "templates"
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			return fmt.Errorf("templates directory not found")
		}

		patterns := []string{
			filepath.Join(templatesDir, "*.yaml"),
			filepath.Join(templatesDir, "generics", "*.yaml"),
		}
		var files []string
		for _, pattern := range patterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("failed to read templates directory: %w", err)
			}
			files = append(files, matches...)
		}

		return withDB(func(db *database.DB) error {
			loaded := 0
			for _, file := range files {
				if strings.HasSuffix(file, "README.yaml") {
					continue
				}

				data, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("⚠ Failed to read %s: %v\n", file, err)
					continue
				}

				template, err := models.ParseTemplateYAML(data, true)
				if err != nil {
					fmt.Printf("⚠ Failed to parse %s: %v\n", file, err)
					continue
				}

				existing, _ := database.GetTemplateByName(db.Conn(), template.Name)
				if existing != nil {
					if !existing.IsBuiltin {
						fmt.Printf("⊘ Template '%s' exists as a custom template, skipping\n", template.Name)
						continue
					}

					template.ID = existing.ID
					template.CreatedAt = existing.CreatedAt
					if err := database.UpdateTemplate(db.Conn(), template); err != nil {
						fmt.Printf("⚠ Failed to update %s: %v\n", template.Name, err)
						continue
					}

					if existing.Version != template.Version {
						fmt.Printf("✓ Updated built-in template: %s (v%s → v%s)\n", template.Name, existing.Version, template.Version)
					} else {
						fmt.Printf("✓ Refreshed built-in template: %s (v%s)\n", template.Name, template.Version)
					}
					loaded++
					continue
				}

				if err := database.CreateTemplate(db.Conn(), template); err != nil {
					fmt.Printf("⚠ Failed to load %s: %v\n", template.Name, err)
					continue
				}

				fmt.Printf("✓ Loaded built-in template: %s (v%s)\n", template.Name, template.Version)
				loaded++
			}

			fmt.Printf("\nSynced %d built-in template(s)\n", loaded)
			return nil
		})
	},
}

func printStructureTree(node *models.ShapeNode, indent string) {
	if node == nil {
		fmt.Println(indent + "(none)")
		return
	}
	line := fmt.Sprintf("%s[%s] %s", indent, node.Kind, node.DisplayTitle())
	if node.Repeat != nil && node.Repeat.Count != "" {
		line += fmt.Sprintf(" × %s", node.Repeat.Count)
	}
	if node.OutputRequired != "" {
		line += fmt.Sprintf(" → %s", node.OutputRequired)
	}
	fmt.Println(line)
	if node.Description != "" {
		fmt.Printf("%s  %s\n", indent, node.Description)
	}
	for _, child := range node.ChildNodes() {
		printStructureTree(&child, indent+"  ")
	}
}

func countStructureNodes(node *models.ShapeNode) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, child := range node.ChildNodes() {
		count += countStructureNodes(&child)
	}
	return count
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
