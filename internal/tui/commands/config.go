package commands

import (
	"ssh-messer/internal/config"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func LoadAllConfigsCmd() tea.Cmd {
	pkg.Logger.Trace().Msg("[Commands] LoadAllConfigsCmd Start")
	return func() tea.Msg {
		configs, err := config.LoadTomlConfigsFromHomeDir()
		if err != nil {
			return messages.AppErrMsg{
				Error:   err,
				IsFatal: false,
			}
		}

		pkg.Logger.Info().Int("configs_count", len(configs)).Msg("[Commands] LoadAllConfigsCmd Completed")
		return messages.ConfigLoadedMsg{
			Configs: configs,
			Err:     nil,
		}
	}
}
