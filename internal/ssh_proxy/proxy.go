package ssh_proxy

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"ssh-messer/internal/pubsub"
	"ssh-messer/pkg"

	"golang.org/x/crypto/ssh"
)

// SSH Hops 跳板状态更新事件
// ------------------------------------------------------------
var statusBroker = pubsub.NewBroker[SSHStatusUpdateEvent]()

type SSHStatusUpdateEvent struct {
	ConfigName string
	Status     SSHProxyStatus
}

func GetStatusBroker() *pubsub.Broker[SSHStatusUpdateEvent] {
	return statusBroker
}

func (p *SSHHopsProxy) publishStatus() {
	update := SSHStatusUpdateEvent{
		ConfigName: p.configName,
		Status:     p.Status,
	}
	statusBroker.Publish(pubsub.UpdatedEvent, update)
}

// updateStatus 封装状态更新和发布逻辑
func (p *SSHHopsProxy) updateStatus(updater func(*SSHProxyStatus)) {
	updater(&p.Status)
	p.publishStatus()
}

// ============================================================

// SSHHopsProxy 函数
// ------------------------------------------------------------
func NewSSHHopsProxy(configName string, hopsConfigs []SSHHopConfig, healthCheckInterval time.Duration, services []SSHService, localPort string) *SSHHopsProxy {
	status := SSHProxyStatus{
		IsConnecting:         false,
		IsConnected:          false,
		ConnectedAt:          time.Time{},
		IsChecking:           false,
		CheckedAt:            time.Time{},
		CurrentInfo:          "",
		LastError:            nil,
		ReconnectAttempts:    0,
		LastReconnectAttempt: time.Time{},
	}

	proxy := &SSHHopsProxy{
		configName:          configName,
		hopsConfigs:         hopsConfigs,
		client:              nil,
		Status:              status,
		serviceProxy:        nil,
		healthStop:          nil,
		services:            services,
		localPort:           localPort,
		healthCheckInterval: healthCheckInterval,
	}

	proxy.publishStatus()

	return proxy
}

func (p *SSHHopsProxy) GetHopsConfigs() []SSHHopConfig {
	return p.hopsConfigs
}

func GetHopDisplayName(hopConfig SSHHopConfig) string {
	if hopConfig.Alias != nil && *hopConfig.Alias != "" {
		return *hopConfig.Alias
	}
	if hopConfig.Host != nil {
		return *hopConfig.Host
	}
	return "Unknown"
}

// ============================================================

