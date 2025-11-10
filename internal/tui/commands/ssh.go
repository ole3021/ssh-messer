package commands

import (
	"time"

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

		// 从配置读取健康检查间隔，如果未配置则使用默认值 30 秒
		healthCheckInterval := 30 * time.Second
		if config.HealthCheckIntervalSecs != nil {
			healthCheckInterval = time.Duration(*config.HealthCheckIntervalSecs) * time.Second
		}

		// 从配置获取 services 和 localPort
		services := config.SSHServices
		localPort := ""
		if config.LocalHttpPort != nil {
			localPort = *config.LocalHttpPort
		}

		sshProxy := ssh_proxy.NewSSHHopsProxy(configName, config.SSHHops, healthCheckInterval, services, localPort)
		appState.SetSSHProxy(configName, sshProxy)

		go sshProxy.Connect()
		return nil
	}
}
