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

// Play holds all books for running
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
		return ErrEmptyCommand
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
					env:  clientEnv + `export PLAY_HOST="` + MaskUserHostWithPasswd(host) + `";`,
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

		// hook for ssh client
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
		return ErrEmptyClient
	}

	// Run commands defined by target sequentially.
	// TODO: should gather all outputs and calc stats
	for _, cmd := range commands {
		// build book(s) from command.
		books, err := play.createBooks(clients, cmd, clientEnv)
		if err != nil {
			Errorf("%v\n", errors.Wrapf(err, "creating book %v failed", cmd))

			continue
		}

		// Run books sequentially.
		for _, book := range books {
			var (
				writers []io.Writer
				bookWg  sync.WaitGroup
			)

			// Run books on the provided clients.
			for _, client := range book.clients {
				err := client.Run(book)
				if err != nil {
					Errorf("%s%v\n", PadStringWithTimestamp(client.Prompt(), maxPromptLen), errors.Wrapf(err, "running book %v failed", book))

					continue
				}

				// Copy over book's STDOUT.
				bookWg.Add(1)
				go func(c Client) {
					defer bookWg.Done()

					err := pcopy(
						os.Stdout,
						prefixer.New(c.Stdout(), PadStringWithTimestamp(c.Prompt(), maxPromptLen)),
						pinfo)
					if err != nil && err != io.EOF {
						// TODO: io.Copy() should not return io.EOF at all.
						// Upstream bug? Or prefixer.WriteTo() bug?
						Errorf("%s%v\n", PadStringWithTimestamp(c.Prompt(), maxPromptLen), errors.Wrap(err, "reading STDOUT failed"))
					}
				}(client)

				// Copy over book's STDERR.
				bookWg.Add(1)
				go func(c Client) {
					defer bookWg.Done()

					err := pcopy(
						os.Stderr,
						prefixer.New(c.Stderr(), PadStringWithTimestamp(c.Prompt(), maxPromptLen)),
						perror)
					if err != nil && err != io.EOF {
						Errorf("%s%v\n", PadStringWithTimestamp(c.Prompt(), maxPromptLen), errors.Wrap(err, "reading STDERR failed"))
					}
				}(client)

				writers = append(writers, client.Stdin())
			}

			// Copy over book's STDIN.
			if book.input != nil {
				go func() {
					writer := io.MultiWriter(writers...)

					_, err := io.Copy(writer, book.input)
					if err != nil && err != io.EOF {
						Errorf("%v\n", errors.Wrap(err, "writing STDIN failed"))
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

						for _, client := range book.clients {
							err := client.Signal(sig)
							if err != nil {
								Errorf("%s%v\n", PadStringWithTimestamp(client.Prompt(), maxPromptLen), errors.Wrapf(err, "sending signal %v failed", sig))
							}
						}
					}
				}
			}()

			// Wait for all I/O operations first.
			bookWg.Wait()

			// Make sure each client finishes the book, return on failure.
			for _, client := range book.clients {
				bookWg.Add(1)

				go func(c Client) {
					defer bookWg.Done()

					prompt := PadStringWithTimestamp(c.Prompt(), maxPromptLen)

					err := c.Wait()
					if err != nil {
						// TODO: Store all the errors, and print them after Wait().
						if e, ok := err.(*ssh.ExitError); ok && e.ExitStatus() != 15 {
							Errorf("%s%v\n%sexit status %v\n", prompt, e, prompt, e.ExitStatus())
						} else {
							Errorf("%s%v\n", prompt, err)
						}
					} else {
						Infof("%sDone!\n", prompt)
					}
				}(client)
			}

			// Wait for all commands to finish.
			bookWg.Wait()

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
