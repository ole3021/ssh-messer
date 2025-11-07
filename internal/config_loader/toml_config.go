package config_loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const homeConfigFolder = ".ssh_messer"

// LoadTomlProxyConfig 加载单个 TOML 配置文件
func LoadTomlProxyConfig(filename string, dir ...string) (*TomlConfig, error) {
	var fullPath string

	if len(dir) > 0 {
		fullPath = filepath.Join(dir[0], filename)
	} else {
		fullPath = filepath.Join("configs", filename)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", fullPath)
	}

	var proxyConfig TomlConfig

	// 加载 TOML 文件
	if _, err := toml.DecodeFile(fullPath, &proxyConfig); err != nil {
		return nil, fmt.Errorf("failed to decode TOML file %s: %w", fullPath, err)
	}

	return &proxyConfig, nil
}

// LoadTomlConfigsFromHomeDir 从用户主目录加载所有 TOML 配置
func LoadTomlConfigsFromHomeDir() (map[string]*TomlConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, homeConfigFolder)
	configs := make(map[string]*TomlConfig)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return configs, nil
	}

	files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
	if err != nil {
		return configs, nil
	}

	for _, file := range files {
		config, err := LoadTomlProxyConfig(filepath.Base(file), configDir)
		if err != nil {
			continue // 跳过有错误的文件
		}

		configs[filepath.Base(file)] = config
	}

	return configs, nil
}
