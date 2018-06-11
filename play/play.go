package play

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"

	"github.com/goware/prefixer"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Play holds all tasks for running
type Play struct {
	config *Playfile
	prompt bool
	debug  bool
}

// New returns *Play with config
func New(config *Playfile) (*Play, error) {
	return &Play{
		config: config,
	}, nil
}

// Run runs set of commands on multiple hosts defined by network sequentially.
// TODO: This megamoth method needs a big refactor and should be split
//       to multiple smaller methods.
func (play *Play) Run(network *Network, envs EnvVars, commands ...*Command) error {
	if len(commands) == 0 {
		return errors.New("no command to be run")
	}

	// Create bastion for every host (either SSH or Localhost).
	var bastion *SSHClient
	if network.Bastion != "" {
		bastion = &SSHClient{}
		if err := bastion.Connect(network.Bastion); err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s: connecting to bastion failed", network.Bastion))
		}
	}

	clientCh := make(chan Client, len(network.Hosts))
	clientEnv := envs.AsExport()

	var (
		wg sync.WaitGroup
	)
	for i, host := range network.Hosts {
		wg.Add(1)

		go func(i int, host string) {
			defer wg.Done()

			switch host {
			case "localhost", "127.0.0.1": // localhost client
				local := &LocalClient{
					env: clientEnv + `export PLAY_HOST="` + host + `";`,
				}

				local.Connect(host)

				clientCh <- local

			default: // ssh client
				remote := &SSHClient{
					env:  clientEnv + `export PLAY_HOST="` + host + `";`,
					user: network.User,
				}

				if bastion != nil {
					remote.ConnectWith(host, bastion.dialThrough)
				} else {
					remote.Connect(host)
				}

				clientCh <- remote

			}
		}(i, host)
	}
	wg.Wait()

	close(clientCh)

	var (
		clients      []Client
		maxPromptLen = 0
	)
	for client := range clientCh {
		lastErr := client.LastError()
		if lastErr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", errors.Wrap(lastErr, "connecting failed"))

			continue
		}

		if remote, ok := client.(*SSHClient); ok {
			defer remote.Close()
		}

		clients = append(clients, client)

		prompt := client.Prompt()
		if len(prompt) > maxPromptLen {
			maxPromptLen = len(prompt)
		}
	}

	if len(clients) == 0 {
		return errors.New("no client available")
	}

	// Run commands defined by target sequentially.
	for _, cmd := range commands {
		// Translate command into task(s).
		tasks, err := play.createTasks(clients, cmd, clientEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", errors.Wrap(err, fmt.Sprintf("creating task %v failed", cmd)))

			continue
		}

		// Run tasks sequentially.
		for _, task := range tasks {
			var (
				writers []io.Writer
				taskWg  sync.WaitGroup
			)

			// Run tasks on the provided clients.
			for _, client := range task.clients {
				err := client.Run(task)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s%v\n", PadStringWithTimestamp(client.Prompt(), maxPromptLen), errors.Wrap(err, fmt.Sprintf("running task %v failed", task)))

					continue
				}

				// Copy over tasks's STDOUT.
				taskWg.Add(1)
				go func(c Client) {
					defer taskWg.Done()

					_, err := io.Copy(os.Stdout, prefixer.New(c.Stdout(), PadStringWithTimestamp(c.Prompt(), maxPromptLen)))
					if err != nil && err != io.EOF {
						// TODO: io.Copy() should not return io.EOF at all.
						// Upstream bug? Or prefixer.WriteTo() bug?
						fmt.Fprintf(os.Stderr, "%s%v\n", PadStringWithTimestamp(c.Prompt(), maxPromptLen), errors.Wrap(err, "reading STDOUT failed"))
					}
				}(client)

				// Copy over tasks's STDERR.
				taskWg.Add(1)
				go func(c Client) {
					defer taskWg.Done()

					_, err := io.Copy(os.Stderr, prefixer.New(c.Stderr(), PadStringWithTimestamp(c.Prompt(), maxPromptLen)))
					if err != nil && err != io.EOF {
						fmt.Fprintf(os.Stderr, "%s%v\n", PadStringWithTimestamp(c.Prompt(), maxPromptLen), errors.Wrap(err, "reading STDERR failed"))
					}
				}(client)

				writers = append(writers, client.Stdin())
			}

			// Copy over task's STDIN.
			if task.input != nil {
				go func() {
					writer := io.MultiWriter(writers...)

					_, err := io.Copy(writer, task.input)
					if err != nil && err != io.EOF {
						fmt.Fprintf(os.Stderr, "%v\n", errors.Wrap(err, "writing STDIN failed"))
					}

					// TODO: Use MultiWriteCloser (not in Stdlib), so we can writer.Close() instead?
					for _, client := range clients {
						client.Close()
					}
				}()
			}

			// Catch OS signals and pass them to all active clients.
			trap := make(chan os.Signal, 1)
			signal.Notify(trap, os.Interrupt)
			go func() {
				for {
					select {
					case sig, ok := <-trap:
						if !ok {
							return
						}

						for _, client := range task.clients {
							err := client.Signal(sig)
							if err != nil {
								fmt.Fprintf(os.Stderr, "%s%v\n", PadStringWithTimestamp(client.Prompt(), maxPromptLen), errors.Wrap(err, "sending signal failed"))
							}
						}
					}
				}
			}()

			// Wait for all I/O operations first.
			taskWg.Wait()

			// Make sure each client finishes the task, return on failure.
			for _, client := range task.clients {
				taskWg.Add(1)

				go func(c Client) {
					defer taskWg.Done()

					if err := c.Wait(); err != nil {
						prompt := c.Prompt()

						// TODO: Store all the errors, and print them after Wait().
						if e, ok := err.(*ssh.ExitError); ok && e.ExitStatus() != 15 {
							fmt.Fprintf(os.Stderr, "%s%v\n%sexit status %v\n", PadStringWithTimestamp(prompt, maxPromptLen), e, PadStringWithTimestamp(prompt, maxPromptLen), e.ExitStatus())
						} else {
							fmt.Fprintf(os.Stderr, "%s%v\n", PadStringWithTimestamp(prompt, maxPromptLen), err)
						}
					}
				}(client)
			}

			// Wait for all commands to finish.
			taskWg.Wait()

			// Stop catching signals for the currently active clients.
			signal.Stop(trap)
			close(trap)
		}
	}

	return nil
}

func (play *Play) Debug(value bool) {
	play.debug = value
}

func (play *Play) Prompt(value bool) {
	play.prompt = value
}
