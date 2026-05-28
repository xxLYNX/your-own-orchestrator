package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"yoo/internal/database"
	"yoo/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab constants
const (
	TabRecords = iota
	TabSteps
	TabArtifacts
)

// TemplatedNoteModel is the master view that integrates all three shapes
type TemplatedNoteModel struct {
	noteID         int64
	note           *database.Note
	noteTemplate   *models.NoteTemplate
	template       *models.Template
	currentTab     int // 0=records, 1=steps, 2=artifacts
	recordsModel   *RecordsTableModel
	stepsModel     *StepsViewModel
	artifactsModel *ArtifactsViewModel
	db             *sql.DB
	width, height  int
	err            error
	quitting       bool
}

// ArtifactsViewModel is a placeholder for the artifacts view
// TODO: Implement full artifacts view in a separate file
type ArtifactsViewModel struct {
	noteTemplateID int64
	artifacts      []*models.Artifact
	cursor         int
	width, height  int
	db             *sql.DB
	err            error
}

// NewTemplatedNoteModel creates a new templated note model
func NewTemplatedNoteModel(db *sql.DB, noteID int64) (*TemplatedNoteModel, error) {
	// Load note
	note, err := database.GetNoteByID(db, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load note: %w", err)
	}

	// Check if note is templated
	if !note.IsTemplated {
		return nil, fmt.Errorf("note is not templated")
	}

	// Load note template association
	noteTemplate, err := database.GetNoteTemplate(db, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load note template: %w", err)
	}

	// Load template definition
	template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Determine which tab to show first based on template shape
	defaultTab := TabSteps
	if template.Definition.RecordSchema != nil {
		defaultTab = TabRecords
	}

	model := &TemplatedNoteModel{
		noteID:       noteID,
		note:         note,
		noteTemplate: noteTemplate,
		template:     template,
		currentTab:   defaultTab,
		db:           db,
		width:        80,
		height:       24,
	}

	// Initialize child models based on template shape
	if err := model.initializeChildModels(); err != nil {
		return nil, err
	}

	return model, nil
}

// initializeChildModels initializes the appropriate child models based on template shape
func (m *TemplatedNoteModel) initializeChildModels() error {
	// Initialize records model if template has record schema
	if m.template.Definition.RecordSchema != nil {
		records, err := database.ListTemplateRecords(m.db, m.noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to load records: %w", err)
		}

		recordsModel := NewRecordsTableModel(m.noteTemplate.ID, records, m.template.Definition.RecordSchema)
		m.recordsModel = &recordsModel
	}

	// Initialize steps model if template has steps
	if len(m.template.Definition.Steps) > 0 {
		stepsModel, err := NewStepsViewModel(m.db, m.noteID, m.noteTemplate.ID, m.template)
		if err != nil {
			return fmt.Errorf("failed to initialize steps model: %w", err)
		}
		m.stepsModel = stepsModel
	}

	// Initialize artifacts model (always present - inputs/outputs)
	artifacts, err := database.ListArtifacts(m.db, m.noteTemplate.ID)
	if err != nil {
		return fmt.Errorf("failed to load artifacts: %w", err)
	}

	m.artifactsModel = &ArtifactsViewModel{
		noteTemplateID: m.noteTemplate.ID,
		artifacts:      artifacts,
		cursor:         0,
		db:             m.db,
	}

	return nil
}

