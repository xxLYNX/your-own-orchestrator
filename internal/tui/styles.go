package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ==================== Color Palette ====================

var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#5C4099")

	// Status colors
	ColorSuccess = lipgloss.Color("#04B575")
	ColorWarning = lipgloss.Color("#F59E0B")
	ColorError   = lipgloss.Color("#EF4444")
	ColorInfo    = lipgloss.Color("#3B82F6")

	// State colors
	ColorPending    = lipgloss.Color("#F59E0B")
	ColorInProgress = lipgloss.Color("#3B82F6")
	ColorCompleted  = lipgloss.Color("#04B575")
	ColorDraft      = lipgloss.Color("#6B7280")

	// Neutral colors
	ColorMuted      = lipgloss.Color("#666666")
	ColorSubtle     = lipgloss.Color("#999999")
	ColorBackground = lipgloss.Color("#1A1A1A")
	ColorForeground = lipgloss.Color("#FFFFFF")
	ColorBorder     = lipgloss.Color("#444444")

	// Semantic colors
	ColorHighlight = lipgloss.Color("#7D56F4")
	ColorDimmed    = lipgloss.Color("#4A4A4A")
)

// ==================== Title Styles ====================

var (
	// TitleStyle is for main page titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// TitleWithBorderStyle is for main titles with decorative border
	TitleWithBorderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1).
				MarginBottom(1)

	// SectionHeaderStyle is for section headers
	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	// SubtitleStyle is for subtitles and descriptions
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true).
			MarginBottom(1)

	// EmphasisStyle is for emphasized text
	EmphasisStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorForeground)
)

// ==================== Table Styles ====================

var (
	// TableHeaderStyle is for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	// TableRowStyle is for regular table rows
	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Padding(0, 1)

	// TableRowSelectedStyle is for selected table rows
	TableRowSelectedStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorForeground).
				Bold(true).
				Padding(0, 1)

	// TableRowAlternateStyle is for alternating row colors
	TableRowAlternateStyle = lipgloss.NewStyle().
				Background(ColorDimmed).
				Foreground(ColorForeground).
				Padding(0, 1)

	// TableCellStyle is for individual table cells
	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// TableBorderStyle defines table borders
	TableBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(1, 2)
)

// ==================== Progress Styles ====================

var (
	// ProgressBarFilledStyle is for the filled portion of progress bars
	ProgressBarFilledStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Background(ColorSuccess)

	// ProgressBarEmptyStyle is for the empty portion of progress bars
	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorDimmed).
				Background(ColorDimmed)

	// ProgressPercentStyle is for percentage display
	ProgressPercentStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSuccess)

	// ProgressTextStyle is for progress text labels
	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)
)

// ==================== Status Styles ====================

var (
	// StatusPendingStyle is for pending status badges
	StatusPendingStyle = lipgloss.NewStyle().
				Foreground(ColorPending).
				Background(lipgloss.Color("#3D2E00")).
				Bold(true).
				Padding(0, 1)

	// StatusInProgressStyle is for in-progress status badges
	StatusInProgressStyle = lipgloss.NewStyle().
				Foreground(ColorInProgress).
				Background(lipgloss.Color("#001F3D")).
				Bold(true).
				Padding(0, 1)

	// StatusCompletedStyle is for completed status badges
	StatusCompletedStyle = lipgloss.NewStyle().
				Foreground(ColorCompleted).
				Background(lipgloss.Color("#002A1F")).
				Bold(true).
				Padding(0, 1)

	// StatusDraftStyle is for draft status badges
	StatusDraftStyle = lipgloss.NewStyle().
				Foreground(ColorDraft).
				Background(lipgloss.Color("#2A2A2A")).
				Bold(true).
				Padding(0, 1)

	// StatusErrorStyle is for error status badges
	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Background(lipgloss.Color("#3D0000")).
				Bold(true).
				Padding(0, 1)
)

// ==================== Help Styles ====================

var (
	// HelpStyle is for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	// HelpWithBorderStyle is for help text with border
	HelpWithBorderStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1).
				MarginTop(1)

	// HelpKeyStyle is for key binding keys
	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// HelpDescStyle is for key binding descriptions
	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	// HelpSeparatorStyle is for separating key bindings
	HelpSeparatorStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
)

