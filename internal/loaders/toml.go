package loaders

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type TomlConfig struct {
	SSHHops         []TomlConfigSSH     `toml:"ssh_hops"`
	Services        []TomlConfigService `toml:"services"`
	LocalHttpPort   *string             `toml:"local_http_port,omitempty"`
	LocalDockerPort *string             `toml:"local_docker_port,omitempty"`
}

type TomlConfigSSH struct {
	Order          *int    `toml:"order"`
	Host           *string `toml:"host"`
	Port           *int    `toml:"port"`
	AuthType       *string `toml:"authType"`
	PrivateKeyPath *string `toml:"privateKeyPath"`
	Passphrase     *string `toml:"passphrase"`
	User           *string `toml:"user"`
	Alias          *string `toml:"alias,omitempty"`
	TimeoutSec     *int    `toml:"timeoutSec,omitempty"`
}

type TomlConfigService struct {
	Host      *string `toml:"host"`
	Port      *string `toml:"port"`
	Subdomain *string `toml:"subdomain"`
	Alias     *string `toml:"alias,omitempty"`
}

func LoadTomlProxyConfig(filename string, dir ...string) (*TomlConfig, error) {
	var fullPath string

	if len(dir) > 0 {
		fullPath = filepath.Join(dir[0], filename)
	} else {
		fullPath = filepath.Join("configs", filename)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("📄❔ 配置文件不存在: %s", fullPath)
	}

	var proxyConfig TomlConfig

	// 加载 TOML 文件
	if _, err := toml.DecodeFile(fullPath, &proxyConfig); err != nil {
		log.Fatal(err)

		return nil, err
	}

	return &proxyConfig, nil
}
