package config_loader

import (
	"ssh-messer/internal/ssh_proxy"
)

// TomlConfig TOML 配置结构（保留原有结构）
type TomlConfig struct {
	Name                    *string                  `toml:"name,omitempty"`
	SSHHops                 []ssh_proxy.SSHHopConfig `toml:"ssh_hops"`
	SSHServices             []ssh_proxy.SSHService   `toml:"services"`
	LocalHttpPort           *string                  `toml:"local_http_port,omitempty"`
	LocalDockerPort         *string                  `toml:"local_docker_port,omitempty"`
	HealthCheckIntervalSecs *int                     `toml:"health_check_interval,omitempty"`
}
