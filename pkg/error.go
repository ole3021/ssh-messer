package pkg

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	meta "ssh-messer"
)

const (
	colorPurple = "\033[35m" // 紫色
	colorRed    = "\033[31m" // 红色
	colorReset  = "\033[0m"  // 重置颜色
)

func HandleTerminalError(err error, logFilePath string) {
	if err != nil {
		fmt.Println("#################################################################################")
		fmt.Printf("%sOops, something went wrong!%s\n", colorRed, colorReset)
		fmt.Printf("%s⚠️  Error: %s%s\n", colorRed, err.Error(), colorReset)
		fmt.Println("#################################################################################")
		fmt.Println("Extra Debug Information:")
		// TODO: fix emoji size alignment issue
		fmt.Printf("%-4s %-9s %s\n", "🏷️", "Version:", meta.Version)
		fmt.Printf("%-4s %-9s %s/%s\n", "🖥️", "OS:", runtime.GOOS, runtime.GOARCH)
		if hostname, err := os.Hostname(); err == nil {
			fmt.Printf("%-2s %-8s %s\n", "🏠", "Hostname:", hostname)
		}
		if currentUser, err := user.Current(); err == nil {
			fmt.Printf("%-2s %-8s %s\n", "👤", "Username:", currentUser.Username)
		}
		if logFilePath != "" {
			fmt.Printf("%-2s %-8s %s\n", "📄", "Logfile:", logFilePath)
			fmt.Printf("%s>> Please attach the log file with above information when reporting the issue. <<%s\n", colorPurple, colorReset)

		} else {
			fmt.Printf("%s>> Please attach above information when reporting the issue. <<%s\n", colorRed, colorReset)
		}
		fmt.Println("#################################################################################")
		fmt.Println("Report any issue or feature request through the following channels:")
		fmt.Println("📧  Email: ", meta.Email)
		fmt.Println("🔗  GitHub: ", meta.Repository)
		fmt.Println("#################################################################################")

		os.Exit(0)
	}
}
