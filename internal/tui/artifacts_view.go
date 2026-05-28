package tui

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"yoo/internal/database"
	"yoo/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArtifactsViewModel is the Bubble Tea model for the artifacts view
type ArtifactsViewModel struct {
	db             *sql.DB
	noteID         int64
	noteTemplateID int64
	artifacts      []*models.Artifact
	template       *models.Template
	cursor         int
	viewMode       string // "list", "add", "details", "delete"
	width          int
	height         int

	// Form fields for adding new artifacts
	formField        int // 0=name, 1=type, 2=artifact_type, 3=value, 4=description, 5=required
	formName         string
	formType         string // file, url, text, folder
	formArtifactType string // input, output
	formValue        string
	formDescription  string
	formRequired     bool

	embedded   bool
	err        error
	successMsg string
	quitting   bool
}

// NewArtifactsViewModel creates a new artifacts view model
func NewArtifactsViewModel(db *sql.DB, noteID int64, noteTemplateID int64, template *models.Template) (*ArtifactsViewModel, error) {
	artifacts, err := database.ListArtifacts(db, noteTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	return &ArtifactsViewModel{
		db:               db,
		noteID:           noteID,
		noteTemplateID:   noteTemplateID,
		artifacts:        artifacts,
		template:         template,
		cursor:           0,
		viewMode:         "list",
		width:            80,
		height:           24,
		formArtifactType: "input",
		formType:         "file",
		formRequired:     false,
	}, nil
}

// ShowArtifacts launches the Bubble Tea TUI to display the artifacts view
func ShowArtifacts(db *sql.DB, noteID int64, noteTemplateID int64, template *models.Template) error {
	model, err := NewArtifactsViewModel(db, noteID, noteTemplateID, template)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// IsCapturingInput reports whether keyboard input should go to this view exclusively.
func (m *ArtifactsViewModel) IsCapturingInput() bool {
	return m.viewMode != "list"
}

// SetEmbedded configures compact rendering for use inside the templated note view.
func (m *ArtifactsViewModel) SetEmbedded(v bool) {
	m.embedded = v
}

// Init initializes the model
func (m ArtifactsViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ArtifactsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.viewMode {
		case "list":
			return m.handleListInput(msg)
		case "add":
			return m.handleAddInput(msg)
		case "details":
			return m.handleDetailsInput(msg)
		case "delete":
			return m.handleDeleteInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// handleListInput handles keyboard input in list mode
func (m ArtifactsViewModel) handleListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if !m.embedded {
			m.quitting = true
			return m, tea.Quit
		}

	case "q", "esc":
		if m.embedded {
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.successMsg = ""
		}

	case "down", "j":
		if m.cursor < len(m.artifacts)-1 {
			m.cursor++
			m.successMsg = ""
		}

	case "enter", "o":
		// Open artifact (file or URL)
		if len(m.artifacts) > 0 && m.cursor < len(m.artifacts) {
			artifact := m.artifacts[m.cursor]
			if err := m.openArtifact(artifact); err != nil {
				m.err = err
			} else {
				m.successMsg = fmt.Sprintf("Opened %s", artifact.Name)
			}
		}

	case "a":
		// Add new artifact
		m.viewMode = "add"
		m.resetForm()
		m.successMsg = ""

	case "d":
		// Delete selected artifact
		if len(m.artifacts) > 0 && m.cursor < len(m.artifacts) {
			m.viewMode = "delete"
			m.successMsg = ""
		}

	case "v":
		// View artifact details
		if len(m.artifacts) > 0 && m.cursor < len(m.artifacts) {
			m.viewMode = "details"
			m.successMsg = ""
		}
	}

	return m, nil
}

// handleAddInput handles keyboard input in add mode
func (m ArtifactsViewModel) handleAddInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = "list"
		m.err = nil
		return m, nil

	case "tab", "down":
		m.formField = (m.formField + 1) % 6

	case "shift+tab", "up":
		m.formField--
		if m.formField < 0 {
			m.formField = 5
		}

	case "enter":
		// Save artifact
		if m.formName == "" {
			m.err = fmt.Errorf("artifact name is required")
			return m, nil
		}

		artifact := &models.Artifact{
			NoteTemplateID: m.noteTemplateID,
			ArtifactType:   m.formArtifactType,
			Name:           m.formName,
			Type:           m.formType,
			Value:          m.formValue,
			Description:    m.formDescription,
			Required:       m.formRequired,
		}

		if err := database.CreateArtifact(m.db, artifact); err != nil {
			m.err = err
			return m, nil
		}

		// Reload artifacts
		artifacts, err := database.ListArtifacts(m.db, m.noteTemplateID)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.artifacts = artifacts
		m.viewMode = "list"
		m.successMsg = fmt.Sprintf("Added artifact: %s", m.formName)
		m.err = nil
		return m, nil

	case "backspace":
		m.deleteChar()

	case "space":
		if m.formField == 5 { // Required field
			m.formRequired = !m.formRequired
		} else {
			m.addChar(" ")
		}

	default:
		// Handle typing
		if len(msg.String()) == 1 {
			m.addChar(msg.String())
		}
	}

	return m, nil
}

