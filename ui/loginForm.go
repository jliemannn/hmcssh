package ui

import (
	"fmt"
	"strings"

	"hmcssh/ssh"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var bannerText = `
$$\                                                   $$\
$$ |                                                  $$ |
$$$$$$$\  $$$$$$\$$$$\   $$$$$$$\  $$$$$$$\  $$$$$$$\ $$$$$$$\
$$  __$$\ $$  _$$  _$$\ $$  _____|$$  _____|$$  _____|$$  __$$\
$$ |  $$ |$$ / $$ / $$ |$$ /      \$$$$$$\  \$$$$$$\  $$ |  $$ |
$$ |  $$ |$$ | $$ | $$ |$$ |       \____$$\  \____$$\ $$ |  $$ |
$$ |  $$ |$$ | $$ | $$ |\$$$$$$$\ $$$$$$$  |$$$$$$$  |$$ |  $$ |
\__|  \__|\__| \__| \__| \_______|\_______/ \_______/ \__|  \__|
`

type LoginForm struct {
	inputs       []textinput.Model
	focusIndex   int
	errorMsg     string
	connecting   bool
	sshConnector *ssh.Connector
	width        int
	height       int
}

func (f *LoginForm) SetSize(width, height int) {
	f.width = width
	f.height = height
}

func NewLoginForm() *LoginForm {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "192.168.1.100"
	inputs[0].Focus()
	inputs[0].CharLimit = 100
	inputs[0].Width = 40
	inputs[0].Prompt = "Host: "

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "22"
	inputs[1].CharLimit = 5
	inputs[1].Width = 40
	inputs[1].Prompt = "Port: "

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "username"
	inputs[2].CharLimit = 50
	inputs[2].Width = 40
	inputs[2].Prompt = "User: "

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password"
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].CharLimit = 100
	inputs[3].Width = 40
	inputs[3].Prompt = "Pass: "

	return &LoginForm{
		inputs:       inputs,
		focusIndex:   0,
		sshConnector: ssh.NewConnector(),
	}
}

func (f *LoginForm) Update(msg tea.Msg) (*LoginForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				f.focusIndex--
			} else {
				f.focusIndex++
			}

			if f.focusIndex > len(f.inputs)-1 {
				f.focusIndex = 0
			} else if f.focusIndex < 0 {
				f.focusIndex = len(f.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(f.inputs))
			for i := 0; i < len(f.inputs); i++ {
				if i == f.focusIndex {
					cmds[i] = f.inputs[i].Focus()
					f.inputs[i].PromptStyle = lipgloss.NewStyle().Background(lipgloss.Color("7"))
					//f.inputs[i].TextStyle = lipgloss.NewStyle().Background(lipgloss.Color("7"))
					f.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					f.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				} else {
					f.inputs[i].Blur()
					f.inputs[i].PromptStyle = lipgloss.NewStyle()
					f.inputs[i].TextStyle = lipgloss.NewStyle()
				}
			}
			return f, tea.Batch(cmds...)

		case "enter":
			if !f.connecting {
				f.connecting = true
				f.errorMsg = ""

				host := f.inputs[0].Value()
				port := f.inputs[1].Value()
				user := f.inputs[2].Value()
				pass := f.inputs[3].Value()

				if port == "" {
					port = "22"
				}

				return f, f.sshConnector.Connect(host, port, user, pass)
			}
		}
	}

	cmd := f.updateInputs(msg)
	return f, cmd
}

func (f *LoginForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		f.inputs[i], cmds[i] = f.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (f *LoginForm) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		MarginBottom(2)

	b.WriteString(titleStyle.Render(bannerText))
	b.WriteString("\n\n")

	for i := range f.inputs {
		b.WriteString(f.inputs[i].View())
		b.WriteString("\n\n")
	}

	if f.connecting {
		connectingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		b.WriteString(connectingStyle.Render("Connecting..."))
		b.WriteString("\n\n")
	}

	if f.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", f.errorMsg)))
		b.WriteString("\n\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render("Tab: next field • Enter: connect • Ctrl+C: quit"))

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("7"))

	content := containerStyle.Render(b.String())

	if f.width > 0 && f.height > 0 {
		return lipgloss.Place(f.width, f.height, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}

func (f *LoginForm) SetError(msg string) {
	f.errorMsg = msg
	f.connecting = false
}
