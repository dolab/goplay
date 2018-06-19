package books

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/dolab/goplay/play"
	"github.com/dolab/logger"
	"github.com/golib/cli"
)

var (
	Ansible *_Ansible
)

type _Ansible struct{}

func (_ *_Ansible) Setup(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		filename := ctx.String("inventory")
		if filename == "" {
			cli.ShowCommandHelp(ctx, "setup")

			return cli.NewExitError("Both playfile and inventory are required", 04)
		}
		filename = abspath(filename)

		pfile, err := play.NewPlayfileFromFile(playfile)
		if err != nil {
			return err
		}

		// resolve all hosts
		all, ok := pfile.Networks.Get("all")
		if !ok {
			return cli.NewExitError("Network named with all does not exist", 04)
		}

		hostname := ctx.String("hostname")
		if hostname == "" {
			hostname = "kodoe"
		}

		ip2name := map[string][]string{}
		for i, host := range all.Hosts {
			ip2name[host] = []string{fmt.Sprintf("%s-%d", hostname, i+1), host}
		}

		// try to resolve ansilbe version for ssh config
		ansibleVersion := "2"

		cmd := exec.Command("bash", "-c", "ansible --version")
		if data, err := cmd.Output(); err == nil {
			matches := rversion.FindStringSubmatch(string(data))
			if len(matches) == 2 {
				ansibleVersion = matches[1]
			} else {
				log.Warnf("CANNOT resolve ansible version, use default of %v.0", ansibleVersion)
			}
		} else {
			log.Errorf("Failed to resolve ansible version: %v (default to %v.0)", err, ansibleVersion)
		}

		buf := bytes.NewBuffer(nil)
		for _, name := range pfile.Networks.Names {
			buf.WriteString(fmt.Sprintf("[%s]\n", name))

			network, ok := pfile.Networks.Get(name)
			if !ok {
				continue
			}

			for _, host := range network.Hosts {
				if ansibleVersion >= defaultVersion {
					buf.WriteString(fmt.Sprintf("%s ansible_host=%s ansible_port=%v ansible_user=%s ansible_ssh_private_key_file=%s\n", ip2name[host][0], host, network.Port, network.User, network.IdentityFile))
				} else {
					buf.WriteString(fmt.Sprintf("%s ansible_ssh_host=%s ansible_ssh_port=%v ansible_ssh_user=%s ansible_ssh_private_key_file=%s\n", ip2name[host][0], host, network.Port, network.User, network.IdentityFile))
				}
			}

			buf.WriteRune('\n')
		}

		return ioutil.WriteFile(filename, buf.Bytes(), 0644)
	}
}
