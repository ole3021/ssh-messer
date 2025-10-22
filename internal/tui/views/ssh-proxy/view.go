package sshproxy

import (
	"fmt"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/views/styles"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderSSHProxyView(v *SSHProxyView) string {
	width, height := v.UIState.Width, v.UIState.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// 组合主内容和状态栏
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSSHProxyContent(v),
		renderStatusBar(v),
	)
}

// renderSSHProxyContent 渲染 SSH 代理主内容
func renderSSHProxyContent(v *SSHProxyView) string {
	currentConfigName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(currentConfigName)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render(currentConfigName),
		"",
		styles.ItemStyle.Render("当前状态:"),
		styles.ItemStyle.Render(getSSHStateString(sshInfo.SSHConnectionState)),
		"",
		styles.ItemStyle.Render("HTTP Proxy Logs:"),
		styles.ItemStyle.Render(strings.Join(sshInfo.HTTPProxyLogs, "\n")),
		"",
		styles.ItemStyle.Render("Docker Proxy Logs:"),
		styles.ItemStyle.Render(strings.Join(sshInfo.DockerProxyLogs, "\n")),
	)

	return lipgloss.Place(v.UIState.Width, v.UIState.Height, lipgloss.Left, lipgloss.Top, content)
}

// renderStatusBar 渲染状态栏
func renderStatusBar(v *SSHProxyView) string {
	// 获取 SSH 状态
	currentConfigName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(currentConfigName)
	connectionState := sshInfo.SSHConnectionState

	// 根据状态选择样式和图标

	var sshStatusStyle lipgloss.Style
	var statusIcon string

	switch connectionState {
	case messages.Connected:
		sshStatusStyle = styles.SSHStatusConnectedStyle
		statusIcon = "●"
	case messages.Connecting:
		sshStatusStyle = styles.SSHStatusConnectingStyle
		icons := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		index := sshInfo.SSHConnectionProcess % 10
		statusIcon = fmt.Sprintf(" %s", icons[index])
	case messages.Disconnected:
		sshStatusStyle = styles.SSHStatusDisconnectedStyle
		statusIcon = "○"
	case messages.Error:
		sshStatusStyle = styles.SSHStatusErrorStyle
		statusIcon = "!"
	default:
		sshStatusStyle = styles.SSHStatusDisconnectedStyle
		statusIcon = "?"
	}

	// 左侧 SSH 状态
	leftStatus := sshStatusStyle.Render(fmt.Sprintf("%s %s", statusIcon, getSSHStateString(connectionState)))

	// 右侧配置名称
	rightStatus := styles.ConfigNameStyle.Render(fmt.Sprintf(" %s < ", currentConfigName))

	// 计算中间填充空间
	leftWidth := lipgloss.Width(leftStatus)
	rightWidth := lipgloss.Width(rightStatus)
	gap := v.UIState.Width - leftWidth - rightWidth
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

	return styles.StatusBarStyle.Width(v.UIState.Width).Render(statusBar)
}

// getSSHStateString 获取 SSH 状态字符串
func getSSHStateString(state messages.SSHConnectState) string {
	switch state {
	case messages.Disconnected:
		return "Disconnected"
	case messages.Connecting:
		return "Connecting"
	case messages.Connected:
		return "Connected"
	case messages.Error:
		return "Error"
	default:
		return "Unknown"
	}
}
