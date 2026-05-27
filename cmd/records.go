package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var (
	recordNoteID    int64
	recordIndex     int
	recordFields    []string
	recordFilter    string
	recordFormat    string
	recordStatus    string
	recordOutput    string
	interactiveMode bool
)

// recordsCmd represents the records command
var recordsCmd = &cobra.Command{
	Use:     "records",
	Aliases: []string{"record", "rec"},
	Short:   "Manage template records (log-style data)",
	Long: `Manage template records for log-style repeating data.

Records allow you to store structured, repeating data directly in the database
instead of external files. Perfect for tracking job applications, contacts,
expenses, or any other tabular data.

Examples:
  yoo records add 42 --field company="Acme Corp" --field position="Developer"
  yoo records list 42
  yoo records edit 42 3 --field status="Interview"
  yoo records export 42 --format csv > jobs.csv`,
}

// recordsAddCmd adds a new record to a templated note
var recordsAddCmd = &cobra.Command{
	Use:   "add <note-id> [flags]",
	Short: "Add a new record to a templated note",
	Long: `Add a new record to a templated note.

You can provide field values using --field flags, or use interactive mode
to be prompted for each field.

Examples:
  yoo records add 42 --field company="Acme Corp" --field position="Developer"
  yoo records add 42 --interactive`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note and verify it's templated
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or template not found: %w", err)
		}

		// Get the template definition
		template, err := database.GetTemplateByID(db.Conn(), noteTemplate.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		// Check if template has a record schema
		if template.Definition.RecordSchema == nil {
			return fmt.Errorf("template '%s' does not support records (no record_schema defined)", template.Name)
		}

		// Parse field values from flags
		data := make(map[string]interface{})
		for _, fieldStr := range recordFields {
			parts := strings.SplitN(fieldStr, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid field format '%s', expected name=value", fieldStr)
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			data[name] = value
		}

		// Interactive mode
		if interactiveMode || len(recordFields) == 0 {
			fmt.Printf("Adding record to note: %s\n", note.Title)
			fmt.Printf("Template: %s\n\n", template.Name)

			reader := bufio.NewReader(os.Stdin)
			for _, field := range template.Definition.RecordSchema.Fields {
				// Skip if already provided via flag
				if _, exists := data[field.Name]; exists {
					continue
				}

				prompt := fmt.Sprintf("%s", field.Name)
				if field.Description != "" {
					prompt += fmt.Sprintf(" (%s)", field.Description)
				}
				if field.Required {
					prompt += " [required]"
				}
				if field.Default != "" {
					prompt += fmt.Sprintf(" [default: %s]", field.Default)
				}
				if field.Type == "enum" && len(field.Values) > 0 {
					prompt += fmt.Sprintf(" [%s]", strings.Join(field.Values, ", "))
				}
				prompt += ": "

				fmt.Print(prompt)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input == "" && field.Default != "" {
					input = field.Default
				}

				if input == "" && field.Required {
					return fmt.Errorf("required field '%s' cannot be empty", field.Name)
				}

				if input != "" {
					// Type conversion
					switch field.Type {
					case "integer":
						intVal, err := strconv.Atoi(input)
						if err != nil {
							return fmt.Errorf("field '%s' must be an integer", field.Name)
						}
						data[field.Name] = intVal
					case "boolean":
						boolVal, err := strconv.ParseBool(input)
						if err != nil {
							return fmt.Errorf("field '%s' must be true/false", field.Name)
						}
						data[field.Name] = boolVal
					default:
						data[field.Name] = input
					}
				}
			}
		}

		// Validate record data
		if err := template.Definition.RecordSchema.ValidateRecord(data); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Get next record index
		nextIndex, err := database.GetNextRecordIndex(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to get next record index: %w", err)
		}

		// Create the record
		record := &models.TemplateRecord{
			NoteTemplateID: noteTemplate.ID,
			RecordIndex:    nextIndex,
			Data:           data,
			Status:         "draft",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := database.CreateTemplateRecord(db.Conn(), record); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		fmt.Printf("\n✓ Record #%d added successfully\n", nextIndex)
		return nil
	},
}

// recordsListCmd lists all records for a note
var recordsListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List all records for a templated note",
	Long: `List all records for a templated note in table format.

Examples:
  yoo records list 42
  yoo records list 42 --filter status=Applied
  yoo records list 42 --status complete`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get records
		records, err := database.ListTemplateRecords(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		// Filter by status if specified
		if recordStatus != "" {
			filtered := make([]*models.TemplateRecord, 0)
			for _, rec := range records {
				if rec.Status == recordStatus {
					filtered = append(filtered, rec)
				}
			}
			records = filtered
		}

		if len(records) == 0 {
			fmt.Printf("No records found for note: %s\n", note.Title)
			fmt.Println("\nTo add a record, use: yoo records add " + args[0])
			return nil
		}

		// Display records in table format
		fmt.Printf("Records for: %s\n\n", note.Title)

		// Get field names from first record
		if len(records) > 0 && len(records[0].Data) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Header row
			fmt.Fprint(w, "#\t")
			fieldNames := make([]string, 0)
			for name := range records[0].Data {
				fieldNames = append(fieldNames, name)
			}
			fmt.Fprint(w, strings.Join(fieldNames, "\t"))
			fmt.Fprint(w, "\tStatus\n")

			// Separator
			fmt.Fprint(w, "---\t")
			for range fieldNames {
				fmt.Fprint(w, "---\t")
			}
			fmt.Fprint(w, "---\n")

			// Data rows
			for _, rec := range records {
				fmt.Fprintf(w, "%d\t", rec.RecordIndex)
				for _, fieldName := range fieldNames {
					value := rec.Data[fieldName]
					valueStr := fmt.Sprintf("%v", value)
					if len(valueStr) > 30 {
						valueStr = valueStr[:27] + "..."
					}
					fmt.Fprintf(w, "%s\t", valueStr)
				}
				fmt.Fprintf(w, "%s\n", rec.Status)
			}
			w.Flush()
		}

		fmt.Printf("\nTotal: %d record(s)\n", len(records))
		return nil
	},
}

