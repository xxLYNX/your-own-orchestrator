package tui

import (
	"fmt"
	"strings"
	"time"

	"yoo/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents different view states
type ViewMode int

const (
	ViewModeTable ViewMode = iota
	ViewModeAdd
	ViewModeEdit
	ViewModeDelete
	ViewModeSearch
)

// RecordsTableModel represents the records table view
type RecordsTableModel struct {
	noteID        int64
	records       []*models.TemplateRecord
	recordSchema  *models.RecordSchema
	cursor        int
	page          int
	perPage       int
	filter        string
	statusFilter  string // "all", "draft", "in_progress", "complete"
	width, height int
	viewMode      ViewMode
	editingRecord *models.TemplateRecord
	formData      map[string]interface{}
	formCursor    int
	searchInput   string
	err           error
	quitting      bool
}

// NewRecordsTableModel creates a new records table model
func NewRecordsTableModel(noteID int64, records []*models.TemplateRecord, schema *models.RecordSchema) RecordsTableModel {
	return RecordsTableModel{
		noteID:       noteID,
		records:      records,
		recordSchema: schema,
		cursor:       0,
		page:         0,
		perPage:      20,
		filter:       "",
		statusFilter: "all",
		viewMode:     ViewModeTable,
		formData:     make(map[string]interface{}),
		searchInput:  "",
	}
}

// Init initializes the model
func (m RecordsTableModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m RecordsTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.viewMode {
		case ViewModeTable:
			return m.handleTableInput(msg)
		case ViewModeAdd, ViewModeEdit:
			return m.handleFormInput(msg)
		case ViewModeDelete:
			return m.handleDeleteInput(msg)
		case ViewModeSearch:
			return m.handleSearchInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// handleTableInput handles keyboard input in table view mode
func (m RecordsTableModel) handleTableInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filteredRecords := m.getFilteredRecords()
	maxPage := (len(filteredRecords) - 1) / m.perPage
	if maxPage < 0 {
		maxPage = 0
	}

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			// If cursor moves to previous page
			if m.cursor < m.page*m.perPage {
				m.page--
			}
		}

	case "down", "j":
		if m.cursor < len(filteredRecords)-1 {
			m.cursor++
			// If cursor moves to next page
			if m.cursor >= (m.page+1)*m.perPage {
				m.page++
			}
		}

	case "left", "h", "pgup":
		if m.page > 0 {
			m.page--
			m.cursor = m.page * m.perPage
		}

	case "right", "l", "pgdown":
		if m.page < maxPage {
			m.page++
			m.cursor = m.page * m.perPage
		}

	case "home", "g":
		m.cursor = 0
		m.page = 0

	case "end", "G":
		if len(filteredRecords) > 0 {
			m.cursor = len(filteredRecords) - 1
			m.page = maxPage
		}

	case "a":
		// Start adding a new record
		m.viewMode = ViewModeAdd
		m.formData = make(map[string]interface{})
		m.formCursor = 0
		m.initializeFormData()

	case "e", "enter":
		// Edit selected record
		if len(filteredRecords) > 0 && m.cursor < len(filteredRecords) {
			m.viewMode = ViewModeEdit
			m.editingRecord = filteredRecords[m.cursor]
			m.formData = make(map[string]interface{})
			m.formCursor = 0
			// Copy existing data
			for k, v := range m.editingRecord.Data {
				m.formData[k] = v
			}
		}

	case "d":
		// Delete selected record
		if len(filteredRecords) > 0 && m.cursor < len(filteredRecords) {
			m.viewMode = ViewModeDelete
		}

	case "/":
		// Start search/filter mode
		m.viewMode = ViewModeSearch
		m.searchInput = m.filter

	case "f":
		// Cycle through status filters
		switch m.statusFilter {
		case "all":
			m.statusFilter = "draft"
		case "draft":
			m.statusFilter = "in_progress"
		case "in_progress":
			m.statusFilter = "complete"
		case "complete":
			m.statusFilter = "all"
		}
		m.cursor = 0
		m.page = 0

	case "c":
		// Clear filters
		m.filter = ""
		m.statusFilter = "all"
		m.cursor = 0
		m.page = 0
	}

	return m, nil
}

