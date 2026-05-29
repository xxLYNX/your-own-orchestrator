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

// OrchestratorModel navigates the fractal structure tree for a templated note.
type OrchestratorModel struct {
	noteID         int64
	note           *database.Note
	noteTemplate   *models.NoteTemplate
	template       *models.Template
	structure      *models.ShapeNode
	inputs         map[string]interface{}
	navStack       []models.NavContext
	cursor         int
	recordsModel   *RecordsTableModel
	stepsModel     *StepsViewModel
	artifactsModel *ArtifactsViewModel
	panelMode      string
	focusChecklist bool
	db             *sql.DB
	width, height  int
	done           bool
	err            error
	quitting       bool
}

// NewOrchestratorModel creates a tree navigator for a templated note.
func NewOrchestratorModel(db *sql.DB, noteID int64) (*OrchestratorModel, error) {
	ctx, err := database.LoadTemplatedNoteContext(db, noteID)
	if err != nil {
		return nil, err
	}
	if !ctx.Note.IsTemplated {
		return nil, fmt.Errorf("note is not templated")
	}

	comp, err := ctx.Template.Definition.GetStructure()
	if err != nil {
		return nil, err
	}

	m := &OrchestratorModel{
		noteID:       noteID,
		note:         ctx.Note,
		noteTemplate: ctx.NoteTemplate,
		template:     ctx.Template,
		structure:    comp,
		inputs:       ctx.NoteTemplate.TemplateData.Inputs,
		navStack:     []models.NavContext{{Path: []string{comp.ID}}},
		db:           db,
		width:        80,
		height:       24,
	}

	if err := m.initChildPanels(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *OrchestratorModel) initChildPanels() error {
	if logNode := m.structure.FindFirstLogNode(); logNode != nil {
		records, err := database.ListTemplateRecords(m.db, m.noteTemplate.ID, nil)
		if err != nil {
			return err
		}
		rm := NewRecordsTableModel(m.db, m.noteTemplate.ID, records, logNode.RecordSchema)
		rm.SetEmbedded(true)
		rm.SetRepeatFilter(nil)
		m.recordsModel = &rm
	}

	if m.template.Definition.HasProcedureShape() {
		if err := database.EnsureShapeStates(m.db, m.noteTemplate.ID, m.template, m.inputs); err != nil {
			return err
		}
		stepsModel, err := NewStepsViewModel(m.db, m.noteID, m.noteTemplate.ID, m.template, m.inputs)
		if err != nil {
			return err
		}
		stepsModel.SetEmbedded(true)
		if err := stepsModel.SetScope(m.currentNav().Path, m.currentNav().RepeatStack, m.currentNode()); err != nil {
			return err
		}
		m.stepsModel = stepsModel
		m.focusChecklist = stepsModel.HasChecklist()
	}

	if m.template.Definition.HasArtifactShape() {
		artifactsModel, err := NewArtifactsViewModel(m.db, m.noteID, m.noteTemplate.ID, m.template)
		if err != nil {
			return err
		}
		artifactsModel.SetEmbedded(true)
		m.artifactsModel = artifactsModel
	}
	return nil
}

func (m *OrchestratorModel) currentNav() models.NavContext {
	if len(m.navStack) == 0 {
		return models.NavContext{Path: []string{m.structure.ID}}
	}
	return m.navStack[len(m.navStack)-1]
}

func (m *OrchestratorModel) currentNode() *models.ShapeNode {
	return m.structure.FindByPath(m.currentNav().Path)
}

func (m *OrchestratorModel) Init() tea.Cmd {
	return nil
}

func (m *OrchestratorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.panelMode != "" {
		return m.updatePanel(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleInput(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeChildPanels()
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}
	return m, nil
}

func (m *OrchestratorModel) resizeChildPanels() {
	h := m.height - 12
	if h < 8 {
		h = 8
	}
	w := m.width - 4
	if w < 40 {
		w = 40
	}
	if m.recordsModel != nil {
		m.recordsModel.width = w
		m.recordsModel.height = h
	}
	if m.stepsModel != nil {
		m.stepsModel.width = w
		m.stepsModel.height = h
	}
	if m.artifactsModel != nil {
		m.artifactsModel.width = w
		m.artifactsModel.height = h
	}
}

func (m *OrchestratorModel) updatePanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.panelMode {
	case "log":
		if m.recordsModel != nil {
			updated, _ := m.recordsModel.Update(msg)
			if rm, ok := updated.(RecordsTableModel); ok {
				*m.recordsModel = rm
			}
		}
	case "steps":
		if m.stepsModel != nil {
			updated, _ := m.stepsModel.Update(msg)
			if sm, ok := updated.(*StepsViewModel); ok {
				m.stepsModel = sm
				if key, ok := msg.(tea.KeyMsg); ok && (key.String() == " " || key.String() == "enter") {
					_ = m.persistProgress()
				}
			}
		}
	case "artifacts":
		if m.artifactsModel != nil {
			updated, _ := m.artifactsModel.Update(msg)
			if am, ok := updated.(ArtifactsViewModel); ok {
				*m.artifactsModel = am
			}
		}
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		if !m.panelCapturingInput() {
			switch key.String() {
			case "q":
				m.panelMode = ""
			case "esc":
				return m.requestExit()
			}
		}
	}
	return m, nil
}

func (m *OrchestratorModel) panelCapturingInput() bool {
	switch m.panelMode {
	case "log":
		if m.recordsModel != nil {
			return m.recordsModel.IsCapturingInput()
		}
	case "steps":
		if m.stepsModel != nil {
			return m.stepsModel.IsCapturingInput()
		}
	case "artifacts":
		if m.artifactsModel != nil {
			return m.artifactsModel.IsCapturingInput()
		}
	}
	return false
}

func (m *OrchestratorModel) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.requestExit()
	case "esc":
		return m.requestExit()
	case "h":
		m.drillOut()
		m.syncStepsScope()
	case "l", "enter", "o":
		m.drillIntoSelection()
		m.syncStepsScope()
	case "tab":
		if m.stepsModel != nil && m.stepsModel.HasChecklist() && m.listCount() > 0 {
			m.focusChecklist = !m.focusChecklist
		}
	case " ":
		if m.stepsModel != nil && m.stepsModel.HasChecklist() {
			if err := m.stepsModel.ToggleCursor(); err != nil {
				m.err = err
			} else if err := m.persistProgress(); err != nil {
				m.err = err
			}
			return m, nil
		}
	case "x":
		if m.stepsModel != nil && m.focusChecklist {
			if err := m.stepsModel.SetCursorTerminalStatus(models.StatusSkipped); err != nil {
				m.err = err
			} else if err := m.persistProgress(); err != nil {
				m.err = err
			}
			return m, nil
		}
	case "!":
		if m.stepsModel != nil && m.focusChecklist {
			if err := m.stepsModel.SetCursorTerminalStatus(models.StatusFailed); err != nil {
				m.err = err
			} else if err := m.persistProgress(); err != nil {
				m.err = err
			}
			return m, nil
		}
	case "u":
		if m.stepsModel != nil && m.focusChecklist {
			if err := m.stepsModel.SetCursorTerminalStatus(models.StatusNotStarted); err != nil {
				m.err = err
			} else if err := m.persistProgress(); err != nil {
				m.err = err
			}
			return m, nil
		}
	case "up", "k":
		if m.focusChecklist && m.stepsModel != nil && m.stepsModel.HasChecklist() {
			m.stepsModel.MoveCursor(-1)
		} else if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.focusChecklist && m.stepsModel != nil && m.stepsModel.HasChecklist() {
			m.stepsModel.MoveCursor(1)
		} else if m.cursor < m.listCount()-1 {
			m.cursor++
		}
	case "r":
		if m.recordsModel != nil {
			if err := m.refreshRecords(); err != nil {
				m.err = err
				return m, nil
			}
			m.panelMode = "log"
		}
	case "p":
		if m.stepsModel != nil {
			if err := m.refreshStepsScope(); err != nil {
				m.err = err
				return m, nil
			}
			m.panelMode = "steps"
		}
	case "f":
		if m.artifactsModel != nil {
			m.panelMode = "artifacts"
		}
	}
	return m, nil
}