// SSH Hops 跳板连接
// ------------------------------------------------------------
func (p *SSHHopsProxy) Connect() {
	pkg.Logger.Debug().Str("config_name", p.configName).Int("hops_count", len(p.hopsConfigs)).Msg("[SSHHopsProxy] 连接开始")
	p.updateStatus(func(s *SSHProxyStatus) {
		s.IsConnecting = true
		s.IsConnected = false
		s.CurrentInfo = "正在连接到 SSH 跳板..."
		s.LastError = nil
	})

	sort.Slice(p.hopsConfigs, func(i, j int) bool {
		orderI := 0
		orderJ := 0
		if p.hopsConfigs[i].Order != nil {
			orderI = *p.hopsConfigs[i].Order
		}
		if p.hopsConfigs[j].Order != nil {
			orderJ = *p.hopsConfigs[j].Order
		}
		return orderI < orderJ
	})

	var currentClient *ssh.Client
	for i, hopConfig := range p.hopsConfigs {
		port := 22
		if hopConfig.Port != nil {
			port = *hopConfig.Port
		}
		sshAddress := *hopConfig.Host + ":" + strconv.Itoa(port)

		var aliasName string
		if hopConfig.Alias != nil && *hopConfig.Alias != "" {
			aliasName = *hopConfig.Alias
		} else {
			aliasName = *hopConfig.Host
		}

		pkg.Logger.Debug().Str("config_name", p.configName).Int("hop_index", i+1).Int("total_hops", len(p.hopsConfigs)).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] 正在连接 hop")

		sshClientConfig, err := transformSSHHopsConfigToSSHClientConfig(hopConfig)
		if err != nil {
			// 关闭已建立的所有连接
			if currentClient != nil {
				currentClient.Close()
			}
			pkg.Logger.Error().Err(err).Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Msg("[SSHHopsProxy] 配置 SSH 跳板失败")
			p.updateStatus(func(s *SSHProxyStatus) {
				s.LastError = err
				s.CurrentInfo = fmt.Sprintf("配置 SSH 跳板 %s 失败: %v", aliasName, err)
				s.IsConnecting = false
				s.IsConnected = false
			})
			return
		}

		if i == 0 {
			// 第一个 hop：直接连接到第一台服务器
			p.updateStatus(func(s *SSHProxyStatus) {
				s.CurrentInfo = fmt.Sprintf("正在连接到 SSH 跳板 %d/%d: %s", i+1, len(p.hopsConfigs), aliasName)
			})

			currentClient, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
			if err != nil {
				pkg.Logger.Error().Err(err).Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] 连接 hop 失败")
				p.updateStatus(func(s *SSHProxyStatus) {
					s.LastError = err
					s.CurrentInfo = fmt.Sprintf("连接到 SSH 跳板 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
					s.IsConnecting = false
					s.IsConnected = false
				})
				return
			}
			pkg.Logger.Info().Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] hop 连接成功")
		} else {
			// 后续 hop：通过前一个 client 的 Dial 方法连接到下一台服务器
			p.updateStatus(func(s *SSHProxyStatus) {
				s.CurrentInfo = fmt.Sprintf("正在连接到 SSH 跳板 %d/%d: %s", i+1, len(p.hopsConfigs), aliasName)
			})

			// 通过前一个 client 建立到下一个 hop 的 TCP 连接
			conn, err := currentClient.Dial("tcp", sshAddress)
			if err != nil {
				// 关闭已建立的所有连接
				if currentClient != nil {
					currentClient.Close()
				}
				pkg.Logger.Error().Err(err).Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] 通过前一个 hop 连接失败")
				p.updateStatus(func(s *SSHProxyStatus) {
					s.LastError = err
					s.CurrentInfo = fmt.Sprintf("连接到 SSH 跳板 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
					s.IsConnecting = false
					s.IsConnected = false
				})
				return
			}

			// 基于这个 TCP 连接创建新的 SSH 客户端连接
			nconn, chans, reqs, err := ssh.NewClientConn(conn, sshAddress, sshClientConfig)
			if err != nil {
				conn.Close()
				if currentClient != nil {
					currentClient.Close()
				}
				pkg.Logger.Error().Err(err).Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] 创建 SSH 客户端连接失败")
				p.updateStatus(func(s *SSHProxyStatus) {
					s.LastError = err
					s.CurrentInfo = fmt.Sprintf("创建 SSH 客户端连接 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
					s.IsConnecting = false
					s.IsConnected = false
				})
				return
			}

			// 创建新的 SSH 客户端（前一个 client 会自动通过连接链保持）
			// 注意：不要关闭 currentClient，因为它被新 client 使用
			currentClient = ssh.NewClient(nconn, chans, reqs)
			pkg.Logger.Info().Str("config_name", p.configName).Int("hop_index", i+1).Str("alias", aliasName).Str("address", sshAddress).Msg("[SSHHopsProxy] hop 连接成功")
		}
	}

	// 保存最终的 client（最后一个 hop 的 client）
	p.client = currentClient

	// 所有跳板连接成功
	pkg.Logger.Info().Str("config_name", p.configName).Int("total_hops", len(p.hopsConfigs)).Msg("[SSHHopsProxy] 所有 hop 连接成功")
	p.updateStatus(func(s *SSHProxyStatus) {
		s.CurrentInfo = ""
		s.IsConnecting = false
		s.IsConnected = true
		s.LastError = nil
		s.ConnectedAt = time.Now()
	})

	// 启动健康检查循环
	p.StartHealthCheck()

	// 如果配置了 services 和 localPort，自动启动 services 代理
	if len(p.services) > 0 && p.localPort != "" {
		if err := p.StartServices(p.services, p.localPort); err != nil {
			pkg.Logger.Error().Err(err).Str("config_name", p.configName).Msg("[SSHHopsProxy] 连接后启动服务代理失败")
		}
	}
}

func (p *SSHHopsProxy) Disconnect() {
	pkg.Logger.Debug().Str("config_name", p.configName).Msg("[SSHHopsProxy] 断开连接")
	// 停止健康检查
	p.stopHealthCheck()

	// 停止 services 代理
	p.StopServices()

	if p.client != nil {
		p.client.Close()
		p.client = nil
	}

	p.updateStatus(func(s *SSHProxyStatus) {
		s.IsConnecting = false
		s.IsConnected = false
		s.ConnectedAt = time.Time{}
		s.IsChecking = false
		s.CheckedAt = time.Time{}
		s.CurrentInfo = ""
		s.LastError = nil
	})
}

// ============================================================

// SSH Client 连接状态检查。
// ------------------------------------------------------------
func (p *SSHHopsProxy) StartHealthCheck() {
	if p.healthStop != nil {
		return // 已经在运行
	}

	pkg.Logger.Debug().Str("config_name", p.configName).Dur("interval", p.healthCheckInterval).Msg("[SSHHopsProxy] 启动健康检查")
	p.healthStop = make(chan struct{})
	go p.healthCheckLoop()
}

