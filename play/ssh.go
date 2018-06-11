package play

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	sshAuthMethod     ssh.AuthMethod
	sshAuthMethodOnce sync.Once

	resolveSSHAuthMethod = func() {
		var signers []ssh.Signer

		// If there's a running SSH Agent, try to use its Private keys.
		sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			agent := agent.NewClient(sock)
			signers, _ = agent.Signers()
		}

		// Try to read user's SSH private keys form the standard paths.
		files, _ := filepath.Glob(os.Getenv("HOME") + "/.ssh/id_*")
		for _, file := range files {
			// Skip public keys.
			if strings.HasSuffix(file, ".pub") {
				continue
			}

			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}

			signer, err := ssh.ParsePrivateKey(data)
			if err != nil {
				continue
			}

			signers = append(signers, signer)
		}

		sshAuthMethod = ssh.PublicKeys(signers...)
	}
)

// SSHDialFunc can dial a ssh server and return a client
type SSHDialFunc func(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// SSHClient is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	conn        *ssh.Client
	sess        *ssh.Session
	env         string //export FOO="bar"; export BAR="baz";
	host        string
	user        string
	passwd      string
	symbol      string
	stdin       io.WriteCloser
	stdout      io.Reader
	stderr      io.Reader
	lastError   error
	isConnected bool
	isOpened    bool
	running     bool
}

// Connect creates SSH connection to a specified host.
// It expects the host of the form "[ssh://][user:passwd@]host[:port]".
func (c *SSHClient) Connect(host string) error {
	return c.ConnectWith(host, ssh.Dial)
}

// ConnectWith creates a SSH connection to a specified host.
// It will use dialer to establish the connection.
// TODO: Split Signers to its own method.
func (c *SSHClient) ConnectWith(host string, dialer SSHDialFunc) error {
	if c.isConnected {
		return ErrConnected
	}

	c.lastError = c.parseHost(host)
	if c.lastError != nil {
		return c.lastError
	}

	var clientConfig *ssh.ClientConfig
	if c.passwd != "" {
		clientConfig = &ssh.ClientConfig{
			User: c.user,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.passwd),
			},
		}

	} else {
		sshAuthMethodOnce.Do(resolveSSHAuthMethod)

		clientConfig = &ssh.ClientConfig{
			User: c.user,
			Auth: []ssh.AuthMethod{
				sshAuthMethod,
			},
		}
	}
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	c.conn, c.lastError = dialer("tcp", c.host, clientConfig)
	if c.lastError != nil {
		c.lastError = ErrConnect{c.user, c.host, c.lastError.Error()}

		return c.lastError
	}

	c.isConnected = true

	return nil
}

// Run runs the task.Run command on remote host.
func (c *SSHClient) Run(task *Task) error {
	if c.running {
		return ErrRunning
	}
	if c.isOpened {
		return ErrOpened
	}

	var sess *ssh.Session

	sess, c.lastError = c.conn.NewSession()
	if c.lastError != nil {
		return c.lastError
	}

	c.stdin, c.lastError = sess.StdinPipe()
	if c.lastError != nil {
		return c.lastError
	}

	c.stdout, c.lastError = sess.StdoutPipe()
	if c.lastError != nil {
		return c.lastError
	}

	c.stderr, c.lastError = sess.StderrPipe()
	if c.lastError != nil {
		return c.lastError
	}

	if task.tty {
		c.lastError = c.createPseudoTerm(sess)
		if c.lastError != nil {
			c.lastError = ErrTask{task, fmt.Sprintf("Create pseudo terminal failed: %s", c.lastError.Error())}

			return c.lastError
		}
	}

	// Start the remote command.
	c.lastError = sess.Start(c.env + task.run)
	if c.lastError != nil {
		c.lastError = ErrTask{task, c.lastError.Error()}

		return c.lastError
	}

	c.sess = sess
	c.isOpened = true
	c.running = true

	return nil
}

// Wait waits until the remote command finishes and exits.
// NOTE: It closes the SSH session.
func (c *SSHClient) Wait() error {
	if !c.running {
		return ErrNotRunning
	}

	defer c.sess.Close()

	c.lastError = c.sess.Wait()

	c.isOpened = false
	c.running = false

	return c.lastError
}

func (c *SSHClient) Write(p []byte) (n int, err error) {
	n, err = c.stdin.Write(p)
	if err != nil {
		c.lastError = err
	}

	return
}

// Close closes the underlying SSH connection and session.
func (c *SSHClient) Close() error {
	if c.isOpened {
		c.stdin.Close()
		c.sess.Close()
	}

	if !c.isConnected {
		return ErrNotConnected
	}

	c.lastError = c.conn.Close()

	c.isConnected = false
	c.isOpened = false
	c.running = false

	return c.lastError
}

func (c *SSHClient) Stdin() io.WriteCloser {
	return c.stdin
}

func (c *SSHClient) Stdout() io.Reader {
	return c.stdout
}

func (c *SSHClient) Stderr() io.Reader {
	return c.stderr
}

func (c *SSHClient) Signal(sig os.Signal) error {
	if !c.isOpened {
		return ErrNotOpened
	}

	switch sig {
	case os.Interrupt:
		// TODO: Turns out that .Signal(ssh.SIGHUP) doesn't work for me.
		// Instead, sending \x03 to the remote session works for me,
		// which sounds like something that should be fixed/resolved
		// upstream in the golang.org/x/crypto/ssh pkg.
		// https://github.com/golang/go/issues/4115#issuecomment-66070418
		c.stdin.Write([]byte("\x03"))

		c.lastError = c.sess.Signal(ssh.SIGINT)

		return c.lastError

	default:
		return fmt.Errorf("Signal %v is not supported", sig)
	}
}

func (c *SSHClient) Prompt() string {
	symbol := c.symbol
	if symbol == "" {
		symbol = "-"
	}

	return fmt.Sprintf("[%s@%s] %s ", c.user, c.host, symbol)
}

func (c *SSHClient) LastError() error {
	return c.lastError
}

// dialThrough will create a new connection from the ssh server sc is connected to.
// NOTE: dialThrough is a SSHDialer.
func (c *SSHClient) dialThrough(net, addr string, config *ssh.ClientConfig) (client *ssh.Client, err error) {
	nc, err := c.conn.Dial(net, addr)
	if err != nil {
		return
	}

	sc, cc, rc, err := ssh.NewClientConn(nc, addr, config)
	if err != nil {
		return
	}

	client = ssh.NewClient(sc, cc, rc)
	return
}

// parseHost parses and normalizes <user>@<host:port> from a given string.
func (c *SSHClient) parseHost(host string) error {
	c.host = host

	// Remove extra "ssh://" schema
	if strings.HasPrefix(c.host, "ssh://") {
		c.host = strings.TrimPrefix(c.host, "ssh://")
	}

	if at := strings.Index(c.host, "@"); at != -1 {
		user2pass := strings.SplitN(c.host[:at], ":", 2)
		c.user = user2pass[0]
		if len(user2pass) == 2 {
			c.passwd = user2pass[1]
		}

		c.host = c.host[at+1:]
	}

	// Add default user, if not set
	if c.user == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}

		c.user = u.Username
	}

	if strings.Index(c.host, "/") != -1 {
		return ErrConnect{c.user, c.host, "unexpected slash in the host."}
	}

	// Add default port, if not set
	if strings.Index(c.host, ":") == -1 {
		c.host += ":22"
	}

	return nil
}

func (c *SSHClient) createPseudoTerm(sess *ssh.Session) error {
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	return sess.RequestPty("xterm", 80, 40, modes)
}
