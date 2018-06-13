package books

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/dolab/goplay/play"
	"github.com/dolab/logger"
	"github.com/golib/cli"
)

var (
	SSH *_SSH
)

type _SSH struct{}

func (_ *_SSH) Generate() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return nil
	}
}

func (_ *_SSH) Setup(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) (err error) {
		// hosts definitions
		filename := ctx.String("hosts")
		if filename == "" {
			cli.ShowSubcommandHelp(ctx)

			return cli.NewExitError("hosts is required", 04)
		}
		if strings.HasPrefix(filename, "~/") {
			cu, err := user.Current()
			if err == nil {
				filename = strings.Replace(filename, "~", cu.HomeDir, 1)
			}
		}

		lines, err := ioutil.ReadFile(path.Clean(filename))
		if err != nil {
			return
		}

		// public key file
		keyfile := path.Clean(ctx.GlobalString("keyfile"))
		if strings.HasPrefix(keyfile, "~/") {
			cu, err := user.Current()
			if err == nil {
				keyfile = strings.Replace(keyfile, "~", cu.HomeDir, 1)
			}
		}

		publicKey, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return
		}

		var (
			hosts []string
			errs  []error
		)
		for i, line := range strings.Split(string(lines), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if isValidUserHostWithPasswd(line) {
				hosts = append(hosts, line)
			} else {
				errs = append(errs, fmt.Errorf("Line %d host format is invalid: %s", i, line))
			}
		}
		if len(errs) > 0 {
			return cli.NewMultiError(errs...)
		}

		network := play.Network{
			Hosts: hosts,
		}
		envs := play.EnvVars{
			{
				Key:   "PUB_KEY",
				Value: string(publicKey),
			},
		}
		command := play.Command{
			Name:  "setup ssh trust",
			Desc:  "append public key to remote hosts",
			Run:   "mkdir -p ~/.ssh && echo $PUB_KEY >> ~/.ssh/authorized_keys",
			Stdin: true,
		}

		player, err := play.New(nil)
		if err != nil {
			return
		}
		player.Prompt(ctx.GlobalBool("prompt"))
		player.Debug(ctx.GlobalBool("debug"))

		return player.Run(&network, envs, &command)
	}
}

func isValidUserHostWithPasswd(host string) bool {
	user2host := strings.SplitN(host, "@", 2)
	if len(user2host) != 2 {
		return false
	}

	ipv4 := strings.Split(user2host[1], ".")
	if len(ipv4) < 4 {
		return false
	}

	for _, n := range ipv4 {
		i, err := strconv.Atoi(n)
		if err != nil || i < 0 || i > 255 {
			return false
		}
	}

	return true
}
