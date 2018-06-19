package play

import (
	"io"
	"os"
)

// VERSION defines current package version
const (
	VERSION = "1.0.0"
)

// Client is an interface wraps the connection/sessions.
type Client interface {
	io.WriteCloser

	Connect(host string) error
	Run(book *Book) error
	Wait() error
	Stdin() io.WriteCloser
	Stderr() io.Reader
	Stdout() io.Reader
	Signal(os.Signal) error
	Prompt() string
	LastError() error
}
