package config

import (
	"fmt"
	"os"
	"path/filepath"

	"ssh-messer/pkg"

	"github.com/BurntSushi/toml"
)

const homeConfigFolder = ".ssh_messer"

func LoadTomlProxyConfig(filename string, dir ...string) (*MesserConfig, error) {
	var fullPath string

	if len(dir) > 0 {
		fullPath = filepath.Join(dir[0], filename)
	} else {
		fullPath = filepath.Join("configs", filename)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		pkg.Logger.Error().Str("file", fullPath).Msg("[toml::loadTomlProxyConfig] Config file not exists")
		return nil, fmt.Errorf("config file not exists: %s", fullPath)
	}

	var proxyConfig MesserConfig

	if _, err := toml.DecodeFile(fullPath, &proxyConfig); err != nil {
		pkg.Logger.Error().Err(err).Str("file", fullPath).Msg("[toml::loadTomlProxyConfig] Failed to decode config file")
		return nil, fmt.Errorf("failed to decode config file %s: %w", fullPath, err)
	}

	pkg.Logger.Info().Str("file", fullPath).Msg("[toml::loadTomlProxyConfig] Loaded config file")
	return &proxyConfig, nil
}

func LoadTomlConfigsFromHomeDir() (map[string]*MesserConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		pkg.Logger.Error().Err(err).Msg("[toml::LoadTomlConfigsFromHomeDir] Failed to get home directory")
		return nil, err
	}
	pkg.Logger.Info().Str("dir", homeConfigFolder).Msg("[toml::LoadTomlConfigsFromHomeDir] Load configs from home directory")

	configDir := filepath.Join(homeDir, homeConfigFolder)
	configs := make(map[string]*MesserConfig)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		pkg.Logger.Warn().Str("dir", configDir).Msg("[toml::LoadTomlConfigsFromHomeDir] Config directory not exists")
		return configs, nil
	}

	files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
	if err != nil {
		pkg.Logger.Error().Err(err).Str("dir", configDir).Msg("[toml::LoadTomlConfigsFromHomeDir] Failed to find config files")
		return configs, nil
	}

	successCount := 0
	for _, file := range files {
		config, err := LoadTomlProxyConfig(filepath.Base(file), configDir)
		if err != nil {
			pkg.Logger.Warn().Err(err).Str("file", file).Msg("[toml::LoadTomlConfigsFromHomeDir] Failed to load config file, continue")
			continue
		}

		configs[filepath.Base(file)] = config
		successCount++
	}

	pkg.Logger.Info().Int("success_count", successCount).Int("total_count", len(files)).Msg("[toml::LoadTomlConfigsFromHomeDir] Loaded config files")
	return configs, nil
}
