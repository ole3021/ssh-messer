package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type MesserRenderer interface {
	GetCurrentConfigName() string
	GetSSHConnectionState() string
	GetConfigChoices() []string
	GetHTTPProxyLogs() []string
	GetDockerProxyLogs() []string
	GetConnectionProcess() int
	GetCurrentInfo() string
	GetWidth() int
	GetHeight() int
}

var (
	MainPopupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border)
)

func RenderMesserView(renderer MesserRenderer) string {
	width, height := renderer.GetWidth(), renderer.GetHeight()
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// 组合主内容和状态栏
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderMesserContent(renderer),
		renderStatusBar(renderer),
	)
}

func renderMesserContent(renderer MesserRenderer) string {
	width, height := renderer.GetWidth(), renderer.GetHeight()
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	contentHeight := height - 1

	currentConfigName := renderer.GetCurrentConfigName()
	currentInfo := renderer.GetCurrentInfo()
	httpProxyLogs := renderer.GetHTTPProxyLogs()
	dockerProxyLogs := renderer.GetDockerProxyLogs()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		TitleStyle.Render(currentConfigName),
		"",
		ItemStyle.Render("当前状态:"),
		ItemStyle.Render(currentInfo),
		"",
		ItemStyle.Render("HTTP Proxy Logs:"),
		ItemStyle.Render(strings.Join(httpProxyLogs, "\n")),
		"",
		ItemStyle.Render("Docker Proxy Logs:"),
		ItemStyle.Render(strings.Join(dockerProxyLogs, "\n")),
	)

	return lipgloss.Place(width, contentHeight, lipgloss.Left, lipgloss.Top, content)

}

func renderStatusBar(renderer MesserRenderer) string {
	width := renderer.GetWidth()
	if width == 0 {
		width = 80
	}

	// 获取 SSH 状态
	sshState := renderer.GetSSHConnectionState()
	configName := renderer.GetCurrentConfigName()

	// 根据状态选择样式和图标
	var sshStatusStyle lipgloss.Style
	var statusIcon string

	switch sshState {
	case "Connected":
		sshStatusStyle = SSHStatusConnectedStyle
		statusIcon = "●"
	case "Connecting":
		sshStatusStyle = SSHStatusConnectingStyle
		icons := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		index := renderer.GetConnectionProcess() % 10
		statusIcon = fmt.Sprintf(" %s", icons[index])
	case "Disconnected":
		sshStatusStyle = SSHStatusDisconnectedStyle
		statusIcon = "○"
	case "Error":
		sshStatusStyle = SSHStatusErrorStyle
		statusIcon = "!"
	default:
		sshStatusStyle = SSHStatusDisconnectedStyle
		statusIcon = "?"
	}

	// 左侧 SSH 状态
	leftStatus := sshStatusStyle.Render(fmt.Sprintf("%s %s", statusIcon, sshState))

	// 右侧配置名称
	rightStatus := ConfigNameStyle.Render(fmt.Sprintf(" %s < ", configName))

	// 计算中间填充空间
	leftWidth := lipgloss.Width(leftStatus)
	rightWidth := lipgloss.Width(rightStatus)
	gap := width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	// 中间空白区域
	middleSpace := StatusBarStyle.
		Width(gap).
		Render(fmt.Sprintf(" %s ", renderer.GetCurrentInfo()))

	// 组合状态栏
	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStatus,
		middleSpace,
		rightStatus,
	)

	return StatusBarStyle.Width(width).Render(statusBar)
}
