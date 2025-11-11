package ssh_proxy

import (
	"fmt"
	"net"
	"os"
	"time"

	"ssh-messer/pkg"

	"golang.org/x/crypto/ssh"
)

func transformSSHHopsConfigToSSHClientConfig(sshHopConfig SSHHopConfig) (*ssh.ClientConfig, error) {
	var aliasName string
	if sshHopConfig.Alias != nil && *sshHopConfig.Alias != "" {
		aliasName = *sshHopConfig.Alias
	} else if sshHopConfig.Host != nil {
		aliasName = *sshHopConfig.Host
	} else {
		aliasName = "Unknown"
	}

	authType := "unknown"
	if sshHopConfig.AuthType != nil {
		authType = *sshHopConfig.AuthType
	}
	pkg.Logger.Debug().Str("alias", aliasName).Str("auth_type", authType).Msg("[SSHHelper] 开始转换 SSH 配置")

	var clientConfig = &ssh.ClientConfig{}

	if sshHopConfig.User == nil {
		pkg.Logger.Error().Str("alias", aliasName).Msg("[SSHHelper] 配置转换失败: 缺少用户")
		return nil, fmt.Errorf("user is required for ssh hop config")
	}
	clientConfig.User = *sshHopConfig.User

	if sshHopConfig.AuthType == nil {
		pkg.Logger.Error().Str("alias", aliasName).Msg("[SSHHelper] 配置转换失败: 缺少认证类型")
		return nil, fmt.Errorf("auth type is required for ssh hop config: %+v", sshHopConfig)
	}

	switch *sshHopConfig.AuthType {
	case "privateKeyWithPassphrase", "privateKey":
		signer, err := parsePrivateKey(sshHopConfig)
		if err != nil {
			pkg.Logger.Error().Err(err).Str("alias", aliasName).Msg("[SSHHelper] 配置转换失败: 私钥解析失败")
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	case "password":
		if sshHopConfig.Passphrase == nil {
			pkg.Logger.Error().Str("alias", aliasName).Msg("[SSHHelper] 配置转换失败: 密码认证缺少密码")
			return nil, fmt.Errorf("passphrase is required for password auth")
		}
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*sshHopConfig.Passphrase),
		}
	default:
		pkg.Logger.Error().Str("alias", aliasName).Str("auth_type", *sshHopConfig.AuthType).Msg("[SSHHelper] 配置转换失败: 不支持的认证类型")
		return nil, fmt.Errorf("unsupported auth type: %s", *sshHopConfig.AuthType)
	}

	timeout := 30 * time.Second
	if sshHopConfig.TimeoutSec != nil {
		timeout = time.Duration(*sshHopConfig.TimeoutSec) * time.Second
	}
	clientConfig.Timeout = timeout
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	pkg.Logger.Debug().Str("alias", aliasName).Str("user", clientConfig.User).Dur("timeout", timeout).Msg("[SSHHelper] SSH 配置转换成功")
	return clientConfig, nil
}

func parsePrivateKey(sshHopConfig SSHHopConfig) (ssh.Signer, error) {
	var aliasName string
	if sshHopConfig.Alias != nil && *sshHopConfig.Alias != "" {
		aliasName = *sshHopConfig.Alias
	} else if sshHopConfig.Host != nil {
		aliasName = *sshHopConfig.Host
	} else {
		aliasName = "Unknown"
	}

	if sshHopConfig.PrivateKeyPath == nil {
		pkg.Logger.Error().Str("alias", aliasName).Msg("[SSHHelper] 私钥解析失败: 缺少私钥路径")
		return nil, fmt.Errorf("private key path is required")
	}

	pkg.Logger.Debug().Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 开始解析私钥")

	privateKey, err := os.ReadFile(*sshHopConfig.PrivateKeyPath)
	if err != nil {
		pkg.Logger.Error().Err(err).Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 私钥解析失败: 无法读取私钥文件")
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	var signer ssh.Signer
	if sshHopConfig.Passphrase != nil {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(*sshHopConfig.Passphrase))
		if err != nil {
			pkg.Logger.Error().Err(err).Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 私钥解析失败: 带密码的私钥解析失败")
			return nil, err
		}
		pkg.Logger.Debug().Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 私钥解析成功: 带密码的私钥")
	} else {
		signer, err = ssh.ParsePrivateKey(privateKey)
		if err != nil {
			pkg.Logger.Error().Err(err).Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 私钥解析失败: 无密码私钥解析失败")
			return nil, err
		}
		pkg.Logger.Debug().Str("alias", aliasName).Str("key_path", *sshHopConfig.PrivateKeyPath).Msg("[SSHHelper] 私钥解析成功: 无密码私钥")
	}

	return signer, nil
}

// CheckPortAvailable 检查指定端口是否可用（未被占用）
func CheckPortAvailable(port string) error {
	if port == "" {
		return nil // 如果端口为空，跳过检查
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("端口 %s 已被占用: %w", port, err)
	}
	listener.Close()
	return nil
}
