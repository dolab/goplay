package main

import (
	"os"

	"github.com/dolab/goplay/books"
	"github.com/dolab/goplay/play"
	"github.com/dolab/logger"
	"github.com/golib/cli"
)

var (
	root = "~/.goplay"

	playfile string
	keyfile  string
	log      *logger.Logger
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
			Name:  "ssh",
			Usage: "ssh management for ansible",
			Subcommands: []cli.Command{
				{
					Name:  "init",
					Usage: "generate SSH RSA Private/Public Key pair",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Supply `NAME` of ssh key file",
							Value: "ansible",
						},
						cli.IntFlag{
							Name:  "bit-size",
							Usage: "Supply security of ssh keypair",
							Value: 4096,
						},
					},
					Action: books.SSH.Init(log),
				},
				{
					Name:  "setup",
					Usage: "setup ssh trust of all hosts",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "hostfile",
							Usage: "Supply hosts list `FILE`, formed in user:passwd@ip for each line",
						},
					},
					Action: books.SSH.Setup(log),
				},
			},
		},
		{
			Name:  "ansible",
			Usage: "ansible initializations",
			Subcommands: []cli.Command{
				{
					Name:  "setup",
					Usage: "setup ansible inventory",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "inventory",
							Usage: "Supply path to ansible inventory file",
							Value: "~/.goplay/ansible_hosts",
						},
						cli.StringFlag{
							Name:  "hostname",
							Usage: "Supply hostname prefix for all hosts",
							Value: "kodoe",
						},
					},
					Action: books.Ansible.Setup(log),
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("OK!")
	}
}