// handleDetailsInput handles keyboard input in details mode
func (m ArtifactsViewModel) handleDetailsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "v":
		m.viewMode = "list"
	}
	return m, nil
}

// handleDeleteInput handles keyboard input in delete confirmation mode
func (m ArtifactsViewModel) handleDeleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.viewMode = "list"

	case "y", "enter":
		// Delete the artifact
		if len(m.artifacts) > 0 && m.cursor < len(m.artifacts) {
			artifact := m.artifacts[m.cursor]
			if err := database.DeleteArtifact(m.db, m.noteTemplateID, artifact.Name); err != nil {
				m.err = err
				m.viewMode = "list"
				return m, nil
			}

			// Reload artifacts
			artifacts, err := database.ListArtifacts(m.db, m.noteTemplateID)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.artifacts = artifacts

			// Adjust cursor if needed
			if m.cursor >= len(m.artifacts) && m.cursor > 0 {
				m.cursor--
			}

			m.successMsg = fmt.Sprintf("Deleted artifact: %s", artifact.Name)
			m.viewMode = "list"
		}
	}
	return m, nil
}

// View renders the view
func (m ArtifactsViewModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.viewMode {
	case "add":
		return m.renderAddView()
	case "details":
		return m.renderDetailsView()
	case "delete":
		return m.renderDeleteView()
	default:
		return m.renderListView()
	}
}

// renderListView renders the main list view
func (m ArtifactsViewModel) renderListView() string {
	var s strings.Builder

	// Title
	if !m.embedded {
		title := TitleStyle.Render("📦 Artifacts")
		s.WriteString(title)
		s.WriteString("\n\n")

		if m.template != nil {
			templateInfo := SubtitleStyle.Render(fmt.Sprintf("Template: %s", m.template.Name))
			s.WriteString(templateInfo)
			s.WriteString("\n\n")
		}
	}

	// Group artifacts by type
	inputs := []*models.Artifact{}
	outputs := []*models.Artifact{}
	for _, artifact := range m.artifacts {
		if artifact.ArtifactType == "input" {
			inputs = append(inputs, artifact)
		} else {
			outputs = append(outputs, artifact)
		}
	}

	// Render inputs
	s.WriteString(SectionHeaderStyle.Render("INPUTS:"))
	s.WriteString("\n")
	if len(inputs) == 0 {
		s.WriteString(EmptyState("  No input artifacts"))
		s.WriteString("\n")
	} else {
		for _, artifact := range inputs {
			s.WriteString(m.renderArtifactLine(artifact, m.getGlobalIndex(artifact)))
			s.WriteString("\n")
		}
	}
	s.WriteString("\n")

	// Render outputs
	s.WriteString(SectionHeaderStyle.Render("OUTPUTS:"))
	s.WriteString("\n")
	if len(outputs) == 0 {
		s.WriteString(EmptyState("  No output artifacts"))
		s.WriteString("\n")
	} else {
		for _, artifact := range outputs {
			s.WriteString(m.renderArtifactLine(artifact, m.getGlobalIndex(artifact)))
			s.WriteString("\n")
		}
	}
	s.WriteString("\n")

	// Summary
	requiredCount, providedCount := m.getArtifactStats()
	summaryText := fmt.Sprintf("Summary: %d/%d required artifacts provided", providedCount, requiredCount)
	if providedCount >= requiredCount {
		s.WriteString(SuccessMessageStyle.Render(summaryText))
	} else {
		s.WriteString(WarningMessageStyle.Render(summaryText))
	}
	s.WriteString("\n\n")

	// Error or success message
	if m.err != nil {
		s.WriteString(ErrorMessageStyle.Render(fmt.Sprintf("Error: %s", m.err)))
		s.WriteString("\n\n")
		m.err = nil // Clear error after displaying
	} else if m.successMsg != "" {
		s.WriteString(SuccessMessageStyle.Render(m.successMsg))
		s.WriteString("\n\n")
	}

	// Help
	if !m.embedded {
		help := KeyBindings(
			"j/k", "navigate",
			"enter/o", "open",
			"a", "add",
			"d", "delete",
			"v", "details",
			"q/esc", "back",
		)
		s.WriteString(HelpStyle.Render(help))
	}

	return s.String()
}

