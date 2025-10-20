package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"ssh-messer/internal/loaders"
	"ssh-messer/internal/proxy"
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
		log.Fatalf("请使用 -c 参数提供配置文件名称\n用法: %s -c <配置文件名称>", os.Args[0])
		return
	}

	if !*shell && !*docker && !*http {
		log.Fatalf("请使用 -shell, -docker, -http 参数至少一个")
		return
	}

	configFileName := *configFile
	if !strings.HasSuffix(strings.ToLower(configFileName), ".toml") {
		configFileName += ".toml"
	}

	fmt.Printf("📄📄 配置文件 [%s] 加载成功 📄📄 \n", configFileName)

	// 2. 加载配置文件
	proxyConfig, err := loaders.LoadTomlProxyConfig(configFileName)
	if err != nil {
		fmt.Println("📄❌ 配置文件加载失败:", err)
		return
	}

	// 3. 创建 ssh hops 客户端
	sshHopsClient, err := proxy.CreateSSHHopsClient(proxyConfig.SSHHops)
	if err != nil {
		fmt.Println("❌ SSH 跳转客户端连接失败:", err)
		return
	}
	fmt.Println("🦘🦘 SSH 跳转客户端连接成功 🦘🦘")
	defer sshHopsClient.Close()

	if *http {
		// 创建HTTP服务代理
		serviceProxy := proxy.NewtHttpServiceProxyServer(*proxyConfig.LocalHttpPort, proxyConfig.Services, sshHopsClient)
		go serviceProxy.Start()

		fmt.Printf("🔗🔗 HTTP服务代理启动成功  🔗🔗\n")
		for _, service := range proxyConfig.Services {
			fmt.Printf("🔗🔗 [%-20s] => http://%s.localhost:%s\n", *service.Alias, *service.Subdomain, *proxyConfig.LocalHttpPort)
		}
	}

	if *shell {
		// 启动交互式Shell
		go proxy.StartInteractiveShell(sshHopsClient)

		fmt.Printf("🐚🐚 SSH交互式Shell启动成功 🐚🐚\n")
		fmt.Println("🐚🐚 输入命令执行，输入 'exit' 退出 🐚🐚")
	}

	if *docker {
		// 创建Docker TCP代理
		dockerTCPProxy := proxy.NewDockerTCPProxy(
			sshHopsClient,
			*proxyConfig.LocalDockerPort, // 本地TCP端口
			"/var/run/docker.sock",       // 远程Docker Socket
		)

		// 启动Docker TCP代理
		go func() {
			if err := dockerTCPProxy.Start(); err != nil {
				fmt.Printf("🐳🔴 Docker TCP代理启动失败: %v\n", err)
			}
		}()

		fmt.Printf("🐳🐳 Docker TCP代理启动成功  🐳🐳\n")
		fmt.Printf("🐳 查看远程容器: DOCKER_HOST=tcp://localhost:%s docker ps \n", *proxyConfig.LocalDockerPort)
	}

	// 等待信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// fmt.Println("\n正在关闭代理服务器...")
	<-c
}
