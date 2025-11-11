package ssh_proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ssh-messer/internal/pubsub"
	"ssh-messer/pkg"

	"golang.org/x/crypto/ssh"
)

// Service Proxy 日志事件
// ------------------------------------------------------------
const (
	MaxErrorMessageLength = 200 // 错误消息最大长度
)

var (
	serviceProxyLogBroker = pubsub.NewBroker[ServiceProxyLogEvent]()
	requestIDCounter      atomic.Uint64
)

// GetServiceProxyLogBroker 获取 Service Proxy 日志 broker
func GetServiceProxyLogBroker() *pubsub.Broker[ServiceProxyLogEvent] {
	return serviceProxyLogBroker
}

// generateRequestID 生成唯一的请求 ID
func generateRequestID() string {
	counter := requestIDCounter.Add(1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), counter)
}

// truncateErrorMessage 截断错误消息到指定长度
func truncateErrorMessage(msg string) string {
	if len([]rune(msg)) <= MaxErrorMessageLength {
		return msg
	}
	runes := []rune(msg)
	if MaxErrorMessageLength > 3 {
		return string(runes[:MaxErrorMessageLength-3]) + "..."
	}
	return string(runes[:MaxErrorMessageLength])
}

// ============================================================

// ServiceProxy 统一的 HTTP 反向代理服务器
// ------------------------------------------------------------
type ServiceProxy struct {
	configName      string
	localPort       string
	services        []SSHService
	sshClient       *ssh.Client
	getClientForHop func(int) *ssh.Client // 根据 hopOrder 获取对应的 SSH client
	server          *http.Server
	serviceMap      map[string]*SSHService // subdomain -> service mapping
	mu              sync.RWMutex
	stopped         bool
}

// NewServiceProxy 创建新的 Service Proxy
// serviceMap 在构建后只读，确保并发安全
func NewServiceProxy(configName string, localPort string, services []SSHService, sshClient *ssh.Client) *ServiceProxy {
	return NewServiceProxyWithHopSelector(configName, localPort, services, sshClient, nil)
}

// NewServiceProxyWithHopSelector 创建新的 Service Proxy，支持根据 hopOrder 选择 client
func NewServiceProxyWithHopSelector(configName string, localPort string, services []SSHService, defaultClient *ssh.Client, getClientForHop func(int) *ssh.Client) *ServiceProxy {
	serviceMap := make(map[string]*SSHService)
	for i := range services {
		service := &services[i]
		if service.Subdomain != nil && *service.Subdomain != "" {
			serviceMap[*service.Subdomain] = service
		}
	}

	return &ServiceProxy{
		configName:      configName,
		localPort:       localPort,
		services:        services,
		sshClient:       defaultClient,
		getClientForHop: getClientForHop,
		serviceMap:      serviceMap,
		stopped:         false,
	}
}

// ============================================================

// Start ReverseProxy 启动 处理 停止 逻辑
// ------------------------------------------------------------
func (sp *ServiceProxy) StartReverseProxy() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.stopped {
		return fmt.Errorf("service proxy is stopped")
	}

	if sp.server != nil {
		return fmt.Errorf("service proxy is already running")
	}

	pkg.Logger.Debug().Str("config_name", sp.configName).Str("port", sp.localPort).Int("services_count", len(sp.services)).Msg("[ServiceProxy] 开始启动服务代理")

	mux := http.NewServeMux()
	mux.HandleFunc("/", sp.handleReverseProxyRequest)

	sp.server = &http.Server{
		Addr:    ":" + sp.localPort,
		Handler: mux,
	}

	go func() {
		if err := sp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg.Logger.Error().Err(err).Str("config_name", sp.configName).Str("port", sp.localPort).Msg("[ServiceProxy] 服务代理服务器错误")
		}
	}()

	pkg.Logger.Info().Str("config_name", sp.configName).Str("port", sp.localPort).Msg("[ServiceProxy] 服务代理启动成功")
	return nil
}

