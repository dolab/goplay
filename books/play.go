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
		return nil
	}
}
