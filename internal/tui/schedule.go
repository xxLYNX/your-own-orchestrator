package tui

import (
	"fmt"
	"strings"
	"time"

	"yoo/internal/database"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScheduleModel is the Bubble Tea model for the schedule view
type ScheduleModel struct {
	notes      []*database.Note
	cursor     int
	date       time.Time
	width      int
	height     int
	err        error
	quitting   bool
	addingNote bool
	noteInput  string
}

// NewScheduleModel creates a new schedule model
func NewScheduleModel(date time.Time, notes []*database.Note) ScheduleModel {
	return ScheduleModel{
		notes:      notes,
		cursor:     0,
		date:       date,
		addingNote: false,
		noteInput:  "",
	}
}

// ShowSchedule launches the Bubble Tea TUI to display the schedule
func ShowSchedule(notes []*database.Note, date time.Time) error {
	model := NewScheduleModel(date, notes)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}

// Init initializes the model
func (m ScheduleModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ScheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// handleNormalInput handles keyboard input in normal mode
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
		// Start adding a new note
		m.addingNote = true
		m.noteInput = ""

	case "enter", " ":
		// Toggle completion status
		if len(m.notes) > 0 && m.cursor < len(m.notes) {
			if m.notes[m.cursor].Status == "completed" {
				m.notes[m.cursor].Status = "pending"
			} else {
				m.notes[m.cursor].Status = "completed"
			}
			// In a real app, this would trigger a database update
		}

	case "d":
		// Delete the current note
		if len(m.notes) > 0 && m.cursor < len(m.notes) {
			m.notes = append(m.notes[:m.cursor], m.notes[m.cursor+1:]...)
			if m.cursor >= len(m.notes) && m.cursor > 0 {
				m.cursor--
			}
			// In a real app, this would trigger a database delete
		}
	}

	return m, nil
}

// handleAddNoteInput handles keyboard input when adding a note
func (m ScheduleModel) handleAddNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingNote = false
		m.noteInput = ""

	case "enter":
		if m.noteInput != "" {
			// Create new note
			newNote := &database.Note{
				ID:          int64(len(m.notes) + 1),
				Title:       m.noteInput,
				ScheduledAt: m.date,
				Status:      "pending",
				Priority:    0,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			m.notes = append(m.notes, newNote)
			m.addingNote = false
			m.noteInput = ""
			// In a real app, this would trigger a database insert
		}

	case "backspace":
		if len(m.noteInput) > 0 {
			m.noteInput = m.noteInput[:len(m.noteInput)-1]
		}

	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.noteInput += msg.String()
		}
	}

	return m, nil
}

// View renders the TUI
func (m ScheduleModel) View() string {
	if m.quitting {
		return "Thanks for using yoo!\n"
	}

	if m.addingNote {
		return m.renderAddNoteView()
	}

	return m.renderScheduleView()
}

// renderScheduleView renders the main schedule view
func (m ScheduleModel) renderScheduleView() string {
	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	dateStr := m.date.Format("Monday, January 2, 2006")
	s.WriteString(headerStyle.Render(fmt.Sprintf("📅 Schedule for %s", dateStr)))
	s.WriteString("\n\n")

	// Notes list
	if len(m.notes) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
		s.WriteString(emptyStyle.Render("No notes for this day. Press 'a' to add one."))
		s.WriteString("\n\n")
	} else {
		for i, note := range m.notes {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checkbox := "☐"
			titleStyle := lipgloss.NewStyle()
			if note.Status == "completed" {
				checkbox = "☑"
				titleStyle = titleStyle.
					Foreground(lipgloss.Color("#666666")).
					Strikethrough(true)
			} else {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#FFFFFF"))
			}

			line := fmt.Sprintf("%s %s %s", cursor, checkbox, note.Title)
			if m.cursor == i {
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("#7D56F4")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Render(line)
			} else {
				line = titleStyle.Render(line)
			}

			s.WriteString(line)
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Footer with help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1)

	helpText := "↑/k: up • ↓/j: down • enter/space: toggle • a: add • d: delete • q: quit"
	s.WriteString(helpStyle.Render(helpText))

	return s.String()
}

// renderAddNoteView renders the add note input view
func (m ScheduleModel) renderAddNoteView() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	s.WriteString(titleStyle.Render("Add New Note"))
	s.WriteString("\n\n")

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(50)

	s.WriteString(inputStyle.Render(m.noteInput + "█"))
	s.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	s.WriteString(helpStyle.Render("enter: save • esc: cancel"))

	return s.String()
}