// handleRequest 处理 HTTP 请求
func (sp *ServiceProxy) handleReverseProxyRequest(w http.ResponseWriter, r *http.Request) {
	sp.mu.RLock()
	stopped := sp.stopped
	serviceMap := sp.serviceMap
	sp.mu.RUnlock()

	if stopped {
		http.Error(w, "Service proxy is not available", http.StatusServiceUnavailable)
		return
	}

	// 从 Host header 提取 subdomain
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	// 尝试通过 subdomain 路由
	var targetService *SSHService
	subdomain := extractSubdomain(host)
	if subdomain != "" {
		targetService = serviceMap[subdomain]
	}

	// 如果 subdomain 路由失败，尝试通过路径路由
	if targetService == nil {
		path := r.URL.Path
		for _, service := range sp.services {
			if service.Alias != nil && strings.HasPrefix(path, "/"+*service.Alias+"/") {
				targetService = &service
				// 重写路径，移除 alias 前缀
				r.URL.Path = strings.TrimPrefix(path, "/"+*service.Alias)
				break
			}
		}
	}

	if targetService == nil {
		http.Error(w, "No service found for this request", http.StatusNotFound)
		return
	}

	// 根据 service 的 HopOrder 选择对应的 SSH client
	sp.mu.RLock()
	getClientForHop := sp.getClientForHop
	defaultClient := sp.sshClient
	sp.mu.RUnlock()

	var sshClient *ssh.Client
	if targetService.HopOrder != nil && *targetService.HopOrder > 0 && getClientForHop != nil {
		// 使用指定 hopOrder 的 client
		sshClient = getClientForHop(*targetService.HopOrder)
		if sshClient == nil {
			// 如果指定的 hopOrder client 不存在，使用默认 client
			sshClient = defaultClient
		}
	} else {
		// 使用默认 client
		sshClient = defaultClient
	}

	if sshClient == nil {
		http.Error(w, "SSH client is not available", http.StatusServiceUnavailable)
		return
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 生成请求 ID
	requestID := generateRequestID()

	// 获取服务别名
	serviceAlias := "unknown"
	if targetService.Alias != nil {
		serviceAlias = *targetService.Alias
	} else if targetService.Subdomain != nil {
		serviceAlias = *targetService.Subdomain
	}

	// 在请求开始时发送日志（StatusCode 为 0 表示请求中）
	requestEvent := ServiceProxyLogEvent{
		RequestID:    requestID,
		ConfigName:   sp.configName,
		ServiceAlias: serviceAlias,
		Method:       r.Method,
		URL:          r.URL.String(),
		StatusCode:   0, // 0 表示请求中
		ResponseSize: 0,
		Timestamp:    startTime,
		IsUpdate:     false,
	}
	serviceProxyLogBroker.Publish(pubsub.UpdatedEvent, requestEvent)

	// 创建响应写入器来捕获状态码和响应大小
	responseWriter := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		responseSize:   0,
	}

	// 构建服务配置
	serviceConfig := buildServiceConfig(targetService)
	remoteURL, err := url.Parse(serviceConfig.scheme + "://" + serviceConfig.remoteAddr)
	if err != nil {
		// 计算请求用时
		duration := time.Since(startTime)
		http.Error(w, fmt.Sprintf("Invalid remote address: %v", err), http.StatusInternalServerError)
		// 发送错误响应日志
		errorEvent := ServiceProxyLogEvent{
			RequestID:    requestID,
			ConfigName:   sp.configName,
			ServiceAlias: serviceAlias,
			Method:       r.Method,
			URL:          r.URL.String(),
			StatusCode:   http.StatusInternalServerError,
			ResponseSize: 0,
			Timestamp:    startTime,
			IsUpdate:     true,
			Duration:     duration,
		}
		serviceProxyLogBroker.Publish(pubsub.UpdatedEvent, errorEvent)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(remoteURL)

	// 自定义 Transport 以通过 SSH 隧道
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 使用配置中的 scheme 和 host
		req.URL.Scheme = serviceConfig.scheme
		req.URL.Host = serviceConfig.remoteAddr
		// 重写 Host header（关键：用于 SNI 和虚拟主机识别）
		req.Host = serviceConfig.remoteHost
	}
	proxy.Transport = &sshTransport{
		sshClient:     sshClient,
		useTLS:        serviceConfig.useTLS,
		tlsServerName: serviceConfig.tlsServerName,
	}

	// 设置错误处理器以捕获错误消息
	var errorMessage string
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		errorMessage = truncateErrorMessage(err.Error())
		http.Error(w, errorMessage, http.StatusBadGateway)
	}

	// 执行代理请求
	proxy.ServeHTTP(responseWriter, r)

	// 计算请求用时
	duration := time.Since(startTime)

	// 在响应返回后发送更新日志
	responseEvent := ServiceProxyLogEvent{
		RequestID:    requestID,
		ConfigName:   sp.configName,
		ServiceAlias: serviceAlias,
		Method:       r.Method,
		URL:          r.URL.String(),
		StatusCode:   responseWriter.statusCode,
		ResponseSize: responseWriter.responseSize,
		Timestamp:    startTime,
		IsUpdate:     true,
		ErrorMessage: errorMessage,
		Duration:     duration,
	}

	serviceProxyLogBroker.Publish(pubsub.UpdatedEvent, responseEvent)
}

