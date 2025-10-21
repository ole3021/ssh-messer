package welcome

import (
	"ssh-messer/internal/tui/components/config"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/views"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// WelcomeView 欢迎视图
type WelcomeView struct {
	views.BaseView
	State           *models.WelcomeViewState
	ConfigComponent *config.ConfigComponent
	UI              *models.UIState
	App             *models.AppState
}

// NewWelcomeView 创建新的欢迎视图
func NewWelcomeView(uiState *models.UIState, appState *models.AppState) *WelcomeView {
	return &WelcomeView{
		BaseView:        views.BaseView{Type: models.WelcomeView},
		State:           models.NewWelcomeViewState(),
		ConfigComponent: config.NewConfigComponent(uiState, config.RenderModePopup),
		UI:              uiState,
		App:             appState,
	}
}

// Init 初始化欢迎视图
func (v *WelcomeView) Init() tea.Cmd {
	return tickWelcomeAnimation()
}

// Update 处理消息并更新视图状态
func (v *WelcomeView) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.WelcomeTick:
		return v.handleWelcomeTick()
	case messages.ConfigSelected:
		return v.handleConfigSelected(msg)
	case messages.ConfigCancelled:
		return v.handleConfigCancelled()
	}

	// 将所有其他消息委托给 ConfigComponent
	newComp, cmd := v.ConfigComponent.Update(msg)
	v.ConfigComponent = newComp.(*config.ConfigComponent)
	return v, cmd
}

// View 渲染欢迎视图
func (v *WelcomeView) View() string {
	return RenderWelcomeView(v.State, v.ConfigComponent, v.UI)
}

// handleWelcomeTick 处理欢迎动画 tick
func (v *WelcomeView) handleWelcomeTick() (views.View, tea.Cmd) {
	v.State.WelcomeAnimationProgress += 3
	if v.State.WelcomeAnimationProgress >= 100 {
		return v, nil
	}
	return v, tickWelcomeAnimation()
}

// handleConfigSelected 处理配置选择
func (v *WelcomeView) handleConfigSelected(msg messages.ConfigSelected) (views.View, tea.Cmd) {
	v.App.SetCurrentConfig(msg.ConfigName)

	// 初始化 SSH 信息
	sshInfo := messages.SSHInfo{
		SSHConnectionState:   int(models.Connecting),
		SSHConnectionProcess: 0,
		HTTPProxyLogs:        []string{},
		DockerProxyLogs:      []string{},
	}
	v.App.SetSSHInfo(msg.ConfigName, sshInfo)

	// 发送视图切换消息
	return v, func() tea.Msg {
		return messages.ViewChangeMsg{TargetView: int(models.SSHProxyView)}
	}
}

// handleConfigCancelled 处理配置取消
func (v *WelcomeView) handleConfigCancelled() (views.View, tea.Cmd) {
	// 配置取消，保持在欢迎页面
	return v, nil
}

// tickWelcomeAnimation 创建欢迎动画 tick 命令
func tickWelcomeAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return messages.WelcomeTick{}
	})
}
