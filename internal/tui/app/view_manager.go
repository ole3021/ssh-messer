package app

import (
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"
	sshproxy "ssh-messer/internal/tui/views/ssh-proxy"
	"ssh-messer/internal/tui/views/welcome"

	tea "github.com/charmbracelet/bubbletea"
)

// handleViewChange 处理视图切换
func (m *AppModel) handleViewChangeMsg(msg messages.ViewChangeMsg) (tea.Model, tea.Cmd) {
	// 清理当前视图
	if m.CurrentView != nil {
		cleanupCmd := m.CurrentView.Cleanup()
		if cleanupCmd != nil {
			// 执行清理命令
			cleanupCmd()
		}
	}

	// 切换到新视图
	switch types.ViewEnum(msg.TargetView) {
	case types.SSHProxyView:
		m.CurrentView = sshproxy.NewSSHProxyView(*m.UIState, *m.AppState)
		return m, m.CurrentView.Init()
	case types.WelcomeView:
		m.CurrentView = welcome.NewWelcomeView(m.UIState, m.AppState)
		return m, m.CurrentView.Init()
	default:
		return m, nil
	}
}