// ==================== Border Styles ====================

var (
	// BorderNormalStyle is a normal border
	BorderNormalStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder)

	// BorderRoundedStyle is a rounded border
	BorderRoundedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder)

	// BorderDoubleStyle is a double-line border
	BorderDoubleStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(ColorBorder)

	// BorderThickStyle is a thick border
	BorderThickStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(ColorBorder)

	// BorderHighlightStyle is a highlighted border
	BorderHighlightStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)
)

// ==================== Layout Styles ====================

var (
	// PaddingStyle adds padding
	PaddingStyle = lipgloss.NewStyle().Padding(1, 2)

	// MarginStyle adds margin
	MarginStyle = lipgloss.NewStyle().Margin(1, 2)

	// CenterStyle centers content
	CenterStyle = lipgloss.NewStyle().Align(lipgloss.Center)

	// LeftStyle aligns content to the left
	LeftStyle = lipgloss.NewStyle().Align(lipgloss.Left)

	// RightStyle aligns content to the right
	RightStyle = lipgloss.NewStyle().Align(lipgloss.Right)

	// BoxStyle is a basic box container
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			Margin(1, 0)

	// PanelStyle is for larger panels with content
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Margin(0, 0, 1, 0)
)

// ==================== List Styles ====================

var (
	// ListItemStyle is for list items
	ListItemStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			PaddingLeft(2)

	// ListItemSelectedStyle is for selected list items
	ListItemSelectedStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorForeground).
				Bold(true).
				PaddingLeft(1)

	// ListItemCompletedStyle is for completed list items
	ListItemCompletedStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Strikethrough(true).
				PaddingLeft(2)

	// ListBulletStyle is for list bullets
	ListBulletStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// ==================== Form Styles ====================

var (
	// InputStyle is for text input fields
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	// InputFocusedStyle is for focused input fields
	InputFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Padding(0, 1).
				Bold(true)

	// LabelStyle is for form labels
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Bold(true).
			MarginRight(1)

	// PlaceholderStyle is for placeholder text
	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)
)

// ==================== Message Styles ====================

var (
	// SuccessMessageStyle is for success messages
	SuccessMessageStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true).
				Padding(0, 1)

	// ErrorMessageStyle is for error messages
	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true).
				Padding(0, 1)

	// WarningMessageStyle is for warning messages
	WarningMessageStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Padding(0, 1)

	// InfoMessageStyle is for info messages
	InfoMessageStyle = lipgloss.NewStyle().
				Foreground(ColorInfo).
				Bold(true).
				Padding(0, 1)
)

// ==================== Helper Functions ====================

// StatusBadge returns a styled status badge based on the status string
func StatusBadge(status string) string {
	status = strings.ToLower(status)
	switch status {
	case "pending":
		return StatusPendingStyle.Render("⏳ PENDING")
	case "in_progress", "in-progress", "inprogress":
		return StatusInProgressStyle.Render("▶ IN PROGRESS")
	case "completed", "complete", "done":
		return StatusCompletedStyle.Render("✓ COMPLETED")
	case "draft":
		return StatusDraftStyle.Render("📝 DRAFT")
	case "error", "failed":
		return StatusErrorStyle.Render("✗ ERROR")
	default:
		return StatusDraftStyle.Render(strings.ToUpper(status))
	}
}

// ProgressBar creates a visual progress bar
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ProgressBarEmptyStyle.Render(strings.Repeat(" ", width))
	}

	percentage := float64(current) / float64(total)
	if percentage > 1.0 {
		percentage = 1.0
	}

	filledWidth := int(float64(width) * percentage)
	emptyWidth := width - filledWidth

	filled := ProgressBarFilledStyle.Render(strings.Repeat("█", filledWidth))
	empty := ProgressBarEmptyStyle.Render(strings.Repeat("░", emptyWidth))

	return filled + empty
}

