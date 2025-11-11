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

type ServicePage struct {
	Name *string `toml:"name"`
	URL  *string `toml:"url"`
}

type SSHService struct {
	Host          *string       `toml:"host"`
	Port          *string       `toml:"port"`
	Subdomain     *string       `toml:"subdomain"`
	Alias         *string       `toml:"alias,omitempty"`
	UseTLS        *bool         `toml:"use_tls,omitempty"`
	TLSServerName *string       `toml:"tls_server_name,omitempty"`
	RemoteHost    *string       `toml:"remote_host,omitempty"`
	Pages         []ServicePage `toml:"pages,omitempty"`
	HopOrder      *int          `toml:"hopOrder,omitempty"`
}

type SSHHopsProxy struct {
	configName          string
	hopsConfigs         []SSHHopConfig
	client              *ssh.Client
	hopClients          map[int]*ssh.Client // 存储不同 hopOrder 对应的 SSH client
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