// ShowTemplatedNote launches the Bubble Tea TUI for a templated note
func ShowTemplatedNote(db *sql.DB, noteID int64) error {
	model, err := NewTemplatedNoteModel(db, noteID)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Init initializes the model
func (m TemplatedNoteModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m TemplatedNoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleInput(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update child model dimensions
		// Reserve space for header (6 lines) and footer (3 lines)
		childHeight := m.height - 9
		childWidth := m.width - 4

		if m.recordsModel != nil {
			m.recordsModel.width = childWidth
			m.recordsModel.height = childHeight
		}
		if m.stepsModel != nil {
			m.stepsModel.width = childWidth
			m.stepsModel.height = childHeight
		}
		if m.artifactsModel != nil {
			m.artifactsModel.width = childWidth
			m.artifactsModel.height = childHeight
		}

		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	// Delegate to active child model
	return m.updateActiveChild(msg)
}

// handleInput handles keyboard input for tab navigation
func (m TemplatedNoteModel) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		return m, tea.Quit

	// Tab navigation with numbers
	case "1":
		if m.recordsModel != nil {
			m.currentTab = TabRecords
		}
		return m, nil

	case "2":
		if m.stepsModel != nil {
			m.currentTab = TabSteps
		}
		return m, nil

	case "3":
		if m.artifactsModel != nil {
			m.currentTab = TabArtifacts
		}
		return m, nil

	// Tab navigation with letters
	case "r":
		if m.recordsModel != nil {
			m.currentTab = TabRecords
		}
		return m, nil

	case "s":
		if m.stepsModel != nil {
			m.currentTab = TabSteps
		}
		return m, nil

	case "a":
		if m.artifactsModel != nil {
			m.currentTab = TabArtifacts
		}
		return m, nil

	// Tab key cycles through tabs
	case "tab":
		m.currentTab = m.nextAvailableTab()
		return m, nil

	case "shift+tab":
		m.currentTab = m.previousAvailableTab()
		return m, nil

	default:
		// Pass through to active child model
		return m.updateActiveChild(msg)
	}
}

// nextAvailableTab returns the next available tab
func (m *TemplatedNoteModel) nextAvailableTab() int {
	start := m.currentTab
	for i := 0; i < 3; i++ {
		next := (start + i + 1) % 3
		if m.isTabAvailable(next) {
			return next
		}
	}
	return m.currentTab
}

// previousAvailableTab returns the previous available tab
func (m *TemplatedNoteModel) previousAvailableTab() int {
	start := m.currentTab
	for i := 0; i < 3; i++ {
		prev := (start - i - 1 + 3) % 3
		if m.isTabAvailable(prev) {
			return prev
		}
	}
	return m.currentTab
}

// isTabAvailable checks if a tab is available
func (m *TemplatedNoteModel) isTabAvailable(tab int) bool {
	switch tab {
	case TabRecords:
		return m.recordsModel != nil
	case TabSteps:
		return m.stepsModel != nil
	case TabArtifacts:
		return m.artifactsModel != nil
	default:
		return false
	}
}

// updateActiveChild delegates update to the active child model
func (m TemplatedNoteModel) updateActiveChild(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentTab {
	case TabRecords:
		if m.recordsModel != nil {
			updated, c := m.recordsModel.Update(msg)
			if updatedModel, ok := updated.(RecordsTableModel); ok {
				*m.recordsModel = updatedModel
				cmd = c
			}
		}

	case TabSteps:
		if m.stepsModel != nil {
			updated, c := m.stepsModel.Update(msg)
			if updatedModel, ok := updated.(*StepsViewModel); ok {
				m.stepsModel = updatedModel
				cmd = c
			}
		}

	case TabArtifacts:
		if m.artifactsModel != nil {
			updated, c := m.updateArtifactsModel(msg)
			if updatedModel, ok := updated.(*ArtifactsViewModel); ok {
				m.artifactsModel = updatedModel
				cmd = c
			}
		}
	}

	return m, cmd
}

// View renders the TUI
func (m TemplatedNoteModel) View() string {
	if m.quitting {
		return SuccessMessageStyle.Render("✓ Saved!") + "\n"
	}

	if m.err != nil {
		return ErrorMessageStyle.Render("Error: "+m.err.Error()) + "\n\n" +
			HelpStyle.Render("Press q to quit")
	}

	var s strings.Builder

	// Header with note title and template info
	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	// Tab indicators
	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	// Separator
	s.WriteString(Divider(m.width))
	s.WriteString("\n")

	// Active tab content
	s.WriteString(m.renderActiveTab())
	s.WriteString("\n")

	// Footer with help
	s.WriteString(Divider(m.width))
	s.WriteString("\n")
	s.WriteString(m.renderFooter())

	return s.String()
}

// renderHeader renders the note title and template information
func (m TemplatedNoteModel) renderHeader() string {
	var s strings.Builder

	// Note title
	title := TitleStyle.Render(m.note.Title)
	s.WriteString(title)
	s.WriteString("\n")

	// Template info
	templateInfo := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Render(fmt.Sprintf("Template: %s v%s", m.template.Name, m.template.Version))
	s.WriteString(templateInfo)
	s.WriteString("\n")

	// Progress summary
	s.WriteString(m.renderProgressSummary())

	return s.String()
}

