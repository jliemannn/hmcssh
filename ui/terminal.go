package ui

import (
	"strings"

	"hmcssh/ssh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	sshLib "golang.org/x/crypto/ssh"
)

type Terminal struct {
	session    *sshLib.Session
	lines      []string
	width      int
	height     int
	scrollPos  int
	sshHandler *ssh.Handler
	input      string
	cursorPos  int
	rawMode    bool
	rawOutput  string
}

type OutputMsg struct {
	Data []byte
}

func NewTerminal() *Terminal {
	return &Terminal{
		lines:      make([]string, 0),
		sshHandler: ssh.NewHandler(),
		input:      "",
		cursorPos:  0,
		rawMode:    false,
		rawOutput:  "",
	}
}

func (t *Terminal) SetSize(width, height int) {
	t.width = width
	t.height = height
	if t.session != nil {
		t.sshHandler.ResizeTerminal(t.session, width, height)
	}
}

func (t *Terminal) SetSession(session *sshLib.Session) {
	t.session = session
	if err := t.sshHandler.InitializeStreams(t.session); err != nil {
		return
	}
}

func (t *Terminal) HasSession() bool {
	return t.session != nil
}

func (t *Terminal) Update(msg tea.Msg) (*Terminal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKeyPress(msg)
	case OutputMsg:
		t.processOutput(msg.Data)
		return t, t.ReadOutput()
	}
	return t, nil
}

func (t *Terminal) processOutput(data []byte) {
	output := string(data)

	if strings.Contains(output, "\x1b[?1049h") ||
		strings.Contains(output, "\x1b[?47h") ||
		strings.Contains(output, "\x1b[?1h\x1b=") {
		t.rawMode = true
		t.rawOutput = ""
	}

	if strings.Contains(output, "\x1b[?1049l") ||
		strings.Contains(output, "\x1b[?47l") {
		t.rawMode = false
		t.rawOutput = ""
		return
	}

	if t.rawMode {
		t.rawOutput += output
		return
	}

	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "")

	if strings.Contains(output, "\x1b[H\x1b[2J") || strings.Contains(output, "\x1b[2J\x1b[H") || strings.Contains(output, "\x1b[2J") {
		t.lines = make([]string, 0)
		output = strings.ReplaceAll(output, "\x1b[H\x1b[2J", "")
		output = strings.ReplaceAll(output, "\x1b[2J\x1b[H", "")
		output = strings.ReplaceAll(output, "\x1b[2J", "")
		t.scrollPos = 0
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasSuffix(trimmed, "$") || strings.HasSuffix(trimmed, "#") || strings.HasSuffix(trimmed, ">") {
			continue
		}

		if strings.Contains(line, t.input) && t.input != "" {
			continue
		}

		t.lines = append(t.lines, line)
	}

	maxLines := 10000
	if len(t.lines) > maxLines {
		t.lines = t.lines[len(t.lines)-maxLines:]
	}
}

func (t *Terminal) handleKeyPress(msg tea.KeyMsg) (*Terminal, tea.Cmd) {
	if t.session == nil {
		return t, nil
	}

	if t.rawMode {
		return t.handleRawModeKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		t.sshHandler.SendInput(t.session, "\x03")
		t.input = ""
		t.cursorPos = 0
		return t, nil

	case "ctrl+d":
		t.sshHandler.SendInput(t.session, "\x04")
		return t, nil

	case "enter":
		if t.input != "" {
			t.sshHandler.SendInput(t.session, t.input+"\n")
		} else {
			t.sshHandler.SendInput(t.session, "\n")
		}
		t.input = ""
		t.cursorPos = 0

	case "backspace":
		if t.cursorPos > 0 {
			t.input = t.input[:t.cursorPos-1] + t.input[t.cursorPos:]
			t.cursorPos--
		}

	case "tab":
		t.input = t.input[:t.cursorPos] + "\t" + t.input[t.cursorPos:]
		t.cursorPos++

	case "up":
		t.sshHandler.SendInput(t.session, "\x1b[A")

	case "down":
		t.sshHandler.SendInput(t.session, "\x1b[B")

	case "left":
		if t.cursorPos > 0 {
			t.cursorPos--
		}

	case "right":
		if t.cursorPos < len(t.input) {
			t.cursorPos++
		}

	case "home":
		t.cursorPos = 0

	case "end":
		t.cursorPos = len(t.input)

	case "delete":
		if t.cursorPos < len(t.input) {
			t.input = t.input[:t.cursorPos] + t.input[t.cursorPos+1:]
		}

	case "pgup":
		if t.scrollPos > 0 {
			t.scrollPos--
		}

	case "pgdown":
		t.scrollPos++

	default:
		if len(msg.Runes) > 0 {
			char := string(msg.Runes)
			t.input = t.input[:t.cursorPos] + char + t.input[t.cursorPos:]
			t.cursorPos++
		}
	}

	return t, nil
}

