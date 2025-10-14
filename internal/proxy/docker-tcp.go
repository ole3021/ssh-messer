// 在 proxy 包中创建 docker-tcp-proxy.go
package proxy

import (
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type DockerTCPProxy struct {
	sshClient    *ssh.Client
	listener     net.Listener
	localPort    string
	remoteSocket string
	isRunning    bool
}

func NewDockerTCPProxy(sshClient *ssh.Client, localPort, remoteSocket string) *DockerTCPProxy {
	return &DockerTCPProxy{
		sshClient:    sshClient,
		localPort:    localPort,
		remoteSocket: remoteSocket,
		isRunning:    false,
	}
}

func (d *DockerTCPProxy) Start() error {
	// 创建TCP监听器
	listener, err := net.Listen("tcp", ":"+d.localPort)
	if err != nil {
		return fmt.Errorf("🐳🔴 创建TCP监听器失败: %v", err)
	}
	d.listener = listener
	d.isRunning = true

	fmt.Printf("🐳 本地端口: %s -> 远程Socket: %s\n", d.localPort, d.remoteSocket)

	// 处理连接
	for d.isRunning {
		conn, err := listener.Accept()
		if err != nil {
			if d.isRunning {
				log.Printf("🐳🔴 接受连接失败: %v", err)
			}
			continue
		}

		go d.handleConnection(conn)
	}

	return nil
}

func (d *DockerTCPProxy) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	// 通过SSH隧道连接到远程Docker Socket
	remoteConn, err := d.sshClient.Dial("unix", d.remoteSocket)
	if err != nil {
		log.Printf("🐳🔴 连接远程Docker Socket失败: %v", err)
		return
	}
	defer remoteConn.Close()

	// 双向数据转发
	done := make(chan struct{}, 2)

	// 本地 -> 远程
	go func() {
		defer func() { done <- struct{}{} }()
		_, err := io.Copy(remoteConn, localConn)
		if err != nil {
			log.Printf("🐳🔴 本地到远程数据转发失败: %v", err)
		}
	}()

	// 远程 -> 本地
	go func() {
		defer func() { done <- struct{}{} }()
		_, err := io.Copy(localConn, remoteConn)
		if err != nil {
			log.Printf("🐳🔴 远程到本地数据转发失败: %v", err)
		}
	}()

	// 等待任一方向完成
	<-done
}

func (d *DockerTCPProxy) Stop() error {
	d.isRunning = false
	if d.listener != nil {
		return d.listener.Close()
	}
	return nil
}
