package app

import (
	"ssh-messer/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// handleGlobalMsg 处理全局消息
func (m *AppModel) handleGlobalMsg(msg tea.Msg) (bool, tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case tea.KeyMsg:
		return m.handleGlobalKeyPressMsg(msg)
	case messages.AppError:
		return m.handleAppErrorMsg(msg)
	default:
		return false, m, nil
	}
}

// handleGlobalKeyPress 处理全局键盘输入
func (m *AppModel) handleGlobalKeyPressMsg(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "q":
		// 执行清理操作
		if m.CurrentView != nil {
			m.CurrentView.Cleanup()
		}
		return true, m, tea.Quit
	default:
		return false, m, nil
	}
}

// handleWindowSize 处理窗口大小变化
func (m *AppModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (bool, tea.Model, tea.Cmd) {
	m.UIState.Width = msg.Width
	m.UIState.Height = msg.Height
	return true, m, nil
}

// handleAppError 处理应用错误
func (m *AppModel) handleAppErrorMsg(appErr messages.AppError) (bool, tea.Model, tea.Cmd) {
	m.AppState.Error = appErr
	return true, m, nil
}
