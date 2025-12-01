package messer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"ssh-messer/internal/config"
	"ssh-messer/internal/pubsub"
	"ssh-messer/pkg"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"sync"

	"golang.org/x/crypto/ssh"
)

func NewMesserHops(config *config.MesserConfig, eventsCtx context.Context) *MesserHops {
	broker := pubsub.NewBroker[MesserEventType, any]()

	return &MesserHops{
		Config:          config,
		SSHClients:      make(map[int]*MesserClient),
		ReverseServices: make(map[string]ReverseService),
		transportCache:  make(map[string]*sshTransport), // 初始化缓存

		Broker: broker,
		SubCh:  broker.Subscribe(eventsCtx),
	}
}

func (h *MesserHops) updateHopStatus(hopOrder int, status ClientStatus, info string, error error) {

	client, ok := h.SSHClients[hopOrder]
	if !ok {
		client = &MesserClient{
			Status:    status,
			UpdatedAt: time.Now(),
		}
		h.SSHClients[hopOrder] = client
	}
	// Update Client Status if exist
	if status == StatusConnected {
		client.ConnectedAt = time.Now()
	}
	client.Status = status
	client.UpdatedAt = time.Now()

	h.CurrentInfo = info
	h.LastError = error

	// Publish SSH Status Update Event
	pkg.Logger.Debug().Str("config", h.Config.Name).Str("info", info).Msg("[hops:updateHopStatus] Publishing SSH Status Update Event")
	h.Broker.Publish(EveSSHStatusUpdate, EveSSHStatusUpdatePayload{Info: info, Error: error})
}

func (h *MesserHops) ConnectHops() {
	pkg.Logger.Info().Str("config_name", h.Config.Name).Msg("[hops:connect] Connect Start")

	sortSSHHopsByOrderAsc(h.Config.SSHHops)

	var preClient *ssh.Client
	var lastHopOrder int = len(h.Config.SSHHops) - 1
	for i, hop := range h.Config.SSHHops {
		pkg.Logger.Debug().Int("order", hop.Order).Str("name", hop.Name).Msg("[hops:connect] Connecting to hop")
		connectingInfo := fmt.Sprintf("%d / %d: Connecting to hop %s", i+1, len(h.Config.SSHHops), hop.Name)
		h.updateHopStatus(hop.Order, StatusConnecting, connectingInfo, nil)

		client, err := connectSSHHop(hop, preClient)
		if err != nil {
			pkg.Logger.Error().Err(err).Str("hop", hop.Host+":"+strconv.Itoa(hop.Port)).Str("Name", hop.Name).Msg("[hops:connect] Failed to connect to hop")
			failedInfo := fmt.Sprintf("%d / %d: Failed to connect to hop %s", i+1, len(h.Config.SSHHops), hop.Name)
			h.updateHopStatus(hop.Order, StatusDisconnected, failedInfo, err)
			return
		}

		preClient = client
		lastHopOrder = hop.Order
		// Update existing MesserClient or create new one
		messerClient, exists := h.SSHClients[hop.Order]
		if !exists {
			messerClient = &MesserClient{}
			h.SSHClients[hop.Order] = messerClient
		}
		messerClient.SSHClient = client
		h.updateHopStatus(hop.Order, StatusConnected, "", nil)

		pkg.Logger.Info().Int("order", hop.Order).Str("name", hop.Name).Msg("[hops:connect] Connected to hop")
	}

	h.initReverseServicesPages(lastHopOrder)
	go h.CheckHealthLoop()

	// 如果配置了 ReverseServices 和 LocalHttpPort，自动启动反向代理
	if len(h.Config.ReverseServices) > 0 && h.Config.LocalHttpPort != "" {
		if err := h.StartReverseProxy(); err != nil {
			pkg.Logger.Error().Err(err).Str("config_name", h.Config.Name).Msg("[MesserHops] 自动启动反向代理失败")
		}
	}
}

