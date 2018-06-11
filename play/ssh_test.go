package play

import (
	"testing"

	"github.com/golib/assert"
)

func Test_SSHClient(t *testing.T) {
	client := &SSHClient{}

	assert.Implements(t, (*Client)(nil), client)
}
