package config

type SSHConfig struct {
	Order          int    `toml:"order"`
	Host           string `toml:"host"`
	Port           int    `toml:"port"`
	User           string `toml:"user"`
	AuthType       string `toml:"authType"`
	PrivateKeyPath string `toml:"privateKeyPath,omitempty"`
	Passphrase     string `toml:"passphrase,omitempty"`
	Password       string `toml:"password,omitempty"`
	Name           string `toml:"name,omitempty"`
	TimeoutSec     int    `toml:"timeoutSec,omitempty"`
}

type ReverseServicePageConfig struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

type ReverseServiceConfig struct {
	Host           string                     `toml:"host"`
	Port           string                     `toml:"port"`
	Subdomain      string                     `toml:"subdomain"`
	Name           string                     `toml:"name,omitempty"`
	UseTLS         bool                       `toml:"use_tls,omitempty"`
	TLSServerName  string                     `toml:"tls_server_name,omitempty"`
	RemoteHost     string                     `toml:"remote_host,omitempty"`
	Pages          []ReverseServicePageConfig `toml:"pages,omitempty"`
	CustomHopOrder int                        `toml:"hopOrder,omitempty"`
}

type MesserConfig struct {
	Name                    string                 `toml:"name,omitempty"`
	Version                 string                 `toml:"version,omitempty"`
	LocalHttpPort           string                 `toml:"local_http_port,omitempty"`
	LocalDockerPort         string                 `toml:"local_docker_port,omitempty"`
	HealthCheckIntervalSecs int                    `toml:"health_check_interval_secs,omitempty"`
	SSHHops                 []SSHConfig            `toml:"ssh_hops"`
	ReverseServices         []ReverseServiceConfig `toml:"services"`
}
