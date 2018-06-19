package play

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
)

// LocalClient is a wrapper over the local host.
type LocalClient struct {
	cmd       *exec.Cmd
	env       string // export FOO="bar"; export BAR="baz";
	user      string
	symbol    string
	stdin     io.WriteCloser
	stdout    io.Reader
	stderr    io.Reader
	lastError error
	running   bool
}

// NewLocalClient creates a local client with given env, and
// export PLAY_HOST=localhost
func NewLocalClient(env string) *LocalClient {
	if env[len(env)-1] != ';' {
		env += ";"
	}

	return &LocalClient{
		env: env + `export PLAY_HOST="localhost";`,
	}
}

// Connect returns current user for local
func (c *LocalClient) Connect(_ string) error {
	cu, err := user.Current()
	if err != nil {
		c.lastError = err
	} else {
		c.user = cu.Username
	}

	return c.lastError
}

// Run execs book on local
func (c *LocalClient) Run(book *Book) error {
	if c.running {
		return ErrRunning
	}

	cmd := exec.Command("bash", "-c", c.env+book.run)

	c.cmd = cmd

	c.stdin, c.lastError = cmd.StdinPipe()
	if c.lastError != nil {
		return c.lastError
	}

	c.stdout, c.lastError = cmd.StdoutPipe()
	if c.lastError != nil {
		return c.lastError
	}

	c.stderr, c.lastError = cmd.StderrPipe()
	if c.lastError != nil {
		return c.lastError
	}

	c.lastError = c.cmd.Start()
	if c.lastError != nil {
		c.lastError = ErrBook{book, c.lastError.Error()}

		return c.lastError
	}

	c.running = true

	return nil
}

// Wait waits book to finish or return from error
func (c *LocalClient) Wait() error {
	if !c.running {
		return ErrNotRunning
	}

	c.lastError = c.cmd.Wait()

	c.running = false

	return c.lastError
}

func (c *LocalClient) Write(p []byte) (n int, err error) {
	n, err = c.stdin.Write(p)
	if err != nil {
		c.lastError = err
	}

	return
}

func (c *LocalClient) Close() error {
	c.lastError = c.stdin.Close()

	return c.lastError
}

func (c *LocalClient) Stdin() io.WriteCloser {
	return c.stdin
}

func (c *LocalClient) Stdout() io.Reader {
	return c.stdout
}

func (c *LocalClient) Stderr() io.Reader {
	return c.stderr
}

func (c *LocalClient) Signal(sig os.Signal) error {
	c.lastError = c.cmd.Process.Signal(sig)

	return c.lastError
}

func (c *LocalClient) Prompt() string {
	symbol := c.symbol
	if symbol == "" {
		symbol = ">>>"
	}

	return fmt.Sprintf("[%s@localhost] %s ", c.user, symbol)
}

func (c *LocalClient) LastError() error {
	return c.lastError
}
