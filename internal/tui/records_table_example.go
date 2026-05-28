package tui

import (
	"time"

	"yoo/internal/models"
)

// ExampleRecordsTable demonstrates how to use the RecordsTableModel
//
// This example shows:
// 1. Creating a record schema
// 2. Populating sample records
// 3. Launching the interactive table view
func ExampleRecordsTable() error {
	// Define a record schema for tracking daily work logs
	schema := &models.RecordSchema{
		Fields: []models.RecordField{
			{
				Name:        "date",
				Type:        "date",
				Description: "Date of the log entry",
				Required:    true,
			},
			{
				Name:        "task",
				Type:        "text",
				Description: "Description of the task",
				Required:    true,
			},
			{
				Name:        "hours",
				Type:        "integer",
				Description: "Number of hours worked",
				Required:    true,
				Default:     "8",
			},
			{
				Name:        "category",
				Type:        "enum",
				Description: "Task category",
				Required:    true,
				Values:      []string{"development", "testing", "documentation", "meetings", "other"},
				Default:     "development",
			},
			{
				Name:        "notes",
				Type:        "text",
				Description: "Additional notes",
				Required:    false,
			},
		},
	}

	// Create sample records
	records := []*models.TemplateRecord{
		{
			ID:             1,
			NoteTemplateID: 1,
			RecordIndex:    1,
			Data: map[string]interface{}{
				"date":     "2024-01-15",
				"task":     "Implemented user authentication",
				"hours":    6,
				"category": "development",
				"notes":    "Used JWT tokens",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:             2,
			NoteTemplateID: 1,
			RecordIndex:    2,
			Data: map[string]interface{}{
				"date":     "2024-01-16",
				"task":     "Code review and bug fixes",
				"hours":    4,
				"category": "testing",
				"notes":    "Fixed 3 critical bugs",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:             3,
			NoteTemplateID: 1,
			RecordIndex:    3,
			Data: map[string]interface{}{
				"date":     "2024-01-17",
				"task":     "Write API documentation",
				"hours":    3,
				"category": "documentation",
				"notes":    "In progress",
			},
			Status:    "in_progress",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:             4,
			NoteTemplateID: 1,
			RecordIndex:    4,
			Data: map[string]interface{}{
				"date":     "2024-01-18",
				"task":     "Team planning meeting",
				"hours":    2,
				"category": "meetings",
				"notes":    "",
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Launch the interactive table view
	return ShowRecordsTable(1, records, schema)
}

// ExampleRecordsTableWithDatabase demonstrates integration with database operations
//
// In a real application, you would:
// 1. Fetch records from the database
// 2. Present them in the table
// 3. Handle CRUD operations by calling database methods
func ExampleRecordsTableWithDatabase() {
	/*
		// Example integration with database layer

		// 1. Fetch the note template to get the schema
		db := database.NewDB(...)
		noteTemplate, err := db.GetNoteTemplate(noteID)
		if err != nil {
			return err
		}

		// 2. Get the record schema from the template
		schema := noteTemplate.Template.Definition.RecordSchema
		if schema == nil {
			return fmt.Errorf("note template does not have a record schema")
		}

		// 3. Fetch all records for this note
		records, err := db.GetTemplateRecords(noteID)
		if err != nil {
			return err
		}

		// 4. Launch the interactive table view
		return ShowRecordsTable(noteID, records, schema)

		// Note: The current implementation handles CRUD operations in-memory.
		// To persist changes, you would need to:
		// - Modify the Update() method to emit custom messages for CRUD operations
		// - Handle those messages in your main application
		// - Call the appropriate database methods
	*/
}

// Example of extending the RecordsTableModel for database integration
//
// You can extend the model to support real database operations:
/*
type RecordsTableWithDB struct {
	RecordsTableModel
	db *database.DB
}

func (m *RecordsTableWithDB) addRecord() {
	// Create the record
	newRecord := &models.TemplateRecord{
		NoteTemplateID: m.noteID,
		RecordIndex:    len(m.records) + 1,
		Data:           m.formData,
		Status:         "draft",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	if err := m.db.CreateTemplateRecord(newRecord); err != nil {
		m.err = err
		return
	}

	// Add to local list
	m.records = append(m.records, newRecord)
	m.formData = make(map[string]interface{})
}

func (m *RecordsTableWithDB) updateRecord() {
	if m.editingRecord == nil {
		return
	}

	// Update the record
	m.editingRecord.Data = m.formData
	m.editingRecord.UpdatedAt = time.Now()

	// Save to database
	if err := m.db.UpdateTemplateRecord(m.editingRecord); err != nil {
		m.err = err
		return
	}

	m.editingRecord = nil
	m.formData = make(map[string]interface{})
}

func (m *RecordsTableWithDB) deleteRecord(record *models.TemplateRecord) {
	// Delete from database
	if err := m.db.DeleteTemplateRecord(record.ID); err != nil {
		m.err = err
		return
	}

	// Remove from local list
	for i, r := range m.records {
		if r.ID == record.ID {
			m.records = append(m.records[:i], m.records[i+1:]...)
			break
		}
	}
}
*/

// ExampleRecordSchemaPatterns shows common record schema patterns
func ExampleRecordSchemaPatterns() {
	// Pattern 1: Daily Log
	dailyLog := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "date", Type: "date", Required: true},
			{Name: "entry", Type: "text", Required: true},
			{Name: "mood", Type: "enum", Values: []string{"great", "good", "okay", "bad"}, Required: false},
		},
	}

	// Pattern 2: Contact List
	contactList := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "name", Type: "text", Required: true},
			{Name: "email", Type: "text", Required: true},
			{Name: "phone", Type: "text", Required: false},
			{Name: "company", Type: "text", Required: false},
			{Name: "notes", Type: "text", Required: false},
		},
	}

	// Pattern 3: Task Tracker
	taskTracker := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "task", Type: "text", Required: true},
			{Name: "priority", Type: "enum", Values: []string{"low", "medium", "high", "urgent"}, Required: true},
			{Name: "due_date", Type: "date", Required: false},
			{Name: "assigned_to", Type: "text", Required: false},
			{Name: "completed", Type: "boolean", Required: true, Default: "false"},
		},
	}

	// Pattern 4: Expense Tracker
	expenseTracker := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "date", Type: "date", Required: true},
			{Name: "description", Type: "text", Required: true},
			{Name: "amount", Type: "integer", Required: true},
			{Name: "category", Type: "enum", Values: []string{"food", "transport", "utilities", "entertainment", "other"}, Required: true},
			{Name: "receipt_url", Type: "url", Required: false},
		},
	}

	// Pattern 5: Reading List
	readingList := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "title", Type: "text", Required: true},
			{Name: "author", Type: "text", Required: true},
			{Name: "pages", Type: "integer", Required: false},
			{Name: "rating", Type: "enum", Values: []string{"1", "2", "3", "4", "5"}, Required: false},
			{Name: "date_finished", Type: "date", Required: false},
			{Name: "notes", Type: "text", Required: false},
		},
	}

	// Pattern 6: Workout Log
	workoutLog := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "date", Type: "date", Required: true},
			{Name: "exercise", Type: "text", Required: true},
			{Name: "sets", Type: "integer", Required: true},
			{Name: "reps", Type: "integer", Required: true},
			{Name: "weight", Type: "integer", Required: false},
			{Name: "duration_minutes", Type: "integer", Required: false},
			{Name: "notes", Type: "text", Required: false},
		},
	}

	// Use any of these schemas
	_ = dailyLog
	_ = contactList
	_ = taskTracker
	_ = expenseTracker
	_ = readingList
	_ = workoutLog
}

