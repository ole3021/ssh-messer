package messer

import (
	"net/http"
	"ssh-messer/internal/config"
	"ssh-messer/internal/pubsub"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type ClientStatus int

const (
	StatusDisconnected ClientStatus = iota
	StatusConnecting
	StatusConnected
	StatusChecking
)

type ReverseService struct {
	Subdomain     string
	SSHClient     *ssh.Client
	Host          string
	Port          string
	UseTLS        bool
	TLSServerName string
	RemoteHost    string
	Name          string
}

type ReversePages struct {
	Name string
	URL  string
}

type MesserClient struct {
	SSHClient   *ssh.Client
	Status      ClientStatus
	ConnectedAt time.Time
	UpdatedAt   time.Time
}

type MesserHops struct {
	Config          *config.MesserConfig
	SSHClients      map[int]*MesserClient
	ReverseServices map[string]ReverseService
	ReversePages    []ReversePages
	CurrentInfo     string
	LastError       error

	// pubsub
	Broker *pubsub.Broker[MesserEventType, any]
	SubCh  <-chan pubsub.Event[MesserEventType, any]

	// reverse proxy
	server  *http.Server
	mu      sync.RWMutex
	stopped bool

	// 添加 sshTransport 缓存，key 格式: "sshClient指针:useTLS:tlsServerName"
	transportCache   map[string]*sshTransport
	transportCacheMu sync.RWMutex
}

// PubSub Event Types
type MesserEventType string

const (
	EveSSHStatusUpdate MesserEventType = "messer_event_ssh_status_update"
	EveServiceProxyLog MesserEventType = "messer_event_service_proxy_log"
)

type EveSSHStatusUpdatePayload struct {
	Info  string
	Error error
}

type EveSSHServiceProxyLogPayload struct {
	RequestID  string
	Method     string
	URL        string
	StatusCode int
	Duration   time.Duration
	Error      error
}

// ServiceConfig 服务配置信息
type ServiceConfig struct {
	scheme        string
	remoteAddr    string
	remoteHost    string
	tlsServerName string
	useTLS        bool
}
