package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderPromptView renders the main capture screen layout using the design doc layout
// This provides an alternative rendering approach with the header/separator/footer design
func RenderPromptView(m *Model) string {
	var sections []string

	// header section
	header := renderHeader(m)
	sections = append(sections, header)

	// separator line
	sections = append(sections, renderSeparator(m.width))

	// main content area
	content := renderContent(m)
	sections = append(sections, content)

	// footer section
	footer := renderFooter(m.width)
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the top bar with title and tool selector
func renderHeader(m *Model) string {
	title := HeaderStyle.Render("crumb")
	toolSelector := m.toolSelect.View()

	// calculate padding to push tool selector to the right
	titleWidth := lipgloss.Width(title)
	toolWidth := lipgloss.Width(toolSelector)
	availableWidth := m.width - 4 // account for border padding
	padding := availableWidth - titleWidth - toolWidth

	if padding < 1 {
		padding = 1
	}

	headerContent := title + strings.Repeat(" ", padding) + toolSelector

	headerStyle := lipgloss.NewStyle().
		Width(m.width - 2).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

// renderSeparator renders a horizontal separator line
func renderSeparator(width int) string {
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Surface)).
		Width(width)

	return separatorStyle.Render(strings.Repeat("─", width))
}

// renderContent renders the main form content
func renderContent(m *Model) string {
	var lines []string

	// add vertical padding
	lines = append(lines, "")

	// prompt field (index 0) - first and most important
	promptSection := renderPromptFieldAlt(m)
	lines = append(lines, promptSection)
	lines = append(lines, "")

	// output field (index 1)
	outputSection := renderOutputFieldAlt(m)
	lines = append(lines, outputSection)
	lines = append(lines, "")

	// tags field (index 3)
	tagsSection := renderTagsFieldAlt(m)
	lines = append(lines, tagsSection)
	lines = append(lines, "")

	contentStyle := lipgloss.NewStyle().
		Width(m.width - 2).
		Padding(0, 2)

	return contentStyle.Render(strings.Join(lines, "\n"))
}

// renderPromptFieldAlt renders the prompt input area (index 0)
func renderPromptFieldAlt(m *Model) string {
	label := LabelStyle.Render("→ Prompt:")
	if m.focusIndex != 0 {
		label = LabelStyle.Render("Prompt:")
	}

	var borderStyle lipgloss.Style
	if m.focusIndex == 0 {
		borderStyle = FocusedBorderStyle
	} else {
		borderStyle = BorderStyle
	}

	content := m.prompt.View()

	inputBox := borderStyle.
		Width(m.width - 8).
		Padding(1).
		Render(content)

	return label + "\n" + inputBox
}

// renderTagsFieldAlt renders the tags input (index 3)
func renderTagsFieldAlt(m *Model) string {
	label := LabelStyle.Render("Tags: ")
	if m.focusIndex == 3 {
		label = LabelStyle.Render("→ Tags: ")
	}

	tagList := m.tags.View()

	return label + tagList + " " + HelpStyle.Render("(enter to add, backspace to remove)")
}

// renderOutputFieldAlt renders the optional output textarea (index 1)
func renderOutputFieldAlt(m *Model) string {
	label := LabelStyle.Render("Paste Output:")
	if m.focusIndex == 1 {
		label = LabelStyle.Render("→ Paste Output:")
	}

	var borderStyle lipgloss.Style
	if m.focusIndex == 1 {
		borderStyle = FocusedBorderStyle
	} else {
		borderStyle = BorderStyle
	}

	content := m.output.View()

	inputBox := borderStyle.
		Width(m.width - 8).
		Padding(1).
		Render(content)

	return label + " " + HelpStyle.Render("(optional)") + "\n" + inputBox
}

// renderFooter renders the bottom help bar
func renderFooter(width int) string {
	shortcuts := []string{
		"[Tab] next field",
		"[Ctrl+S] save",
		"[Esc] cancel",
		"[?] help",
	}

	footerContent := strings.Join(shortcuts, "  ")

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Subtext)).
		Width(width - 2).
		Padding(0, 2)

	// separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Surface)).
		Width(width).
		Render(strings.Repeat("─", width))

	return separator + "\n" + footerStyle.Render(footerContent)
}
