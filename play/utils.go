package play

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	user2pass = regexp.MustCompile(`[\w.+-_]+?:.+?@`)
)

func ResolveLocalPath(cwd, path, env string) (string, error) {
	// Check if file exists first. Use bash to resolve $ENV_VARs.
	cmd := exec.Command("bash", "-c", env+" echo -n "+path)
	cmd.Dir = cwd

	filename, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "resolving path failed")
	}

	return string(filename), nil
}

func MaskUserHostWithPasswd(host string) string {
	return user2pass.ReplaceAllStringFunc(host, func(passwd string) string {
		u2p := strings.SplitN(passwd, ":", 2)

		return u2p[0] + ":***@"
	})
}

func PadStringWithTimestamp(s string, n int) string {
	ts := time.Now().Format("2006/01/02 15:04:05")

	if len(s) < n {
		s = strings.Repeat(" ", n-len(s)) + s
	}

	return fmt.Sprintf("%s - %s", ts, s)
}
