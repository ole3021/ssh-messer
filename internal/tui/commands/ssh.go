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

		// 如果配置了 local_http_port 且有 services，在启动前检查端口是否可用
		if localPort != "" && len(services) > 0 {
			if err := ssh_proxy.CheckPortAvailable(localPort); err != nil {
				pkg.Logger.Error().Err(err).Str("configName", configName).Str("port", localPort).Msg("[InitSSHProxy] 端口检查失败")
				return messages.AppErrMsg{
					Error:   err,
					IsFatal: true,
				}
			}
			pkg.Logger.Info().Str("configName", configName).Str("port", localPort).Msg("[InitSSHProxy] 端口检查通过")
		}

		sshProxy := ssh_proxy.NewSSHHopsProxy(configName, config.SSHHops, healthCheckInterval, services, localPort)
		appState.SetSSHProxy(configName, sshProxy)

		go sshProxy.Connect()
		return nil
	}
}
