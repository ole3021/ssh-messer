package messages

import (
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/proxy"

	tea "github.com/charmbracelet/bubbletea"
)

// ==================== 应用错误相关消息 ====================

// AppError 消息：应用错误
type AppError struct {
	Error   *error
	IsFatal bool
}

// ==================== 配置相关消息 ====================

// LoadConfigs 消息：配置文件加载完成
type LoadConfigs struct {
	Configs map[string]loaders.TomlConfig
	Err     error
}

// ConfigSelected 消息：用户选择了一个配置
type ConfigSelected struct {
	ConfigName string
}

// ConfigCancelled 消息：用户取消了配置选择
type ConfigCancelled struct{}

// ==================== SSH 连接相关消息 ====================

// SSHClientResult 消息：SSH 客户端连接结果
type SSHClientResult struct {
	Result proxy.SSHClientResultChan
}

// SSHProcessResult 消息：SSH 进程状态结果
type SSHProcessResult struct {
	Result proxy.SSHProcessChan
}

type SSHConnectState int

const (
	Disconnected SSHConnectState = iota
	Connecting
	Connected
	Error
)

// SSHInfo 结构体：SSH 连接信息
type SSHInfo struct {
	SSHClient               interface{}         // use interface{} to avoid circular dependency
	SSHServicesReverseProxy *proxy.ServiceProxy // use interface{} to avoid circular dependency
	SSHConnectionState      SSHConnectState
	SSHConnectionProcess    int
	CurrentInfo             string
	HTTPProxyLogs           []string
	DockerProxyLogs         []string
}

// ==================== 动画相关消息 ====================

// WelcomeTick 消息：欢迎页面动画 tick
type WelcomeTick struct{}

// SSHConnectionTick 消息：SSH 连接动画 tick
type SSHConnectionTick struct{}

// ==================== 视图切换相关消息 ====================

// ViewChangeMsg 消息：请求切换视图
type ViewChangeMsg struct {
	TargetView int // 使用 int 避免循环依赖
}

type ProxyRequestResult struct {
	Result proxy.ProxyRequestResult
}

// 确保所有消息类型都实现了 tea.Msg 接口
var _ tea.Msg = AppError{}
var _ tea.Msg = LoadConfigs{}
var _ tea.Msg = ConfigSelected{}
var _ tea.Msg = ConfigCancelled{}
var _ tea.Msg = SSHClientResult{}
var _ tea.Msg = SSHProcessResult{}
var _ tea.Msg = WelcomeTick{}
var _ tea.Msg = SSHConnectionTick{}
var _ tea.Msg = ViewChangeMsg{}
var _ tea.Msg = ProxyRequestResult{}
