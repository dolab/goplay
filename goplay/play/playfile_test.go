package play

import (
	"testing"

	"github.com/golib/assert"
)

var (
	playfile = `---
version: 1.0.0

# Global variables
user: &user
  user: root
  identity_file: ~/.ssh/testing_rsa

all_hosts: &all_hosts
  <<: *user
  hosts:
    - 127.0.0.1

app_hosts: &app_hosts
  <<: *user
  hosts:
    - 127.0.0.1

db_hosts: &db_hosts
  <<: *user
  hosts:
    - 127.0.0.1

pfd_hosts: &pfd_hosts
  <<: *user
  hosts:
    - 127.0.0.1

ebd_hosts: &ebd_hosts
  <<: *user
  hosts:
    - 127.0.0.1

networks:
  all:
    <<: *all_hosts

  app:
    <<: *app_hosts

  db:
    <<: *db_hosts

  pfd:
    <<: *pfd_hosts

  ebd:
    <<: *ebd_hosts

commands:
  echo:
    desc: Print some env vars
    run: echo $PLAY_NETWORK

  date:
    desc: Print OS name and current date/time
    run: uname -a; date

  assets:
    uploads:
      vimrc:
        src: ~/.vimrc
        dst: /home/deploy/.vimrc
      bash_profile:
        src: ~/.bash_profile
        dst: /home/deploy/.bash_profile

books:
  all:
    - echo
    - date
`
)

func Test_NewPlayfile(t *testing.T) {
	assertion := assert.New(t)

	pfile, err := NewPlayfile([]byte(playfile))
	assertion.Nil(err)
	assertion.Equal("1.0.0", pfile.Version)

	// check global envs
	assertion.Empty(pfile.Envs.AsExport())

	// check networks
	for _, name := range pfile.Networks.Names {
		network, ok := pfile.Networks.Get(name)
		assertion.True(ok)
		assertion.Equal("root", network.User)
		assertion.Equal("~/.ssh/testing_rsa", network.IdentityFile)
		assertion.Equal([]string{"127.0.0.1"}, network.Hosts)
	}

	// check commands
	for _, name := range pfile.Commands.Names {
		cmd, ok := pfile.Commands.Get(name)
		assertion.True(ok)
		if len(cmd.Uploads) > 0 {
			assertion.Equal(2, len(cmd.Uploads))
		} else {
			assertion.NotEmpty(cmd.Run)
		}
	}

	// check books
	for _, name := range pfile.Books.Names {
		book, ok := pfile.Books.Get(name)
		assertion.True(ok)
		assertion.Equal([]string{"echo", "date"}, book)
	}
}
