package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"yoo/internal/database"

	tea "github.com/charmbracelet/bubbletea"
)

// ScheduleModel is the Bubble Tea model for the schedule view.
type ScheduleModel struct {
	db         *sql.DB
	notes      []*database.Note
	cursor     int
	date       time.Time
	width      int
	height     int
	detail     *OrchestratorModel
	err        error
	quitting   bool
	addingNote bool
	noteInput  string
}

// NewScheduleModel creates a new schedule model.
func NewScheduleModel(db *sql.DB, date time.Time, notes []*database.Note) ScheduleModel {
	return ScheduleModel{
		db:         db,
		notes:      notes,
		cursor:     0,
		date:       date,
		addingNote: false,
		noteInput:  "",
	}
}

// ShowSchedule launches the Bubble Tea TUI to display the schedule.
func ShowSchedule(db *sql.DB, notes []*database.Note, date time.Time) error {
	model := NewScheduleModel(db, date, notes)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m ScheduleModel) Init() tea.Cmd {
	return nil
}

func (m ScheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.detail != nil {
		return m.updateDetail(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.addingNote {
			return m.handleAddNoteInput(msg)
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

func (m ScheduleModel) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.detail.Update(msg)
	detail, ok := updated.(*OrchestratorModel)
	if !ok {
		return m, cmd
	}

	m.detail = detail
	if detail.done {
		m.detail = nil
		refreshed, reloadCmd := m.reloadNotes()
		return refreshed, reloadCmd
	}

	return m, cmd
}

func (m ScheduleModel) reloadNotes() (ScheduleModel, tea.Cmd) {
	notes, err := database.GetNotesByDate(m.db, m.date)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.notes = notes
	if m.cursor >= len(m.notes) {
		if len(m.notes) == 0 {
			m.cursor = 0
		} else {
			m.cursor = len(m.notes) - 1
		}
	}

	return m, nil
}

func (m ScheduleModel) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.notes)-1 {
			m.cursor++
		}

	case "a", "n":
		m.addingNote = true
		m.noteInput = ""

	case "enter":
		if len(m.notes) > 0 && m.cursor < len(m.notes) {
			note := m.notes[m.cursor]
			if note.IsTemplated {
				detail, err := NewOrchestratorModel(m.db, note.ID)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.detail = detail
				return m, nil
			}
			return m.toggleNoteCompletion()
		}

	case " ":
		if len(m.notes) > 0 && m.cursor < len(m.notes) {
			note := m.notes[m.cursor]
			if note.IsTemplated {
				detail, err := NewOrchestratorModel(m.db, note.ID)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.detail = detail
				return m, nil
			}
		}
		return m.toggleNoteCompletion()

	case "d":
		if len(m.notes) > 0 && m.cursor < len(m.notes) {
			note := m.notes[m.cursor]
			if err := database.DeleteNote(m.db, note.ID); err != nil {
				m.err = err
				return m, nil
			}
			m.notes = append(m.notes[:m.cursor], m.notes[m.cursor+1:]...)
			if m.cursor >= len(m.notes) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m ScheduleModel) toggleNoteCompletion() (ScheduleModel, tea.Cmd) {
	if len(m.notes) == 0 || m.cursor >= len(m.notes) {
		return m, nil
	}

	note := m.notes[m.cursor]
	if note.Status == "completed" {
		note.Status = "pending"
	} else {
		note.Status = "completed"
	}

	if err := database.UpdateNote(m.db, note); err != nil {
		m.err = err
	}

	return m, nil
}

func (m ScheduleModel) handleAddNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingNote = false
		m.noteInput = ""

	case "enter":
		if m.noteInput != "" {
			newNote := &database.Note{
				Title:       m.noteInput,
				ScheduledAt: m.date,
				Status:      "pending",
				Priority:    0,
			}
			if err := database.CreateNote(m.db, newNote); err != nil {
				m.err = err
				return m, nil
			}

			m.notes = append(m.notes, newNote)
			m.cursor = len(m.notes) - 1
			m.addingNote = false
			m.noteInput = ""
		}

	case "backspace":
		if len(m.noteInput) > 0 {
			m.noteInput = m.noteInput[:len(m.noteInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.noteInput += msg.String()
		}
	}

	return m, nil
}

func (m ScheduleModel) View() string {
	if m.detail != nil {
		return m.detail.View()
	}

	if m.quitting {
		return SuccessMessageStyle.Render("Thanks for using yoo!") + "\n"
	}

	if m.err != nil {
		return ErrorMessageStyle.Render("Error: "+m.err.Error()) + "\n\n" +
			HelpStyle.Render("Press q to quit")
	}

	if m.addingNote {
		return m.renderAddNoteView()
	}

	return m.renderScheduleView()
}

func (m ScheduleModel) renderScheduleView() string {
	var s strings.Builder

	dateStr := m.date.Format("Monday, January 2, 2006")
	s.WriteString(TitleWithBorderStyle.Render(fmt.Sprintf("📅 Schedule for %s", dateStr)))
	s.WriteString("\n\n")

	if len(m.notes) == 0 {
		s.WriteString(EmptyState("No notes for this day. Press 'a' to add one."))
		s.WriteString("\n\n")
	} else {
		for i, note := range m.notes {
			s.WriteString(m.renderNoteLine(i, note))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	helpText := "↑/k ↓/j: navigate • enter: open/toggle • space: toggle • a: add • d: delete • q: quit"
	s.WriteString(HelpWithBorderStyle.Render(helpText))

	return s.String()
}

func (m ScheduleModel) renderNoteLine(index int, note *database.Note) string {
	selected := m.cursor == index
	completed := note.Status == "completed"

	cursor := " "
	if selected {
		cursor = ">"
	}

	checkbox := "☐"
	if completed {
		checkbox = "☑"
	}

	title := note.Title
	if note.IsTemplated {
		progress := int(note.TemplateProgress * 100)
		marker := ">"
		if progress >= 100 {
			marker = "☑"
		}
		title = fmt.Sprintf("%s  %s structured %d%%", note.Title, marker, progress)
	}

	line := fmt.Sprintf("%s %s %s", cursor, checkbox, title)

	switch {
	case selected:
		return ListItemSelectedStyle.Render(line)
	case completed:
		return ListItemCompletedStyle.Render(line)
	default:
		return ListItemStyle.Render(line)
	}
}

func (m ScheduleModel) renderAddNoteView() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("Add New Note"))
	s.WriteString("\n\n")
	s.WriteString(InputFocusedStyle.Width(50).Render(m.noteInput + "█"))
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("enter: save • esc: cancel"))

	return s.String()
}
