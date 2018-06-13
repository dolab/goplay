package main

import (
	"os"

	"github.com/dolab/goplay/books"
	"github.com/dolab/goplay/play"
	"github.com/dolab/logger"
	"github.com/golib/cli"
)

var (
	playfile string
	keyfile  string

	log *logger.Logger
)

func init() {
	log, _ = logger.New("stdout")
	log.SetColor(true)
	log.SetFlag(3)
}

func main() {
	app := cli.NewApp()
	app.Name = "goplay"
	app.Usage = "goplay -h"
	app.Version = play.VERSION
	app.Authors = []cli.Author{
		{
			Name:  "Spring MC",
			Email: "Heresy.MC@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "playfile",
			Usage:       "Supply play configuration `FILE`",
			Value:       "~/.goplay/playfile.yml",
			Destination: &playfile,
		},
		cli.StringFlag{
			Name:        "keyfile",
			Usage:       "Supply ssh key `FILE`",
			Value:       "~/.ssh/id_rsa.pub",
			Destination: &keyfile,
		},
		cli.BoolFlag{
			Name:  "prompt",
			Usage: "Print info(s) while running playbook(s)",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Run playbook(s) with debug mode",
		},
	}
	// app.Action = func(ctx *cli.Context) error {
	// 	log.Info(playfile)
	// 	log.Info(keyfile)

	// 	return nil
	// }

	app.Commands = []cli.Command{
		{
			Name:    "ssh",
			Aliases: []string{"ssh"},
			Usage:   "ssh management for ansible",
			Subcommands: []cli.Command{
				{
					Name:    "generate",
					Aliases: []string{"gen"},
					Usage:   "generate SSH RSA Private/Public Key pair",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Supply name of ssh key file",
						},
					},
					Action: books.SSH.Generate(),
				},
				{
					Name:  "setup",
					Usage: "setup ssh trust of all hosts",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "hosts",
							Usage: "Supply hosts list `FILE`, formed in user:passwd@ip for each line",
						},
					},
					Action: books.SSH.Setup(log),
				},
			},
		},
		{
			Name:    "play",
			Aliases: []string{"run"},
			Usage:   "run playbook(s)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "group",
					Usage: "Supply group(s) for playbook(s) target",
				},
				cli.StringFlag{
					Name:  "ssh-config",
					Usage: "Supply ssh config for custom SSH config",
				},
			},
			Action: books.Play.Run(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("OK!")
	}
}
