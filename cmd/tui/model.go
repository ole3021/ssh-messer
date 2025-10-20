package main

import (
	"ssh-messer/internal/loaders"

	"golang.org/x/crypto/ssh"
)

// ConfigModel represents the configuration selection model
type ConfigModel struct {
	Choices         []string         `json:"choices"`
	Cursor          int              `json:"cursor"`
	Selected        map[int]struct{} `json:"selected"`
	WelcomeProgress int              `json:"welcome_progress"`
}

type SSHInfo struct {
	SSHClient            *ssh.Client     `json:"ssh_client"`
	SSHConnectionState   SSHConnectState `json:"ssh_connection_state"`
	SSHConnectionProcess int             `json:"ssh_connection_process"`
	HTTPProxyLogs        []string        `json:"http_proxy_logs"`
	DockerProxyLogs      []string        `json:"docker_proxy_logs"`
}

type WelcomeViewModel struct {
	WelcomeAnimationProgress int `json:"welcome_animation_progress"`
}

type MainViewModel struct {
}

type ConfigViewModel struct {
	ConfigNames  []string `json:"config_names"`
	Cursor       int      `json:"cursor"`
	SelectedName string   `json:"selected_name"`
	Error        string   `json:"error"`
}

// AppModel is the main application model that can hold different view models
type AppModel struct {
	// Configs
	Configs map[string]loaders.TomlConfig `json:"configs"`
	// SSH Messer Models
	CurrentConfigName string             `json:"current_config_name"`
	SSHInfos          map[string]SSHInfo `json:"ssh_info"`
	// View Related Model
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	CurrentView ViewEnum `json:"current_view"`
	CurrentInfo string   `json:"current_info"`
	// View Models
	WelcomeViewModel WelcomeViewModel `json:"welcome_model"`
	ConfigViewModel  ConfigViewModel  `json:"config_model"`
	MainViewModel    MainViewModel    `json:"main_model"`
}
