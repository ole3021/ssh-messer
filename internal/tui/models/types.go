package models

import (
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/proxy"
	"ssh-messer/internal/tui/messages"
)

// ViewEnum 视图枚举类型
type ViewEnum int

const (
	WelcomeView ViewEnum = iota
	SSHProxyView
)

// SSHConnectState SSH 连接状态
type SSHConnectState int

const (
	Disconnected SSHConnectState = iota
	Connecting
	Connected
	Error
)

// AppState 应用业务状态
type AppState struct {
	// 配置相关
	Configs           map[string]loaders.TomlConfig
	CurrentConfigName string
	SSHInfos          map[string]messages.SSHInfo
	Error             messages.AppError
}

// UIState 用户界面状态
type UIState struct {
	Width  int
	Height int
}

// WelcomeViewState 欢迎视图状态
type WelcomeViewState struct {
	WelcomeAnimationProgress int
}

// ConfigViewState 配置选择视图状态
type ConfigViewState struct {
	ConfigNames  []string
	Cursor       int
	SelectedName string
	Error        string
}

// SSHProxyViewState SSH 代理视图状态
type SSHProxyViewState struct {
	SSHClientChan  chan proxy.SSHClientResultChan
	SSHProcessChan chan proxy.SSHProcessChan
}
