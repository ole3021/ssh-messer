package services

import (
	"fmt"
	"os"
	"path/filepath"
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/tui/messages"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadConfigsFromHomeDir 从用户主目录加载配置文件
func LoadConfigsFromHomeDir() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return messages.AppError{Error: &err, IsFatal: true}
		}

		configDir := filepath.Join(homeDir, ".ssh_messer")
		configs := make(map[string]loaders.TomlConfig)

		// 读取目录下的所有 toml 文件
		files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
		if err != nil {
			return messages.AppError{Error: &err, IsFatal: true}
		}

		var errors []string
		for _, file := range files {
			config, err := loaders.LoadTomlProxyConfig(filepath.Base(file), configDir)
			if err != nil {
				// 收集错误信息
				errors = append(errors, fmt.Sprintf("加载配置文件 %s 失败: %v", filepath.Base(file), err))
				continue
			}

			// 提取配置名称（文件名去掉扩展名）
			name := strings.TrimSuffix(filepath.Base(file), ".toml")
			configs[name] = *config
		}

		// 如果有错误，返回错误信息
		if len(errors) > 0 {
			err := fmt.Errorf("部分配置文件加载失败: %s", strings.Join(errors, "; "))
			return messages.AppError{
				Error:   &err,
				IsFatal: false,
			}
		}

		return messages.LoadConfigs{Configs: configs}
	}
}
