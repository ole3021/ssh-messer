package config_loader

import (
	"fmt"
	"os"
	"path/filepath"

	"ssh-messer/pkg"

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

	pkg.Logger.Debug().Str("file", fullPath).Msg("[ConfigLoader] 开始加载配置文件")

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		pkg.Logger.Error().Str("file", fullPath).Msg("[ConfigLoader] 配置文件不存在")
		return nil, fmt.Errorf("配置文件不存在: %s", fullPath)
	}

	var proxyConfig TomlConfig

	// 加载 TOML 文件
	if _, err := toml.DecodeFile(fullPath, &proxyConfig); err != nil {
		pkg.Logger.Error().Err(err).Str("file", fullPath).Msg("[ConfigLoader] 配置文件加载失败")
		return nil, fmt.Errorf("failed to decode TOML file %s: %w", fullPath, err)
	}

	pkg.Logger.Info().Str("file", fullPath).Msg("[ConfigLoader] 配置文件加载成功")
	return &proxyConfig, nil
}

// LoadTomlConfigsFromHomeDir 从用户主目录加载所有 TOML 配置
func LoadTomlConfigsFromHomeDir() (map[string]*TomlConfig, error) {
	pkg.Logger.Debug().Msg("[ConfigLoader] 开始从主目录加载配置")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		pkg.Logger.Error().Err(err).Msg("[ConfigLoader] 获取用户主目录失败")
		return nil, err
	}

	configDir := filepath.Join(homeDir, homeConfigFolder)
	configs := make(map[string]*TomlConfig)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		pkg.Logger.Debug().Str("dir", configDir).Msg("[ConfigLoader] 配置目录不存在")
		return configs, nil
	}

	files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
	if err != nil {
		pkg.Logger.Error().Err(err).Str("dir", configDir).Msg("[ConfigLoader] 查找配置文件失败")
		return configs, nil
	}

	pkg.Logger.Debug().Int("file_count", len(files)).Str("dir", configDir).Msg("[ConfigLoader] 找到的配置文件数量")

	successCount := 0
	for _, file := range files {
		config, err := LoadTomlProxyConfig(filepath.Base(file), configDir)
		if err != nil {
			// 记录错误但不中断加载过程
			pkg.Logger.Warn().Err(err).Str("file", file).Msg("[ConfigLoader] 加载配置文件失败")
			continue
		}

		configs[filepath.Base(file)] = config
		successCount++
	}

	pkg.Logger.Info().Int("success_count", successCount).Int("total_count", len(files)).Msg("[ConfigLoader] 成功加载的配置数量")
	return configs, nil
}
