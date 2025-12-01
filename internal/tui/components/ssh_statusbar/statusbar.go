package ssh_statusbar

import (
	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// StatusBarCmp SSH 状态栏组件接口
type StatusBarCmp interface {
	util.Model
	layout.Sizeable
	SetAppState(appState *types.AppState)
}

// statusBarCmp SSH 状态栏组件实现
type statusBarCmp struct {
	width, height int
	appState      *types.AppState
}

// New 创建新的状态栏组件
func New() StatusBarCmp {
	return &statusBarCmp{
		appState: nil, // 将在 Phase 6 中设置
	}
}

// SetAppState 设置应用状态（将在 Phase 6 中使用）
func (s *statusBarCmp) SetAppState(appState *types.AppState) {
	s.appState = appState
}

func (s *statusBarCmp) Init() tea.Cmd {
	return nil
}

func (s *statusBarCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	}
	return s, nil
}

func (s *statusBarCmp) View() string {
	if s.appState == nil || s.appState.CurrentConfigFileName == "" {
		return "No configuration selected"
	}

	// 优先显示应用错误信息（包括端口检查失败等致命错误）
	if s.appState.Error.Error != nil {
		return "Error: " + s.appState.Error.Error.Error()
	}

	proxy := s.appState.GetMesserHops(s.appState.CurrentConfigFileName)
	if proxy == nil {
		return "SSH Proxy not initialized"
	}

	return proxy.CurrentInfo
}

func (s *statusBarCmp) SetSize(width, height int) tea.Cmd {
	s.width = width
	s.height = height
	return nil
}

func (s *statusBarCmp) GetSize() (int, int) {
	return s.width, s.height
}
