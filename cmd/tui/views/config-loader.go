package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Config popup specific styles
	ConfigPopupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border)
	// Foreground(Text)
)

// ConfigSelectionRenderer defines the interface for config selection rendering
type ConfigSelectionRenderer interface {
	GetConfigChoices() []string
	GetConfigCursor() int
	GetWidth() int
	GetHeight() int
}

// RenderConfigSimpleList renders a clean, modern configuration selection using lipgloss
// This is the main function used by the application
func RenderConfigSimpleList(renderer ConfigSelectionRenderer) string {
	choices := renderer.GetConfigChoices()
	cursor := renderer.GetConfigCursor()
	width := renderer.GetWidth()

	// Build options using lipgloss conditional styling (similar to welcome.go approach)
	var options []string
	for i, choice := range choices {
		// Use lipgloss conditional styling
		style := ItemStyle
		if cursor == i {
			style = SelectedStyle
		}

		cursorChar := " "
		if cursor == i {
			cursorChar = ">"
		}

		// 确保所有选项的文本部分都对齐
		option := style.Render(fmt.Sprintf("%s %s", cursorChar, choice))
		options = append(options, option)
	}

	content := lipgloss.Place(
		width/2, 10,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width/2, 3, lipgloss.Center, lipgloss.Center, TitleStyle.Render("Select Proxy Configuration")),
			lipgloss.Place(width/2, len(options)+4, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, options...)),
		),
		lipgloss.WithWhitespaceBackground(Bg),
	)

	// Apply popup styling
	return ConfigPopupStyle.Render(content)
}
