package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"ssh-messer/internal/config_loader"
	// "ssh-messer/internal/proxy" // TODO: 此包已被重构，需要更新实现
	"ssh-messer/pkg"
)

func main() {
	// 1. 接收命令行参数
	var configFile = flag.String("c", "", "配置文件名称（必需）")
	var shell = flag.Bool("shell", false, "启动交互式Shell")
	var docker = flag.Bool("docker", false, "启动Docker TCP代理")
	var http = flag.Bool("http", false, "启动HTTP服务代理")
	flag.Parse()

	// 检查必需参数
	if *configFile == "" {
		fmt.Println("请使用 -c 参数提供配置文件名称\n用法: ssh-messher -c <配置文件路径>")
		return
	}

	if !*shell && !*docker && !*http {
		fmt.Println("请使用 -shell, -docker, -http 参数至少一个")
		return
	}

	configFileName := *configFile
	if !strings.HasSuffix(strings.ToLower(configFileName), ".toml") {
		configFileName += ".toml"
	}

	pkg.Logger.Info().Str("filename", configFileName).Msg("📄📄 配置文件加载成功")

	// 2. 加载配置文件
	_, err := config_loader.LoadTomlProxyConfig(configFileName)
	if err != nil {
		pkg.Logger.Error().Str("filename", configFileName).Err(err).Msg("📄❌ 配置文件加载失败")
		return
	}

	// TODO: 以下功能需要重构以使用新的 internal/ssh_proxy 包
	// 3. 创建 ssh hops 客户端
	// sshHopsClient, err := proxy.CreateSSHHopsClient(proxyConfig.SSHHops)
	// if err != nil {
	// 	pkg.Logger.Error().Err(err).Msg("❌ SSH 跳转客户端连接失败")
	// 	return
	// }
	// pkg.Logger.Info().Msg("🦘🦘 SSH 跳转客户端连接成功 🦘🦘")
	// defer sshHopsClient.Close()

	// if *http {
	// 	// 创建HTTP服务代理
	// 	serviceProxy := proxy.NewtHttpServiceProxyServer(*proxyConfig.LocalHttpPort, proxyConfig.Services, sshHopsClient)
	// 	go serviceProxy.Start()

	// 	pkg.Logger.Info().Msg("🔗🔗 HTTP服务代理启动成功  🔗🔗")
	// 	for _, service := range proxyConfig.Services {
	// 		pkg.Logger.Info().Msgf("🔗🔗 [%-20s] => http://%s.localhost:%s", *service.Alias, *service.Subdomain, *proxyConfig.LocalHttpPort)
	// 	}
	// }

	// if *shell {
	// 	// 启动交互式Shell
	// 	go proxy.StartInteractiveShell(sshHopsClient)

	// 	fmt.Printf("🐚🐚 SSH交互式Shell启动成功 🐚🐚\n")
	// 	fmt.Println("🐚🐚 输入命令执行，输入 'exit' 退出 🐚🐚")
	// }

	// if *docker {
	// 	// 创建Docker TCP代理
	// 	dockerTCPProxy := proxy.NewDockerTCPProxy(
	// 		sshHopsClient,
	// 		*proxyConfig.LocalDockerPort, // 本地TCP端口
	// 		"/var/run/docker.sock",       // 远程Docker Socket
	// 	)

	// 	// 启动Docker TCP代理
	// 	go func() {
	// 		if err := dockerTCPProxy.Start(); err != nil {
	// 			pkg.Logger.Error().Err(err).Msg("🐳🔴 Docker TCP代理启动失败")
	// 		}
	// 	}()

	// 	pkg.Logger.Info().Str("local_docker_port", *proxyConfig.LocalDockerPort).Msg("🐳🐳 Docker TCP代理启动成功  🐳🐳")
	// 	pkg.Logger.Info().Str("local_docker_port", *proxyConfig.LocalDockerPort).Msg("🐳 查看远程容器: DOCKER_HOST=tcp://localhost:%s docker ps")
	// }

	fmt.Println("⚠️  console 模式功能正在重构中，请使用 TUI 模式 (cmd/tui)")

	// 等待信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// fmt.Println("\n正在关闭代理服务器...")
	<-c
}
