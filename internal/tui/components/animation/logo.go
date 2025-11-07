package animation

import (
	"strings"
	"time"

	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/styles"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
)

const (
	logoFps       = 60
	frameDuration = time.Second / time.Duration(logoFps)
	maxStep       = 100
	asciiArt      = `
    ███╗   ███████╗███████╗██╗  ██╗    ███╗   ███╗███████╗███████╗███████╗███████╗██████╗ 
   ████║   ██╔════╝██╔════╝██║  ██║    ████╗ ████║██╔════╝██╔════╝██╔════╝██╔════╝██╔══██╗
  ██╔██║   ███████╗███████╗███████║    ██╔████╔██║█████╗  ███████╗█████╗  █████╗  ██████╔╝
 ██╔╝██║   ╚════██║╚════██║██╔══██║    ██║╚██╔╝██║██╔══╝  ╚════██║██╔══╝  ██╔══╝  ██╔══██╗
██╔╝ ██║   ███████║███████║██║  ██║    ██║ ╚═╝ ██║███████╗███████║███████╗███████╗██║  ██║
╚═╝  ╚═╝   ╚══════╝╚══════╝╚═╝  ╚═╝    ╚═╝     ╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`
)

type AppLogoCmp interface {
	util.Model
	layout.Sizeable
}

type appLogoCmp struct {
	width, height int
	step          int
	animType      types.AnimationType
}

// NewLogo creates a new logo component instance.
func NewLogo() AppLogoCmp {
	return &appLogoCmp{
		step:     0,
		animType: types.AnimationTypeLogo,
	}
}

func (a *appLogoCmp) Init() tea.Cmd {
	return a.Step()
}

func (a *appLogoCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.StepMsg:
		if msg.AnimType != a.animType {
			return a, nil
		}

		a.step++
		if a.step > maxStep {
			a.step = maxStep
			return a, nil
		}

		return a, a.Step()
	}

	return a, nil
}

func (a *appLogoCmp) View() string {
	return renderASCIIArt(a.step)
}

// Step returns a command that triggers the next step in the animation.
func (a *appLogoCmp) Step() tea.Cmd {
	return tea.Tick(frameDuration, func(t time.Time) tea.Msg {
		return messages.StepMsg{AnimType: a.animType}
	})
}

func (a *appLogoCmp) SetSize(width, height int) tea.Cmd {
	a.width = width
	a.height = height
	return nil
}

func (a *appLogoCmp) GetSize() (int, int) {
	return a.width, a.height
}

// renderASCIIArt renders the ASCII art with animation based on step.
func renderASCIIArt(step int) string {
	lines := strings.Split(asciiArt, "\n")

	// 计算每行应该显示多少字符
	var result []string
	for i, line := range lines {
		// 每行有不同的延迟，创造波浪效果
		lineDelay := i * 15 // 每行延迟15%
		lineProgress := step - lineDelay
		if lineProgress < 0 {
			lineProgress = 0
		}

		visibleChars := (lineProgress * len(line)) / (100 - lineDelay)
		if visibleChars > len(line) {
			visibleChars = len(line)
		}

		// 逐字符显示，带闪烁效果
		var visibleLine string
		for j, char := range line {
			if j < visibleChars {
				// 添加闪烁效果
				if (step/5)%2 == 0 {
					visibleLine += styles.AsciiStyle.Render(string(char))
				} else {
					visibleLine += styles.AsciiStyle.Foreground(styles.Primary).Render(string(char))
				}
			} else {
				visibleLine += " "
			}
		}
		result = append(result, visibleLine)
	}

	return strings.Join(result, "\n")
}