// ProgressBarWithPercentage creates a progress bar with percentage display
func ProgressBarWithPercentage(current, total int, width int) string {
	if total == 0 {
		return ProgressBar(0, 1, width) + " " + ProgressPercentStyle.Render("0%")
	}

	percentage := float64(current) / float64(total) * 100.0
	if percentage > 100.0 {
		percentage = 100.0
	}

	bar := ProgressBar(current, total, width)
	percent := ProgressPercentStyle.Render(fmt.Sprintf("%.1f%%", percentage))

	return bar + " " + percent
}

// ProgressBarWithLabel creates a progress bar with label and percentage
func ProgressBarWithLabel(label string, current, total int, width int) string {
	labelText := ProgressTextStyle.Render(label + ":")
	bar := ProgressBarWithPercentage(current, total, width)
	return labelText + " " + bar
}

// Checkbox returns a styled checkbox (checked or unchecked)
func Checkbox(checked bool) string {
	if checked {
		return lipgloss.NewStyle().Foreground(ColorSuccess).Render("☑")
	}
	return lipgloss.NewStyle().Foreground(ColorMuted).Render("☐")
}

// Bullet returns a styled bullet point
func Bullet() string {
	return ListBulletStyle.Render("•")
}

// Arrow returns a styled arrow indicator
func Arrow() string {
	return lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("→")
}

// Cursor returns a styled cursor for selected items
func Cursor() string {
	return lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(">")
}

// Divider returns a horizontal divider line
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("─", width))
}

// ThickDivider returns a thick horizontal divider line
func ThickDivider(width int) string {
	return lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("━", width))
}

// KeyBinding formats a key binding for help text
func KeyBinding(key, description string) string {
	return HelpKeyStyle.Render(key) +
		HelpSeparatorStyle.Render(": ") +
		HelpDescStyle.Render(description)
}

// KeyBindings formats multiple key bindings with separator
func KeyBindings(bindings ...string) string {
	var parts []string
	for i := 0; i < len(bindings); i += 2 {
		if i+1 < len(bindings) {
			parts = append(parts, KeyBinding(bindings[i], bindings[i+1]))
		}
	}
	return strings.Join(parts, HelpSeparatorStyle.Render(" • "))
}

// Priority returns a styled priority indicator
func Priority(level int) string {
	switch level {
	case 3, 4, 5:
		return lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("!!!")
	case 2:
		return lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("!!")
	case 1:
		return lipgloss.NewStyle().Foreground(ColorInfo).Bold(true).Render("!")
	default:
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("-")
	}
}

// Tag returns a styled tag
func Tag(text string) string {
	return lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Background(lipgloss.Color("#2A1F3D")).
		Padding(0, 1).
		Render("#" + text)
}

// DateBadge returns a styled date badge
func DateBadge(dateStr string) string {
	return lipgloss.NewStyle().
		Foreground(ColorInfo).
		Background(lipgloss.Color("#001F3D")).
		Bold(true).
		Padding(0, 1).
		Render("📅 " + dateStr)
}

// TimeBadge returns a styled time badge
func TimeBadge(timeStr string) string {
	return lipgloss.NewStyle().
		Foreground(ColorInfo).
		Background(lipgloss.Color("#001F3D")).
		Bold(true).
		Padding(0, 1).
		Render("🕐 " + timeStr)
}

// EmptyState returns a styled empty state message
func EmptyState(message string) string {
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true).
		Align(lipgloss.Center).
		Render(message)
}

// LoadingSpinner returns a loading spinner character based on frame
func LoadingSpinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render(spinners[frame%len(spinners)])
}

// SuccessIcon returns a success checkmark icon
func SuccessIcon() string {
	return lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("✓")
}

// ErrorIcon returns an error X icon
func ErrorIcon() string {
	return lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("✗")
}

// WarningIcon returns a warning icon
func WarningIcon() string {
	return lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("⚠")
}

// InfoIcon returns an info icon
func InfoIcon() string {
	return lipgloss.NewStyle().Foreground(ColorInfo).Bold(true).Render("ℹ")
}

// Truncate truncates a string to a maximum length with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PadRight pads a string to a specific width with spaces
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadLeft pads a string to a specific width with spaces on the left
func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// Center centers a string within a specific width
func Center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
