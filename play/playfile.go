package play

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Playfile represents the play configuration YAML file.
type Playfile struct {
	Networks Networks `yaml:"networks"`
	Commands Commands `yaml:"commands"`
	Targets  Targets  `yaml:"targets"`
	Env      EnvVars  `yaml:"env"`
	Version  string   `yaml:"version"`
}

// NewPlayfile parses configuration file and returns Playfile or error.
func NewPlayfile(data []byte) (*Playfile, error) {
	var config Playfile

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// API backward compatibility. Will be deprecated in v1.0.
	switch config.Version {
	case "":
		config.Version = "1.0"

	case "1.0":
		// ignore

	default:
		return nil, ErrPlayfileVersion{"Invalid Playfile version: " + config.Version}
	}

	return &config, nil
}

// Network is group of hosts with extra custom env vars.
type Network struct {
	Hosts     []string `yaml:"hosts"`
	Env       EnvVars  `yaml:"env"`
	Inventory string   `yaml:"inventory"`
	Bastion   string   `yaml:"bastion"` // Jump host for the environment

	// Should these live on Hosts too? We'd have to change []string to struct, even in Playfile.
	User         string `yaml:"user"`
	IdentityFile string `yaml:"identity_file"`
}

// ParseInventory runs the inventory command, if provided, and appends
// the command's output lines to the manually defined list of hosts.
func (n Network) ParseInventory() ([]string, error) {
	if n.Inventory == "" {
		return nil, nil
	}

	cmd := exec.Command("/bin/sh", "-c", n.Inventory)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, n.Env.Slice()...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(output)

	var hosts []string
	for {
		host, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		host = strings.TrimSpace(host)
		// skip empty lines and comments
		if host == "" || host[:1] == "#" {
			continue
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

// Networks is a list of user-defined networks
type Networks struct {
	Names []string
	nets  map[string]Network
}

func (n *Networks) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&n.nets)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	n.Names = make([]string, len(items))
	for i, item := range items {
		n.Names[i] = item.Key.(string)
	}

	return nil
}

func (n *Networks) Get(name string) (Network, bool) {
	net, ok := n.nets[name]
	return net, ok
}

// Command represents command(s) to be run remotely.
type Command struct {
	Name    string   `yaml:"-"`       // Command name.
	Desc    string   `yaml:"desc"`    // Command description.
	Run     string   `yaml:"run"`     // Command(s) to be run remotelly.
	Script  string   `yaml:"script"`  // Load command(s) from script and run it remotelly.
	Upload  []Upload `yaml:"upload"`  // See Upload struct.
	Serial  int      `yaml:"serial"`  // Max number of clients processing a task in parallel.
	Locally bool     `yaml:"locally"` // Command(s) to be run locally.
	Stdin   bool     `yaml:"stdin"`   // Attach localhost STDOUT to remote commands' STDIN?
	Once    bool     `yaml:"once"`    // The command should be run "once" (randomly on one host only).
}

// Commands is a list of user-defined commands
type Commands struct {
	Names []string
	cmds  map[string]Command
}

func (c *Commands) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&c.cmds)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	c.Names = make([]string, len(items))
	for i, item := range items {
		c.Names[i] = item.Key.(string)
	}

	return nil
}

func (c *Commands) Get(name string) (Command, bool) {
	cmd, ok := c.cmds[name]
	return cmd, ok
}

// Targets is a list of user-defined targets
type Targets struct {
	Names   []string
	targets map[string][]string
}

func (t *Targets) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&t.targets)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	t.Names = make([]string, len(items))
	for i, item := range items {
		t.Names[i] = item.Key.(string)
	}

	return nil
}

func (t *Targets) Get(name string) ([]string, bool) {
	cmds, ok := t.targets[name]
	return cmds, ok
}

// Upload represents file copy operation from localhost Src path to remote Dst
// path of every host in a given Network.
type Upload struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
	Exc string `yaml:"exclude"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Key   string
	Value string
}

func (e EnvVar) String() string {
	return e.Key + `=` + e.Value
}

// AsExport returns the environment variable as a bash export statement
func (e EnvVar) AsExport() string {
	return `export ` + e.Key + `="` + e.Value + `";`
}

// EnvVars is a list of environment variables that maps to a YAML map,
// but maintains order, enabling late variables to reference early variables.
type EnvVars []*EnvVar

func (e EnvVars) Slice() []string {
	envs := make([]string, len(e))
	for i, env := range e {
		envs[i] = env.String()
	}
	return envs
}

func (e *EnvVars) UnmarshalYAML(unmarshal func(interface{}) error) error {
	items := []yaml.MapItem{}

	err := unmarshal(&items)
	if err != nil {
		return err
	}

	*e = make(EnvVars, 0, len(items))

	for _, v := range items {
		e.Set(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
	}

	return nil
}

// Set key to be equal value in this list.
func (e *EnvVars) Set(key, value string) {
	for i, v := range *e {
		if v.Key == key {
			(*e)[i].Value = value
			return
		}
	}

	*e = append(*e, &EnvVar{
		Key:   key,
		Value: value,
	})
}

func (e *EnvVars) ResolveValues() error {
	if len(*e) == 0 {
		return nil
	}

	exports := ""
	for i, v := range *e {
		exports += v.AsExport()

		cmd := exec.Command("bash", "-c", exports+"echo -n "+v.Value+";")
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.Dir = cwd
		resolvedValue, err := cmd.Output()
		if err != nil {
			return errors.Wrapf(err, "resolving env var %v failed", v.Key)
		}

		(*e)[i].Value = string(resolvedValue)
	}

	return nil
}

func (e *EnvVars) AsExport() string {
	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	exports := ``
	for _, v := range *e {
		exports += v.AsExport() + " "
	}

	return exports
}
