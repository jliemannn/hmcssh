package ssh

import (
	tea "github.com/charmbracelet/bubbletea"
	sshLib "golang.org/x/crypto/ssh"
)

type ConnectionSuccessMsg struct {
	Session *sshLib.Session
}

type ConnectionErrorMsg struct {
	Err error
}

type Connector struct {
	client *Client
}

func NewConnector() *Connector {
	return &Connector{
		client: NewClient(),
	}
}

func (c *Connector) Connect(host, port, user, password string) tea.Cmd {
	return func() tea.Msg {
		session, err := c.client.CreateSession(host, port, user, password)
		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}
		return ConnectionSuccessMsg{Session: session}
	}
}
