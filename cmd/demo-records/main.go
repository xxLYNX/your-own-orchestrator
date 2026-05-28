package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"yoo/internal/models"
	"yoo/internal/tui"
)

func main() {
	fmt.Println("=== Records Table Demo ===")
	fmt.Println("This demo shows the interactive records table with sample data.")
	fmt.Println()

	// Select a demo scenario
	scenario := "work-log"
	if len(os.Args) > 1 {
		scenario = os.Args[1]
	}

	var schema *models.RecordSchema
	var records []*models.TemplateRecord

	switch scenario {
	case "work-log":
		schema, records = createWorkLogDemo()
		fmt.Println("Scenario: Work Log Tracker")
	case "contacts":
		schema, records = createContactsDemo()
		fmt.Println("Scenario: Contact List")
	case "tasks":
		schema, records = createTasksDemo()
		fmt.Println("Scenario: Task Tracker")
	case "expenses":
		schema, records = createExpensesDemo()
		fmt.Println("Scenario: Expense Tracker")
	default:
		fmt.Printf("Unknown scenario: %s\n", scenario)
		fmt.Println("Available scenarios: work-log, contacts, tasks, expenses")
		os.Exit(1)
	}

	fmt.Printf("Total records: %d\n", len(records))
	fmt.Println()
	fmt.Println("Keyboard shortcuts:")
	fmt.Println("  j/k or ↑/↓ - Navigate")
	fmt.Println("  a - Add new record")
	fmt.Println("  e/Enter - Edit record")
	fmt.Println("  d - Delete record")
	fmt.Println("  f - Filter by status")
	fmt.Println("  / - Search")
	fmt.Println("  c - Clear filters")
	fmt.Println("  q/Esc - Quit")
	fmt.Println()
	fmt.Println("Press any key to start...")
	fmt.Scanln()

	// Launch the interactive table
	if err := tui.ShowRecordsTable(1, records, schema); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nDemo completed!")
}

func createWorkLogDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := &models.RecordSchema{
		Fields: []models.RecordField{
			{
				Name:        "date",
				Type:        "date",
				Description: "Date of work",
				Required:    true,
			},
			{
				Name:        "task",
				Type:        "text",
				Description: "Task description",
				Required:    true,
			},
			{
				Name:        "hours",
				Type:        "integer",
				Description: "Hours worked",
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

	records := []*models.TemplateRecord{
		{
			ID:             1,
			NoteTemplateID: 1,
			RecordIndex:    1,
			Data: map[string]interface{}{
				"date":     "2024-01-15",
				"task":     "Implemented user authentication system",
				"hours":    6,
				"category": "development",
				"notes":    "Used JWT tokens with refresh mechanism",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-72 * time.Hour),
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
				"notes":    "Fixed 3 critical bugs in payment module",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:             3,
			NoteTemplateID: 1,
			RecordIndex:    3,
			Data: map[string]interface{}{
				"date":     "2024-01-17",
				"task":     "Write API documentation",
				"hours":    5,
				"category": "documentation",
				"notes":    "Documented REST endpoints and authentication flow",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:             4,
			NoteTemplateID: 1,
			RecordIndex:    4,
			Data: map[string]interface{}{
				"date":     "2024-01-18",
				"task":     "Sprint planning meeting",
				"hours":    2,
				"category": "meetings",
				"notes":    "Planned features for next sprint",
			},
			Status:    "in_progress",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			ID:             5,
			NoteTemplateID: 1,
			RecordIndex:    5,
			Data: map[string]interface{}{
				"date":     "2024-01-19",
				"task":     "Implement search functionality",
				"hours":    0,
				"category": "development",
				"notes":    "To be started",
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:             6,
			NoteTemplateID: 1,
			RecordIndex:    6,
			Data: map[string]interface{}{
				"date":     "2024-01-19",
				"task":     "Performance optimization",
				"hours":    0,
				"category": "development",
				"notes":    "Database query optimization needed",
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return schema, records
}

func createContactsDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "name", Type: "text", Required: true},
			{Name: "email", Type: "text", Required: true},
			{Name: "phone", Type: "text", Required: false},
			{Name: "company", Type: "text", Required: false},
			{Name: "notes", Type: "text", Required: false},
		},
	}

	records := []*models.TemplateRecord{
		{
			ID:             1,
			NoteTemplateID: 1,
			RecordIndex:    1,
			Data: map[string]interface{}{
				"name":    "Alice Johnson",
				"email":   "alice@techcorp.com",
				"phone":   "+1-555-0101",
				"company": "TechCorp",
				"notes":   "Senior developer, contact for API questions",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:             2,
			NoteTemplateID: 1,
			RecordIndex:    2,
			Data: map[string]interface{}{
				"name":    "Bob Smith",
				"email":   "bob@startupinc.io",
				"phone":   "+1-555-0102",
				"company": "StartupInc",
				"notes":   "CTO, interested in our product",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:             3,
			NoteTemplateID: 1,
			RecordIndex:    3,
			Data: map[string]interface{}{
				"name":    "Carol White",
				"email":   "carol@freelance.com",
				"phone":   "",
				"company": "Freelance",
				"notes":   "Designer, working on UI project",
			},
			Status:    "in_progress",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			ID:             4,
			NoteTemplateID: 1,
			RecordIndex:    4,
			Data: map[string]interface{}{
				"name":    "David Chen",
				"email":   "david@example.com",
				"phone":   "",
				"company": "",
				"notes":   "Need to follow up",
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return schema, records
}

func createTasksDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "task", Type: "text", Required: true},
			{
				Name:     "priority",
				Type:     "enum",
				Values:   []string{"low", "medium", "high", "urgent"},
				Required: true,
				Default:  "medium",
			},
			{Name: "due_date", Type: "date", Required: false},
			{Name: "assigned_to", Type: "text", Required: false},
			{
				Name:     "completed",
				Type:     "boolean",
				Required: true,
				Default:  "false",
			},
		},
	}

	records := []*models.TemplateRecord{
		{
			ID:             1,
			NoteTemplateID: 1,
			RecordIndex:    1,
			Data: map[string]interface{}{
				"task":        "Fix login bug",
				"priority":    "urgent",
				"due_date":    "2024-01-20",
				"assigned_to": "Alice",
				"completed":   false,
			},
			Status:    "in_progress",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:             2,
			NoteTemplateID: 1,
			RecordIndex:    2,
			Data: map[string]interface{}{
				"task":        "Update dependencies",
				"priority":    "medium",
				"due_date":    "2024-01-25",
				"assigned_to": "Bob",
				"completed":   true,
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:             3,
			NoteTemplateID: 1,
			RecordIndex:    3,
			Data: map[string]interface{}{
				"task":        "Write unit tests",
				"priority":    "high",
				"due_date":    "2024-01-22",
				"assigned_to": "Carol",
				"completed":   false,
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return schema, records
}

func createExpensesDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := &models.RecordSchema{
		Fields: []models.RecordField{
			{Name: "date", Type: "date", Required: true},
			{Name: "description", Type: "text", Required: true},
			{Name: "amount", Type: "integer", Required: true},
			{
				Name:     "category",
				Type:     "enum",
				Values:   []string{"food", "transport", "utilities", "entertainment", "other"},
				Required: true,
			},
			{Name: "receipt_url", Type: "url", Required: false},
		},
	}

	records := []*models.TemplateRecord{
		{
			ID:             1,
			NoteTemplateID: 1,
			RecordIndex:    1,
			Data: map[string]interface{}{
				"date":        "2024-01-15",
				"description": "Lunch with client",
				"amount":      45,
				"category":    "food",
				"receipt_url": "https://example.com/receipt1.pdf",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-72 * time.Hour),
		},
		{
			ID:             2,
			NoteTemplateID: 1,
			RecordIndex:    2,
			Data: map[string]interface{}{
				"date":        "2024-01-16",
				"description": "Taxi to airport",
				"amount":      65,
				"category":    "transport",
				"receipt_url": "",
			},
			Status:    "complete",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:             3,
			NoteTemplateID: 1,
			RecordIndex:    3,
			Data: map[string]interface{}{
				"date":        "2024-01-17",
				"description": "Office electricity bill",
				"amount":      120,
				"category":    "utilities",
				"receipt_url": "https://example.com/receipt3.pdf",
			},
			Status:    "in_progress",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:             4,
			NoteTemplateID: 1,
			RecordIndex:    4,
			Data: map[string]interface{}{
				"date":        "2024-01-18",
				"description": "Team dinner",
				"amount":      0,
				"category":    "food",
				"receipt_url": "",
			},
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return schema, records
}
