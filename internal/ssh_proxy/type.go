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
	configName          string
	hopsConfigs         []SSHHopConfig
	client              *ssh.Client
	Status              SSHProxyStatus
	serviceProxy        *ServiceProxy
	healthStop          chan struct{}
	services            []SSHService
	localPort           string
	healthCheckInterval time.Duration
}

type SSHProxyStatus struct {
	IsConnecting         bool
	IsConnected          bool
	ConnectedAt          time.Time
	IsChecking           bool
	CheckedAt            time.Time
	CurrentInfo          string
	LastError            error
	ReconnectAttempts    int
	LastReconnectAttempt time.Time
}

// UpdateSSHProxyStatus 更新 SSH 代理状态
func (s *SSHHopsProxy) UpdateSSHProxyStatus(status SSHProxyStatus) {
	s.Status = status
}

// ServiceProxyLogEvent Service 代理日志事件
type ServiceProxyLogEvent struct {
	RequestID    string // 请求 ID，用于匹配请求和响应
	ConfigName   string
	ServiceAlias string
	Method       string
	URL          string
	StatusCode   int // 0 表示请求中，>0 表示响应状态码
	ResponseSize int64
	Timestamp    time.Time
	IsUpdate     bool          // true 表示这是更新事件，false 表示新请求
	ErrorMessage string        // 错误消息（如果有）
	Duration     time.Duration // 请求用时
}
