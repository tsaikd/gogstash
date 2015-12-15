package gogstash

import (
	"github.com/codegangsta/cli"
	"github.com/tsaikd/gogstash/config"
)

var (
	logger = config.Logger
)

func actionWrapper(action func(context *cli.Context) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		if err := action(context); err != nil {
			logger.Errorln(err)
		}
	}
}
