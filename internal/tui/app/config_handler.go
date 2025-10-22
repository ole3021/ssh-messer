package app

import (
	"fmt"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"

	tea "github.com/charmbracelet/bubbletea"
)

// handleLoadConfigsMsg 处理配置加载消息
func (m *AppModel) handleLoadConfigsMsg(loadMsg messages.LoadConfigs) (tea.Model, tea.Cmd) {
	if loadMsg.Err == nil {
		// 更新应用状态中的配置
		m.AppState.SetConfigs(loadMsg.Configs)
		// 调试信息
		if len(loadMsg.Configs) == 0 {
			return m, func() tea.Msg {
				return messages.AppError{Error: &[]error{fmt.Errorf("没有找到配置文件")}[0], IsFatal: false}
			}
		}
	} else {
		return m, func() tea.Msg {
			return messages.AppError{Error: &loadMsg.Err, IsFatal: true}
		}
	}
	return m, nil
}

// handleConfigSelected 处理配置选择
func (m *AppModel) handleConfigSelectedMsg(msg messages.ConfigSelected) (tea.Model, tea.Cmd) {
	m.AppState.SetCurrentConfigName(msg.ConfigName)

	// 初始化 SSH 信息
	sshInfo := messages.SSHInfo{
		SSHConnectionState:   messages.Connecting,
		SSHConnectionProcess: 0,
		HTTPProxyLogs:        []string{},
		DockerProxyLogs:      []string{},
	}
	m.AppState.SetSSHInfo(msg.ConfigName, sshInfo)

	// 发送视图切换消息
	return m, func() tea.Msg {
		return messages.ViewChangeMsg{TargetView: int(types.SSHProxyView)}
	}
}