func (m *OrchestratorModel) drillOut() {
	if len(m.navStack) > 1 {
		m.navStack = m.navStack[:len(m.navStack)-1]
		m.cursor = 0
	}
}

func (m *OrchestratorModel) syncStepsScope() {
	if m.stepsModel == nil {
		return
	}
	if err := m.refreshStepsScope(); err != nil {
		m.err = err
		return
	}
	m.focusChecklist = m.stepsModel.HasChecklist()
}

func (m *OrchestratorModel) persistProgress() error {
	progress, err := database.ComputeTemplateProgress(m.db, m.noteTemplate.ID, m.template, m.inputs)
	if err != nil {
		return err
	}
	m.note.TemplateProgress = progress
	return database.UpdateTemplateProgress(m.db, m.noteID, progress)
}

func (m *OrchestratorModel) requestExit() (tea.Model, tea.Cmd) {
	m.done = true
	return m, nil
}

func (m *OrchestratorModel) listCount() int {
	node := m.currentNode()
	if node == nil {
		return 0
	}
	if models.NeedsRepeatPicker(node, m.currentNav().RepeatStack, m.inputs) {
		return node.EffectiveRepeatCount(m.inputs)
	}
	return len(node.NavigableChildren())
}

func (m *OrchestratorModel) drillIntoSelection() {
	node := m.currentNode()
	if node == nil {
		return
	}

	nav := m.currentNav()

	if models.NeedsRepeatPicker(node, nav.RepeatStack, m.inputs) {
		count := node.EffectiveRepeatCount(m.inputs)
		if m.cursor < 0 || m.cursor >= count {
			return
		}
		newStack := nav.RepeatStack.WithFrame(node.ID, m.cursor+1)
		if err := database.EnsureRepeatScope(m.db, m.noteTemplate.ID, m.structure, node, newStack, m.inputs); err != nil {
			m.err = err
			return
		}
		m.navStack = append(m.navStack, models.NavContext{
			Path:        append(append([]string{}, nav.Path...), node.ID),
			RepeatStack: newStack,
		})
		m.cursor = 0
		return
	}

	switch node.Kind {
	case models.ShapeLog:
		if err := m.refreshRecords(); err != nil {
			m.err = err
			return
		}
		m.panelMode = "log"
		return
	case models.ShapeArtifact:
		m.panelMode = "artifacts"
		return
	}

	children := node.NavigableChildren()
	if m.cursor < 0 || m.cursor >= len(children) {
		return
	}
	child := children[m.cursor]
	m.navStack = append(m.navStack, models.NavContext{
		Path:        append(append([]string{}, nav.Path...), child.ID),
		RepeatStack: append(models.RepeatStack(nil), nav.RepeatStack...),
	})
	m.cursor = 0
}

