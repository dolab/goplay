package play

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

// Book represents a set of commands to be run.
type Book struct {
	clients []Client
	run     string
	input   io.Reader
	once    bool
	tty     bool
}

func (play *Play) createBooks(clients []Client, cmd *Command, env string) (books []*Book, err error) {
	var allBooks []*Book

	// Upload.
	// Always run upload first.
	if len(cmd.Uploads) > 0 {
		uploadBooks, uploadErr := play.createUploadBooks(clients, cmd.Uploads, env)
		if uploadErr != nil {
			err = errors.Wrap(uploadErr, "can't create upload book")
			return
		}

		allBooks = append(allBooks, uploadBooks...)
	}

	// Script.
	// Read the script as a multi lines input command.
	if cmd.Script != "" {
		data, ioerr := ioutil.ReadFile(path.Clean(cmd.Script))
		if ioerr != nil {
			err = errors.Wrap(ioerr, "can't read script")
			return
		}

		// overwrite clients with local host connection
		if cmd.Locally {
			local := NewLocalClient(env)
			local.Connect("localhost")

			clients = []Client{local}
		}

		scriptBooks, scriptErr := play.createShellBooks(clients, string(data), env, cmd.Stdin)
		if scriptErr != nil {
			err = errors.Wrap(scriptErr, "can't create script book: "+string(data))
			return
		}

		allBooks = append(allBooks, scriptBooks...)
	}

	// Command.
	if cmd.Run != "" {
		// overwrite clients with local host connection
		if cmd.Locally {
			local := NewLocalClient(env)
			local.Connect("localhost")

			clients = []Client{local}
		}

		shellBooks, shellErr := play.createShellBooks(clients, cmd.Run, env, cmd.Stdin)
		if shellErr != nil {
			err = errors.Wrap(shellErr, "can't create shell book: "+cmd.Run)
			return
		}

		allBooks = append(allBooks, shellBooks...)
	}

	for _, book := range allBooks {
		switch {
		case cmd.Once: // TODO: try cmd over all of clients until one success?
			book.clients = []Client{clients[0]}
			book.once = cmd.Once

			books = append(books, book)

		case cmd.Serial > 0: // Each "serial" book client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}

				copy := *book
				copy.clients = clients[i:j]

				books = append(books, &copy)
			}

		default:
			book.clients = clients

			books = append(books, book)
		}
	}

	return
}

func (play *Play) createUploadBooks(clients []Client, uploads map[string]Upload, env string) (books []*Book, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		err = errors.Wrap(err, "os.Getwd() failed")
		return
	}

	for _, upload := range uploads {
		uploadFile, uploadErr := ResolveLocalPath(cwd, upload.Src, env)
		if uploadErr != nil {
			err = errors.Wrap(uploadErr, "resolve local path failed: "+upload.Src)
			return
		}

		uploadTarReader, uploadTarErr := NewTarStreamReader(cwd, uploadFile, upload.Filter)
		if uploadTarErr != nil {
			err = errors.Wrap(uploadTarErr, "create tar stream of local path failed: "+upload.Src)
			return
		}

		book := Book{
			run:   RemoteTarCommand(upload.Dst),
			input: uploadTarReader,
			tty:   false,
		}

		books = append(books, &book)
	}

	return
}

func (play *Play) createShellBooks(clients []Client, shell, env string, stdin bool) (books []*Book, err error) {
	book := Book{
		run: shell,
		tty: true,
	}
	if stdin {
		book.input = os.Stdin
	}
	if play.debug {
		book.run = "set -x;" + book.run
	}

	books = append(books, &book)

	return
}
