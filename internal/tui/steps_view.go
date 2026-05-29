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

// StepsViewModel is the Bubble Tea model for the steps checklist view
type StepsViewModel struct {
	db             *sql.DB
	noteID         int64
	noteTemplateID int64
	steps          []*models.ShapeState
	template       *models.Template
	inputs         map[string]interface{}
	runtime        *database.ShapeRuntime
	cursor         int
	showingDetails bool
	addingNote     bool
	noteInput      string
	width          int
	height         int
	embedded       bool
	scopePath      []string
	repeatStack    models.RepeatStack
	scopeNode      *models.ShapeNode
	checklistNode  *models.ShapeNode
	shapeState     *models.ShapeState
	checklistItems []models.ChecklistItemView
	useShapeState  bool
	err            error
	quitting       bool
}

// NewStepsViewModel creates a new steps view model
func NewStepsViewModel(db *sql.DB, noteID int64, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) (*StepsViewModel, error) {
	steps, err := database.ListTopLevelProcedureStates(db, noteTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load procedure states: %w", err)
	}

	runtime, err := database.LoadShapeRuntime(db, noteTemplateID, template, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to load shape runtime: %w", err)
	}

	return &StepsViewModel{
		db:             db,
		noteID:         noteID,
		noteTemplateID: noteTemplateID,
		steps:          steps,
		template:       template,
		inputs:         inputs,
		runtime:        runtime,
		cursor:         0,
		showingDetails: false,
		addingNote:     false,
		noteInput:      "",
		width:          80,
		height:         24,
	}, nil
}

// IsCapturingInput reports whether keyboard input should go to this view exclusively.
func (m *StepsViewModel) IsCapturingInput() bool {
	return m.addingNote || m.showingDetails
}

// SetScope loads checklist or step state for the current structure node.
func (m *StepsViewModel) SetScope(path []string, stack models.RepeatStack, node *models.ShapeNode) error {
	m.scopePath = append([]string{}, path...)
	m.repeatStack = stack
	m.scopeNode = node
	m.useShapeState = false
	m.checklistItems = nil
	m.shapeState = nil
	m.checklistNode = nil
	m.cursor = 0

	if err := m.refreshRuntime(); err != nil {
		return err
	}

	if node == nil {
		steps, err := database.ListTopLevelProcedureStates(m.db, m.noteTemplateID)
		if err != nil {
			return err
		}
		m.steps = steps
		return nil
	}

	structure, err := m.template.Definition.GetStructure()
	if err != nil {
		return err
	}

	if checklist := node.ChecklistForScope(); checklist != nil {
		ids := structure.PathTo(checklist.ID)
		if len(ids) == 0 {
			return fmt.Errorf("checklist not found in structure tree")
		}
		shapePath := models.ShapePath(ids)
		state, err := database.GetShapeState(m.db, m.noteTemplateID, shapePath, stack)
		if err != nil {
			return err
		}
		if state == nil {
			return fmt.Errorf("checklist state not initialized for %s", shapePath)
		}
		m.shapeState = state
		m.checklistNode = checklist
		m.checklistItems = models.ChecklistItemsFromState(checklist, state)
		m.useShapeState = true
		return nil
	}

	steps, err := database.ListProcedureStatesForScope(m.db, m.noteTemplateID, stack, path)
	if err != nil {
		return err
	}
	m.steps = steps
	return nil
}

func (m *StepsViewModel) refreshRuntime() error {
	runtime, err := database.LoadShapeRuntime(m.db, m.noteTemplateID, m.template, m.inputs)
	if err != nil {
		return err
	}
	m.runtime = runtime
	return nil
}

func (m *StepsViewModel) isStateBlocked(state *models.ShapeState) bool {
	if m.runtime == nil || state == nil {
		return false
	}
	return m.runtime.IsBlocked(state)
}

// SetEmbedded configures compact rendering for use inside the templated note view.
func (m *StepsViewModel) SetEmbedded(v bool) {
	m.embedded = v
}