// renderProgressSummary renders the progress summary line
func (m TemplatedNoteModel) renderProgressSummary() string {
	var parts []string

	// Steps progress
	if m.stepsModel != nil && len(m.stepsModel.steps) > 0 {
		completed := 0
		for _, step := range m.stepsModel.steps {
			if step.Completed {
				completed++
			}
		}
		total := len(m.stepsModel.steps)
		percentage := 0
		if total > 0 {
			percentage = (completed * 100) / total
		}
		stepInfo := fmt.Sprintf("%d/%d steps (%d%%)", completed, total, percentage)
		parts = append(parts, stepInfo)
	}

	// Records count
	if m.recordsModel != nil {
		recordCount := len(m.recordsModel.records)
		recordInfo := fmt.Sprintf("%d records", recordCount)
		parts = append(parts, recordInfo)
	}

	// Artifacts count
	if m.artifactsModel != nil {
		inputCount := 0
		outputCount := 0
		for _, artifact := range m.artifactsModel.artifacts {
			if artifact.ArtifactType == "input" {
				inputCount++
			} else if artifact.ArtifactType == "output" {
				outputCount++
			}
		}
		artifactInfo := fmt.Sprintf("%d/%d artifacts", outputCount, inputCount+outputCount)
		parts = append(parts, artifactInfo)
	}

	if len(parts) == 0 {
		return ""
	}

	progressText := "Progress: " + strings.Join(parts, " | ")
	return lipgloss.NewStyle().
		Foreground(ColorInfo).
		Render(progressText)
}

// renderTabs renders the tab indicators
func (m TemplatedNoteModel) renderTabs() string {
	var tabs []string

	// Records tab
	if m.recordsModel != nil {
		if m.currentTab == TabRecords {
			tabs = append(tabs, lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Background(lipgloss.Color("#2A2A2A")).
				Padding(0, 2).
				Render("[Records]"))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(ColorSubtle).
				Padding(0, 2).
				Render("Records"))
		}
	}

	// Steps tab
	if m.stepsModel != nil {
		if m.currentTab == TabSteps {
			tabs = append(tabs, lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Background(lipgloss.Color("#2A2A2A")).
				Padding(0, 2).
				Render("[Steps]"))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(ColorSubtle).
				Padding(0, 2).
				Render("Steps"))
		}
	}

	// Artifacts tab
	if m.artifactsModel != nil {
		if m.currentTab == TabArtifacts {
			tabs = append(tabs, lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Background(lipgloss.Color("#2A2A2A")).
				Padding(0, 2).
				Render("[Artifacts]"))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(ColorSubtle).
				Padding(0, 2).
				Render("Artifacts"))
		}
	}

	return strings.Join(tabs, " ")
}

// renderActiveTab renders the content of the active tab
func (m TemplatedNoteModel) renderActiveTab() string {
	switch m.currentTab {
	case TabRecords:
		if m.recordsModel != nil {
			return m.recordsModel.View()
		}

	case TabSteps:
		if m.stepsModel != nil {
			return m.stepsModel.View()
		}

	case TabArtifacts:
		if m.artifactsModel != nil {
			return m.renderArtifactsView()
		}
	}

	return EmptyState("No content available")
}

// renderFooter renders the help footer
func (m TemplatedNoteModel) renderFooter() string {
	var shortcuts []string

	// Tab navigation shortcuts
	if m.recordsModel != nil {
		shortcuts = append(shortcuts, "1/r: records")
	}
	if m.stepsModel != nil {
		shortcuts = append(shortcuts, "2/s: steps")
	}
	if m.artifactsModel != nil {
		shortcuts = append(shortcuts, "3/a: artifacts")
	}

	shortcuts = append(shortcuts, "tab: switch")
	shortcuts = append(shortcuts, "q/esc: back")

	// Add tab-specific shortcuts
	switch m.currentTab {
	case TabRecords:
		if m.recordsModel != nil {
			shortcuts = append(shortcuts, "j/k: navigate")
			shortcuts = append(shortcuts, "enter: edit")
		}

	case TabSteps:
		if m.stepsModel != nil {
			shortcuts = append(shortcuts, "j/k: navigate")
			shortcuts = append(shortcuts, "space: toggle")
		}

	case TabArtifacts:
		if m.artifactsModel != nil {
			shortcuts = append(shortcuts, "j/k: navigate")
			shortcuts = append(shortcuts, "enter: view")
		}
	}

	helpText := strings.Join(shortcuts, " • ")
	return HelpStyle.Render(helpText)
}

