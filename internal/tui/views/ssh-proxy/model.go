package sshproxy

import (
	"fmt"
	"ssh-messer/internal/proxy"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/views"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SSHProxyView struct {
	views.BaseView
	UIState             *types.UIState
	AppState            *types.AppState
	SSHClientResultChan chan proxy.SSHClientResultChan
	SSHProcessChan      chan proxy.SSHProcessChan
	ProxyRequestChan    chan proxy.ProxyRequestResult
}

func NewSSHProxyView(uiState types.UIState, appState types.AppState) *SSHProxyView {
	return &SSHProxyView{
		BaseView:            views.BaseView{Type: types.SSHProxyView},
		UIState:             &uiState,
		AppState:            &appState,
		SSHClientResultChan: make(chan proxy.SSHClientResultChan),
		SSHProcessChan:      make(chan proxy.SSHProcessChan),
		ProxyRequestChan:    make(chan proxy.ProxyRequestResult),
	}
}

func (v *SSHProxyView) Init() tea.Cmd {
	configs := v.AppState.GetConfigs()
	currentConfigName := v.AppState.GetCurrentConfigName()

	if config, exists := configs[currentConfigName]; exists {
		go proxy.AsyncCreateSSHHopsClient(config.SSHHops, v.SSHClientResultChan, &v.SSHProcessChan)
	}

	return tea.Batch(
		tickSSHConnectionAnimation(),
		waitForSSHClientResult(v.SSHClientResultChan),
		waitForSSHProcessResult(v.SSHProcessChan),
		waitForProxyRequestResult(v.ProxyRequestChan),
	)
}

func (v *SSHProxyView) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SSHConnectionTick:
		return v.handleSSHConnectionTick()
	case messages.SSHClientResult:
		return v.handleSSHClientResult(msg)
	case messages.SSHProcessResult:
		return v.handleSSHProcessResult(msg)
	case messages.ProxyRequestResult:
		return v.handleProxyRequestResult(msg)
	default:
		return v, nil
	}
}

func (v *SSHProxyView) View() string {
	return RenderSSHProxyView(v)
}

func (v *SSHProxyView) Cleanup() tea.Cmd {
	if v.SSHClientResultChan != nil {
		close(v.SSHClientResultChan)
		v.SSHClientResultChan = nil
	}
	if v.SSHProcessChan != nil {
		close(v.SSHProcessChan)
		v.SSHProcessChan = nil
	}
	if v.ProxyRequestChan != nil {
		close(v.ProxyRequestChan)
		v.ProxyRequestChan = nil
	}
	return nil
}

func (v *SSHProxyView) handleSSHConnectionTick() (views.View, tea.Cmd) {
	configName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(configName)

	if sshInfo.SSHConnectionState == messages.Connecting {
		sshInfo.SSHConnectionProcess++
		v.AppState.SetSSHInfo(configName, sshInfo)
		// 只返回动画 tick，不处理消息监听
		return v, tickSSHConnectionAnimation()
	}

	// 连接完成后，停止动画 tick
	return v, nil
}

func (v *SSHProxyView) handleSSHClientResult(msg messages.SSHClientResult) (views.View, tea.Cmd) {
	configName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(configName)
	currentConfig := v.AppState.GetConfigs()[configName]
	// 直接使用类型安全的结果
	result := msg.Result
	if result.Error != nil {
		v.AppState.Error = messages.AppError{Error: &result.Error, IsFatal: false}

		// 处理错误
		sshInfo.SSHConnectionState = messages.Error
		sshInfo.CurrentInfo = ""
		v.AppState.SetSSHInfo(configName, sshInfo)
		return v, nil
	}

	// 更新 SSH 客户端
	sshInfo.SSHClient = result.Client
	sshInfo.SSHServicesReverseProxy = proxy.NewtHttpServiceProxyServer(*currentConfig.LocalHttpPort, currentConfig.Services, result.Client)
	sshInfo.SSHConnectionState = messages.Connected
	sshInfo.CurrentInfo = ""
	v.AppState.SetSSHInfo(configName, sshInfo)

	return v, startSSHServicesReverseProxy(v)
}

func (v *SSHProxyView) handleSSHProcessResult(msg messages.SSHProcessResult) (views.View, tea.Cmd) {
	configName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(configName)

	result := msg.Result
	if result.Error != nil {
		v.AppState.Error = messages.AppError{Error: &result.Error, IsFatal: false}
		sshInfo.SSHConnectionState = messages.Error
		sshInfo.CurrentInfo = ""
		v.AppState.SetSSHInfo(configName, sshInfo)
		return v, nil
	}

	sshInfo.SSHConnectionProcess = result.CompletedHopsCount
	sshInfo.CurrentInfo = result.Message
	v.AppState.SetSSHInfo(configName, sshInfo)

	if result.CompletedHopsCount >= result.TotalHopsCount {
		sshInfo.SSHConnectionState = messages.Connected
		sshInfo.CurrentInfo = ""
		v.AppState.SetSSHInfo(configName, sshInfo)
	}

	return v, waitForSSHProcessResult(v.SSHProcessChan)
}

func tickSSHConnectionAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return messages.SSHConnectionTick{}
	})
}

func waitForSSHClientResult(ch chan proxy.SSHClientResultChan) tea.Cmd {
	return func() tea.Msg {
		result := <-ch // 阻塞等待
		return messages.SSHClientResult{Result: result}
	}
}

func waitForSSHProcessResult(ch chan proxy.SSHProcessChan) tea.Cmd {
	return func() tea.Msg {
		result := <-ch // 阻塞等待
		return messages.SSHProcessResult{Result: result}
	}
}

func waitForProxyRequestResult(ch chan proxy.ProxyRequestResult) tea.Cmd {
	return func() tea.Msg {
		result := <-ch // 阻塞等待
		return messages.ProxyRequestResult{Result: result}
	}
}

func startSSHServicesReverseProxy(v *SSHProxyView) tea.Cmd {
	go func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		// 检查 SSH 连接状态
		configName := v.AppState.GetCurrentConfigName()
		sshInfo := v.AppState.GetSSHInfo(configName)

		if sshInfo.SSHClient == nil {
			return
		}

		v.AppState.GetSSHInfo(v.AppState.GetCurrentConfigName()).SSHServicesReverseProxy.AsyncStart(v.ProxyRequestChan)
	}()
	return nil
}

func (v *SSHProxyView) handleProxyRequestResult(msg messages.ProxyRequestResult) (views.View, tea.Cmd) {
	configName := v.AppState.GetCurrentConfigName()
	sshInfo := v.AppState.GetSSHInfo(configName)

	result := msg.Result
	logEntry := fmt.Sprintf("[%s] %s %s - %d (%.2fs)",
		result.StartTime.Format("15:04:05"),
		result.Method,
		result.URL,
		result.StatusCode,
		result.Duration.Seconds())

	sshInfo.HTTPProxyLogs = append(sshInfo.HTTPProxyLogs, logEntry)

	if len(sshInfo.HTTPProxyLogs) > 100 {
		sshInfo.HTTPProxyLogs = sshInfo.HTTPProxyLogs[len(sshInfo.HTTPProxyLogs)-100:]
	}

	v.AppState.SetSSHInfo(configName, sshInfo)

	return v, waitForProxyRequestResult(v.ProxyRequestChan)
}