func (h *MesserHops) initReverseServicesPages(lastHopOrder int) {
	for _, service := range h.Config.ReverseServices {
		hopOrder := lastHopOrder
		if service.CustomHopOrder > 0 {
			hopOrder = service.CustomHopOrder
		}

		var sshClient *ssh.Client
		if client, exists := h.SSHClients[hopOrder]; exists && client.SSHClient != nil {
			sshClient = client.SSHClient
		} else if lastHopOrder > 0 {
			if client, exists := h.SSHClients[lastHopOrder]; exists && client.SSHClient != nil {
				sshClient = client.SSHClient
			}
		}

		reverseService := ReverseService{
			Subdomain:     service.Subdomain,
			SSHClient:     sshClient,
			Host:          service.Host,
			Port:          service.Port,
			UseTLS:        service.UseTLS,
			TLSServerName: service.TLSServerName,
			RemoteHost:    service.RemoteHost,
			Name:          service.Name,
		}
		h.ReverseServices[service.Subdomain] = reverseService

		pkg.Logger.Info().Str("subdomain", service.Subdomain).Str("URL", "http://"+service.Subdomain+"."+"localhost"+":"+h.Config.LocalHttpPort).Msg("[hops:initReverseServicesPages] Initialized reverse page")
		for _, page := range service.Pages {
			h.ReversePages = append(h.ReversePages, ReversePages{
				Name: page.Name,
				// TODO: Use better way to generate correct URL
				URL: "http://" + service.Subdomain + "." + "localhost" + ":" + h.Config.LocalHttpPort + page.Path,
			})

			pkg.Logger.Info().Str("subdomain", service.Subdomain).Str("name", page.Name).Str("url", h.ReversePages[len(h.ReversePages)-1].URL).Msg("[hops:initReverseServicesPages] Initialized reverse page")
		}
	}
}

func (h *MesserHops) CheckHealthLoop() {
	if h.Config.HealthCheckIntervalSecs == 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(h.Config.HealthCheckIntervalSecs) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.CheckHealth()
	}
}

func (h *MesserHops) CheckHealth() {
	pkg.Logger.Info().Str("config_name", h.Config.Name).Msg("[hops:checkHealth] Checking health")
	requireReConnect := false

	orders := getSortedSSHClientOrders(h.SSHClients)

	for _, order := range orders {
		client := h.SSHClients[order]
		h.updateHopStatus(order, StatusChecking, "", nil)

		conn := client.SSHClient.Conn
		// 检查连接是否正常
		if conn == nil {
			h.updateHopStatus(order, StatusDisconnected, "SSH connection is disconnected", nil)
			// 触发重连
			requireReConnect = true
			break
		}

		// 尝试创建SSH会话
		session, err := client.SSHClient.NewSession()
		if err != nil {
			h.updateHopStatus(order, StatusDisconnected, "Failed to create SSH session", err)
			// 触发重连
			requireReConnect = true
			break
		}
		defer session.Close()

		err = session.Run("echo 'health_check'")
		if err != nil {
			h.updateHopStatus(order, StatusDisconnected, "Failed to execute SSH command", err)
			// 触发重连
			requireReConnect = true
			break
		}

		h.updateHopStatus(order, StatusConnected, "", nil)
	}

	if requireReConnect {
		pkg.Logger.Info().Str("config_name", h.Config.Name).Msg("[hops:checkHealth] re-connect")
		go h.ConnectHops()
	}
}

func (h *MesserHops) StartReverseProxy() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return fmt.Errorf("reverse proxy is stopped")
	}

	if h.server != nil {
		return fmt.Errorf("reverse proxy is already running")
	}

	if h.Config.LocalHttpPort == "" {
		return fmt.Errorf("local http port is not configured")
	}

	if len(h.ReverseServices) == 0 {
		return fmt.Errorf("no reverse services configured")
	}

	pkg.Logger.Debug().Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Int("services_count", len(h.ReverseServices)).Msg("[MesserHops] 开始启动反向代理")

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleReverseProxyRequest)

	h.server = &http.Server{
		Addr:    ":" + h.Config.LocalHttpPort,
		Handler: mux,
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg.Logger.Error().Err(err).Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Msg("[MesserHops] 反向代理服务器错误")
		}
	}()

	pkg.Logger.Info().Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Msg("[MesserHops] 反向代理启动成功")
	return nil
}

