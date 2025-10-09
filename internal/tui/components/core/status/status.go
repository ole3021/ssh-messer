package status

import (
	"time"

	"ssh-messer/internal/tui/util"

	"github.com/charmbracelet/bubbles/v2/help"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type StatusCmp interface {
	util.Model
	ToggleFullHelp()
	SetKeyMap(keyMap help.KeyMap)
}

type statusCmp struct {
	info       util.InfoMsg
	width      int
	messageTTL time.Duration
	help       help.Model
	keyMap     help.KeyMap
}

// clearMessageCmd is a command that clears status messages after a timeout
func (m *statusCmp) clearMessageCmd(ttl time.Duration) tea.Cmd {
	return tea.Tick(ttl, func(time.Time) tea.Msg {
		return util.ClearStatusMsg{}
	})
}

func (m *statusCmp) Init() tea.Cmd {
	return nil
}

func (m *statusCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.width > 2 {
			m.help.Width = m.width - 2
		}
		return m, nil

	// Handle status info
	case util.InfoMsg:
		m.info = msg
		ttl := msg.TTL
		if ttl == 0 {
			ttl = m.messageTTL
		}
		return m, m.clearMessageCmd(ttl)
	case util.ClearStatusMsg:
		m.info = util.InfoMsg{}
	}
	return m, nil
}

func (m *statusCmp) View() string {
	if m.keyMap == nil {
		return "" // 如果 keyMap 未设置，返回空字符串
	}
	return m.help.View(m.keyMap)
}

func (m *statusCmp) ToggleFullHelp() {
	m.help.ShowAll = !m.help.ShowAll
}

func (m *statusCmp) SetKeyMap(keyMap help.KeyMap) {
	m.keyMap = keyMap
}

func NewStatusCmp() StatusCmp {
	helpModel := help.New()
	helpModel.Width = 80 // 设置默认宽度，避免初始化问题

	return &statusCmp{
		messageTTL: 3 * time.Second,
		help:       helpModel,
		keyMap:     nil, // 初始化为 nil，稍后通过 SetKeyMap 设置
	}
}
