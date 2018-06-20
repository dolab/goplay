package books

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/dolab/goplay/play"
	"github.com/dolab/logger"
	"github.com/golib/cli"
	"golang.org/x/crypto/ssh"
)

var (
	SSH *_SSH
)

type _SSH struct{}

func (_ *_SSH) Init(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := ctx.String("name")
		if name == "" {
			name = "ansible"
		}

		bitSize := ctx.Int("bit-size")
		if bitSize%1024 != 0 || bitSize == 0 {
			bitSize = 4096
		}

		sshKey, err := generatePrivateKey(bitSize)
		if err != nil {
			log.Errorf("generate ssh private key (%d bits): %v", bitSize, err)

			return err
		}

		privateKey := encodePrivateKeyToPEM(sshKey)

		publicKey, err := generatePublicKey(&sshKey.PublicKey)
		if err != nil {
			log.Errorf("generate ssh public key: %v", err)

			return err
		}

		if err := ioutil.WriteFile(path.Join(absroot, name+"_rsa"), privateKey, 0600); err != nil {
			log.Errorf("ioutil.WriteFile(%s, ?, 0600): %v", path.Join(absroot, name+"_rsa"), err)

			return err
		}

		if err := ioutil.WriteFile(path.Join(absroot, name+"_rsa.pub"), publicKey, 0600); err != nil {
			log.Errorf("ioutil.WriteFile(%s, ?, 0600): %v", path.Join(absroot, name+"_rsa.pub"), err)

			return err
		}

		return nil
	}
}

func (_ *_SSH) Setup(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) (err error) {
		// hosts definitions
		hostfile := ctx.String("hostfile")
		if hostfile == "" {
			cli.ShowSubcommandHelp(ctx)

			return cli.NewExitError("hostfile is required", 04)
		}
		hostfile = abspath(hostfile)

		lines, err := ioutil.ReadFile(hostfile)
		if err != nil {
			log.Errorf("ioutil.ReadFile(%s): %v", hostfile, err)

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

		// public key file
		keyfile := ctx.GlobalString("keyfile")
		if keyfile == "" {
			keyfile = identityfile
		}
		keyfile = abspath(keyfile)

		publicKey, err := ioutil.ReadFile(keyfile)
		if err != nil {
			log.Errorf("ioutil.ReadFile(%s): %v", keyfile, err)

			return
		}

		// goplay
		envs := play.EnvVars{
			{
				Key:   "PUB_KEY",
				Value: string(publicKey),
			},
		}
		network := play.Network{
			Hosts: hosts,
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

		err = player.Run(&network, envs, &command)
		if err != nil {
			return
		}

		// generate default playfile
		for i, host := range hosts {
			hosts[i] = strings.SplitN(host, "@", 2)[1]
		}

		var buf bytes.Buffer
		err = playfiletpl.Execute(&buf, map[string]string{
			"hosts":         "- " + strings.Join(hosts, "\n    - "),
			"identity_file": keyfile,
		})
		if err != nil {
			log.Errorf("playfile.Execute(): %v", err)

			return err
		}

		err = ioutil.WriteFile(playfile, buf.Bytes(), 0755)
		if err != nil {
			log.Errorf("ioutil.WriteFile(%s): %v", playfile, err)

			return err
		}

		return nil
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

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return pubKeyBytes, nil
}