// Init initializes the model
func (m StepsViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m StepsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.addingNote {
			return m.handleNoteInput(msg)
		}
		return m.handleNormalInput(msg)

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

// handleNormalInput handles keyboard input in normal mode
func (m StepsViewModel) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if !m.embedded {
			m.quitting = true
			return m, tea.Quit
		}

	case "q", "esc":
		if m.embedded {
			if m.showingDetails {
				m.showingDetails = false
			}
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.showingDetails = false
		}

	case "down", "j":
		if m.cursor < m.rowCount()-1 {
			m.cursor++
			m.showingDetails = false
		}

	case "enter", " ":
		if m.rowCount() > 0 && m.cursor < m.rowCount() {
			if err := m.toggleRow(m.cursor); err != nil {
				m.err = err
			}
		}

	case "n":
		// Start adding a note to the current step
		m.addingNote = true
		m.noteInput = ""
		if len(m.steps) > 0 && m.cursor < len(m.steps) {
			// Pre-populate with existing note if any
			m.noteInput = m.steps[m.cursor].Notes
		}

	case "v":
		// Toggle detailed view
		m.showingDetails = !m.showingDetails
	}

	return m, nil
}

// handleNoteInput handles keyboard input when adding a note
func (m StepsViewModel) handleNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingNote = false
		m.noteInput = ""

	case "enter":
		if len(m.steps) > 0 && m.cursor < len(m.steps) {
			state := m.steps[m.cursor]
			if err := database.UpdateShapeNotes(m.db, state, m.noteInput); err != nil {
				m.err = err
			} else {
				state.Notes = m.noteInput
			}
		}
		m.addingNote = false
		m.noteInput = ""

	case "backspace":
		if len(m.noteInput) > 0 {
			m.noteInput = m.noteInput[:len(m.noteInput)-1]
		}

	case "ctrl+u":
		// Clear entire line
		m.noteInput = ""

	case "ctrl+w":
		// Delete last word
		if len(m.noteInput) > 0 {
			parts := strings.Fields(m.noteInput)
			if len(parts) > 0 {
				m.noteInput = strings.Join(parts[:len(parts)-1], " ")
				if len(m.noteInput) > 0 {
					m.noteInput += " "
				}
			}
		}

	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.noteInput += msg.String()
		}
	}

	return m, nil
}

func (m *StepsViewModel) rowCount() int {
	if m.useShapeState {
		return len(m.checklistItems)
	}
	return len(m.steps)
}

func (m *StepsViewModel) toggleRow(index int) error {
	if m.useShapeState {
		if index < 0 || index >= len(m.checklistItems) {
			return nil
		}
		item := m.checklistItems[index]
		if err := database.ToggleChecklistItem(m.db, m.runtime, m.shapeState, item.ItemID, !item.Completed); err != nil {
			return err
		}
		m.checklistItems = models.ChecklistItemsFromState(m.checklistNode, m.shapeState)
		return nil
	}
	if index < 0 || index >= len(m.steps) {
		return nil
	}
	return m.toggleStepCompletion(m.steps[index])
}

// RenderInlineSection renders the progress bar and checklist for embedding in the orchestrator.
func (m StepsViewModel) RenderInlineSection() string {
	if m.rowCount() == 0 {
		return ""
	}
	var s strings.Builder
	s.WriteString(m.renderProgressBar())
	s.WriteString("\n\n")
	s.WriteString(m.renderStepsList())
	return s.String()
}

// ToggleCursor toggles completion for the item under the cursor.
func (m *StepsViewModel) ToggleCursor() error {
	if m.rowCount() == 0 || m.cursor >= m.rowCount() {
		return nil
	}
	return m.toggleRow(m.cursor)
}

