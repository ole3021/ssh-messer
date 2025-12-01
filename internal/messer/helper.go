package messer

import (
	"fmt"
	"os"
	"sort"
	"ssh-messer/internal/config"
	"ssh-messer/pkg"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

func (s ClientStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting"
	case StatusConnected:
		return "Connected"
	case StatusChecking:
		return "Checking"
	default:
		return "Unknown"
	}
}

func sortSSHHopsByOrderAsc(hops []config.SSHConfig) {
	sort.Slice(hops, func(i, j int) bool {
		return hops[i].Order < hops[j].Order
	})
}

func getSortedSSHClientOrders(clients map[int]*MesserClient) []int {
	orders := make([]int, 0, len(clients))
	for order := range clients {
		orders = append(orders, order)
	}
	sort.Ints(orders)
	return orders
}

func transSSHConfigToDialInfo(sshConfig config.SSHConfig) (string, *ssh.ClientConfig, error) {
	var dialConfig = &ssh.ClientConfig{}

	// ssh.Addr
	if sshConfig.Host == "" || sshConfig.Port == 0 {
		return "", nil, fmt.Errorf("host and port are required for ssh hop config")
	}
	sshAddress := sshConfig.Host + ":" + strconv.Itoa(sshConfig.Port)

	// ssh.User
	if sshConfig.User == "" {
		return "", nil, fmt.Errorf("user is required for ssh hop config")
	}
	dialConfig.User = sshConfig.User

	// ssh.Auth
	if sshConfig.AuthType == "" {
		return "", nil, fmt.Errorf("auth type is required for ssh hop config")
	}

	switch sshConfig.AuthType {
	case "privateKeyWithPassphrase", "privateKey":
		signer, err := parsePrivateKey(sshConfig.PrivateKeyPath, sshConfig.Passphrase)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		dialConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	case "password":
		if sshConfig.Passphrase == "" {
			return "", nil, fmt.Errorf("passphrase is required for password auth")
		}
		dialConfig.Auth = []ssh.AuthMethod{
			ssh.Password(sshConfig.Passphrase),
		}
	default:
		return "", nil, fmt.Errorf("unsupported auth type: %s", sshConfig.AuthType)
	}

	// ssh.Timeout
	timeout := 30 * time.Second
	if sshConfig.TimeoutSec != 0 {
		timeout = time.Duration(sshConfig.TimeoutSec) * time.Second
	}
	dialConfig.Timeout = timeout
	dialConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return sshAddress, dialConfig, nil
}

func parsePrivateKey(keyPath string, passphrase string) (ssh.Signer, error) {
	// Expand home directory
	keyFilePath, err := pkg.ExpandHomeDir(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand private key file path: %v", err)
	}

	// Load private key file
	privateKey, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	// Parse private key
	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key with passphrase: %v", err)
		}
	} else {
		signer, err = ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	return signer, nil
}

func connectSSHHop(sshConfig config.SSHConfig, hopClient *ssh.Client) (*ssh.Client, error) {
	var client *ssh.Client
	var err error

	sshAddress, sshClientConfig, err := transSSHConfigToDialInfo(sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to transform ssh config to dial info: %v", err)
	}

	if hopClient == nil {
		// Connect directly to ssh without hop
		client, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to dial ssh: %v", err)
		}
	} else {
		// Connect through hop client
		tcpConn, err := hopClient.Dial("tcp", sshAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to dial ssh through hop client: %v", err)
		}
		newConn, chans, reqs, err := ssh.NewClientConn(tcpConn, sshAddress, sshClientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create new client conn: %v", err)
		}
		client = ssh.NewClient(newConn, chans, reqs)
	}

	return client, nil
}
