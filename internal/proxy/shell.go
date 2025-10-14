package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// 启动交互式Shell
func StartInteractiveShell(client *ssh.Client) {

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("🐚 SSH> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		if command == "exit" || command == "quit" {
			fmt.Println("🐚🐚 退出SSH Shell 🐚🐚")
			break
		}

		// 执行命令
		result, err := executeSSHCommand(client, command)
		if err != nil {
			fmt.Printf("🐚🔴 命令执行失败: %v\n", err)
		} else {
			fmt.Print(result)
		}
	}
}

// 执行SSH命令
func executeSSHCommand(client *ssh.Client, command string) (string, error) {
	// 创建SSH会话
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 设置输出缓冲区
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// 执行命令
	err = session.Run(command)

	// 组合输出
	var output bytes.Buffer
	if stdout.Len() > 0 {
		output.WriteString("STDOUT:\n")
		output.Write(stdout.Bytes())
	}
	if stderr.Len() > 0 {
		output.WriteString("\nSTDERR:\n")
		output.Write(stderr.Bytes())
	}

	if err != nil {
		return output.String(), fmt.Errorf("🐚🔴 命令执行错误: %v", err)
	}

	return output.String(), nil
}
