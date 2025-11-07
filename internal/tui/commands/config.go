package commands

import (
	"ssh-messer/internal/config_loader"
	"ssh-messer/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// LoadAllConfigs 加载所有配置（系统配置 + TOML 配置）
func LoadAllConfigs() tea.Cmd {
	return func() tea.Msg {
		configs, err := config_loader.LoadTomlConfigsFromHomeDir()
		if err != nil {
			return messages.AppErrMsg{
				Error:   err,
				IsFatal: false,
			}
		}

		return messages.LoadConfigsMsg{
			Configs: configs,
			Err:     nil,
		}
	}
}
