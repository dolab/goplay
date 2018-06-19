package books

import (
	"github.com/golib/cli"
)

var (
	Config *_Config
)

type _Config struct{}

func (_ *_Config) Init() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return nil
	}
}