// Integration Guide:
//
// To integrate the RecordsTable into your application:
//
// 1. From a CLI command:
//    ```
//    func recordsCmd(noteID int64) error {
//        // Fetch records from database
//        records, err := db.GetTemplateRecords(noteID)
//        if err != nil {
//            return err
//        }
//
//        // Get schema
//        noteTemplate, err := db.GetNoteTemplate(noteID)
//        if err != nil {
//            return err
//        }
//
//        schema := noteTemplate.Template.Definition.RecordSchema
//
//        // Show the table
//        return tui.ShowRecordsTable(noteID, records, schema)
//    }
//    ```
//
// 2. From within another TUI view:
//    ```
//    case tea.KeyMsg:
//        switch msg.String() {
//        case "r":
//            // Switch to records table view
//            return NewRecordsTableModel(m.noteID, m.records, m.schema), nil
//        }
//    ```
//
// 3. Keyboard shortcuts reference:
//    - Navigation:
//      * j/k or ↑/↓: Move cursor up/down
//      * h/l or ←/→: Previous/next page
//      * g/G or Home/End: First/last record
//      * PgUp/PgDown: Page up/down
//
//    - Actions:
//      * a: Add new record
//      * e or Enter: Edit selected record
//      * d: Delete selected record (with confirmation)
//      * f: Filter by status (cycles through: all → draft → in_progress → complete)
//      * /: Search/filter records
//      * c: Clear all filters
//
//    - Navigation:
//      * q or Esc: Quit/back to previous view
//
// 4. Customization:
//    - Adjust `perPage` in the model to change pagination size
//    - Modify status values in the switch statements to match your workflow
//    - Extend the form input handling to support rich field editors
//    - Add custom key bindings by extending handleTableInput()
//
// 5. Database Integration Pattern:
//    ```
//    // Define custom messages for database operations
//    type recordCreatedMsg struct{ record *models.TemplateRecord }
//    type recordUpdatedMsg struct{ record *models.TemplateRecord }
//    type recordDeletedMsg struct{ id int64 }
//
//    // Emit these messages from the model
//    func (m RecordsTableModel) createRecordCmd() tea.Cmd {
//        return func() tea.Msg {
//            // Call database
//            record, err := db.CreateRecord(m.formData)
//            if err != nil {
//                return err
//            }
//            return recordCreatedMsg{record}
//        }
//    }
//
//    // Handle them in Update()
//    case recordCreatedMsg:
//        m.records = append(m.records, msg.record)
//        m.viewMode = ViewModeTable
//    ```
