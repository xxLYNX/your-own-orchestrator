package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var (
	recordFields    []string
	recordFilter    string
	recordFormat    string
	recordStatus    string
	recordOutput    string
	interactiveMode bool
)

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
		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			schema := ctx.Template.Definition.RecordSchema
			if schema == nil {
				return fmt.Errorf("template '%s' does not support records (no record_schema defined)", ctx.Template.Name)
			}

			data := make(map[string]interface{})
			if err := schema.ApplyFieldFlags(data, recordFields); err != nil {
				return err
			}

			if interactiveMode || len(recordFields) == 0 {
				fmt.Printf("Adding record to note: %s\n", ctx.Note.Title)
				fmt.Printf("Template: %s\n\n", ctx.Template.Name)
				reader := bufio.NewReader(os.Stdin)
				if err := schema.PromptForMissingFields(data, reader); err != nil {
					return err
				}
			}

			if err := schema.ValidateRecord(data); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			nextIndex, err := database.GetNextRecordIndex(db.Conn(), ctx.NoteTemplate.ID, nil)
			if err != nil {
				return fmt.Errorf("failed to get next record index: %w", err)
			}

			record := &models.TemplateRecord{
				NoteTemplateID: ctx.NoteTemplate.ID,
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
		})
	},
}

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
		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			records, err := database.ListTemplateRecords(db.Conn(), ctx.NoteTemplate.ID, nil)
			if err != nil {
				return fmt.Errorf("failed to list records: %w", err)
			}

			records = models.FilterTemplateRecords(records, recordStatus, recordFilter)

			if len(records) == 0 {
				fmt.Printf("No records found for note: %s\n", ctx.Note.Title)
				fmt.Println("\nTo add a record, use: yoo records add " + args[0])
				return nil
			}

			fmt.Printf("Records for: %s\n\n", ctx.Note.Title)
			printRecordsTable(records, ctx.Template.Definition.RecordSchema)
			fmt.Printf("\nTotal: %d record(s)\n", len(records))
			return nil
		})
	},
}

var recordsEditCmd = &cobra.Command{
	Use:   "edit <note-id> <record-index> [flags]",
	Short: "Edit an existing record",
	Long: `Edit an existing record by updating field values.

Examples:
  yoo records edit 42 3 --field status="Interview"
  yoo records edit 42 5 --field company="New Company" --field position="Senior Dev"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseNoteIDArg(args[0])
		if err != nil {
			return err
		}
		recordIdx, err := parseRecordIndexArg(args[1])
		if err != nil {
			return err
		}
		if len(recordFields) == 0 {
			return fmt.Errorf("no fields specified (use --field name=value)")
		}

		return runForTemplatedNote(noteID, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			record, err := database.GetTemplateRecord(db.Conn(), ctx.NoteTemplate.ID, nil, recordIdx)
			if err != nil {
				return fmt.Errorf("failed to get record: %w", err)
			}

			schema := ctx.Template.Definition.RecordSchema
			if err := schema.ApplyFieldFlags(record.Data, recordFields); err != nil {
				return err
			}
			if err := schema.ValidateRecord(record.Data); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			if err := database.UpdateTemplateRecord(db.Conn(), record); err != nil {
				return fmt.Errorf("failed to update record: %w", err)
			}

			fmt.Printf("✓ Record #%d updated successfully\n", recordIdx)
			return nil
		})
	},
}

var recordsDeleteCmd = &cobra.Command{
	Use:   "delete <note-id> <record-index>",
	Short: "Delete a record",
	Long: `Delete a record from a templated note.

Example:
  yoo records delete 42 5`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseNoteIDArg(args[0])
		if err != nil {
			return err
		}
		recordIdx, err := parseRecordIndexArg(args[1])
		if err != nil {
			return err
		}

		return runForTemplatedNote(noteID, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			if err := database.DeleteTemplateRecord(db.Conn(), ctx.NoteTemplate.ID, nil, recordIdx); err != nil {
				return fmt.Errorf("failed to delete record: %w", err)
			}
			fmt.Printf("✓ Record #%d deleted successfully\n", recordIdx)
			return nil
		})
	},
}

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
		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			records, err := database.ListTemplateRecords(db.Conn(), ctx.NoteTemplate.ID, nil)
			if err != nil {
				return fmt.Errorf("failed to list records: %w", err)
			}
			if len(records) == 0 {
				return fmt.Errorf("no records to export")
			}

			var writer *os.File
			if recordOutput != "" {
				f, err := os.Create(recordOutput)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer func() { _ = f.Close() }()
				writer = f
			} else {
				writer = os.Stdout
			}

			switch recordFormat {
			case "json":
				encoder := json.NewEncoder(writer)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(records); err != nil {
					return fmt.Errorf("failed to encode JSON: %w", err)
				}

			case "csv":
				fieldNames := ctx.Template.Definition.RecordSchema.FieldNames()
				if len(fieldNames) == 0 {
					for name := range records[0].Data {
						fieldNames = append(fieldNames, name)
					}
				}

				csvWriter := csv.NewWriter(writer)
				header := append([]string{"index", "status"}, fieldNames...)
				if err := csvWriter.Write(header); err != nil {
					return fmt.Errorf("failed to write CSV header: %w", err)
				}

				for _, rec := range records {
					row := []string{
						strconv.Itoa(rec.RecordIndex),
						rec.Status,
					}
					for _, fieldName := range fieldNames {
						row = append(row, fmt.Sprintf("%v", rec.Data[fieldName]))
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
		})
	},
}

func init() {
	rootCmd.AddCommand(recordsCmd)

	recordsCmd.AddCommand(recordsAddCmd)
	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsEditCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)
	recordsCmd.AddCommand(recordsExportCmd)

	recordsAddCmd.Flags().StringArrayVarP(&recordFields, "field", "f", []string{}, "Field value (name=value)")
	recordsAddCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "Interactive mode (prompt for each field)")

	recordsListCmd.Flags().StringVar(&recordStatus, "status", "", "Filter by status (draft, in_progress, complete)")
	recordsListCmd.Flags().StringVar(&recordFilter, "filter", "", "Filter expression (e.g., status=Applied)")

	recordsEditCmd.Flags().StringArrayVarP(&recordFields, "field", "f", []string{}, "Field value to update (name=value)")

	recordsExportCmd.Flags().StringVar(&recordFormat, "format", "csv", "Export format (csv or json)")
	recordsExportCmd.Flags().StringVarP(&recordOutput, "output", "o", "", "Output file (default: stdout)")
}