func (h *MesserHops) handleReverseProxyRequest(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	stopped := h.stopped
	reverseServices := h.ReverseServices
	h.mu.RUnlock()

	if stopped {
		http.Error(w, "Reverse proxy is not available", http.StatusServiceUnavailable)
		return
	}

	// 从 Host header 提取 subdomain
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	// 尝试通过 subdomain 路由
	var targetService *ReverseService
	subdomain := extractSubdomain(host)
	if subdomain != "" {
		if service, exists := reverseServices[subdomain]; exists {
			targetService = &service
		}
	}

	// 如果 subdomain 路由失败，尝试通过路径路由到 pages
	if targetService == nil {
		path := r.URL.Path
		for _, service := range reverseServices {
			// 检查是否匹配某个 page 的路径
			for _, page := range h.Config.ReverseServices {
				if page.Subdomain == service.Subdomain {
					for _, pageConfig := range page.Pages {
						if strings.HasPrefix(path, pageConfig.Path) {
							targetService = &service
							break
						}
					}
					if targetService != nil {
						break
					}
				}
			}
			if targetService != nil {
				break
			}
		}
	}

	if targetService == nil {
		http.Error(w, "No service found for this request", http.StatusNotFound)
		return
	}

	// 获取 SSH client
	sshClient := targetService.SSHClient
	if sshClient == nil {
		// 如果服务的 SSH client 不存在，尝试从 SSHClients 获取
		// 根据服务的 CustomHopOrder 或使用最后一个 hop
		var lastHopOrder int
		for order := range h.SSHClients {
			if order > lastHopOrder {
				lastHopOrder = order
			}
		}

		// 查找对应的服务配置以获取 CustomHopOrder
		var hopOrder int
		for _, service := range h.Config.ReverseServices {
			if service.Subdomain == targetService.Subdomain {
				if service.CustomHopOrder > 0 {
					hopOrder = service.CustomHopOrder
				} else {
					hopOrder = lastHopOrder
				}
				break
			}
		}

		if client, exists := h.SSHClients[hopOrder]; exists && client.SSHClient != nil {
			sshClient = client.SSHClient
		} else if lastHopOrder > 0 {
			if client, exists := h.SSHClients[lastHopOrder]; exists && client.SSHClient != nil {
				sshClient = client.SSHClient
			}
		}
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
	serviceAlias := targetService.Name
	if serviceAlias == "" {
		serviceAlias = targetService.Subdomain
	}
	if serviceAlias == "" {
		serviceAlias = "unknown"
	}

	// 查找对应的服务配置
	var serviceConfig *config.ReverseServiceConfig
	for i := range h.Config.ReverseServices {
		if h.Config.ReverseServices[i].Subdomain == targetService.Subdomain {
			serviceConfig = &h.Config.ReverseServices[i]
			break
		}
	}

	if serviceConfig == nil {
		http.Error(w, "Service configuration not found", http.StatusInternalServerError)
		return
	}

	// 构建服务配置
	serviceCfg := buildServiceConfig(serviceConfig)
	remoteURL, err := url.Parse(serviceCfg.scheme + "://" + serviceCfg.remoteAddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid remote address: %v", err), http.StatusInternalServerError)
		return
	}

	// 创建响应写入器来捕获状态码和响应大小
	responseWriter := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		responseSize:   0,
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(remoteURL)

	// 自定义 Transport 以通过 SSH 隧道
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 使用配置中的 scheme 和 host
		req.URL.Scheme = serviceCfg.scheme
		req.URL.Host = serviceCfg.remoteAddr
		// 重写 Host header（关键：用于 SNI 和虚拟主机识别）
		req.Host = serviceCfg.remoteHost
	}

	// 获取或创建缓存的 sshTransport 实例
	transportKey := fmt.Sprintf("%p:%t:%s", sshClient, serviceCfg.useTLS, serviceCfg.tlsServerName)

	h.transportCacheMu.RLock()
	sshTrans, exists := h.transportCache[transportKey]
	h.transportCacheMu.RUnlock()

	if !exists {
		h.transportCacheMu.Lock()
		// 双重检查
		if sshTrans, exists = h.transportCache[transportKey]; !exists {
			if h.transportCache == nil {
				h.transportCache = make(map[string]*sshTransport)
			}
			sshTrans = &sshTransport{
				sshClient:     sshClient,
				useTLS:        serviceCfg.useTLS,
				tlsServerName: serviceCfg.tlsServerName,
			}
			h.transportCache[transportKey] = sshTrans
		}
		h.transportCacheMu.Unlock()
	}

	proxy.Transport = sshTrans

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

	// 构建错误对象
	var proxyErr error
	if errorMessage != "" {
		proxyErr = fmt.Errorf("%s", errorMessage)
	}

	// 发布日志事件
	payload := EveSSHServiceProxyLogPayload{
		RequestID:  requestID,
		Method:     r.Method,
		URL:        r.URL.String(),
		StatusCode: responseWriter.statusCode,
		Duration:   duration,
		Error:      proxyErr,
	}

	h.Broker.Publish(EveServiceProxyLog, payload)
}