// MoveCursor moves the checklist/step cursor by delta.
func (m *StepsViewModel) MoveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= m.rowCount() {
		m.cursor = m.rowCount() - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// HasChecklist returns whether scoped checklist items are loaded.
func (m StepsViewModel) HasChecklist() bool {
	return m.rowCount() > 0
}

// toggleStepCompletion toggles the completion status of a procedure shape state.
func (m *StepsViewModel) toggleStepCompletion(state *models.ShapeState) error {
	return database.ToggleShapeComplete(m.db, m.runtime, state, !state.Completed)
}

// View renders the TUI
func (m StepsViewModel) View() string {
	if m.quitting {
		return SuccessMessageStyle.Render("✓ Steps saved!") + "\n"
	}

	if m.err != nil {
		return ErrorMessageStyle.Render("Error: "+m.err.Error()) + "\n\n" +
			HelpStyle.Render("Press q to quit")
	}

	if m.addingNote {
		return m.renderNoteInput()
	}

	if m.showingDetails {
		return m.renderDetailedView()
	}

	return m.renderStepsView()
}

// renderStepsView renders the main steps checklist view
func (m StepsViewModel) renderStepsView() string {
	var s strings.Builder

	if !m.embedded {
		header := TitleWithBorderStyle.Render(fmt.Sprintf("📋 %s", m.template.Name))
		s.WriteString(header)
		s.WriteString("\n\n")
	}

	// Progress bar
	s.WriteString(m.renderProgressBar())
	s.WriteString("\n\n")

	// Steps list
	s.WriteString(m.renderStepsList())
	s.WriteString("\n")

	// Selected step details (compact when embedded to avoid viewport overflow)
	if m.rowCount() > 0 && m.cursor < m.rowCount() {
		if m.embedded {
			s.WriteString(m.renderCompactStepPreview())
		} else {
			s.WriteString(m.renderSelectedStepPreview())
		}
		s.WriteString("\n")
	}

	// Footer with help
	if !m.embedded {
		s.WriteString(m.renderHelp())
	}

	return s.String()
}

// renderProgressBar renders the progress indicator
func (m StepsViewModel) renderProgressBar() string {
	var s strings.Builder

	completedCount := 0
	totalCount := m.rowCount()

	if m.useShapeState {
		for _, item := range m.checklistItems {
			if item.Completed {
				completedCount++
			}
		}
	} else {
		for _, state := range m.steps {
			if state.Completed {
				completedCount++
			}
		}
	}

	percentage := 0
	if totalCount > 0 {
		percentage = (completedCount * 100) / totalCount
	}

	// Progress label
	progressLabel := ProgressTextStyle.Render("Progress:")
	s.WriteString(progressLabel)
	s.WriteString(" ")

	// Progress bar (40 characters wide)
	barWidth := 40
	bar := ProgressBar(completedCount, totalCount, barWidth)
	s.WriteString(bar)
	s.WriteString(" ")

	// Progress text
	progressText := ProgressPercentStyle.Render(fmt.Sprintf("%d/%d items (%d%%)", completedCount, totalCount, percentage))
	s.WriteString(progressText)

	return s.String()
}

// renderStepsList renders the list of all steps or scoped checklist items
func (m StepsViewModel) renderStepsList() string {
	var s strings.Builder

	if m.useShapeState {
		for i, item := range m.checklistItems {
			cursor := "  "
			if i == m.cursor {
				cursor = Cursor() + " "
			}

			checkbox := Checkbox(item.Completed)
			if !item.Completed && m.isStateBlocked(m.shapeState) {
				checkbox = lipgloss.NewStyle().Foreground(ColorWarning).Render("🔒")
			}

			var titleStyle lipgloss.Style
			if item.Completed {
				titleStyle = ListItemCompletedStyle
			} else if i == m.cursor {
				titleStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
			} else {
				titleStyle = ListItemStyle
			}

			title := titleStyle.Render(item.Title)
			line := fmt.Sprintf("%s%s %s", cursor, checkbox, title)

			if i == m.cursor {
				plainCheck := "☐"
				if item.Completed {
					plainCheck = "☑"
				}
				plainLine := fmt.Sprintf("> %s %s", plainCheck, item.Title)
				pad := ""
				if m.width > 2 {
					target := m.width - 2
					if w := lipgloss.Width(plainLine); w < target {
						pad = strings.Repeat(" ", target-w)
					}
				}
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("#2A2A2A")).
					Render(line + pad)
			}

			s.WriteString(line)
			s.WriteString("\n")
		}
		return s.String()
	}

	for i, state := range m.steps {
		cursor := "  "
		if i == m.cursor {
			cursor = Cursor() + " "
		}

		checkbox := Checkbox(state.Completed)
		if !state.Completed && m.isStateBlocked(state) {
			checkbox = lipgloss.NewStyle().Foreground(ColorWarning).Render("🔒")
		}

		stepNum := lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Render(fmt.Sprintf("%d.", i+1))

		var titleStyle lipgloss.Style
		if state.Completed {
			titleStyle = ListItemCompletedStyle
		} else if i == m.cursor {
			titleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
		} else {
			titleStyle = ListItemStyle
		}

		title := titleStyle.Render(state.Title)

		dateStr := ""
		if state.Completed && state.CompletedAt != nil {
			dateStr = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Render(fmt.Sprintf("  ✓ Completed %s", *state.CompletedAt))
		}

		line := fmt.Sprintf("%s%s %s %s%s", cursor, checkbox, stepNum, title, dateStr)

		if i == m.cursor {
			plainCheck := "☐"
			if state.Completed {
				plainCheck = "☑"
			}
			plainDate := ""
			if state.Completed && state.CompletedAt != nil {
				plainDate = fmt.Sprintf("  ✓ Completed %s", *state.CompletedAt)
			}
			plainLine := fmt.Sprintf("> %s %d. %s%s", plainCheck, i+1, state.Title, plainDate)

			pad := ""
			if m.width > 2 {
				target := m.width - 2
				if w := lipgloss.Width(plainLine); w < target {
					pad = strings.Repeat(" ", target-w)
				}
			}

			line = lipgloss.NewStyle().
				Background(lipgloss.Color("#2A2A2A")).
				Render(line + pad)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderCompactStepPreview renders a short preview when embedded in the templated note tab.
func (m StepsViewModel) renderCompactStepPreview() string {
	if m.rowCount() == 0 || m.cursor >= m.rowCount() {
		return ""
	}

	if m.useShapeState {
		item := m.checklistItems[m.cursor]
		var s strings.Builder
		header := SectionHeaderStyle.Render("Selected: " + item.Title)
		s.WriteString(header)
		s.WriteString("\n")
		if m.scopeNode != nil && m.scopeNode.Description != "" {
			desc := lipgloss.NewStyle().Foreground(ColorSubtle).Italic(true).Render(m.scopeNode.Description)
			s.WriteString(desc)
			s.WriteString("\n")
		}
		if m.isStateBlocked(m.shapeState) {
			s.WriteString(WarningMessageStyle.Render("🔒 Blocked — complete dependencies first"))
			s.WriteString("\n")
		}
		s.WriteString(HelpStyle.Render("space: toggle • esc: back"))
		return s.String()
	}

	if len(m.steps) == 0 || m.cursor >= len(m.steps) {
		return ""
	}

	state := m.steps[m.cursor]
	var s strings.Builder

	header := SectionHeaderStyle.Render(fmt.Sprintf("Selected: %d. %s", m.cursor+1, state.Title))
	s.WriteString(header)
	s.WriteString("\n")

	if node := m.nodeForState(state); node != nil && node.Description != "" {
		desc := lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true).
			Render(node.Description)
		s.WriteString(desc)
		s.WriteString("\n")
	}
	if m.isStateBlocked(state) {
		s.WriteString(WarningMessageStyle.Render("🔒 Blocked — complete dependencies first"))
		s.WriteString("\n")
	}

	s.WriteString(HelpStyle.Render("v: full details • n: add note"))
	return s.String()
}

func (m *StepsViewModel) nodeForState(state *models.ShapeState) *models.ShapeNode {
	if state == nil {
		return nil
	}
	comp, err := m.template.Definition.GetStructure()
	if err != nil || comp == nil {
		return nil
	}
	return comp.FindByPath(strings.Split(state.ShapePath, "."))
}

// renderSelectedStepPreview renders a preview of the selected procedure state.
func (m StepsViewModel) renderSelectedStepPreview() string {
	if len(m.steps) == 0 || m.cursor >= len(m.steps) {
		return ""
	}

	state := m.steps[m.cursor]
	node := m.nodeForState(state)
	var s strings.Builder

	header := SectionHeaderStyle.Render(fmt.Sprintf("Selected: %d. %s", m.cursor+1, state.Title))
	s.WriteString(header)
	s.WriteString("\n")

	if node != nil && node.Description != "" {
		desc := lipgloss.NewStyle().Foreground(ColorSubtle).Italic(true).Render(node.Description)
		s.WriteString(desc)
		s.WriteString("\n")
	}

	if node != nil && node.EstimatedTime != "" {
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorInfo).Render("Estimated: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSubtle).Render(node.EstimatedTime))
		s.WriteString("\n")
	}

	if state.Notes != "" {
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("Notes:"))
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSubtle).Italic(true).PaddingLeft(2).Render(state.Notes))
		s.WriteString("\n")
	}

	return PanelStyle.Width(m.width - 4).Render(s.String())
}

