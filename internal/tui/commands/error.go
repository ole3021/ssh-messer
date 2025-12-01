package commands

import (
	"ssh-messer/internal/tui/messages"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func ReportAppErrCmd(err error, isFatal bool) tea.Cmd {
	pkg.Logger.Error().Err(err).Msg("[commands::ReportAppErrCmd] Report app error")
	return func() tea.Msg {
		return messages.AppErrMsg{
			Error:   err,
			IsFatal: isFatal,
		}
	}
}
