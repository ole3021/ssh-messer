package util

import tea "github.com/charmbracelet/bubbletea/v2"

// Delegate msg to sub model updated sub model and return commands.
func DelegateMsgToSubModel(cmds []tea.Cmd, msg tea.Msg, model *Model) []tea.Cmd {

	updated, cmd := (*model).Update(msg)
	*model = updated

	cmds = append(cmds, cmd)
	return cmds
}