// renderDetailedView renders the detailed view of the selected procedure state.
func (m StepsViewModel) renderDetailedView() string {
	if len(m.steps) == 0 || m.cursor >= len(m.steps) {
		return EmptyState("No step selected")
	}

	state := m.steps[m.cursor]
	node := m.nodeForState(state)
	var s strings.Builder

	title := TitleStyle.Render(fmt.Sprintf("Step %d: %s", m.cursor+1, state.Title))
	s.WriteString(title)
	s.WriteString("\n\n")

	status := "Pending"
	if state.Completed {
		status = "Completed"
	}
	s.WriteString(StatusBadge(status))
	s.WriteString("\n\n")

	if node != nil && node.Description != "" {
		s.WriteString(SectionHeaderStyle.Render("Description"))
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorForeground).Render(node.Description))
		s.WriteString("\n\n")
	}

	if node != nil && node.EstimatedTime != "" {
		s.WriteString(SectionHeaderStyle.Render("Estimated Time"))
		s.WriteString("\n")
		s.WriteString(TimeBadge(node.EstimatedTime))
		s.WriteString("\n\n")
	}

	if node != nil && node.OutputRequired != "" {
		s.WriteString(SectionHeaderStyle.Render("Output Required"))
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("→ " + node.OutputRequired))
		s.WriteString("\n\n")
	}

	s.WriteString(SectionHeaderStyle.Render("Notes"))
	s.WriteString("\n")
	if state.Notes != "" {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSubtle).Render(state.Notes))
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("(No notes yet. Press 'n' to add a note)"))
	}
	s.WriteString("\n\n")

	if state.Completed && state.CompletedAt != nil {
		s.WriteString(SectionHeaderStyle.Render("Completed"))
		s.WriteString("\n")
		s.WriteString(DateBadge(*state.CompletedAt))
		s.WriteString("\n\n")
	}

	s.WriteString(HelpWithBorderStyle.Render(KeyBindings("v", "back to list", "space", "toggle completion", "n", "add/edit note", "esc/q", "exit")))
	return s.String()
}