// handleFormInput handles keyboard input in add/edit form mode
func (m RecordsTableModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.recordSchema == nil || len(m.recordSchema.Fields) == 0 {
		m.viewMode = ViewModeTable
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeTable
		m.formData = make(map[string]interface{})

	case "up", "shift+tab":
		if m.formCursor > 0 {
			m.formCursor--
		}

	case "down", "tab":
		if m.formCursor < len(m.recordSchema.Fields) {
			m.formCursor++
		}

	case "enter":
		if m.formCursor == len(m.recordSchema.Fields) {
			// Submit form
			if m.viewMode == ViewModeAdd {
				m.addRecord()
			} else if m.viewMode == ViewModeEdit {
				m.updateRecord()
			}
			m.viewMode = ViewModeTable
		} else {
			// Move to next field (in real implementation, would open field editor)
			m.formCursor++
		}
	}

	return m, nil
}

// handleDeleteInput handles keyboard input in delete confirmation mode
func (m RecordsTableModel) handleDeleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm deletion
		filteredRecords := m.getFilteredRecords()
		if len(filteredRecords) > 0 && m.cursor < len(filteredRecords) {
			m.deleteRecord(filteredRecords[m.cursor])
			if m.cursor >= len(m.getFilteredRecords()) && m.cursor > 0 {
				m.cursor--
			}
		}
		m.viewMode = ViewModeTable

	case "n", "N", "esc":
		// Cancel deletion
		m.viewMode = ViewModeTable
	}

	return m, nil
}

// handleSearchInput handles keyboard input in search mode
func (m RecordsTableModel) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeTable

	case "enter":
		m.filter = m.searchInput
		m.cursor = 0
		m.page = 0
		m.viewMode = ViewModeTable

	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}

	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}

	return m, nil
}

// getFilteredRecords returns records filtered by status and search term
func (m RecordsTableModel) getFilteredRecords() []*models.TemplateRecord {
	filtered := []*models.TemplateRecord{}

	for _, record := range m.records {
		// Apply status filter
		if m.statusFilter != "all" && record.Status != m.statusFilter {
			continue
		}

		// Apply search filter
		if m.filter != "" {
			match := false
			// Search in record data
			for _, value := range record.Data {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), strings.ToLower(m.filter)) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		filtered = append(filtered, record)
	}

	return filtered
}

// initializeFormData sets default values for new records
func (m *RecordsTableModel) initializeFormData() {
	if m.recordSchema == nil {
		return
	}

	for _, field := range m.recordSchema.Fields {
		if field.Default != "" {
			m.formData[field.Name] = field.Default
		} else {
			// Set empty values based on type
			switch field.Type {
			case "integer":
				m.formData[field.Name] = 0
			case "boolean":
				m.formData[field.Name] = false
			default:
				m.formData[field.Name] = ""
			}
		}
	}
}

