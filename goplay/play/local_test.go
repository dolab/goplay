package play

import (
	"testing"

	"github.com/golib/assert"
)

func Test_LocalClient(t *testing.T) {
	client := &LocalClient{}

	assert.Implements(t, (*Client)(nil), client)
}
