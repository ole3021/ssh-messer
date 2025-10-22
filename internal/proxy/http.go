package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"ssh-messer/internal/loaders"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ProxyRequestResult 代理请求结果
type ProxyRequestResult struct {
	StartTime    time.Time     // 开始时间
	SubDomain    string        // 子域名
	Method       string        // HTTP 方法
	URL          string        // 请求 URL
	StatusCode   int           // 响应状态码
	Duration     time.Duration // 请求耗时
	Success      bool          // 是否成功
	ErrorMessage string        // 错误信息（如果有）
}

type ServiceProxy struct {
	LocalPort      string
	LocalHost      string
	httpServer     *http.Server
	sshClient      *ssh.Client
	reverseProxies map[string]ReverseProxy
	InfoChan       chan ProxyRequestResult
	isRunning      bool
}

type ReverseProxy struct {
	alias string
	proxy *httputil.ReverseProxy
}

// responseRecorder 用于记录响应状态码
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

func NewtHttpServiceProxyServer(localPort string, services []loaders.TomlConfigService, sshClient *ssh.Client) *ServiceProxy {
	servicesProxy := make(map[string]ReverseProxy)
	for _, service := range services {
		reverseProxy := buildServicesProxy(service, sshClient)
		servicesProxy[*service.Subdomain] = *reverseProxy
	}

	return &ServiceProxy{
		LocalPort:      localPort,
		LocalHost:      "proxy.local",
		reverseProxies: servicesProxy,
		sshClient:      sshClient,
	}
}

func buildServicesProxy(service loaders.TomlConfigService, sshClient *ssh.Client) *ReverseProxy {
	targetURL, _ := url.Parse("http://" + *service.Host + ":" + *service.Port)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 设置Director
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = path.Join(targetURL.Path, req.URL.Path)
		if targetURL.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetURL.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
		}
		if req.URL.Fragment == "" {
			req.URL.Fragment = targetURL.Fragment
		}
		req.Host = targetURL.Host
	}

	// 设置Transport
	remoteAddress := *service.Host + ":" + *service.Port
	proxy.Transport = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			// 添加超时控制
			conn, err := sshClient.Dial("tcp", remoteAddress)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}

	// 设置错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "代理请求失败: "+err.Error(), http.StatusBadGateway)
	}

	return &ReverseProxy{
		alias: *service.Alias,
		proxy: proxy,
	}
}

func (s *ServiceProxy) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleProxyRequest)

	s.httpServer = &http.Server{
		Addr:    ":" + s.LocalPort,
		Handler: mux,
	}

	s.httpServer.ListenAndServe()
}

func (s *ServiceProxy) AsyncStart(infoChan chan ProxyRequestResult) {
	s.InfoChan = infoChan
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleProxyRequest)

	s.httpServer = &http.Server{
		Addr:    ":" + s.LocalPort,
		Handler: mux,
	}

	// 设置运行状态
	s.isRunning = true

	// 添加错误处理，确保服务持续运行
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server error: %v", err)
		s.isRunning = false
	}
}

// IsRunning 检查服务是否正在运行
func (s *ServiceProxy) IsRunning() bool {
	return s.isRunning
}

// Stop 停止服务
func (s *ServiceProxy) Stop() error {
	if s.httpServer != nil {
		s.isRunning = false
		return s.httpServer.Close()
	}
	return nil
}

// getAvailableServices 获取可用的服务列表
func (s *ServiceProxy) getAvailableServices() []string {
	services := make([]string, 0, len(s.reverseProxies))
	for subdomain := range s.reverseProxies {
		services = append(services, subdomain)
	}
	return services
}

func (s *ServiceProxy) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 获取请求的 路径 body header 等信息
	subDomain, err := getSubDomain(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.InfoChan != nil {
		s.InfoChan <- ProxyRequestResult{
			StartTime:  startTime,
			SubDomain:  subDomain,
			Method:     r.Method,
			URL:        r.URL.String(),
			StatusCode: 0,
			Duration:   0,
		}
	}

	reverseProxy := s.reverseProxies[subDomain]
	if reverseProxy == (ReverseProxy{}) {
		http.Error(w, "未找到服务", http.StatusNotFound)
		return
	}

	// 获取代理结果返回的响应状态码
	responseRecorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// 添加超时控制
	done := make(chan bool, 1)
	go func() {
		reverseProxy.proxy.ServeHTTP(responseRecorder, r)
		done <- true
	}()

	select {
	case <-done:
		// 代理请求完成
	case <-time.After(30 * time.Second):
		http.Error(w, "代理请求超时", http.StatusGatewayTimeout)
		return
	}

	statusCode := responseRecorder.statusCode

	duration := time.Since(startTime)
	statusSuccess := statusCode >= 200 && statusCode < 300

	if s.InfoChan != nil {
		s.InfoChan <- ProxyRequestResult{
			StartTime:    startTime,
			SubDomain:    subDomain,
			Method:       r.Method,
			URL:          r.URL.String(),
			StatusCode:   statusCode,
			Duration:     duration,
			Success:      statusSuccess,
			ErrorMessage: "",
		}
	}
}

func getSubDomain(r *http.Request) (string, error) {
	host := r.Host

	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return "", fmt.Errorf("没有子域名")
	}

	subDomain := parts[0]
	return subDomain, nil
}