func (m *OrchestratorModel) recordRepeatScope() models.RepeatStack {
	stack := m.currentNav().RepeatStack
	if len(stack) == 0 {
		return nil
	}
	return stack
}

func (m *OrchestratorModel) refreshRecords() error {
	if m.recordsModel == nil {
		return nil
	}

	scope := m.recordRepeatScope()
	logNode := m.structure.FindFirstLogNode()
	if node := m.currentNode(); node != nil {
		if scoped := node.FindFirstLogNode(); scoped != nil {
			logNode = scoped
		}
	}
	if logNode == nil {
		return fmt.Errorf("no log shape found")
	}

	records, err := database.ListTemplateRecords(m.db, m.noteTemplate.ID, scope)
	if err != nil {
		return err
	}

	m.recordsModel.recordSchema = logNode.RecordSchema
	m.recordsModel.SetRepeatFilter(scope)
	m.recordsModel.ReloadRecords(records)
	return nil
}

func (m *OrchestratorModel) refreshStepsScope() error {
	if m.stepsModel == nil {
		return nil
	}
	return m.stepsModel.SetScope(m.currentNav().Path, m.currentNav().RepeatStack, m.currentNode())
}

func (m *OrchestratorModel) View() string {
	if m.quitting {
		return SuccessMessageStyle.Render("✓ Saved!") + "\n"
	}
	if m.err != nil {
		return ErrorMessageStyle.Render("Error: "+m.err.Error()) + "\n"
	}
	if m.panelMode != "" {
		return m.renderPanelView()
	}

	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	s.WriteString(m.renderOverallProgress())
	s.WriteString("\n")
	s.WriteString(Divider(m.width))
	s.WriteString("\n")
	s.WriteString(m.renderTreeList())
	if m.stepsModel != nil {
		inline := m.stepsModel.RenderInlineSection()
		if inline != "" {
			s.WriteString("\n")
			s.WriteString(Divider(m.width))
			s.WriteString("\n")
			s.WriteString(inline)
		}
	}
	s.WriteString("\n")
	s.WriteString(Divider(m.width))
	s.WriteString("\n")
	s.WriteString(m.renderFooter())
	return s.String()
}