func (t *Terminal) handleRawModeKey(msg tea.KeyMsg) (*Terminal, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		t.sshHandler.SendInput(t.session, "\x03")
	case "ctrl+d":
		t.sshHandler.SendInput(t.session, "\x04")
	case "ctrl+x":
		t.sshHandler.SendInput(t.session, "\x18")
	case "ctrl+o":
		t.sshHandler.SendInput(t.session, "\x0f")
	case "ctrl+w":
		t.sshHandler.SendInput(t.session, "\x17")
	case "ctrl+k":
		t.sshHandler.SendInput(t.session, "\x0b")
	case "ctrl+u":
		t.sshHandler.SendInput(t.session, "\x15")
	case "enter":
		t.sshHandler.SendInput(t.session, "\r")
	case "backspace":
		t.sshHandler.SendInput(t.session, "\x7f")
	case "tab":
		t.sshHandler.SendInput(t.session, "\t")
	case "up":
		t.sshHandler.SendInput(t.session, "\x1b[A")
	case "down":
		t.sshHandler.SendInput(t.session, "\x1b[B")
	case "left":
		t.sshHandler.SendInput(t.session, "\x1b[D")
	case "right":
		t.sshHandler.SendInput(t.session, "\x1b[C")
	case "home":
		t.sshHandler.SendInput(t.session, "\x1b[H")
	case "end":
		t.sshHandler.SendInput(t.session, "\x1b[F")
	case "delete":
		t.sshHandler.SendInput(t.session, "\x1b[3~")
	case "pgup":
		t.sshHandler.SendInput(t.session, "\x1b[5~")
	case "pgdown":
		t.sshHandler.SendInput(t.session, "\x1b[6~")
	default:
		if len(msg.Runes) > 0 {
			t.sshHandler.SendInput(t.session, string(msg.Runes))
		}
	}

	return t, nil
}

func (t *Terminal) View() string {
	if t.session == nil {
		return "No active session"
	}

	if t.rawMode {
		return t.viewRawMode()
	}

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		Width(t.width)

	b.WriteString(headerStyle.Render("HMC SSH Terminal (Ctrl+C: interrupt, Ctrl+D: exit)"))
	b.WriteString("\n")

	visibleLines := t.height - 4

	startLine := len(t.lines) - visibleLines - t.scrollPos
	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + visibleLines
	if endLine > len(t.lines) {
		endLine = len(t.lines)
	}

	for i := startLine; i < endLine; i++ {
		if i >= 0 && i < len(t.lines) {
			b.WriteString(t.lines[i])
			b.WriteString("\n")
		}
	}

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	b.WriteString("\n")
	b.WriteString(promptStyle.Render("$ "))

	displayInput := t.input[:t.cursorPos] + "█" + t.input[t.cursorPos:]
	b.WriteString(inputStyle.Render(displayInput))

	return b.String()
}

func (t *Terminal) viewRawMode() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("205")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		Width(t.width)

	b.WriteString(headerStyle.Render("HMC SSH Terminal - Interactive Mode"))
	b.WriteString("\n")
	b.WriteString(t.rawOutput)

	return b.String()
}

func (t *Terminal) ReadOutput() tea.Cmd {
	return func() tea.Msg {
		if t.session == nil {
			return nil
		}

		buf := make([]byte, 4096)
		n, err := t.sshHandler.ReadOutput(t.session, buf)
		if err != nil {
			return nil
		}

		return OutputMsg{Data: buf[:n]}
	}
}

func (t *Terminal) Close() {
	if t.session != nil {
		t.sshHandler.CloseSession(t.session)
	}
}