// truncateErrorMessage 截断错误消息到指定长度
const MaxErrorMessageLength = 200

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

func (h *MesserHops) StopReverseProxy() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return
	}

	pkg.Logger.Debug().Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Msg("[MesserHops] 开始停止反向代理")

	h.stopped = true

	if h.server != nil {
		err := h.server.Close()
		h.server = nil
		if err != nil && err != http.ErrServerClosed {
			pkg.Logger.Error().Err(err).Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Msg("[MesserHops] 停止反向代理时发生错误")
		}
	}

	// 清理 Transport 缓存
	h.transportCacheMu.Lock()
	if h.transportCache != nil {
		for _, sshTrans := range h.transportCache {
			sshTrans.CloseIdleConnections()
		}
		h.transportCache = make(map[string]*sshTransport)
	}
	h.transportCacheMu.Unlock()

	pkg.Logger.Info().Str("config_name", h.Config.Name).Str("port", h.Config.LocalHttpPort).Msg("[MesserHops] 反向代理已停止")
}

// ============================================================
// 辅助函数
// ============================================================

var requestIDCounter atomic.Uint64

// generateRequestID 生成唯一的请求 ID
func generateRequestID() string {
	counter := requestIDCounter.Add(1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), counter)
}

// extractSubdomain 从 host 中提取 subdomain
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
func buildRemoteAddress(service *config.ReverseServiceConfig) string {
	host := "localhost"
	port := "80"

	if service.Host != "" {
		host = service.Host
	}
	if service.Port != "" {
		port = service.Port
	}

	return host + ":" + port
}

// buildServiceConfig 构建服务配置
func buildServiceConfig(service *config.ReverseServiceConfig) ServiceConfig {
	config := ServiceConfig{
		scheme:     "http",
		remoteAddr: buildRemoteAddress(service),
		useTLS:     false,
	}

	// 确定是否使用 TLS
	if service.UseTLS {
		config.scheme = "https"
		config.useTLS = true
	}

	// 确定远程 Host header
	if service.RemoteHost != "" {
		config.remoteHost = service.RemoteHost
	} else if service.Host != "" {
		config.remoteHost = service.Host
	} else {
		config.remoteHost = "localhost"
	}

	// 确定 TLS ServerName
	if service.TLSServerName != "" {
		config.tlsServerName = service.TLSServerName
	} else if config.useTLS {
		// 如果使用 TLS 但没有指定 ServerName，使用 remoteHost
		config.tlsServerName = config.remoteHost
	}

	return config
}

