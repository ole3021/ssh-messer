package types

import (
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/tui/messages"
)

// ViewEnum 视图类型枚举
type ViewEnum int

const (
	WelcomeView ViewEnum = iota
	SSHProxyView
)

type AppState struct {
	// 配置相关
	Configs           map[string]loaders.TomlConfig
	CurrentConfigName string
	SSHInfos          map[string]messages.SSHInfo
	Error             messages.AppError
}

// NewAppState 创建新的应用状态
func NewAppState() *AppState {
	return &AppState{
		Configs:  make(map[string]loaders.TomlConfig),
		SSHInfos: make(map[string]messages.SSHInfo),
	}
}

// SetConfigs 设置配置信息
func (s *AppState) SetConfigs(configs map[string]loaders.TomlConfig) {
	s.Configs = configs
}

// GetConfigs 获取配置信息
func (s *AppState) GetConfigs() map[string]loaders.TomlConfig {
	return s.Configs
}

// SetSSHInfo 设置 SSH 信息
func (s *AppState) SetSSHInfo(configName string, sshInfo messages.SSHInfo) {
	s.SSHInfos[configName] = sshInfo
}

// GetSSHInfo 获取 SSH 信息
func (s *AppState) GetSSHInfo(configName string) messages.SSHInfo {
	if info, exists := s.SSHInfos[configName]; exists {
		return info
	}
	return messages.SSHInfo{}
}

// SetCurrentConfig 设置当前配置
func (s *AppState) SetCurrentConfigName(configName string) {
	s.CurrentConfigName = configName
}

// GetCurrentConfig 获取当前配置
func (s *AppState) GetCurrentConfigName() string {
	return s.CurrentConfigName
}

// **** UI State ****
// ***********************************************************

type UIState struct {
	Width  int
	Height int
}

func NewUIState() *UIState {
	return &UIState{}
}
