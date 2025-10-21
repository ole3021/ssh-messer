package main

import (
	"fmt"
	"os"

	"ssh-messer/internal/tui/app"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(app.NewAppModel(), tea.WithAltScreen())
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
