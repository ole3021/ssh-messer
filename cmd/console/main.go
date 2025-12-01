package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"ssh-messer/internal/config"
	"ssh-messer/pkg"
)

func main() {
	// 1. 接收命令行参数
	var configFile = flag.String("c", "", "配置文件名称（必需）")
	flag.Parse()

	// 检查必需参数
	if *configFile == "" {
		fmt.Println("请使用 -c 参数提供配置文件名称\n用法: ssh-messher -c <配置文件路径>")
		return
	}

	configFileName := *configFile
	if !strings.HasSuffix(strings.ToLower(configFileName), ".toml") {
		configFileName += ".toml"
	}

	pkg.Logger.Info().Str("filename", configFileName).Msg("📄📄 配置文件加载成功")

	// 2. 加载配置文件
	_, err := config.LoadTomlProxyConfig(configFileName)
	if err != nil {
		pkg.Logger.Error().Str("filename", configFileName).Err(err).Msg("📄❌ 配置文件加载失败")
		return
	}

	fmt.Println("⚠️  console 模式功能正在重构中，请使用 TUI 模式 (cmd/tui)")

	// 等待信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