// renderNoteInput renders the note input view.
func (m StepsViewModel) renderNoteInput() string {
	if len(m.steps) == 0 || m.cursor >= len(m.steps) {
		return EmptyState("No step selected")
	}

	state := m.steps[m.cursor]

	var s strings.Builder

	// Title
	title := TitleStyle.Render(fmt.Sprintf("Add Note to Step %d", m.cursor+1))
	s.WriteString(title)
	s.WriteString("\n\n")

	stepTitle := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Italic(true).
		Render(state.Title)
	s.WriteString(stepTitle)
	s.WriteString("\n\n")

	// Input label
	label := LabelStyle.Render("Note:")
	s.WriteString(label)
	s.WriteString("\n")

	// Input field
	inputWidth := m.width - 8
	if inputWidth < 40 {
		inputWidth = 40
	}

	cursor := "█"
	input := InputFocusedStyle.
		Width(inputWidth).
		Render(m.noteInput + cursor)
	s.WriteString(input)
	s.WriteString("\n\n")

	// Help
	help := HelpStyle.Render(
		KeyBindings(
			"enter", "save",
			"ctrl+u", "clear",
			"ctrl+w", "delete word",
			"esc", "cancel",
		),
	)
	s.WriteString(help)

	return s.String()
}

// renderHelp renders the help footer
func (m StepsViewModel) renderHelp() string {
	help := HelpWithBorderStyle.Render(
		KeyBindings(
			"↑/k", "up",
			"↓/j", "down",
			"space/enter", "toggle",
			"n", "add note",
			"v", "details",
			"q/esc", "quit",
		),
	)
	return help
}
