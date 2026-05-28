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
	_, _ = fmt.Scanln()

	// Launch the interactive table
	if err := tui.ShowRecordsTable(1, records, schema); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nDemo completed!")
}

func createWorkLogDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := newDemoSchema(
		models.RecordField{Name: "date", Type: "date", Description: "Date of work", Required: true},
		models.RecordField{Name: "task", Type: "text", Description: "Task description", Required: true},
		models.RecordField{Name: "hours", Type: "integer", Description: "Hours worked", Required: true, Default: "8"},
		models.RecordField{Name: "category", Type: "enum", Description: "Task category", Required: true, Values: []string{"development", "testing", "documentation", "meetings", "other"}, Default: "development"},
		models.RecordField{Name: "notes", Type: "text", Description: "Additional notes"},
	)

	records := []*models.TemplateRecord{
		newDemoRecord(1, map[string]interface{}{"date": "2024-01-15", "task": "Implemented user authentication system", "hours": 6, "category": "development", "notes": "Used JWT tokens with refresh mechanism"}, "complete", 72*time.Hour),
		newDemoRecord(2, map[string]interface{}{"date": "2024-01-16", "task": "Code review and bug fixes", "hours": 4, "category": "testing", "notes": "Fixed 3 critical bugs in payment module"}, "complete", 48*time.Hour),
		newDemoRecord(3, map[string]interface{}{"date": "2024-01-17", "task": "Write API documentation", "hours": 5, "category": "documentation", "notes": "Documented REST endpoints and authentication flow"}, "complete", 24*time.Hour),
		newDemoRecord(4, map[string]interface{}{"date": "2024-01-18", "task": "Sprint planning meeting", "hours": 2, "category": "meetings", "notes": "Planned features for next sprint"}, "in_progress", 12*time.Hour),
		newDemoRecord(5, map[string]interface{}{"date": "2024-01-19", "task": "Implement search functionality", "hours": 0, "category": "development", "notes": "To be started"}, "draft", 0),
		newDemoRecord(6, map[string]interface{}{"date": "2024-01-19", "task": "Performance optimization", "hours": 0, "category": "development", "notes": "Database query optimization needed"}, "draft", 0),
	}
	return schema, records
}

func createContactsDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := newDemoSchema(
		models.RecordField{Name: "name", Type: "text", Required: true},
		models.RecordField{Name: "email", Type: "text", Required: true},
		models.RecordField{Name: "phone", Type: "text"},
		models.RecordField{Name: "company", Type: "text"},
		models.RecordField{Name: "notes", Type: "text"},
	)
	return schema, []*models.TemplateRecord{
		newDemoRecord(1, map[string]interface{}{"name": "Alice Johnson", "email": "alice@techcorp.com", "phone": "+1-555-0101", "company": "TechCorp", "notes": "Senior developer, contact for API questions"}, "complete", 48*time.Hour),
		newDemoRecord(2, map[string]interface{}{"name": "Bob Smith", "email": "bob@startupinc.io", "phone": "+1-555-0102", "company": "StartupInc", "notes": "CTO, interested in our product"}, "complete", 24*time.Hour),
		newDemoRecord(3, map[string]interface{}{"name": "Carol White", "email": "carol@freelance.com", "phone": "", "company": "Freelance", "notes": "Designer, working on UI project"}, "in_progress", 12*time.Hour),
		newDemoRecord(4, map[string]interface{}{"name": "David Chen", "email": "david@example.com", "phone": "", "company": "", "notes": "Need to follow up"}, "draft", 0),
	}
}

func createTasksDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := newDemoSchema(
		models.RecordField{Name: "task", Type: "text", Required: true},
		models.RecordField{Name: "priority", Type: "enum", Values: []string{"low", "medium", "high", "urgent"}, Required: true, Default: "medium"},
		models.RecordField{Name: "due_date", Type: "date"},
		models.RecordField{Name: "assigned_to", Type: "text"},
		models.RecordField{Name: "completed", Type: "boolean", Required: true, Default: "false"},
	)
	return schema, []*models.TemplateRecord{
		newDemoRecord(1, map[string]interface{}{"task": "Fix login bug", "priority": "urgent", "due_date": "2024-01-20", "assigned_to": "Alice", "completed": false}, "in_progress", 24*time.Hour),
		newDemoRecord(2, map[string]interface{}{"task": "Update dependencies", "priority": "medium", "due_date": "2024-01-25", "assigned_to": "Bob", "completed": true}, "complete", 48*time.Hour),
		newDemoRecord(3, map[string]interface{}{"task": "Write unit tests", "priority": "high", "due_date": "2024-01-22", "assigned_to": "Carol", "completed": false}, "draft", 0),
	}
}

func createExpensesDemo() (*models.RecordSchema, []*models.TemplateRecord) {
	schema := newDemoSchema(
		models.RecordField{Name: "date", Type: "date", Required: true},
		models.RecordField{Name: "description", Type: "text", Required: true},
		models.RecordField{Name: "amount", Type: "integer", Required: true},
		models.RecordField{Name: "category", Type: "enum", Values: []string{"food", "transport", "utilities", "entertainment", "other"}, Required: true},
		models.RecordField{Name: "receipt_url", Type: "url"},
	)
	return schema, []*models.TemplateRecord{
		newDemoRecord(1, map[string]interface{}{"date": "2024-01-15", "description": "Lunch with client", "amount": 45, "category": "food", "receipt_url": "https://example.com/receipt1.pdf"}, "complete", 72*time.Hour),
		newDemoRecord(2, map[string]interface{}{"date": "2024-01-16", "description": "Taxi to airport", "amount": 65, "category": "transport", "receipt_url": ""}, "complete", 48*time.Hour),
		newDemoRecord(3, map[string]interface{}{"date": "2024-01-17", "description": "Office electricity bill", "amount": 120, "category": "utilities", "receipt_url": "https://example.com/receipt3.pdf"}, "in_progress", 24*time.Hour),
		newDemoRecord(4, map[string]interface{}{"date": "2024-01-18", "description": "Team dinner", "amount": 0, "category": "food", "receipt_url": ""}, "draft", 0),
	}
}
