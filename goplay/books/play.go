package books

import (
	"github.com/golib/cli"
)

var (
	Play *_Play
)

type _Play struct{}

func (_ *_Play) Run() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		// if Playfile == "" {
		// 	Playfile = "./Playfile"
		// }
		// data, err := ioutil.ReadFile(resolvePath(Playfile))
		// if err != nil {
		// 	firstErr := err
		// 	data, err = ioutil.ReadFile("./Playfile.yml") // Alternative to ./Playfile.
		// 	if err != nil {
		// 		fmt.Fprintln(os.Stderr, firstErr)
		// 		fmt.Fprintln(os.Stderr, err)
		// 		os.Exit(1)
		// 	}
		// }
		// conf, err := play.NewPlayfile(data)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// 	os.Exit(1)
		// }

		// // Parse network and commands to be run from args.
		// network, commands, err := parseArgs(conf)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// 	os.Exit(1)
		// }

		// // --only flag filters hosts
		// if onlyHosts != "" {
		// 	expr, err := regexp.CompilePOSIX(onlyHosts)
		// 	if err != nil {
		// 		fmt.Fprintln(os.Stderr, err)
		// 		os.Exit(1)
		// 	}

		// 	var hosts []string
		// 	for _, host := range network.Hosts {
		// 		if expr.MatchString(host) {
		// 			hosts = append(hosts, host)
		// 		}
		// 	}
		// 	if len(hosts) == 0 {
		// 		fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts match --only '%v' regexp", onlyHosts))
		// 		os.Exit(1)
		// 	}
		// 	network.Hosts = hosts
		// }

		// // --except flag filters out hosts
		// if exceptHosts != "" {
		// 	expr, err := regexp.CompilePOSIX(exceptHosts)
		// 	if err != nil {
		// 		fmt.Fprintln(os.Stderr, err)
		// 		os.Exit(1)
		// 	}

		// 	var hosts []string
		// 	for _, host := range network.Hosts {
		// 		if !expr.MatchString(host) {
		// 			hosts = append(hosts, host)
		// 		}
		// 	}
		// 	if len(hosts) == 0 {
		// 		fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts left after --except '%v' regexp", onlyHosts))
		// 		os.Exit(1)
		// 	}
		// 	network.Hosts = hosts
		// }

		// // --sshconfig flag location for ssh_config file
		// if sshConfig != "" {
		// 	confHosts, err := sshconfig.ParseSSHConfig(resolvePath(sshConfig))
		// 	if err != nil {
		// 		fmt.Fprintln(os.Stderr, err)
		// 		os.Exit(1)
		// 	}

		// 	// flatten Host -> *SSHHost, not the prettiest
		// 	// but will do
		// 	confMap := map[string]*sshconfig.SSHHost{}
		// 	for _, conf := range confHosts {
		// 		for _, host := range conf.Host {
		// 			confMap[host] = conf
		// 		}
		// 	}

		// 	// check network.Hosts for match
		// 	for _, host := range network.Hosts {
		// 		conf, found := confMap[host]
		// 		if found {
		// 			network.User = conf.User
		// 			network.IdentityFile = resolvePath(conf.IdentityFile)
		// 			network.Hosts = []string{fmt.Sprintf("%s:%d", conf.HostName, conf.Port)}
		// 		}
		// 	}
		// }

		// var vars play.EnvVars
		// for _, val := range append(conf.Env, network.Env...) {
		// 	vars.Set(val.Key, val.Value)
		// }
		// if err := vars.ResolveValues(); err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// 	os.Exit(1)
		// }

		// // Parse CLI --env flag env vars, define $SUP_ENV and override values defined in Playfile.
		// var cliVars play.EnvVars
		// for _, env := range envVars {
		// 	if len(env) == 0 {
		// 		continue
		// 	}
		// 	i := strings.Index(env, "=")
		// 	if i < 0 {
		// 		if len(env) > 0 {
		// 			vars.Set(env, "")
		// 		}
		// 		continue
		// 	}
		// 	vars.Set(env[:i], env[i+1:])
		// 	cliVars.Set(env[:i], env[i+1:])
		// }

		// // SUP_ENV is generated only from CLI env vars.
		// // Separate loop to omit duplicates.
		// supEnv := ""
		// for _, v := range cliVars {
		// 	supEnv += fmt.Sprintf(" -e %v=%q", v.Key, v.Value)
		// }
		// vars.Set("SUP_ENV", strings.TrimSpace(supEnv))

		// // Create new Stackup app.
		// player, err := play.New(conf)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// 	os.Exit(1)
		// }
		// player.Debug(debug)
		// player.Prompt(!disablePrefix)

		// // Run all the commands in the given network.
		// err = player.Run(network, vars, commands...)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// 	os.Exit(1)
		// }
		return nil
	}
}
