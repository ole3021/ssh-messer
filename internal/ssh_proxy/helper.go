package ssh_proxy

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func transformSSHHopsConfigToSSHClientConfig(sshHopConfig SSHHopConfig) (*ssh.ClientConfig, error) {
	var clientConfig = &ssh.ClientConfig{}

	if sshHopConfig.User == nil {
		return nil, fmt.Errorf("user is required for ssh hop config")
	}
	clientConfig.User = *sshHopConfig.User

	if sshHopConfig.AuthType == nil {
		return nil, fmt.Errorf("auth type is required for ssh hop config: %+v", sshHopConfig)
	}

	switch *sshHopConfig.AuthType {
	case "privateKeyWithPassphrase", "privateKey":
		signer, err := parsePrivateKey(sshHopConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	case "password":
		if sshHopConfig.Passphrase == nil {
			return nil, fmt.Errorf("passphrase is required for password auth")
		}
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*sshHopConfig.Passphrase),
		}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", *sshHopConfig.AuthType)
	}

	timeout := 30 * time.Second
	if sshHopConfig.TimeoutSec != nil {
		timeout = time.Duration(*sshHopConfig.TimeoutSec) * time.Second
	}
	clientConfig.Timeout = timeout
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return clientConfig, nil
}

func parsePrivateKey(sshHopConfig SSHHopConfig) (ssh.Signer, error) {
	if sshHopConfig.PrivateKeyPath == nil {
		return nil, fmt.Errorf("private key path is required")
	}

	privateKey, err := os.ReadFile(*sshHopConfig.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	if sshHopConfig.Passphrase != nil {
		return ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(*sshHopConfig.Passphrase))
	}
	return ssh.ParsePrivateKey(privateKey)
}
