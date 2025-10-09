package messages

import "ssh-messer/internal/config_loader"

// LoadConfigsMsg 配置加载完成消息
type LoadConfigsMsg struct {
	Configs map[string]*config_loader.TomlConfig
	Err     error
}

// ConfigLoadedMsg 配置加载完成消息（别名）
type ConfigLoadedMsg = LoadConfigsMsg

// ConfigSelectedMsg 配置选择消息
type ConfigSelectedMsg struct {
	ConfigName string
}
