package ssh_messer

import (
	"fmt"

	"ssh-messer/internal/pubsub"
	"ssh-messer/internal/ssh_proxy"
	"ssh-messer/internal/tui/components/ssh_logs"
	"ssh-messer/internal/tui/components/ssh_sidebar"
	"ssh-messer/internal/tui/components/ssh_statusbar"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

const (
	SideBarWidth      = 35  // 侧边栏宽度
	StatusBarHeight   = 1   // 状态栏高度
	BorderWidth       = 1   // 边框宽度
	LeftRightBorders  = 2   // 左右边框宽度
	TopBottomBorders  = 2   // 上下边框宽度
	CompactModeWidth  = 120 // 紧凑模式宽度阈值
	CompactModeHeight = 30  // 紧凑模式高度阈值
)

type SSHMesserPage interface {
	util.Model
}

type sshMesserPage struct {
	appState *types.AppState
	uiState  *types.UIState

	compact bool

	// Cmponents
	compStatusBar ssh_statusbar.StatusBarCmp
	compSidebar   ssh_sidebar.SidebarCmp
	compLogs      ssh_logs.LogsCmp
}

func New(appState *types.AppState, uiState *types.UIState) SSHMesserPage {
	statusBar := ssh_statusbar.New()
	statusBar.SetAppState(appState)

	return &sshMesserPage{
		appState:      appState,
		uiState:       uiState,
		compStatusBar: statusBar,
		compSidebar:   ssh_sidebar.New(appState),
		compLogs:      ssh_logs.New(appState),
		compact:       false,
	}
}

func (p *sshMesserPage) Init() tea.Cmd {
	return tea.Batch(
		p.compStatusBar.Init(),
		p.compSidebar.Init(),
		p.compLogs.Init(),
	)
}

func (p *sshMesserPage) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.handleCompactMode(msg.Width, msg.Height)

		statusBarHeight := 1
		statusBarWidth := msg.Width
		sidebarHeight := msg.Height - statusBarHeight
		sidebarWidth := SideBarWidth
		var logsHeight int
		var logsWidth int
		if p.compact {
			logsHeight = msg.Height - statusBarHeight
			logsWidth = msg.Width
		} else {
			logsHeight = msg.Height - statusBarHeight
			logsWidth = msg.Width - SideBarWidth
		}

		return p, tea.Batch(p.compLogs.SetSize(logsWidth, logsHeight), p.compSidebar.SetSize(sidebarWidth, sidebarHeight), p.compStatusBar.SetSize(statusBarWidth, statusBarHeight))

	case pubsub.Event[ssh_proxy.SSHStatusUpdate]:
		update := msg.Payload
		p.appState.GetSSHProxy(update.ConfigName).UpdateSSHProxyStatus(update.Status)

		statusText := fmt.Sprintf("[%s] Status: ", update.ConfigName)
		if update.Status.IsConnected {
			statusText += "Connected"
		} else if update.Status.IsConnecting {
			statusText += "Connecting..."
		} else {
			statusText += "Disconnected"
		}
		if update.Status.CurrentInfo != "" {
			statusText += " - " + update.Status.CurrentInfo
		}
		if update.Status.LastError != nil {
			statusText += fmt.Sprintf(" - Error: %v", update.Status.LastError)
		}

		cmd := p.compLogs.AddLog(statusText)
		cmds = append(cmds, cmd)

		s, cmd := p.compStatusBar.Update(msg)
		if updatedStatusBar, ok := s.(ssh_statusbar.StatusBarCmp); ok {
			p.compStatusBar = updatedStatusBar
		}
		cmds = append(cmds, cmd)

		s, cmd = p.compSidebar.Update(msg)
		if updatedSidebar, ok := s.(ssh_sidebar.SidebarCmp); ok {
			p.compSidebar = updatedSidebar
		}
		cmds = append(cmds, cmd)

		s, cmd = p.compLogs.Update(msg)
		if updatedLogs, ok := s.(ssh_logs.LogsCmp); ok {
			p.compLogs = updatedLogs
		}
		cmds = append(cmds, cmd)

		return p, tea.Batch(cmds...)

	default:
		s, cmd := p.compStatusBar.Update(msg)
		if updatedStatusBar, ok := s.(ssh_statusbar.StatusBarCmp); ok {
			p.compStatusBar = updatedStatusBar
		}
		cmds = append(cmds, cmd)

		s, cmd = p.compSidebar.Update(msg)
		if updatedSidebar, ok := s.(ssh_sidebar.SidebarCmp); ok {
			p.compSidebar = updatedSidebar
		}
		cmds = append(cmds, cmd)

		s, cmd = p.compLogs.Update(msg)
		if updatedLogs, ok := s.(ssh_logs.LogsCmp); ok {
			p.compLogs = updatedLogs
		}
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *sshMesserPage) View() string {
	logsWidth, logsHeight := p.compLogs.GetSize()
	sidebarWidth, sidebarHeight := p.compSidebar.GetSize()
	statusBarWidth, statusBarHeight := p.compStatusBar.GetSize()

	logsComponent := lipgloss.NewStyle().
		Width(logsWidth).
		Height(logsHeight).
		Align(lipgloss.Left, lipgloss.Top).
		Render(p.compLogs.View())

	var mainComponent string
	if p.compact {
		mainComponent = logsComponent
	} else {
		sidebarComponent := lipgloss.NewStyle().
			Width(sidebarWidth).
			Height(sidebarHeight).
			Align(lipgloss.Left, lipgloss.Top).
			Render(p.compSidebar.View())
		mainComponent = lipgloss.JoinHorizontal(lipgloss.Left, logsComponent, sidebarComponent)
	}

	statusBarComponent := lipgloss.NewStyle().
		Width(statusBarWidth).
		Height(statusBarHeight).
		Render(p.compStatusBar.View())

	return lipgloss.NewStyle().
		Width(p.uiState.Width).
		Height(p.uiState.Height).
		Render(lipgloss.JoinVertical(lipgloss.Top, mainComponent, statusBarComponent))
}

func (p *sshMesserPage) handleCompactMode(width, height int) {
	if width < CompactModeWidth || height < CompactModeHeight {
		p.compact = true
	} else {
		p.compact = false
	}
}
