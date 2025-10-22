package config

import (
	"fmt"
	"ssh-messer/internal/tui/views/styles"

	"github.com/charmbracelet/lipgloss"
)

// renderPopup 渲染弹窗模式
func renderPopup(m *ConfigComponent) string {
	if len(m.ConfigNames) == 0 {
		return styles.ItemStyle.Render("正在加载配置...")
	}

	// 构建选项
	var options []string
	for i, choice := range m.ConfigNames {
		style := styles.ItemStyle
		if m.Cursor == i {
			style = styles.SelectedStyle
		}

		cursorChar := " "
		if m.Cursor == i {
			cursorChar = ">"
		}

		option := style.Render(fmt.Sprintf("%s %s", cursorChar, choice))
		options = append(options, option)
	}

	content := lipgloss.Place(
		m.UIState.Width/2, 10,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(m.UIState.Width/2, 3, lipgloss.Center, lipgloss.Center, styles.TitleStyle.Render("Select Proxy Configuration")),
			lipgloss.Place(m.UIState.Width/2, len(options)+4, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, options...)),
		),
		lipgloss.WithWhitespaceBackground(styles.Bg),
	)

	// 应用弹窗样式
	return styles.ConfigPopupStyle.Render(content)
}

// renderInline 渲染内联模式
func renderInline(m *ConfigComponent) string {

	if len(m.ConfigNames) == 0 {
		return styles.ItemStyle.Render("正在加载配置...")
	}

	// 构建选项
	var options []string
	for i, choice := range m.ConfigNames {
		style := styles.ItemStyle
		if m.Cursor == i {
			style = styles.SelectedStyle
		}

		cursorChar := " "
		if m.Cursor == i {
			cursorChar = ">"
		}

		option := style.Render(fmt.Sprintf("%s %s", cursorChar, choice))
		options = append(options, option)
	}

	// 内联布局，不包含弹窗样式
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.TitleStyle.Render("Available Configurations:"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, options...),
	)
}
