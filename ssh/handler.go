package ssh

import (
	"io"

	sshLib "golang.org/x/crypto/ssh"
)

type Handler struct {
	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitializeStreams(session *sshLib.Session) error {
	var err error

	h.stdin, err = session.StdinPipe()
	if err != nil {
		return err
	}

	h.stdout, err = session.StdoutPipe()
	if err != nil {
		return err
	}

	h.stderr, err = session.StderrPipe()
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	return nil
}

func (h *Handler) SendInput(session *sshLib.Session, input string) error {
	if h.stdin == nil {
		return nil
	}

	_, err := h.stdin.Write([]byte(input))
	return err
}

func (h *Handler) ReadOutput(session *sshLib.Session, buf []byte) (int, error) {
	if h.stdout == nil {
		return 0, nil
	}

	n, err := h.stdout.Read(buf)
	if n > 0 {
		return n, nil
	}

	if err == nil || err == io.EOF {
		if h.stderr != nil {
			return h.stderr.Read(buf)
		}
	}

	return n, err
}

func (h *Handler) ResizeTerminal(session *sshLib.Session, width, height int) error {
	return session.WindowChange(height, width)
}

func (h *Handler) CloseSession(session *sshLib.Session) error {
	if h.stdin != nil {
		h.stdin.Close()
	}
	return session.Close()
}