// recordsEditCmd edits an existing record
var recordsEditCmd = &cobra.Command{
	Use:   "edit <note-id> <record-index> [flags]",
	Short: "Edit an existing record",
	Long: `Edit an existing record by updating field values.

Examples:
  yoo records edit 42 3 --field status="Interview"
  yoo records edit 42 5 --field company="New Company" --field position="Senior Dev"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		recordIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid record index: %w", err)
		}

		if len(recordFields) == 0 {
			return fmt.Errorf("no fields specified (use --field name=value)")
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get the record
		record, err := database.GetTemplateRecord(db.Conn(), noteTemplate.ID, recordIdx)
		if err != nil {
			return fmt.Errorf("failed to get record: %w", err)
		}

		// Get the template for validation
		template, err := database.GetTemplateByID(db.Conn(), noteTemplate.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		// Parse and update fields
		for _, fieldStr := range recordFields {
			parts := strings.SplitN(fieldStr, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid field format '%s', expected name=value", fieldStr)
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Find field definition for type conversion
			var fieldDef *models.RecordField
			for i := range template.Definition.RecordSchema.Fields {
				if template.Definition.RecordSchema.Fields[i].Name == name {
					fieldDef = &template.Definition.RecordSchema.Fields[i]
					break
				}
			}

			if fieldDef != nil {
				switch fieldDef.Type {
				case "integer":
					intVal, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("field '%s' must be an integer", name)
					}
					record.Data[name] = intVal
				case "boolean":
					boolVal, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("field '%s' must be true/false", name)
					}
					record.Data[name] = boolVal
				default:
					record.Data[name] = value
				}
			} else {
				record.Data[name] = value
			}
		}

		// Validate updated record
		if err := template.Definition.RecordSchema.ValidateRecord(record.Data); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Update the record
		if err := database.UpdateTemplateRecord(db.Conn(), record); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}

		fmt.Printf("✓ Record #%d updated successfully\n", recordIdx)
		return nil
	},
}

// recordsDeleteCmd deletes a record
var recordsDeleteCmd = &cobra.Command{
	Use:   "delete <note-id> <record-index>",
	Short: "Delete a record",
	Long: `Delete a record from a templated note.

Example:
  yoo records delete 42 5`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		recordIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid record index: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Delete the record
		if err := database.DeleteTemplateRecord(db.Conn(), noteTemplate.ID, recordIdx); err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}

		fmt.Printf("✓ Record #%d deleted successfully\n", recordIdx)
		return nil
	},
}

// recordsExportCmd exports records to a file
var recordsExportCmd = &cobra.Command{
	Use:   "export <note-id>",
	Short: "Export records to CSV or JSON",
	Long: `Export records to CSV or JSON format.

By default, outputs to stdout. Use shell redirection to save to a file.

Examples:
  yoo records export 42 --format csv > jobs.csv
  yoo records export 42 --format json > jobs.json
  yoo records export 42 --format csv --output jobs.csv`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get records
		records, err := database.ListTemplateRecords(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		if len(records) == 0 {
			return fmt.Errorf("no records to export")
		}

		// Determine output writer
		var writer *os.File
		if recordOutput != "" {
			f, err := os.Create(recordOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			writer = f
		} else {
			writer = os.Stdout
		}

		// Export based on format
		switch recordFormat {
		case "json":
			encoder := json.NewEncoder(writer)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(records); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}

		case "csv":
			csvWriter := csv.NewWriter(writer)

			// Get field names from first record
			fieldNames := make([]string, 0)
			for name := range records[0].Data {
				fieldNames = append(fieldNames, name)
			}

			// Write header
			header := append([]string{"index", "status"}, fieldNames...)
			if err := csvWriter.Write(header); err != nil {
				return fmt.Errorf("failed to write CSV header: %w", err)
			}

			// Write data rows
			for _, rec := range records {
				row := []string{
					strconv.Itoa(rec.RecordIndex),
					rec.Status,
				}
				for _, fieldName := range fieldNames {
					value := rec.Data[fieldName]
					row = append(row, fmt.Sprintf("%v", value))
				}
				if err := csvWriter.Write(row); err != nil {
					return fmt.Errorf("failed to write CSV row: %w", err)
				}
			}
			csvWriter.Flush()

			if err := csvWriter.Error(); err != nil {
				return fmt.Errorf("CSV writer error: %w", err)
			}

		default:
			return fmt.Errorf("unsupported format '%s' (use csv or json)", recordFormat)
		}

		if recordOutput != "" {
			fmt.Fprintf(os.Stderr, "✓ Exported %d record(s) to %s\n", len(records), recordOutput)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(recordsCmd)

	// Add subcommands
	recordsCmd.AddCommand(recordsAddCmd)
	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsEditCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)
	recordsCmd.AddCommand(recordsExportCmd)

	// Flags for add command
	recordsAddCmd.Flags().StringArrayVarP(&recordFields, "field", "f", []string{}, "Field value (name=value)")
	recordsAddCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "Interactive mode (prompt for each field)")

	// Flags for list command
	recordsListCmd.Flags().StringVar(&recordStatus, "status", "", "Filter by status (draft, in_progress, complete)")
	recordsListCmd.Flags().StringVar(&recordFilter, "filter", "", "Filter expression (e.g., status=Applied)")

	// Flags for edit command
	recordsEditCmd.Flags().StringArrayVarP(&recordFields, "field", "f", []string{}, "Field value to update (name=value)")

	// Flags for export command
	recordsExportCmd.Flags().StringVar(&recordFormat, "format", "csv", "Export format (csv or json)")
	recordsExportCmd.Flags().StringVarP(&recordOutput, "output", "o", "", "Output file (default: stdout)")
}
