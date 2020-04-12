package cmd

import (
	"crawlab/app/master/cmd/start"
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App{
	app:=&cli.App{
		Name:"crawlab",
		Version: "1.0.1.beta",
		Commands: []*cli.Command{
			start.NewCommand(),
		},
	}
	return app
}
