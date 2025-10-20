package proxy

import (
	"fmt"
	"log"
	"os"
	"sort"
	"ssh-messer/internal/loaders"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// Create sshHopsClien
func CreateSSHHopsClient(sshHopsConfigs []loaders.TomlConfigSSH) (*ssh.Client, error) {
	var client *ssh.Client

	// 对sshHopsConfigs 按照order 从小到大进行排序
	sort.Slice(sshHopsConfigs, func(i, j int) bool {
		return *sshHopsConfigs[i].Order < *sshHopsConfigs[j].Order
	})

	log.Printf("🦘 正在连接到 SSH 跳板...")

	for i, sshHopConfig := range sshHopsConfigs {
		sshAddress := *sshHopConfig.Host + ":" + strconv.Itoa(*sshHopConfig.Port|22)
		var aliasName string
		if sshHopConfig.Alias != nil {
			aliasName = *sshHopConfig.Alias
		} else {
			aliasName = sshAddress
		}

		sshClientConfig, err := TransformSSHClientConfig(sshHopConfig)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			// 第一跳：直接连接
			client, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
			if err != nil {
				log.Fatalf("🦘❌ SSH连接 [%s] 失败: %v", sshAddress, err)
				client.Close()
				return nil, err
			}
		} else {
			// 后续跳：通过隧道连接
			conn, err := client.Dial("tcp", sshAddress)
			if err != nil {
				client.Close()
				return nil, err
			}

			nconn, chans, reqs, err := ssh.NewClientConn(conn, sshAddress, sshClientConfig)
			if err != nil {
				log.Fatalf("🦘❌ SSH连接 [%s] 失败: %v", sshAddress, err)
				conn.Close()
				client.Close()
				return nil, err
			}

			client = ssh.NewClient(nconn, chans, reqs)

		}
		log.Printf("🦘 成功连接到 [%s]", aliasName)
	}
	return client, nil
}

type SSHClientResultChan struct {
	Client *ssh.Client
	Error  error
}

type SSHProcessChan struct {
	TotalHopsCount     int
	CompletedHopsCount int
	Message            string
	Error              error
}

func AsyncCreateSSHHopsClient(sshHopsConfigs []loaders.TomlConfigSSH, sshClientChan chan SSHClientResultChan, sshProcessChan *chan SSHProcessChan) {
	var client *ssh.Client

	log.Printf("🦘 开始跳转 SSH: [%v]", sshHopsConfigs)

	// 对sshHopsConfigs 按照order 从小到大进行排序
	sort.Slice(sshHopsConfigs, func(i, j int) bool {
		return *sshHopsConfigs[i].Order < *sshHopsConfigs[j].Order
	})

	for i, sshHopConfig := range sshHopsConfigs {
		// 正确的写法
		port := 22
		if sshHopConfig.Port != nil && *sshHopConfig.Port != 0 {
			port = *sshHopConfig.Port
		}
		sshAddress := *sshHopConfig.Host + ":" + strconv.Itoa(port)
		var aliasName string
		if sshHopConfig.Alias != nil {
			aliasName = *sshHopConfig.Alias
		} else {
			aliasName = sshAddress
		}

		sshClientConfig, err := TransformSSHClientConfig(sshHopConfig)
		if err != nil {
			*sshProcessChan <- SSHProcessChan{
				TotalHopsCount:     len(sshHopsConfigs),
				CompletedHopsCount: i,
				Message:            fmt.Sprintf("SSH配置:[%s] 转换失败: %v", aliasName, err),
				Error:              err,
			}
			sshClientChan <- SSHClientResultChan{
				Client: nil,
				Error:  err,
			}
			return
		}

		*sshProcessChan <- SSHProcessChan{
			TotalHopsCount:     len(sshHopsConfigs),
			CompletedHopsCount: i,
			Message:            fmt.Sprintf("🦘 [%v/%v] 正在跳转 SSH: [%s]", i+1, len(sshHopsConfigs), aliasName),
			Error:              nil,
		}

		if i == 0 {
			// 第一跳：直接连接
			client, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
			if err != nil {
				*sshProcessChan <- SSHProcessChan{
					TotalHopsCount:     len(sshHopsConfigs),
					CompletedHopsCount: i,
					Message:            fmt.Sprintf("SSH连接 [%s] 失败: %v", aliasName, err),
					Error:              err,
				}
				sshClientChan <- SSHClientResultChan{
					Client: nil,
					Error:  err,
				}
				return
			}
		} else {
			// 后续跳：通过隧道连接
			conn, err := client.Dial("tcp", sshAddress)
			if err != nil {
				client.Close()
				*sshProcessChan <- SSHProcessChan{
					TotalHopsCount:     len(sshHopsConfigs),
					CompletedHopsCount: i,
					Message:            fmt.Sprintf("SSH 隧道连接 [%s] 失败: %v", aliasName, err),
					Error:              err,
				}
				sshClientChan <- SSHClientResultChan{
					Client: nil,
					Error:  err,
				}
				return
			}

			nconn, chans, reqs, err := ssh.NewClientConn(conn, sshAddress, sshClientConfig)
			if err != nil {
				conn.Close()
				client.Close()
				*sshProcessChan <- SSHProcessChan{
					TotalHopsCount:     len(sshHopsConfigs),
					CompletedHopsCount: i,
					Message:            fmt.Sprintf("SSH 隧道连接 [%s] 失败: %v", aliasName, err),
					Error:              err,
				}
				sshClientChan <- SSHClientResultChan{
					Client: nil,
					Error:  err,
				}
				return
			}

			client = ssh.NewClient(nconn, chans, reqs)
		}

	}
	*sshProcessChan <- SSHProcessChan{
		TotalHopsCount:     len(sshHopsConfigs),
		CompletedHopsCount: len(sshHopsConfigs),
		Message:            "",
		Error:              nil,
	}
	sshClientChan <- SSHClientResultChan{
		Client: client,
		Error:  nil,
	}
}

// Convert sshHopConfig to sshClientConfig
func TransformSSHClientConfig(sshHopConfig loaders.TomlConfigSSH) (*ssh.ClientConfig, error) {
	var clientConfig = &ssh.ClientConfig{
		User: *sshHopConfig.User,
	}

	if sshHopConfig.AuthType == nil {
		return nil, fmt.Errorf("!!! Auth type is required")
	}

	switch *sshHopConfig.AuthType {
	case "privateKeyWithPassphrase":
		privateKey, err := os.ReadFile(*sshHopConfig.PrivateKeyPath)
		if err != nil {
			log.Fatalf("!!! Failed to read private key file: %v", err)
			return nil, err
		}

		signer, err := ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(*sshHopConfig.Passphrase))
		if err != nil {
			log.Fatalf("Failed to parse private key: %v", err)
			return nil, err
		}
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	case "password":
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*sshHopConfig.Passphrase),
		}
	default:
		return nil, fmt.Errorf("!!! Unsupported auth type: %s", *sshHopConfig.AuthType)
	}

	clientConfig.Timeout = time.Duration(*sshHopConfig.TimeoutSec) * time.Second
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return clientConfig, nil
}
