package play

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

// Task represents a set of commands to be run.
type Task struct {
	clients []Client
	run     string
	input   io.Reader
	once    bool
	tty     bool
}

func (play *Play) createTasks(clients []Client, cmd *Command, env string) (tasks []*Task, err error) {
	var allTasks []*Task

	// Upload.
	// Always run upload first.
	if len(cmd.Upload) > 0 {
		uploadTasks, uploadErr := play.createUploadTasks(clients, cmd.Upload, env)
		if uploadErr != nil {
			err = errors.Wrap(uploadErr, "can't create upload task")
			return
		}

		allTasks = append(allTasks, uploadTasks...)
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

		scriptTasks, scriptErr := play.createShellTasks(clients, string(data), env, cmd.Stdin)
		if scriptErr != nil {
			err = errors.Wrap(scriptErr, "can't create script task: "+string(data))
			return
		}

		allTasks = append(allTasks, scriptTasks...)
	}

	// Command.
	if cmd.Run != "" {
		// overwrite clients with local host connection
		if cmd.Locally {
			local := NewLocalClient(env)
			local.Connect("localhost")

			clients = []Client{local}
		}

		shellTasks, shellErr := play.createShellTasks(clients, cmd.Run, env, cmd.Stdin)
		if shellErr != nil {
			err = errors.Wrap(shellErr, "can't create shell task: "+cmd.Run)
			return
		}

		allTasks = append(allTasks, shellTasks...)
	}

	for _, task := range allTasks {
		switch {
		case cmd.Once: // TODO: try cmd over all of clients until one success?
			task.clients = []Client{clients[0]}
			task.once = cmd.Once

			tasks = append(tasks, task)

		case cmd.Serial > 0: // Each "serial" task client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}

				copy := *task
				copy.clients = clients[i:j]

				tasks = append(tasks, &copy)
			}

		default:
			task.clients = clients

			tasks = append(tasks, task)
		}
	}

	return
}

func (play *Play) createUploadTasks(clients []Client, uploads []Upload, env string) (tasks []*Task, err error) {
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

		uploadTarReader, uploadTarErr := NewTarStreamReader(cwd, uploadFile, upload.Exc)
		if uploadTarErr != nil {
			err = errors.Wrap(uploadTarErr, "create tar stream of local path failed: "+upload.Src)
			return
		}

		task := Task{
			run:   RemoteTarCommand(upload.Dst),
			input: uploadTarReader,
			tty:   false,
		}

		tasks = append(tasks, &task)
	}

	return
}

func (play *Play) createShellTasks(clients []Client, shell, env string, stdin bool) (tasks []*Task, err error) {
	task := Task{
		run: shell,
		tty: true,
	}
	if stdin {
		task.input = os.Stdin
	}
	if play.debug {
		task.run = "set -x;" + task.run
	}

	tasks = append(tasks, &task)

	return
}
