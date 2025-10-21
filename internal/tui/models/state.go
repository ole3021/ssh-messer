package models

import (
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/tui/messages"
)

// NewAppState 创建新的应用状态
func NewAppState() *AppState {
	return &AppState{
		Configs:  make(map[string]loaders.TomlConfig),
		SSHInfos: make(map[string]messages.SSHInfo),
	}
}

// NewUIState 创建新的 UI 状态
func NewUIState() *UIState {
	return &UIState{}
}

// NewWelcomeViewState 创建新的欢迎视图状态
func NewWelcomeViewState() *WelcomeViewState {
	return &WelcomeViewState{}
}

// NewConfigViewState 创建新的配置视图状态
func NewConfigViewState() *ConfigViewState {
	return &ConfigViewState{}
}

// NewSSHProxyViewState 创建新的 SSH 代理视图状态
func NewSSHProxyViewState() *SSHProxyViewState {
	return &SSHProxyViewState{}
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
func (s *AppState) SetCurrentConfig(configName string) {
	s.CurrentConfigName = configName
}

// GetCurrentConfig 获取当前配置
func (s *AppState) GetCurrentConfig() string {
	return s.CurrentConfigName
}
