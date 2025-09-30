package models

import (
	"hmcssh/ssh"
	"hmcssh/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	ScreenLogin screen = iota
	ScreenTerminal
)

type App struct {
	currentScreen screen
	loginForm     *ui.LoginForm
	terminal      *ui.Terminal
	width         int
	height        int
	err           error
}

func NewApp() *App {
	return &App{
		currentScreen: ScreenLogin,
		loginForm:     ui.NewLoginForm(),
		terminal:      ui.NewTerminal(),
	}
}

func (a *App) Init() tea.Cmd {
	return textinput.Blink
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.terminal != nil {
			a.terminal.SetSize(msg.Width, msg.Height)
		}
		if a.loginForm != nil {
			a.loginForm.SetSize(msg.Width, msg.Height)
		}

	case ssh.ConnectionSuccessMsg:
		a.currentScreen = ScreenTerminal
		a.terminal.SetSession(msg.Session)
		return a, a.terminal.ReadOutput()

	case ssh.ConnectionErrorMsg:
		a.err = msg.Err
		a.loginForm.SetError(msg.Err.Error())

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if a.terminal != nil && a.terminal.HasSession() {
				a.terminal.Close()
			}
			return a, tea.Quit
		}
	}

	switch a.currentScreen {
	case ScreenLogin:
		return a.updateLogin(msg)
	case ScreenTerminal:
		return a.updateTerminal(msg)
	}

	return a, nil
}

func (a *App) View() string {
	switch a.currentScreen {
	case ScreenLogin:
		return a.loginForm.View()
	case ScreenTerminal:
		return a.terminal.View()
	}
	return ""
}

func (a *App) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.loginForm, cmd = a.loginForm.Update(msg)
	return a, cmd
}

func (a *App) updateTerminal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.terminal, cmd = a.terminal.Update(msg)
	return a, cmd
}
