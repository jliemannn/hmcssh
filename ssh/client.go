package ssh

import (
	"fmt"
	"time"

	sshLib "golang.org/x/crypto/ssh"
)

type Client struct {
	connection *sshLib.Client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateSession(host, port, user, password string) (*sshLib.Session, error) {
	config := &sshLib.ClientConfig{
		User: user,
		Auth: []sshLib.AuthMethod{
			sshLib.Password(password),
		},
		HostKeyCallback: sshLib.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	client, err := sshLib.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	c.connection = client

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	modes := sshLib.TerminalModes{
		sshLib.ECHO:          1,
		sshLib.TTY_OP_ISPEED: 14400,
		sshLib.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("request pty failed: %w", err)
	}

	return session, nil
}

func (c *Client) CloseConnection() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}
