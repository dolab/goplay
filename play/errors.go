package play

import (
	"errors"
	"fmt"
)

var (
	ErrRunning      = errors.New("Command alreay running.")
	ErrNotRunning   = errors.New("Command not running or has finished.")
	ErrConnected    = errors.New("Client has connected.")
	ErrNotConnected = errors.New("Client not connected.")
	ErrOpened       = errors.New("Session has opened.")
	ErrNotOpened    = errors.New("Session not opened.")
	ErrNotSupported = errors.New("Not supported.")
)

// ErrConnect defines connection error with reason
type ErrConnect struct {
	Host   string
	User   string
	Reason string
}

// Error implements error
func (e ErrConnect) Error() string {
	return fmt.Sprintf(`Connect("%s@%s"): %s`, e.User, e.Host, e.Reason)
}

// ErrTask defines task error
type ErrTask struct {
	Task   *Task
	Reason string
}

func (e ErrTask) Error() string {
	return fmt.Sprintf(`Run task %v: %s`, e.Task, e.Reason)
}

// ErrPlayfileVersion defines error for unsupported playfile version
type ErrPlayfileVersion struct {
	Msg string
}

func (e ErrPlayfileVersion) Error() string {
	return fmt.Sprintf("%s\n\nPlease checking your Playfile version or upgrading your goplay. (The latest version: v%v)", e.Msg, VERSION)
}

// ErrMustUpgrade defines upgrade error for old client
type ErrMustUpgrade struct {
	Msg string
}

func (e ErrMustUpgrade) Error() string {
	return fmt.Sprintf("%s\n\nPlease upgrading goplay by `go get -u github.com/dolab/goplay`", e.Msg)
}
