package welcome

import (
	"fmt"
	meta "ssh-messer"
	app_logo "ssh-messer/internal/tui/components/animation"
	"ssh-messer/internal/tui/components/config_list"
	"ssh-messer/internal/tui/styles"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	copyrightCompHeight = 2
)

type WelcomePage interface {
	util.Model
}

type welcomePage struct {
	appState       *types.AppState
	uiState        *types.UIState
	compLogo       app_logo.AppLogoCmp
	compConfigList config_list.ConfigListCmp
}

func New(appState *types.AppState, uiState *types.UIState) WelcomePage {
	return &welcomePage{
		appState:       appState,
		uiState:        uiState,
		compLogo:       app_logo.NewLogo(),
		compConfigList: config_list.New(),
	}
}

func (p *welcomePage) Init() tea.Cmd {
	return tea.Batch(
		p.compLogo.Init(),
		p.compConfigList.Init(),
	)
}

func (p *welcomePage) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		return p, tea.Batch(p.compLogo.SetSize(msg.Width, msg.Height/2-copyrightCompHeight),
			p.compConfigList.SetSize(msg.Width, msg.Height/2))
	}

	updated, cmd := p.compLogo.Update(msg)
	if updatedLogo, ok := updated.(app_logo.AppLogoCmp); ok {
		p.compLogo = updatedLogo
	}
	cmds = append(cmds, cmd)

	updated, cmd = p.compConfigList.Update(msg)
	if updatedList, ok := updated.(config_list.ConfigListCmp); ok {
		p.compConfigList = updatedList
	}
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *welcomePage) View() string {
	compConfigListWidth, compConfigListHeight := p.compConfigList.GetSize()
	compLogoWidth, compLogoHeight := p.compLogo.GetSize()

	logoComponent := lipgloss.NewStyle().
		Width(compLogoWidth).
		Height(compLogoHeight).
		Align(lipgloss.Center, lipgloss.Center).Render(p.compLogo.View())

	configListComponent := lipgloss.NewStyle().
		Width(compConfigListWidth).
		Height(compConfigListHeight).
		Align(lipgloss.Center, lipgloss.Center).Render(p.compConfigList.View())

	copyrightComponent := lipgloss.NewStyle().
		Foreground(styles.Meta).
		Width(p.uiState.Width).
		Height(copyrightCompHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			fmt.Sprintf("© %s %s • %s", meta.CreateYear, meta.Copyright, meta.Version),
			fmt.Sprintf("%s • %s", meta.License, meta.Repository),
		))

	return lipgloss.NewStyle().
		Width(p.uiState.Width).
		Height(p.uiState.Height).
		Align(lipgloss.Center, lipgloss.Top).
		Render(lipgloss.JoinVertical(lipgloss.Center, logoComponent, configListComponent, copyrightComponent))
}