func (p *SSHHopsProxy) healthCheckLoop() {
	ticker := time.NewTicker(p.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.healthStop:
			return
		case <-ticker.C:
			p.CheckHealth()
		}
	}
}

func (p *SSHHopsProxy) CheckHealth() {
	if p.client == nil {
		return
	}

	p.updateStatus(func(s *SSHProxyStatus) {
		s.IsChecking = true
		s.CurrentInfo = "正在检查 SSH 跳板健康状态..."
	})

	// 检查连接是否正常
	conn := p.client.Conn
	if conn == nil {
		pkg.Logger.Debug().Str("config_name", p.configName).Msg("[SSHHopsProxy] 健康检查失败: SSH 连接已断开")
		p.updateStatus(func(s *SSHProxyStatus) {
			s.LastError = fmt.Errorf("SSH connection is disconnected")
			s.CurrentInfo = ""
			s.IsChecking = false
			s.CheckedAt = time.Now()
			s.IsConnected = false
		})
		// 触发重连
		go p.Reconnect()
		return
	}

	// 尝试创建SSH会话
	session, err := p.client.NewSession()
	if err != nil {
		pkg.Logger.Debug().Err(err).Str("config_name", p.configName).Msg("[SSHHopsProxy] 健康检查失败: 无法创建 SSH 会话")
		p.updateStatus(func(s *SSHProxyStatus) {
			s.LastError = fmt.Errorf("failed to create SSH session: %v", err)
			s.CurrentInfo = ""
			s.IsChecking = false
			s.CheckedAt = time.Now()
			s.IsConnected = false
		})
		// 触发重连
		go p.Reconnect()
		return
	}
	defer session.Close()

	// 执行简单的命令测试连接
	err = session.Run("echo 'health_check'")
	if err != nil {
		pkg.Logger.Debug().Err(err).Str("config_name", p.configName).Msg("[SSHHopsProxy] 健康检查失败: 无法执行 SSH 命令")
		p.updateStatus(func(s *SSHProxyStatus) {
			s.LastError = fmt.Errorf("failed to execute SSH command: %v", err)
			s.CurrentInfo = ""
			s.IsChecking = false
			s.CheckedAt = time.Now()
			s.IsConnected = false
		})
		// 触发重连
		go p.Reconnect()
		return
	}

	pkg.Logger.Debug().Str("config_name", p.configName).Msg("[SSHHopsProxy] 健康检查成功")
	p.updateStatus(func(s *SSHProxyStatus) {
		s.CurrentInfo = ""
		s.IsChecking = false
		s.CheckedAt = time.Now()
		s.IsConnected = true
		s.LastError = nil
	})
}

func (p *SSHHopsProxy) stopHealthCheck() {
	if p.healthStop != nil {
		select {
		case <-p.healthStop:
			// Already closed
		default:
			close(p.healthStop)
		}
		p.healthStop = nil
	}
}

// ============================================================

// Reconnect 重连逻辑（立即重连，无退避策略）
func (p *SSHHopsProxy) Reconnect() {
	if p.Status.IsConnecting {
		return // 已经在重连中
	}

	p.updateStatus(func(s *SSHProxyStatus) {
		s.ReconnectAttempts++
		s.LastReconnectAttempt = time.Now()
		s.CurrentInfo = fmt.Sprintf("正在重连 (尝试 %d)...", s.ReconnectAttempts)
	})

	pkg.Logger.Info().Str("config_name", p.configName).Int("attempt", p.Status.ReconnectAttempts+1).Msg("[SSHHopsProxy] 开始重连")

	// 停止当前的 services 代理
	p.StopServices()

	// 关闭当前连接
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}

	// 重新连接（Connect() 会自动启动 services 代理）
	p.Connect()
}

// StartServices 启动 services 代理
func (p *SSHHopsProxy) StartServices(services []SSHService, localPort string) error {
	if p.client == nil {
		return fmt.Errorf("SSH client is not connected")
	}

	// 保存 services 和 localPort 以便重连后重新启动
	p.services = services
	p.localPort = localPort

	// 停止现有的 services 代理
	p.StopServices()

	// 创建新的 services 代理
	sp := NewServiceProxy(p.configName, localPort, services, p.client)
	p.serviceProxy = sp

	// 启动 services 代理
	return sp.StartReverseProxy()
}

// StopServices 停止 services 代理
func (p *SSHHopsProxy) StopServices() {
	if p.serviceProxy != nil {
		p.serviceProxy.StopReverseProxy()
		p.serviceProxy = nil
	}
}
