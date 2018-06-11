package play

import (
	"io"
	"os"

	"github.com/dolab/colorize"
)

// VERSION defines current package version
const (
	VERSION = "1.0.0"
)

var (
	pinfo  = colorize.New("cyan")
	pwarn  = colorize.New("yellow")
	perror = colorize.New("red")

	pinfoStart, pinfoEnd   = pinfo.Colour()
	pwarnStart, pawrnEnd   = pwarn.Colour()
	perrorStart, perrorEnd = perror.Colour()
)

// Client is an interface wraps the connection/sessions.
type Client interface {
	io.WriteCloser

	Connect(host string) error
	Run(task *Task) error
	Wait() error
	Stdin() io.WriteCloser
	Stderr() io.Reader
	Stdout() io.Reader
	Signal(os.Signal) error
	Prompt() string
	LastError() error
}