// renderArtifactLine renders a single artifact line
func (m ArtifactsViewModel) renderArtifactLine(artifact *models.Artifact, globalIndex int) string {
	var parts []string

	// Cursor
	if globalIndex == m.cursor {
		parts = append(parts, Cursor()+" ")
	} else {
		parts = append(parts, "  ")
	}

	// Status indicator
	statusIcon := m.getArtifactStatus(artifact)
	parts = append(parts, statusIcon+" ")

	// Name (with styling)
	nameStyle := lipgloss.NewStyle().Bold(true)
	if globalIndex == m.cursor {
		nameStyle = nameStyle.Foreground(ColorPrimary)
	}
	name := nameStyle.Render(PadRight(artifact.Name, 20))
	parts = append(parts, name)

	// Type
	typeStyle := lipgloss.NewStyle().Foreground(ColorSubtle)
	artifactType := typeStyle.Render(PadRight(artifact.Type, 8))
	parts = append(parts, artifactType)

	// Value/Path
	valueDisplay := m.getValueDisplay(artifact)
	valueStyle := lipgloss.NewStyle().Foreground(ColorForeground)
	if artifact.Value == "" {
		valueStyle = valueStyle.Foreground(ColorMuted).Italic(true)
	}
	parts = append(parts, valueStyle.Render(PadRight(valueDisplay, 35)))

	// Required badge
	if artifact.Required {
		requiredBadge := lipgloss.NewStyle().
			Foreground(ColorWarning).
			Render("[required]")
		parts = append(parts, requiredBadge)
	}

	return strings.Join(parts, " ")
}

// renderAddView renders the add artifact form
func (m ArtifactsViewModel) renderAddView() string {
	var s strings.Builder

	title := TitleStyle.Render("➕ Add New Artifact")
	s.WriteString(title)
	s.WriteString("\n\n")

	// Form fields
	fields := []struct {
		label string
		value string
		index int
	}{
		{"Name", m.formName, 0},
		{"Type (file/url/text/folder)", m.formType, 1},
		{"Artifact Type (input/output)", m.formArtifactType, 2},
		{"Value/Path", m.formValue, 3},
		{"Description", m.formDescription, 4},
	}

	for _, field := range fields {
		label := LabelStyle.Render(PadRight(field.label+":", 35))

		var value string
		if m.formField == field.index {
			value = InputFocusedStyle.Render(field.value + "▌")
		} else {
			value = InputStyle.Render(field.value)
		}

		s.WriteString(label + " " + value)
		s.WriteString("\n")
	}

	// Required checkbox
	label := LabelStyle.Render(PadRight("Required:", 35))
	checkbox := Checkbox(m.formRequired)
	if m.formField == 5 {
		s.WriteString(label + " " + InputFocusedStyle.Render(checkbox))
	} else {
		s.WriteString(label + " " + checkbox)
	}
	s.WriteString("\n\n")

	// Error message
	if m.err != nil {
		s.WriteString(ErrorMessageStyle.Render(fmt.Sprintf("Error: %s", m.err)))
		s.WriteString("\n\n")
	}

	// Help
	help := KeyBindings(
		"tab", "next field",
		"enter", "save",
		"esc", "cancel",
	)
	s.WriteString(HelpStyle.Render(help))

	return s.String()
}