// addRecord adds a new record (in real implementation, this would call database)
func (m *RecordsTableModel) addRecord() {
	newRecord := &models.TemplateRecord{
		ID:             int64(len(m.records) + 1),
		NoteTemplateID: m.noteID,
		RecordIndex:    len(m.records) + 1,
		Data:           m.formData,
		Status:         "draft",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	m.records = append(m.records, newRecord)
	m.formData = make(map[string]interface{})
}

// updateRecord updates an existing record
func (m *RecordsTableModel) updateRecord() {
	if m.editingRecord != nil {
		m.editingRecord.Data = m.formData
		m.editingRecord.UpdatedAt = time.Now()
		m.editingRecord = nil
	}
	m.formData = make(map[string]interface{})
}

// deleteRecord removes a record
func (m *RecordsTableModel) deleteRecord(record *models.TemplateRecord) {
	for i, r := range m.records {
		if r.ID == record.ID {
			m.records = append(m.records[:i], m.records[i+1:]...)
			break
		}
	}
}

// View renders the TUI
func (m RecordsTableModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.viewMode {
	case ViewModeAdd:
		return m.renderFormView("Add New Record")
	case ViewModeEdit:
		return m.renderFormView("Edit Record")
	case ViewModeDelete:
		return m.renderDeleteView()
	case ViewModeSearch:
		return m.renderSearchView()
	default:
		return m.renderTableView()
	}
}

// renderTableView renders the main table view
func (m RecordsTableModel) renderTableView() string {
	var s strings.Builder

	// Header
	s.WriteString(TitleWithBorderStyle.Render("📊 Records Table"))
	s.WriteString("\n\n")

	// Filter info
	if m.statusFilter != "all" || m.filter != "" {
		filterInfo := ""
		if m.statusFilter != "all" {
			filterInfo += fmt.Sprintf("Status: %s", m.statusFilter)
		}
		if m.filter != "" {
			if filterInfo != "" {
				filterInfo += " | "
			}
			filterInfo += fmt.Sprintf("Search: %s", m.filter)
		}
		s.WriteString(WarningMessageStyle.Render("🔍 Filters: " + filterInfo))
		s.WriteString("\n\n")
	}

	// Get filtered records
	filteredRecords := m.getFilteredRecords()

	if len(filteredRecords) == 0 {
		s.WriteString(EmptyState("No records found. Press 'a' to add one."))
		s.WriteString("\n\n")
	} else {
		// Calculate pagination
		startIdx := m.page * m.perPage
		endIdx := startIdx + m.perPage
		if endIdx > len(filteredRecords) {
			endIdx = len(filteredRecords)
		}
		pageRecords := filteredRecords[startIdx:endIdx]

		// Render table header
		s.WriteString(m.renderTableHeader())
		s.WriteString("\n")

		// Render table rows
		for i, record := range pageRecords {
			globalIdx := startIdx + i
			s.WriteString(m.renderTableRow(record, globalIdx == m.cursor))
			s.WriteString("\n")
		}

		// Pagination info
		totalPages := (len(filteredRecords) + m.perPage - 1) / m.perPage

		s.WriteString("\n")
		s.WriteString(HelpStyle.Render(fmt.Sprintf(
			"Page %d/%d | Records %d-%d of %d",
			m.page+1, totalPages, startIdx+1, endIdx, len(filteredRecords),
		)))
		s.WriteString("\n")
	}

	// Footer with help
	s.WriteString("\n")
	s.WriteString(m.renderHelp())

	return s.String()
}

// renderTableHeader renders the table header row
func (m RecordsTableModel) renderTableHeader() string {
	var columns []string
	columns = append(columns, TableHeaderStyle.Render("#"))
	columns = append(columns, TableHeaderStyle.Render("Status"))

	if m.recordSchema != nil {
		for _, field := range m.recordSchema.Fields {
			// Show first 3-4 key fields
			if len(columns) < 6 {
				columns = append(columns, TableHeaderStyle.Render(field.Name))
			}
		}
	}

	return strings.Join(columns, " ")
}

// renderTableRow renders a single table row
func (m RecordsTableModel) renderTableRow(record *models.TemplateRecord, isSelected bool) string {
	var cellStyle lipgloss.Style
	if isSelected {
		cellStyle = TableRowSelectedStyle
	} else {
		cellStyle = TableRowStyle
	}

	statusIcon := "○"
	switch record.Status {
	case "draft":
		statusIcon = "◐"
	case "in_progress":
		statusIcon = "◑"
	case "complete":
		statusIcon = "●"
	}

	var columns []string
	cursor := " "
	if isSelected {
		cursor = "▶"
	}

	columns = append(columns, cellStyle.Render(fmt.Sprintf("%s %d", cursor, record.RecordIndex)))
	columns = append(columns, cellStyle.Render(fmt.Sprintf("%s %s", statusIcon, record.Status)))

	if m.recordSchema != nil {
		for _, field := range m.recordSchema.Fields {
			if len(columns) < 6 {
				value := record.Data[field.Name]
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 20 {
					valueStr = valueStr[:17] + "..."
				}
				columns = append(columns, cellStyle.Render(valueStr))
			}
		}
	}

	return strings.Join(columns, " ")
}

// renderFormView renders the add/edit form
func (m RecordsTableModel) renderFormView(title string) string {
	var s strings.Builder

	s.WriteString(TitleWithBorderStyle.Render(title))
	s.WriteString("\n\n")

	if m.recordSchema == nil || len(m.recordSchema.Fields) == 0 {
		s.WriteString("No schema defined for records.\n")
		return s.String()
	}

	// Render form fields
	for i, field := range m.recordSchema.Fields {
		var fieldStyle lipgloss.Style
		if i == m.formCursor {
			fieldStyle = InputFocusedStyle
		} else {
			fieldStyle = TableRowStyle
		}

		required := ""
		if field.Required {
			required = "*"
		}

		value := m.formData[field.Name]
		valueStr := fmt.Sprintf("%v", value)

		line := fmt.Sprintf("%s%s (%s): %s", field.Name, required, field.Type, valueStr)
		s.WriteString(fieldStyle.Render(line))
		s.WriteString("\n")
	}

	// Submit button
	var submitStyle lipgloss.Style
	if m.formCursor == len(m.recordSchema.Fields) {
		submitStyle = StatusCompletedStyle
	} else {
		submitStyle = SuccessMessageStyle
	}

	s.WriteString("\n")
	s.WriteString(submitStyle.Render("[Submit]"))
	s.WriteString("\n\n")

	// Help
	s.WriteString(HelpStyle.Render("↑/↓: navigate • enter: edit/submit • esc: cancel"))

	return s.String()
}

// renderDeleteView renders the delete confirmation dialog
func (m RecordsTableModel) renderDeleteView() string {
	var s strings.Builder

	filteredRecords := m.getFilteredRecords()
	if len(filteredRecords) == 0 || m.cursor >= len(filteredRecords) {
		m.viewMode = ViewModeTable
		return m.renderTableView()
	}

	record := filteredRecords[m.cursor]

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorError).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(0, 1)

	s.WriteString(titleStyle.Render("⚠️  Confirm Delete"))
	s.WriteString("\n\n")

	s.WriteString(TableRowStyle.Render(fmt.Sprintf("Are you sure you want to delete record #%d?", record.RecordIndex)))
	s.WriteString("\n")
	s.WriteString(TableRowStyle.Render(fmt.Sprintf("Status: %s", record.Status)))
	s.WriteString("\n\n")

	// Show some record data
	if m.recordSchema != nil && len(m.recordSchema.Fields) > 0 {
		s.WriteString(TableRowStyle.Render("Record data:"))
		s.WriteString("\n")
		for i, field := range m.recordSchema.Fields {
			if i < 3 { // Show first 3 fields
				value := record.Data[field.Name]
				s.WriteString(TableRowStyle.Render(fmt.Sprintf("  %s: %v", field.Name, value)))
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n")

	s.WriteString(ErrorMessageStyle.Render("[Y]es"))
	s.WriteString("  ")
	s.WriteString(SuccessMessageStyle.Render("[N]o"))
	s.WriteString("\n")

	return s.String()
}

// renderSearchView renders the search input dialog
func (m RecordsTableModel) renderSearchView() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("🔍 Search Records"))
	s.WriteString("\n\n")

	inputStyle := InputFocusedStyle.Width(50)

	s.WriteString(inputStyle.Render(m.searchInput + "█"))
	s.WriteString("\n\n")

	s.WriteString(HelpStyle.Render("enter: apply • esc: cancel"))

	return s.String()
}

// renderHelp renders the help/keyboard shortcuts footer
func (m RecordsTableModel) renderHelp() string {
	helpText := "j/k/↑/↓: navigate • enter/e: edit • a: add • d: delete • f: filter status • /: search • c: clear filters • q/esc: quit"
	return HelpWithBorderStyle.Render(helpText)
}

// ShowRecordsTable launches the Bubble Tea TUI to display the records table
func ShowRecordsTable(noteID int64, records []*models.TemplateRecord, schema *models.RecordSchema) error {
	model := NewRecordsTableModel(noteID, records, schema)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