// Stop 停止 Service Proxy 服务器
func (sp *ServiceProxy) StopReverseProxy() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.stopped {
		return nil
	}

	pkg.Logger.Debug().Str("config_name", sp.configName).Str("port", sp.localPort).Msg("[ServiceProxy] 开始停止服务代理")

	sp.stopped = true

	if sp.server != nil {
		err := sp.server.Close()
		sp.server = nil
		if err != nil && err != http.ErrServerClosed {
			pkg.Logger.Error().Err(err).Str("config_name", sp.configName).Str("port", sp.localPort).Msg("[ServiceProxy] 停止服务代理时发生错误")
			return err
		}
	}

	pkg.Logger.Info().Str("config_name", sp.configName).Str("port", sp.localPort).Msg("[ServiceProxy] 服务代理已停止")
	return nil
}

// ============================================================

// extractSubdomain 从 host 中提取 subdomain
// ------------------------------------------------------------
func extractSubdomain(host string) string {
	// 移除端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// 移除 localhost 后缀
	if strings.HasSuffix(host, ".localhost") {
		subdomain := strings.TrimSuffix(host, ".localhost")
		return subdomain
	}

	// 如果 host 包含点，尝试提取第一个部分作为 subdomain
	parts := strings.Split(host, ".")
	if len(parts) > 1 {
		return parts[0]
	}

	return ""
}

// buildRemoteAddress 构建远程服务地址
func buildRemoteAddress(service *SSHService) string {
	host := "localhost"
	port := "80"

	if service.Host != nil && *service.Host != "" {
		host = *service.Host
	}
	if service.Port != nil && *service.Port != "" {
		port = *service.Port
	}

	return host + ":" + port
}

// ServiceConfig 服务配置信息
type ServiceConfig struct {
	scheme        string
	remoteAddr    string
	remoteHost    string
	tlsServerName string
	useTLS        bool
}

// buildServiceConfig 构建服务配置
func buildServiceConfig(service *SSHService) ServiceConfig {
	config := ServiceConfig{
		scheme:     "http",
		remoteAddr: buildRemoteAddress(service),
		useTLS:     false,
	}

	// 确定是否使用 TLS
	if service.UseTLS != nil && *service.UseTLS {
		config.scheme = "https"
		config.useTLS = true
	}

	// 确定远程 Host header
	if service.RemoteHost != nil && *service.RemoteHost != "" {
		config.remoteHost = *service.RemoteHost
	} else if service.Host != nil && *service.Host != "" {
		config.remoteHost = *service.Host
	} else {
		config.remoteHost = "localhost"
	}

	// 确定 TLS ServerName
	if service.TLSServerName != nil && *service.TLSServerName != "" {
		config.tlsServerName = *service.TLSServerName
	} else if config.useTLS {
		// 如果使用 TLS 但没有指定 ServerName，使用 remoteHost
		config.tlsServerName = config.remoteHost
	}

	return config
}

// sshTransport 通过 SSH 隧道传输 HTTP 请求
type sshTransport struct {
	sshClient     *ssh.Client
	useTLS        bool
	tlsServerName string
}

func (t *sshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 从请求 URL 提取目标地址
	targetAddr := req.URL.Host
	if req.URL.Port() == "" {
		if req.URL.Scheme == "https" || t.useTLS {
			targetAddr += ":443"
		} else {
			targetAddr += ":80"
		}
	}

	// 确定使用的 scheme
	scheme := "http"
	if t.useTLS {
		scheme = "https"
	}

	// 创建请求的副本，清除 RequestURI（客户端请求不能设置 RequestURI）
	newReq := req.Clone(req.Context())
	newReq.RequestURI = "" // 清除 RequestURI，这是客户端请求的要求
	newReq.URL.Scheme = scheme
	newReq.URL.Host = targetAddr

	// 创建 HTTP Transport，通过 SSH 客户端建立连接
	// 参考 maancoffee 的实现：让 Transport 自动处理 TLS
	var tlsConfig *tls.Config
	if t.useTLS {
		tlsConfig = &tls.Config{
			ServerName:         t.tlsServerName,
			InsecureSkipVerify: false, // 验证证书
		}
	}

	transport := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			// 通过 SSH 客户端建立到远程服务的 TCP 连接
			conn, err := t.sshClient.Dial("tcp", targetAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to dial remote service through SSH (target: %s): %v", targetAddr, err)
			}
			return conn, nil
		},
		TLSClientConfig: tlsConfig, // 如果使用 TLS，让 Transport 自动处理
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 执行请求
	resp, err := client.Do(newReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}

	return resp, nil
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码和响应大小
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(n)
	return n, err
}

// ============================================================