func (m *OrchestratorModel) renderHeader() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render(m.note.Title))
	s.WriteString("\n")
	shapes := strings.Join(m.structure.ActiveShapeKinds(), " + ")
	meta := fmt.Sprintf("Template: %s v%s  •  shapes: %s", m.template.Name, m.template.Version, shapes)
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSubtle).Render(meta))
	s.WriteString("\n")
	crumb := lipgloss.NewStyle().Foreground(ColorInfo).Render("📍 " + m.currentNav().String())
	s.WriteString(crumb)
	return s.String()
}

func (m *OrchestratorModel) renderOverallProgress() string {
	progress, err := database.ComputeTemplateProgress(m.db, m.noteTemplate.ID, m.template, m.inputs)
	if err != nil {
		return WarningMessageStyle.Render("Progress unavailable")
	}
	m.note.TemplateProgress = progress
	pct := int(progress * 100)
	bar := ProgressBar(pct, 100, 40)
	label := ProgressTextStyle.Render("Overall:")
	pctLabel := ProgressPercentStyle.Render(fmt.Sprintf("%d%%", pct))
	return label + " " + bar + " " + pctLabel
}

func (m *OrchestratorModel) renderTreeList() string {
	node := m.currentNode()
	if node == nil {
		return EmptyState("Nothing to show")
	}

	var s strings.Builder
	title := SectionHeaderStyle.Render(node.DisplayTitle())
	if node.Description != "" {
		title += " — " + lipgloss.NewStyle().Foreground(ColorSubtle).Italic(true).Render(node.Description)
	}
	s.WriteString(title)
	s.WriteString("\n\n")

	if models.NeedsRepeatPicker(node, m.currentNav().RepeatStack, m.inputs) {
		count := node.EffectiveRepeatCount(m.inputs)
		if count == 0 {
			s.WriteString(WarningMessageStyle.Render("Set repeat count via template inputs (e.g. target_count)."))
			s.WriteString("\n")
			return s.String()
		}
		for i := 0; i < count; i++ {
			s.WriteString(m.renderTreeLine(i, fmt.Sprintf("%s #%d", node.DisplayTitle(), i+1), i == m.cursor))
			s.WriteString("\n")
		}
		return s.String()
	}

	children := node.NavigableChildren()
	if len(children) == 0 && !m.focusChecklist {
		kindHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("(%s leaf — press enter to open tools)", node.Kind))
		s.WriteString(kindHint)
		s.WriteString("\n")
		return s.String()
	}

	for i, child := range children {
		label := fmt.Sprintf("[%s] %s", child.Kind, child.DisplayTitle())
		s.WriteString(m.renderTreeLine(i, label, i == m.cursor))
		s.WriteString("\n")
	}
	return s.String()
}

func (m *OrchestratorModel) renderTreeLine(index int, label string, selected bool) string {
	cursor := "  "
	if selected {
		cursor = Cursor() + " "
	}
	line := fmt.Sprintf("%s%s", cursor, label)
	if selected {
		pad := ""
		if m.width > 2 {
			plain := cursor + label
			if w := lipgloss.Width(plain); w < m.width-2 {
				pad = strings.Repeat(" ", m.width-2-w)
			}
		}
		return lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Background(lipgloss.Color("#2A2A2A")).
			Render(line + pad)
	}
	return ListItemStyle.Render(line)
}

func (m *OrchestratorModel) renderPanelView() string {
	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	s.WriteString(Divider(m.width))
	s.WriteString("\n")

	switch m.panelMode {
	case "log":
		if m.recordsModel != nil {
			s.WriteString(m.recordsModel.View())
		}
	case "steps":
		if m.stepsModel != nil {
			s.WriteString(m.stepsModel.View())
		}
	case "artifacts":
		if m.artifactsModel != nil {
			s.WriteString(m.artifactsModel.View())
		}
	}

	s.WriteString("\n")
	s.WriteString(Divider(m.width))
	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("q: back to tree • esc: exit note"))
	return s.String()
}

func (m *OrchestratorModel) renderFooter() string {
	parts := []string{"jk: navigate", "space: toggle", "x/!/u: skip/fail/reset", "h/l: out/in", "tab: tree/checklist", "esc: exit"}
	if m.recordsModel != nil {
		parts = append(parts, "r: log")
	}
	if m.stepsModel != nil {
		parts = append(parts, "p: procedure panel")
	}
	if m.artifactsModel != nil {
		parts = append(parts, "f: artifacts")
	}
	return HelpStyle.Render(strings.Join(parts, " • "))
}
