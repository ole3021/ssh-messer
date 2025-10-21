package sshproxy

import (
	"ssh-messer/internal/proxy"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/views"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SSHProxyView SSH 代理视图
// 负责处理 SSH 连接、代理状态显示和用户交互
type SSHProxyView struct {
	views.BaseView
	State *models.SSHProxyViewState
	UI    *models.UIState
	App   *models.AppState
}

// NewSSHProxyView 创建新的 SSH 代理视图
func NewSSHProxyView(uiState *models.UIState, appState *models.AppState) *SSHProxyView {
	return &SSHProxyView{
		BaseView: views.BaseView{Type: models.SSHProxyView},
		State:    models.NewSSHProxyViewState(),
		UI:       uiState,
		App:      appState,
	}
}

// Init 初始化 SSH 代理视图
func (v *SSHProxyView) Init() tea.Cmd {
	// 创建 channel（修复全局 channel 问题）
	v.State.SSHClientChan = make(chan proxy.SSHClientResultChan)
	v.State.SSHProcessChan = make(chan proxy.SSHProcessChan)

	// 初始化当前信息（通过 SSH 信息状态管理）

	// 启动 SSH 连接
	configs := v.App.GetConfigs()
	currentConfig := v.App.GetCurrentConfig()

	if config, exists := configs[currentConfig]; exists {
		// 启动异步连接
		go proxy.AsyncCreateSSHHopsClient(config.SSHHops, v.State.SSHClientChan, &v.State.SSHProcessChan)
	}

	return tea.Batch(
		tickSSHConnectionAnimation(),
		listenToSSHChannels(v.State.SSHClientChan, v.State.SSHProcessChan),
	)
}

// Update 处理消息并更新视图状态
func (v *SSHProxyView) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SSHConnectionTick:
		return v.handleSSHConnectionTick()
	case messages.SSHClientResult:
		return v.handleSSHClientResult(msg)
	case messages.SSHProcessResult:
		return v.handleSSHProcessResult(msg)
	case tea.KeyMsg:
		return v.handleKeyPress(msg)
	default:
		return v, nil
	}
}

// View 渲染 SSH 代理视图
func (v *SSHProxyView) View() string {
	return RenderSSHProxyView(v.State, v.UI, v.App)
}

// Cleanup 清理资源
func (v *SSHProxyView) Cleanup() tea.Cmd {
	if v.State.SSHClientChan != nil {
		close(v.State.SSHClientChan)
		v.State.SSHClientChan = nil
	}
	if v.State.SSHProcessChan != nil {
		close(v.State.SSHProcessChan)
		v.State.SSHProcessChan = nil
	}
	return nil
}

// handleSSHConnectionTick 处理 SSH 连接动画 tick
func (v *SSHProxyView) handleSSHConnectionTick() (views.View, tea.Cmd) {
	configName := v.App.GetCurrentConfig()
	sshInfo := v.App.GetSSHInfo(configName)

	if models.SSHConnectState(sshInfo.SSHConnectionState) == models.Connecting {
		sshInfo.SSHConnectionProcess++
		v.App.SetSSHInfo(configName, sshInfo)
		return v, tickSSHConnectionAnimation()
	}

	return v, nil
}

// handleSSHClientResult 处理 SSH 客户端结果
func (v *SSHProxyView) handleSSHClientResult(msg messages.SSHClientResult) (views.View, tea.Cmd) {
	configName := v.App.GetCurrentConfig()
	sshInfo := v.App.GetSSHInfo(configName)

	// 直接使用类型安全的结果
	result := msg.Result
	if result.Error != nil {
		v.App.Error = messages.AppError{Error: &result.Error, IsFatal: false}

		// 处理错误
		sshInfo.SSHConnectionState = int(models.Error)
		sshInfo.CurrentInfo = ""
		v.App.SetSSHInfo(configName, sshInfo)
		return v, nil
	}

	// 更新 SSH 客户端
	sshInfo.SSHClient = result.Client
	sshInfo.SSHConnectionState = int(models.Connected)
	sshInfo.CurrentInfo = ""
	v.App.SetSSHInfo(configName, sshInfo)

	return v, listenToSSHChannels(v.State.SSHClientChan, v.State.SSHProcessChan) // 继续监听
}

// handleSSHProcessResult 处理 SSH 进程结果
func (v *SSHProxyView) handleSSHProcessResult(msg messages.SSHProcessResult) (views.View, tea.Cmd) {
	configName := v.App.GetCurrentConfig()
	sshInfo := v.App.GetSSHInfo(configName)

	// 直接使用类型安全的结果
	result := msg.Result
	if result.Error != nil {
		v.App.Error = messages.AppError{Error: &result.Error, IsFatal: false}
		sshInfo.SSHConnectionState = int(models.Error)
		sshInfo.CurrentInfo = ""
		v.App.SetSSHInfo(configName, sshInfo)
		return v, nil
	}

	// 更新连接进度和当前信息
	sshInfo.SSHConnectionProcess = result.CompletedHopsCount
	sshInfo.CurrentInfo = result.Message // 显示连接过程中的消息
	v.App.SetSSHInfo(configName, sshInfo)

	// 如果所有跳都完成了，更新状态
	if result.CompletedHopsCount >= result.TotalHopsCount {
		sshInfo.SSHConnectionState = int(models.Connected)
		sshInfo.CurrentInfo = ""
		v.App.SetSSHInfo(configName, sshInfo)
	}

	return v, listenToSSHChannels(v.State.SSHClientChan, v.State.SSHProcessChan) // 继续监听
}

// handleKeyPress 处理键盘输入
func (v *SSHProxyView) handleKeyPress(msg tea.KeyMsg) (views.View, tea.Cmd) {
	// SSH 代理视图特定的键盘处理
	// 全局快捷键（如 ctrl+c, q）由 App 层处理
	switch msg.String() {
	case "r":
		// 重新连接
		return v, nil
	case "s":
		// 显示状态
		return v, nil
	default:
		return v, nil
	}
}

// tickSSHConnectionAnimation 创建 SSH 连接动画 tick 命令
func tickSSHConnectionAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return messages.SSHConnectionTick{}
	})
}

// listenToSSHChannels 监听 SSH channel
func listenToSSHChannels(sshClientChan chan proxy.SSHClientResultChan, sshProcessChan chan proxy.SSHProcessChan) tea.Cmd {
	return func() tea.Msg {
		select {
		case result := <-sshClientChan:
			return messages.SSHClientResult{Result: result}
		case result := <-sshProcessChan:
			return messages.SSHProcessResult{Result: result}
		}
	}
}