// renderDetailsView renders the artifact details view
func (m ArtifactsViewModel) renderDetailsView() string {
	if len(m.artifacts) == 0 || m.cursor >= len(m.artifacts) {
		return "No artifact selected"
	}

	artifact := m.artifacts[m.cursor]
	var s strings.Builder

	title := TitleStyle.Render("📄 Artifact Details")
	s.WriteString(title)
	s.WriteString("\n\n")

	// Artifact details
	details := []struct {
		label string
		value string
	}{
		{"Name", artifact.Name},
		{"Artifact Type", artifact.ArtifactType},
		{"Type", artifact.Type},
		{"Value", artifact.Value},
		{"Description", artifact.Description},
		{"Required", fmt.Sprintf("%t", artifact.Required)},
		{"Created", artifact.CreatedAt.Format("2006-01-02 15:04:05")},
		{"Updated", artifact.UpdatedAt.Format("2006-01-02 15:04:05")},
	}

	for _, detail := range details {
		label := EmphasisStyle.Render(PadRight(detail.label+":", 20))
		value := detail.value
		if value == "" {
			value = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("(not set)")
		}
		s.WriteString(label + " " + value)
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// File info for file/folder types
	if artifact.Type == "file" || artifact.Type == "folder" {
		s.WriteString(SectionHeaderStyle.Render("File Information:"))
		s.WriteString("\n")

		if artifact.Value != "" {
			expandedPath := expandPath(artifact.Value)
			fileInfo := m.getFileInfo(expandedPath)
			s.WriteString(fileInfo)
		} else {
			s.WriteString(EmptyState("  No path specified"))
		}
		s.WriteString("\n\n")
	}

	// Help
	help := KeyBindings("esc/q/v", "back to list")
	s.WriteString(HelpStyle.Render(help))

	return s.String()
}

// renderDeleteView renders the delete confirmation view
func (m ArtifactsViewModel) renderDeleteView() string {
	if len(m.artifacts) == 0 || m.cursor >= len(m.artifacts) {
		return "No artifact selected"
	}

	artifact := m.artifacts[m.cursor]
	var s strings.Builder

	title := TitleStyle.Render("🗑️  Delete Artifact")
	s.WriteString(title)
	s.WriteString("\n\n")

	warning := WarningMessageStyle.Render(fmt.Sprintf("Are you sure you want to delete '%s'?", artifact.Name))
	s.WriteString(warning)
	s.WriteString("\n\n")

	info := SubtitleStyle.Render(fmt.Sprintf("Type: %s | Value: %s", artifact.Type, artifact.Value))
	s.WriteString(info)
	s.WriteString("\n\n")

	// Help
	help := KeyBindings("y/enter", "confirm", "n/esc", "cancel")
	s.WriteString(HelpStyle.Render(help))

	return s.String()
}

// Helper methods

// getGlobalIndex returns the global index of an artifact in the full list
func (m ArtifactsViewModel) getGlobalIndex(artifact *models.Artifact) int {
	for i, a := range m.artifacts {
		if a.ID == artifact.ID {
			return i
		}
	}
	return -1
}

// getArtifactStatus returns a status icon for an artifact
func (m ArtifactsViewModel) getArtifactStatus(artifact *models.Artifact) string {
	if artifact.Value != "" {
		return SuccessIcon() // ✓
	} else if artifact.Required {
		return WarningIcon() // ⚠
	}
	return lipgloss.NewStyle().Foreground(ColorMuted).Render("○")
}

// getValueDisplay returns a display string for the artifact value
func (m ArtifactsViewModel) getValueDisplay(artifact *models.Artifact) string {
	if artifact.Value == "" {
		return "(not provided)"
	}

	if artifact.Type == "file" || artifact.Type == "folder" {
		// Show relative or abbreviated path
		path := artifact.Value
		if strings.HasPrefix(path, "~/") {
			return path
		}
		if filepath.IsAbs(path) {
			// Try to make it relative to current directory
			if cwd, err := os.Getwd(); err == nil {
				if relPath, err := filepath.Rel(cwd, path); err == nil && !strings.HasPrefix(relPath, "..") {
					return "./" + relPath
				}
			}
		}
		return path
	}

	// Truncate long values
	return Truncate(artifact.Value, 35)
}

// getArtifactStats returns (required count, provided count)
func (m ArtifactsViewModel) getArtifactStats() (int, int) {
	required := 0
	provided := 0

	for _, artifact := range m.artifacts {
		if artifact.Required {
			required++
			if artifact.Value != "" {
				provided++
			}
		}
	}

	return required, provided
}

// getFileInfo returns formatted file information
func (m ArtifactsViewModel) getFileInfo(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrorMessageStyle.Render("  ✗ File does not exist")
		}
		return ErrorMessageStyle.Render(fmt.Sprintf("  ✗ Error: %s", err))
	}

	var s strings.Builder
	s.WriteString(SuccessMessageStyle.Render("  ✓ File exists"))
	s.WriteString("\n")

	s.WriteString("  Path: " + path)
	s.WriteString("\n")

	if info.IsDir() {
		s.WriteString("  Type: Directory")
	} else {
		s.WriteString(fmt.Sprintf("  Type: File | Size: %s", formatFileSize(info.Size())))
	}
	s.WriteString("\n")

	s.WriteString(fmt.Sprintf("  Modified: %s", info.ModTime().Format("2006-01-02 15:04:05")))

	return s.String()
}

