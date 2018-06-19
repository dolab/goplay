package play

import (
	"testing"

	"github.com/golib/assert"
)

func Test_MaskUserHostWithPasswd(t *testing.T) {
	in := "user:passwd@ipv4"
	out := MaskUserHostWithPasswd(in)
	expect := "user:***@ipv4"

	assert.Equal(t, expect, out)
}
