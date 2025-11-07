package types

import (
	"ssh-messer/internal/config_loader"
	"ssh-messer/internal/ssh_proxy"
)

// AppState 应用状态
type AppState struct {
	// 配置相关
	Configs           map[string]*config_loader.TomlConfig
	CurrentConfigName string
	SSHProxies        map[string]*ssh_proxy.SSHHopsProxy

	// APP Error
	Error AppError
}

// NewAppState 创建新的应用状态
func NewAppState() *AppState {
	return &AppState{
		Configs:    make(map[string]*config_loader.TomlConfig),
		SSHProxies: make(map[string]*ssh_proxy.SSHHopsProxy),
	}
}

// SetConfigs 设置配置信息
func (s *AppState) SetConfigs(configs map[string]*config_loader.TomlConfig) {
	s.Configs = configs
}

// GetConfigs 获取配置信息
func (s *AppState) GetConfigs() map[string]*config_loader.TomlConfig {
	return s.Configs
}

func (s *AppState) UpSetConfig(configName string, config *config_loader.TomlConfig) {
	s.Configs[configName] = config
}

func (s *AppState) GetConfig(configName string) *config_loader.TomlConfig {
	return s.Configs[configName]
}

// GetCurrentConfig 获取当前配置
func (s *AppState) GetCurrentConfig() *config_loader.TomlConfig {
	return s.Configs[s.CurrentConfigName]
}

func (s *AppState) SetCurrentConfigName(configName string) {
	s.CurrentConfigName = configName
}

// SetSSHProxy 设置 SSH 代理
func (s *AppState) SetSSHProxy(configName string, sshProxy *ssh_proxy.SSHHopsProxy) {
	s.SSHProxies[configName] = sshProxy
}

// GetSSHProxy 获取 SSH 代理
func (s *AppState) GetSSHProxy(configName string) *ssh_proxy.SSHHopsProxy {
	if proxy, exists := s.SSHProxies[configName]; exists {
		return proxy
	}
	return nil
}

// AppError 应用错误消息
type AppError struct {
	Error   error
	IsFatal bool
}
