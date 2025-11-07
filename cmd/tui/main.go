package main

import (
	"fmt"
	"os"

	"ssh-messer/internal/tui"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	pkg.InitLogger("file")

	model := tui.New()
	p := tea.NewProgram(model)

	// 订阅异步事件
	go model.Subscribe(p)

	if _, err := p.Run(); err != nil {
		fmt.Println("##############################################################################")
		fmt.Printf("Ops, something went wrong: %v !", err)
		fmt.Println("Please contact the developer for support.")
		fmt.Println("Email: ole3021@gmail.com")
		fmt.Println("GitHub: https://github.com/ole3021/ssh-messer/issues")
		fmt.Println("##############################################################################")
		os.Exit(1)
	}
}