// openArtifact opens an artifact (file or URL) with the default application
func (m ArtifactsViewModel) openArtifact(artifact *models.Artifact) error {
	if artifact.Value == "" {
		return fmt.Errorf("no value specified for artifact '%s'", artifact.Name)
	}

	var cmd *exec.Cmd
	value := artifact.Value

	// Handle file/folder types
	if artifact.Type == "file" || artifact.Type == "folder" {
		value = expandPath(value)

		// Check if file exists
		if _, err := os.Stat(value); err != nil {
			return fmt.Errorf("file does not exist: %s", value)
		}
	}

	// Open based on OS
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", value)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", value)
	default: // linux, freebsd, etc.
		// Try multiple options
		for _, opener := range []string{"xdg-open", "gio", "gnome-open", "kde-open"} {
			if _, err := exec.LookPath(opener); err == nil {
				cmd = exec.Command(opener, value)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no suitable opener found for your system")
		}
	}

	return cmd.Start()
}

// resetForm resets the form fields
func (m *ArtifactsViewModel) resetForm() {
	m.formField = 0
	m.formName = ""
	m.formType = "file"
	m.formArtifactType = "input"
	m.formValue = ""
	m.formDescription = ""
	m.formRequired = false
	m.err = nil
}

// addChar adds a character to the current form field
func (m *ArtifactsViewModel) addChar(char string) {
	switch m.formField {
	case 0:
		m.formName += char
	case 1:
		m.formType += char
	case 2:
		m.formArtifactType += char
	case 3:
		m.formValue += char
	case 4:
		m.formDescription += char
	}
}

// deleteChar deletes a character from the current form field
func (m *ArtifactsViewModel) deleteChar() {
	switch m.formField {
	case 0:
		if len(m.formName) > 0 {
			m.formName = m.formName[:len(m.formName)-1]
		}
	case 1:
		if len(m.formType) > 0 {
			m.formType = m.formType[:len(m.formType)-1]
		}
	case 2:
		if len(m.formArtifactType) > 0 {
			m.formArtifactType = m.formArtifactType[:len(m.formArtifactType)-1]
		}
	case 3:
		if len(m.formValue) > 0 {
			m.formValue = m.formValue[:len(m.formValue)-1]
		}
	case 4:
		if len(m.formDescription) > 0 {
			m.formDescription = m.formDescription[:len(m.formDescription)-1]
		}
	}
}

// Utility functions

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
