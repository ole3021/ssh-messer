package commands

import (
	"context"
	"ssh-messer/internal/config"
	"ssh-messer/internal/messer"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/util"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// TODO: Clean up code
func ConnectMesserHopCmd(config *config.MesserConfig, eventsCtx context.Context) tea.Cmd {
	return func() tea.Msg {
		pkg.Logger.Trace().Str("configName", config.Name).Msg("[Commands] ConnectMesserHopCmd Start")
		// TODO: Add Health Check
		// config, exists := appState.GetConfigs()[configName]
		// if !exists {
		// 	return messages.AppErrMsg{
		// 		Error:   messages.ErrConfigNotFound,
		// 		IsFatal: false,
		// 	}
		// }

		// if existingProxy := appState.GetSSHProxy(configName); existingProxy != nil {
		// 	pkg.Logger.Warn().Str("configName", configName).Msg("[InitSSHProxy] SSH proxy already exists")
		// 	return nil
		// }

		// 从配置读取健康检查间隔，如果未配置则使用默认值 30 秒
		// healthCheckInterval := 30 * time.Second
		// if config.HealthCheckIntervalSecs != 0 {
		// 	healthCheckInterval = time.Duration(config.HealthCheckIntervalSecs) * time.Second
		// }

		// 从配置获取 services 和 localPort
		configName := config.Name
		services := config.ReverseServices
		localPort := ""
		if config.LocalHttpPort != "" {
			localPort = config.LocalHttpPort
		}

		// 如果配置了 local_http_port 且有 services，在启动前检查端口是否可用
		if localPort != "" && len(services) > 0 {
			if err := util.CheckPortAvailable(localPort); err != nil {
				pkg.Logger.Error().Err(err).Str("configName", configName).Str("port", localPort).Msg("[InitSSHProxy] 端口检查失败")
				return messages.AppErrMsg{
					Error:   err,
					IsFatal: true,
				}
			}
			pkg.Logger.Info().Str("configName", configName).Str("port", localPort).Msg("[InitSSHProxy] 端口检查通过")
		}

		sshProxy := messer.NewMesserHops(config, eventsCtx)
		sshProxy.ConnectHops()
		pkg.Logger.Info().Str("configName", configName).Msg("[Commands] ConnectMesserHopCmd Completed")

		return nil
	}
}
