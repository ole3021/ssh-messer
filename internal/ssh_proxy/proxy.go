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

var statusBroker = pubsub.NewBroker[SSHStatusUpdate]()

type SSHStatusUpdate struct {
	ConfigName string
	Status     SSHProxyStatus
}

func GetStatusBroker() *pubsub.Broker[SSHStatusUpdate] {
	return statusBroker
}

func NewSSHHopsProxy(configName string, hopsConfigs []SSHHopConfig) *SSHHopsProxy {
	status := SSHProxyStatus{
		IsConnecting: false,
		IsConnected:  false,
		ConnectedAt:  time.Time{},
		IsChecking:   false,
		CheckedAt:    time.Time{},
		CurrentInfo:  "",
		LastError:    nil,
	}

	proxy := &SSHHopsProxy{
		configName:  configName,
		hopsConfigs: hopsConfigs,
		client:      nil,
		Status:      status,
	}

	proxy.publishStatus()

	return proxy
}

func (p *SSHHopsProxy) publishStatus() {
	update := SSHStatusUpdate{
		ConfigName: p.configName,
		Status:     p.Status,
	}
	statusBroker.Publish(pubsub.UpdatedEvent, update)
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

func (p *SSHHopsProxy) Connect() {
	pkg.Logger.Trace().Msg("[SSHHopsProxy] Connect Start")
	p.Status.IsConnecting = true
	p.Status.IsConnected = false
	p.Status.CurrentInfo = "正在连接到 SSH 跳板..."
	p.Status.LastError = nil
	p.publishStatus()

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
		pkg.Logger.Trace().Int("index", i).Str("hopConfig", fmt.Sprintf("%+v", hopConfig)).Msg("[SSHHopsProxy] Connecting...")

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

		sshClientConfig, err := transformSSHHopsConfigToSSHClientConfig(hopConfig)
		if err != nil {
			// 关闭已建立的所有连接
			if currentClient != nil {
				currentClient.Close()
			}
			p.Status.LastError = err
			p.Status.CurrentInfo = fmt.Sprintf("配置 SSH 跳板 %s 失败: %v", aliasName, err)
			p.Status.IsConnecting = false
			p.Status.IsConnected = false
			p.publishStatus()
			return
		}

		if i == 0 {
			// 第一个 hop：直接连接到第一台服务器
			p.Status.CurrentInfo = fmt.Sprintf("正在连接到 SSH 跳板 %d/%d: %s", i+1, len(p.hopsConfigs), aliasName)
			p.publishStatus()

			currentClient, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
			if err != nil {
				p.Status.LastError = err
				p.Status.CurrentInfo = fmt.Sprintf("连接到 SSH 跳板 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
				p.Status.IsConnecting = false
				p.Status.IsConnected = false
				p.publishStatus()
				return
			}
		} else {
			// 后续 hop：通过前一个 client 的 Dial 方法连接到下一台服务器
			p.Status.CurrentInfo = fmt.Sprintf("正在连接到 SSH 跳板 %d/%d: %s", i+1, len(p.hopsConfigs), aliasName)
			p.publishStatus()

			// 通过前一个 client 建立到下一个 hop 的 TCP 连接
			conn, err := currentClient.Dial("tcp", sshAddress)
			if err != nil {
				// 关闭已建立的所有连接
				if currentClient != nil {
					currentClient.Close()
				}
				p.Status.LastError = err
				p.Status.CurrentInfo = fmt.Sprintf("连接到 SSH 跳板 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
				p.Status.IsConnecting = false
				p.Status.IsConnected = false
				p.publishStatus()
				return
			}

			// 基于这个 TCP 连接创建新的 SSH 客户端连接
			nconn, chans, reqs, err := ssh.NewClientConn(conn, sshAddress, sshClientConfig)
			if err != nil {
				conn.Close()
				if currentClient != nil {
					currentClient.Close()
				}
				p.Status.LastError = err
				p.Status.CurrentInfo = fmt.Sprintf("创建 SSH 客户端连接 %d/%d: %s 失败", i+1, len(p.hopsConfigs), aliasName)
				p.Status.IsConnecting = false
				p.Status.IsConnected = false
				p.publishStatus()
				return
			}

			// 创建新的 SSH 客户端（前一个 client 会自动通过连接链保持）
			// 注意：不要关闭 currentClient，因为它被新 client 使用
			currentClient = ssh.NewClient(nconn, chans, reqs)
		}
	}

	// 保存最终的 client（最后一个 hop 的 client）
	p.client = currentClient

	// 所有 hop 连接成功
	lastHopName := ""
	if len(p.hopsConfigs) > 0 {
		lastConfig := p.hopsConfigs[len(p.hopsConfigs)-1]
		if lastConfig.Alias != nil && *lastConfig.Alias != "" {
			lastHopName = *lastConfig.Alias
		} else {
			lastHopName = *lastConfig.Host
		}
	}

	p.Status.CurrentInfo = fmt.Sprintf("已连接到所有 SSH 跳板 (%d/%d)，最终跳板: %s", len(p.hopsConfigs), len(p.hopsConfigs), lastHopName)
	p.Status.IsConnecting = false
	p.Status.IsConnected = true
	p.Status.LastError = nil
	p.Status.ConnectedAt = time.Now()
	p.publishStatus()
}

// Disconnect 断开 SSH 连接
func (p *SSHHopsProxy) Disconnect() {
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}

	p.Status.IsConnecting = false
	p.Status.IsConnected = false
	p.Status.ConnectedAt = time.Time{}
	p.Status.IsChecking = false
	p.Status.CheckedAt = time.Time{}
	p.Status.CurrentInfo = ""
	p.Status.LastError = nil
	p.publishStatus()
}

// CheckHealth 检查 SSH 连接健康状态
func (p *SSHHopsProxy) CheckHealth() {
	if p.client == nil {
		return
	}

	p.Status.IsChecking = true
	p.Status.CurrentInfo = "正在检查 SSH 跳板健康状态..."
	p.publishStatus()

	// 检查连接是否正常
	conn := p.client.Conn
	if conn == nil {
		p.Status.LastError = fmt.Errorf("SSH connection is disconnected")
		p.Status.CurrentInfo = ""
		p.Status.IsChecking = false
		p.Status.CheckedAt = time.Now()
		p.Status.IsConnected = false
		p.publishStatus()
		return
	}

	// 尝试创建SSH会话
	session, err := p.client.NewSession()
	if err != nil {
		p.Status.LastError = fmt.Errorf("failed to create SSH session: %v", err)
		p.Status.CurrentInfo = ""
		p.Status.IsChecking = false
		p.Status.CheckedAt = time.Now()
		p.Status.IsConnected = false
		p.publishStatus()
		return
	}
	defer session.Close()

	// 执行简单的命令测试连接
	err = session.Run("echo 'health_check'")
	if err != nil {
		p.Status.LastError = fmt.Errorf("failed to execute SSH command: %v", err)
		p.Status.CurrentInfo = ""
		p.Status.IsChecking = false
		p.Status.CheckedAt = time.Now()
		p.Status.IsConnected = false
		p.publishStatus()
		return
	}

	p.Status.CurrentInfo = "连接健康"
	p.Status.IsChecking = false
	p.Status.CheckedAt = time.Now()
	p.Status.IsConnected = true
	p.Status.LastError = nil
	p.publishStatus()
}
