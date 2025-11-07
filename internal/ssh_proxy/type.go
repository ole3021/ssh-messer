package ssh_proxy

import (
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHHopConfig struct {
	Order          *int    `toml:"order"`
	Host           *string `toml:"host"`
	Port           *int    `toml:"port"`
	AuthType       *string `toml:"authType"`
	PrivateKeyPath *string `toml:"privateKeyPath"`
	Passphrase     *string `toml:"passphrase"`
	User           *string `toml:"user"`
	Alias          *string `toml:"alias,omitempty"`
	TimeoutSec     *int    `toml:"timeoutSec,omitempty"`
}

type SSHService struct {
	Host      *string `toml:"host"`
	Port      *string `toml:"port"`
	Subdomain *string `toml:"subdomain"`
	Alias     *string `toml:"alias,omitempty"`
}

type SSHHopsProxy struct {
	configName  string
	hopsConfigs []SSHHopConfig
	client      *ssh.Client
	Status      SSHProxyStatus
}

type SSHProxyStatus struct {
	IsConnecting bool
	IsConnected  bool
	ConnectedAt  time.Time
	IsChecking   bool
	CheckedAt    time.Time
	CurrentInfo  string
	LastError    error
}

// UpdateSSHProxyStatus 更新 SSH 代理状态
func (s *SSHHopsProxy) UpdateSSHProxyStatus(status SSHProxyStatus) {
	s.Status = status
}
