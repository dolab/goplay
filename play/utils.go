package play

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/dolab/colorize"
	"github.com/pkg/errors"
)

var (
	pinfo  = colorize.New("cyan")
	pwarn  = colorize.New("yellow")
	perror = colorize.New("red")

	pinfoStart, pinfoEnd   = pinfo.Colour()
	pwarnStart, pwarnEnd   = pwarn.Colour()
	perrorStart, perrorEnd = perror.Colour()

	user2pass = regexp.MustCompile(`[\w.+-_]+?:.+?@`)
)

func pcopy(w io.Writer, r io.Reader, paint colorize.Colorize) (err error) {
	// start, end := paint.Colour()

	// _, err = io.Copy(w, strings.NewReader(start))
	// if err != nil {
	// 	return
	// }

	_, err = io.Copy(w, r)

	// io.Copy(w, strings.NewReader(end))
	return
}

func Infof(format string, v ...interface{}) {
	os.Stdout.WriteString(pinfoStart)
	fmt.Fprintf(os.Stdout, format, v...)
	os.Stdout.WriteString(pinfoEnd)
}

func Warnf(format string, v ...interface{}) {
	os.Stdout.WriteString(pwarnStart)
	fmt.Fprintf(os.Stdout, format, v...)
	os.Stdout.WriteString(pwarnEnd)
}

func Errorf(format string, v ...interface{}) {
	os.Stderr.WriteString(perrorStart)
	fmt.Fprintf(os.Stderr, format, v...)
	os.Stderr.WriteString(perrorEnd)
}

// ResolveLocalPath determines local file path related to cwd
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

// MaskUserHostWithPasswd masks passwd for sensitive
func MaskUserHostWithPasswd(host string) string {
	return user2pass.ReplaceAllStringFunc(host, func(passwd string) string {
		u2p := strings.SplitN(passwd, ":", 2)

		return u2p[0] + ":***@"
	})
}

// PadStringWithTimestamp returns new string leads with time
func PadStringWithTimestamp(s string, n int) string {
	ts := time.Now().Format("2006/01/02 15:04:05")

	if len(s) < n {
		s = strings.Repeat(" ", n-len(s)) + s
	}

	return fmt.Sprintf("%s - %s", ts, s)
}