// formatServiceProxyLog 格式化服务代理日志为字符串
func formatServiceProxyLog(method, url, serviceAlias string, statusCode int, duration time.Duration, errorMsg string) string {
	var logParts []string
	logParts = append(logParts, fmt.Sprintf("[%s]", method))
	logParts = append(logParts, url)
	logParts = append(logParts, fmt.Sprintf("Service: %s", serviceAlias))
	logParts = append(logParts, fmt.Sprintf("Status: %d", statusCode))
	logParts = append(logParts, fmt.Sprintf("Duration: %v", duration))
	if errorMsg != "" {
		logParts = append(logParts, fmt.Sprintf("Error: %s", errorMsg))
	}
	return strings.Join(logParts, " | ")
}

// ============================================================
// SSH Transport
// ============================================================

// sshTransport 通过 SSH 隧道传输 HTTP 请求
type sshTransport struct {
	sshClient     *ssh.Client
	useTLS        bool
	tlsServerName string
	// 添加 Transport 缓存，使用读写锁保护
	transportCache map[string]*http.Transport
	transportMu    sync.RWMutex
}

// getOrCreateTransport 获取或创建缓存的 Transport
// cacheKey 格式: "targetAddr:useTLS:tlsServerName"
func (t *sshTransport) getOrCreateTransport(targetAddr string) (*http.Transport, error) {
	// 构建缓存 key
	cacheKey := fmt.Sprintf("%s:%t:%s", targetAddr, t.useTLS, t.tlsServerName)

	// 先尝试读锁获取
	t.transportMu.RLock()
	if t.transportCache != nil {
		if transport, exists := t.transportCache[cacheKey]; exists {
			t.transportMu.RUnlock()
			return transport, nil
		}
	}
	t.transportMu.RUnlock()

	// 需要创建新的 Transport，使用写锁
	t.transportMu.Lock()
	defer t.transportMu.Unlock()

	// 双重检查，防止并发创建
	if t.transportCache == nil {
		t.transportCache = make(map[string]*http.Transport)
	}
	if transport, exists := t.transportCache[cacheKey]; exists {
		return transport, nil
	}

	// 创建 TLS 配置
	var tlsConfig *tls.Config
	if t.useTLS {
		tlsConfig = &tls.Config{
			ServerName:         t.tlsServerName,
			InsecureSkipVerify: false, // 验证证书
		}
	}

	// 创建新的 Transport，配置连接池参数
	transport := &http.Transport{
		MaxIdleConns:      5,                // 最大空闲连接数，避免过多连接
		MaxConnsPerHost:   32,               // 每个主机的最大连接数
		IdleConnTimeout:   90 * time.Second, // 空闲连接超时
		DisableKeepAlives: false,            // 启用连接复用
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

	// 缓存 Transport
	t.transportCache[cacheKey] = transport

	return transport, nil
}

// RoundTrip 实现通过 SSH 隧道的 HTTP 请求
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

	// 获取或创建缓存的 Transport
	transport, err := t.getOrCreateTransport(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport: %v", err)
	}

	// 创建 HTTP Client，复用 Transport
	// 注意：每次创建新的 Client 是可以的，关键是复用 Transport
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

// CloseIdleConnections 关闭所有空闲连接
// 可以在需要时调用，比如 SSH 连接断开时
func (t *sshTransport) CloseIdleConnections() {
	t.transportMu.RLock()
	defer t.transportMu.RUnlock()

	if t.transportCache != nil {
		for _, transport := range t.transportCache {
			transport.CloseIdleConnections()
		}
	}
}

// ============================================================
// Response Writer
// ============================================================

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
