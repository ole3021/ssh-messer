package sshproxy

import (
	"fmt"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/views/styles"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderSSHProxyView 渲染 SSH 代理视图
func RenderSSHProxyView(state *models.SSHProxyViewState, ui *models.UIState, app *models.AppState) string {
	width, height := ui.Width, ui.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// 组合主内容和状态栏
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSSHProxyContent(state, ui, app),
		renderStatusBar(state, ui, app),
	)
}

// renderSSHProxyContent 渲染 SSH 代理主内容
func renderSSHProxyContent(state *models.SSHProxyViewState, ui *models.UIState, app *models.AppState) string {
	width, height := ui.Width, ui.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	contentHeight := height - 1

	currentConfigName := app.GetCurrentConfig()
	sshInfo := app.GetSSHInfo(currentConfigName)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render(currentConfigName),
		"",
		styles.ItemStyle.Render("当前状态:"),
		styles.ItemStyle.Render(getSSHStateString(models.SSHConnectState(sshInfo.SSHConnectionState))),
		"",
		styles.ItemStyle.Render("HTTP Proxy Logs:"),
		styles.ItemStyle.Render(strings.Join(sshInfo.HTTPProxyLogs, "\n")),
		"",
		styles.ItemStyle.Render("Docker Proxy Logs:"),
		styles.ItemStyle.Render(strings.Join(sshInfo.DockerProxyLogs, "\n")),
	)

	return lipgloss.Place(width, contentHeight, lipgloss.Left, lipgloss.Top, content)
}

// renderStatusBar 渲染状态栏
func renderStatusBar(state *models.SSHProxyViewState, ui *models.UIState, app *models.AppState) string {
	width := ui.Width
	if width == 0 {
		width = 80
	}

	// 获取 SSH 状态
	currentConfigName := app.GetCurrentConfig()
	sshInfo := app.GetSSHInfo(currentConfigName)
	sshState := getSSHStateString(models.SSHConnectState(sshInfo.SSHConnectionState))

	// 根据状态选择样式和图标
	var sshStatusStyle lipgloss.Style
	var statusIcon string

	switch sshState {
	case "Connected":
		sshStatusStyle = styles.SSHStatusConnectedStyle
		statusIcon = "●"
	case "Connecting":
		sshStatusStyle = styles.SSHStatusConnectingStyle
		icons := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		index := sshInfo.SSHConnectionProcess % 10
		statusIcon = fmt.Sprintf(" %s", icons[index])
	case "Disconnected":
		sshStatusStyle = styles.SSHStatusDisconnectedStyle
		statusIcon = "○"
	case "Error":
		sshStatusStyle = styles.SSHStatusErrorStyle
		statusIcon = "!"
	default:
		sshStatusStyle = styles.SSHStatusDisconnectedStyle
		statusIcon = "?"
	}

	// 左侧 SSH 状态
	leftStatus := sshStatusStyle.Render(fmt.Sprintf("%s %s", statusIcon, sshState))

	// 右侧配置名称
	rightStatus := styles.ConfigNameStyle.Render(fmt.Sprintf(" %s < ", currentConfigName))

	// 计算中间填充空间
	leftWidth := lipgloss.Width(leftStatus)
	rightWidth := lipgloss.Width(rightStatus)
	gap := width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	// 中间区域显示当前信息
	currentInfo := sshInfo.CurrentInfo

	// 限制信息长度以适应状态栏，保留一些边距
	maxLength := gap - 4
	if maxLength < 0 {
		maxLength = 0
	}
	if len(currentInfo) > maxLength {
		if maxLength > 3 {
			currentInfo = currentInfo[:maxLength-3] + "..."
		} else {
			currentInfo = "..."
		}
	}

	middleSpace := styles.StatusBarStyle.
		Width(gap).
		Render(fmt.Sprintf(" %s ", currentInfo))

	// 组合状态栏
	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStatus,
		middleSpace,
		rightStatus,
	)

	return styles.StatusBarStyle.Width(width).Render(statusBar)
}

// getSSHStateString 获取 SSH 状态字符串
func getSSHStateString(state models.SSHConnectState) string {
	switch state {
	case models.Disconnected:
		return "Disconnected"
	case models.Connecting:
		return "Connecting"
	case models.Connected:
		return "Connected"
	case models.Error:
		return "Error"
	default:
		return "Unknown"
	}
}