// ==================== Artifacts View (Simple Implementation) ====================

// updateArtifactsModel handles updates for the artifacts model
func (m *TemplatedNoteModel) updateArtifactsModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.artifactsModel == nil {
		return m.artifactsModel, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.artifactsModel.cursor > 0 {
				m.artifactsModel.cursor--
			}

		case "down", "j":
			if m.artifactsModel.cursor < len(m.artifactsModel.artifacts)-1 {
				m.artifactsModel.cursor++
			}

		case "home", "g":
			m.artifactsModel.cursor = 0

		case "end", "G":
			if len(m.artifactsModel.artifacts) > 0 {
				m.artifactsModel.cursor = len(m.artifactsModel.artifacts) - 1
			}
		}

	case tea.WindowSizeMsg:
		m.artifactsModel.width = msg.Width
		m.artifactsModel.height = msg.Height
	}

	return m.artifactsModel, nil
}

// renderArtifactsView renders the artifacts view
func (m *TemplatedNoteModel) renderArtifactsView() string {
	if m.artifactsModel == nil {
		return EmptyState("No artifacts view available")
	}

	var s strings.Builder

	// Header
	s.WriteString(SectionHeaderStyle.Render("📦 Inputs & Outputs"))
	s.WriteString("\n\n")

	if len(m.artifactsModel.artifacts) == 0 {
		s.WriteString(EmptyState("No artifacts defined yet"))
		return s.String()
	}

	// Group artifacts by type
	inputs := []*models.Artifact{}
	outputs := []*models.Artifact{}

	for _, artifact := range m.artifactsModel.artifacts {
		if artifact.ArtifactType == "input" {
			inputs = append(inputs, artifact)
		} else {
			outputs = append(outputs, artifact)
		}
	}

	// Render inputs
	if len(inputs) > 0 {
		s.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			Render("📥 Inputs"))
		s.WriteString("\n")

		for i, artifact := range inputs {
			globalIdx := i
			isSelected := globalIdx == m.artifactsModel.cursor
			s.WriteString(m.renderArtifactRow(artifact, isSelected))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Render outputs
	if len(outputs) > 0 {
		s.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			Render("📤 Outputs"))
		s.WriteString("\n")

		for i, artifact := range outputs {
			globalIdx := len(inputs) + i
			isSelected := globalIdx == m.artifactsModel.cursor
			s.WriteString(m.renderArtifactRow(artifact, isSelected))
			s.WriteString("\n")
		}
	}

	return s.String()
}

// renderArtifactRow renders a single artifact row
func (m *TemplatedNoteModel) renderArtifactRow(artifact *models.Artifact, isSelected bool) string {
	var style lipgloss.Style
	cursor := "  "

	if isSelected {
		style = ListItemSelectedStyle
		cursor = Cursor() + " "
	} else {
		style = ListItemStyle
	}

	// Status indicator
	status := "○"
	if artifact.Value != "" {
		status = "●"
	}

	// Required indicator
	required := ""
	if artifact.Required {
		required = "*"
	}

	// Type badge
	typeStr := ""
	switch artifact.Type {
	case "file":
		typeStr = "📄"
	case "folder":
		typeStr = "📁"
	case "url":
		typeStr = "🔗"
	case "text":
		typeStr = "📝"
	default:
		typeStr = "📦"
	}

	// Value preview
	value := artifact.Value
	if value == "" {
		value = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("(not set)")
	} else if len(value) > 50 {
		value = value[:47] + "..."
	}

	line := fmt.Sprintf("%s%s %s %s%s: %s", cursor, status, typeStr, artifact.Name, required, value)

	if isSelected {
		return style.Width(m.artifactsModel.width - 4).Render(line)
	}

	return style.Render(line)
}
