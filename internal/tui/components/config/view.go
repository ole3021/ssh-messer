package config

import (
	"fmt"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/views/styles"

	"github.com/charmbracelet/lipgloss"
)

// renderPopup 渲染弹窗模式
func renderPopup(state *models.ConfigViewState, ui *models.UIState) string {
	choices := state.ConfigNames
	cursor := state.Cursor
	width := ui.Width

	if len(choices) == 0 {
		return styles.ItemStyle.Render("正在加载配置...")
	}

	// 构建选项
	var options []string
	for i, choice := range choices {
		style := styles.ItemStyle
		if cursor == i {
			style = styles.SelectedStyle
		}

		cursorChar := " "
		if cursor == i {
			cursorChar = ">"
		}

		option := style.Render(fmt.Sprintf("%s %s", cursorChar, choice))
		options = append(options, option)
	}

	content := lipgloss.Place(
		width/2, 10,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width/2, 3, lipgloss.Center, lipgloss.Center, styles.TitleStyle.Render("Select Proxy Configuration")),
			lipgloss.Place(width/2, len(options)+4, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, options...)),
		),
		lipgloss.WithWhitespaceBackground(styles.Bg),
	)

	// 应用弹窗样式
	return styles.ConfigPopupStyle.Render(content)
}

// renderFullscreen 渲染全屏模式
func renderFullscreen(state *models.ConfigViewState, ui *models.UIState) string {
	choices := state.ConfigNames
	cursor := state.Cursor
	width, height := ui.Width, ui.Height

	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	if len(choices) == 0 {
		return lipgloss.Place(
			width, height,
			lipgloss.Center, lipgloss.Center,
			styles.ItemStyle.Render("正在加载配置..."),
			lipgloss.WithWhitespaceBackground(styles.Bg),
		)
	}

	// 构建选项
	var options []string
	for i, choice := range choices {
		style := styles.ItemStyle
		if cursor == i {
			style = styles.SelectedStyle
		}

		cursorChar := " "
		if cursor == i {
			cursorChar = ">"
		}

		option := style.Render(fmt.Sprintf("%s %s", cursorChar, choice))
		options = append(options, option)
	}

	// 全屏布局
	content := lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width, 3, lipgloss.Center, lipgloss.Center, styles.TitleStyle.Render("Select Proxy Configuration")),
			lipgloss.Place(width, len(options)+4, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, options...)),
			lipgloss.Place(width, 2, lipgloss.Center, lipgloss.Center, styles.MetaStyle.Render("Press Enter to select, Esc to cancel")),
		),
		lipgloss.WithWhitespaceBackground(styles.Bg),
	)

	return content
}

// renderInline 渲染内联模式
func renderInline(state *models.ConfigViewState, ui *models.UIState) string {
	choices := state.ConfigNames
	cursor := state.Cursor

	if len(choices) == 0 {
		return styles.ItemStyle.Render("正在加载配置...")
	}

	// 构建选项
	var options []string
	for i, choice := range choices {
		style := styles.ItemStyle
		if cursor == i {
			style = styles.SelectedStyle
		}

		cursorChar := " "
		if cursor == i {
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
