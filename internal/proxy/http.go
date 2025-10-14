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

type ServiceProxy struct {
	LocalPort      string
	LocalHost      string
	httpServer     *http.Server
	sshClient      *ssh.Client
	reverseProxies map[string]ReverseProxy
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
		servicesProxy[*service.Subdomain] = *buildServicesProxy(service, sshClient)
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
			conn, err := sshClient.Dial("tcp", remoteAddress)
			if err != nil {
				log.Printf("🔗❌ SSH 隧道连接失败: %v", err)
				return nil, err
			}
			return conn, nil
		},
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 设置错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("🔗❌ 代理请求失败: %v", err)
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

func (s *ServiceProxy) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 获取请求的 路径 body header 等信息
	subDomain, err := getSubDomain(r)
	if err != nil {
		log.Printf("🔗❌ 获取子域名失败: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("🔗➡️ [%+v]: %+v", subDomain, r.Method+" "+r.URL.String())

	reverseProxy := s.reverseProxies[subDomain]
	if reverseProxy == (ReverseProxy{}) {
		log.Fatalf("🔗❌ 没有子域名对应服务: %+v", subDomain)
		http.Error(w, "未找到服务", http.StatusNotFound)
		return
	}

	// 获取代理结果返回的响应状态码
	responseRecorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
	reverseProxy.proxy.ServeHTTP(responseRecorder, r)
	statusCode := responseRecorder.statusCode

	duration := time.Since(startTime)
	statusSuccess := statusCode >= 200 && statusCode < 300
	statusWarn := statusCode >= 300 && statusCode < 400

	if statusSuccess {
		log.Printf("🔗🟢 %d [%.2f s]", statusCode, duration.Seconds())
	} else if statusWarn {
		log.Printf("🔗🟡 %d [%.2f s]", statusCode, duration.Seconds())
	} else {
		log.Printf("🔗🔴 %d [%.2f s]", statusCode, duration.Seconds())
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
	return parts[0], nil
}
