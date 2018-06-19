package books

import (
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"text/template"
)

var (
	absroot      = "~/.goplay"
	identityfile = path.Join(absroot, "ansible_rsa.pub")
	playfile     = path.Join(absroot, "Playfile.yml")
	playfiletpl  = template.Must(template.ParseFiles("./Playfile.yml"))

	// ansible
	rversion          = regexp.MustCompile(`^ansible +?([\d.]+?)[\d.]*?`)
	rconfig           = regexp.MustCompile(`config file\W*?=\W*?([\w/.]+)`)
	rinventory        = regexp.MustCompile(`#?inventory\W*?=\W*?[\w/.]+`)
	defaultVersion    = "2"
	defaultConfigFile = path.Join(absroot, "ansible.cfg")
)

func init() {
	absroot = abspath(absroot)
	identityfile = abspath(identityfile)
	playfile = abspath(playfile)
	defaultConfigFile = abspath(defaultConfigFile)

	err := os.MkdirAll(absroot, 0755)
	if err != nil {
		panic(err)
	}
}

func abspath(filename string) string {
	filename = path.Clean(filename)

	if !strings.HasPrefix(filename, "~/") {
		return filename
	}

	cu, err := user.Current()
	if err == nil {
		filename = strings.Replace(filename, "~", cu.HomeDir, 1)
	}

	return filename
}
