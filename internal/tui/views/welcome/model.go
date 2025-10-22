package welcome

import (
	"ssh-messer/internal/tui/components/config"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/views"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// WelcomeView 欢迎视图
type WelcomeView struct {
	views.BaseView
	UIState                  *types.UIState
	AppState                 *types.AppState
	WelcomeAnimationProgress int
	ConfigComponent          *config.ConfigComponent
}

// NewWelcomeView 创建新的欢迎视图
func NewWelcomeView(uiState *types.UIState, appState *types.AppState) *WelcomeView {
	return &WelcomeView{
		BaseView:                 views.BaseView{Type: types.WelcomeView},
		UIState:                  uiState,
		AppState:                 appState,
		WelcomeAnimationProgress: 0,
		ConfigComponent:          config.NewConfigComponent(uiState, appState, config.RenderModeInline),
	}
}

// Init 初始化欢迎视图
func (v *WelcomeView) Init() tea.Cmd {
	return tickWelcomeAnimation()
}

// Update 处理消息并更新视图状态
func (v *WelcomeView) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg.(type) {
	case messages.WelcomeTick:
		return v.handleWelcomeTick()
	}

	// 将所有其他消息委托给 ConfigComponent
	newComp, cmd := v.ConfigComponent.Update(msg)
	v.ConfigComponent = newComp.(*config.ConfigComponent)
	return v, cmd
}

// View 渲染欢迎视图
func (v *WelcomeView) View() string {
	return RenderWelcomeView(v.WelcomeAnimationProgress, v.ConfigComponent, v.UIState)
}

// handleWelcomeTick 处理欢迎动画 tick
func (v *WelcomeView) handleWelcomeTick() (views.View, tea.Cmd) {
	v.WelcomeAnimationProgress += 3
	if v.WelcomeAnimationProgress >= 100 {
		return v, nil
	}
	return v, tickWelcomeAnimation()
}

// tickWelcomeAnimation 创建欢迎动画 tick 命令
func tickWelcomeAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return messages.WelcomeTick{}
	})
}
