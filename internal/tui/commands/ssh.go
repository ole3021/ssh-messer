package commands

import (
	"ssh-messer/internal/ssh_proxy"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// InitSSHProxy 初始化 SSH 代理（选择配置后调用）
func InitSSHProxy(appState *types.AppState, configName string) tea.Cmd {
	return func() tea.Msg {
		config, exists := appState.GetConfigs()[configName]
		if !exists {
			return messages.AppErrMsg{
				Error:   messages.ErrConfigNotFound,
				IsFatal: false,
			}
		}

		if existingProxy := appState.GetSSHProxy(configName); existingProxy != nil {
			pkg.Logger.Warn().Str("configName", configName).Msg("[InitSSHProxy] SSH proxy already exists")
			return nil
		}

		sshProxy := ssh_proxy.NewSSHHopsProxy(configName, config.SSHHops)
		appState.SetSSHProxy(configName, sshProxy)

		go sshProxy.Connect()
		return nil
	}
}
